package service

import (
	"context"
	"fmt"

	"github.com/Yogdunana/StarByte/backend/internal/rbac"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/dto"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RoleService 角色服务接口
// 提供角色的增删改查、权限分配、用户查询等业务逻辑，含缓存失效处理。
type RoleService interface {
	// Create 创建新角色，校验编码唯一性和父角色存在性
	Create(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error)
	// GetByID 根据 ID 查询角色详情（含权限 ID 列表）
	GetByID(ctx context.Context, id uuid.UUID) (*dto.RoleDetailResponse, error)
	// List 分页查询角色列表，支持关键字模糊搜索
	List(ctx context.Context, req *dto.ListRoleRequest) ([]dto.RoleListResponse, int64, error)
	// Update 更新角色信息，系统内置角色不可修改状态、编码和父级
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error)
	// Delete 删除角色，系统内置角色和已关联用户的角色不可删除
	Delete(ctx context.Context, id uuid.UUID) error
	// AssignPermissions 为角色分配权限（全量替换），系统内置角色不可修改权限
	AssignPermissions(ctx context.Context, id uuid.UUID, req *dto.AssignPermissionsRequest) error
	// GetRoleUsers 分页查询角色下的用户列表，支持数据权限过滤
	GetRoleUsers(ctx context.Context, id uuid.UUID, page, pageSize int, dataScope *model.DataScopeCondition) ([]dto.RoleUserResponse, int64, error)
}

type roleService struct {
	db             *gorm.DB
	roleRepo       repo.RoleRepo
	permissionRepo repo.PermissionRepo
	cacheService   PermissionCacheService
}

// NewRoleService 创建角色服务
func NewRoleService(db *gorm.DB, roleRepo repo.RoleRepo, permissionRepo repo.PermissionRepo, cacheService PermissionCacheService) RoleService {
	return &roleService{
		db:             db,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		cacheService:   cacheService,
	}
}

// Create 创建角色
func (s *roleService) Create(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	// 检查角色编码是否已存在
	existing, err := s.roleRepo.GetByCode(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("check role code: %w", err)
	}
	if existing != nil {
		return nil, rbac.NewRoleCodeExistsError(req.Code)
	}

	role := &model.Role{
		ID:          uuid.New(),
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
	}

	// 解析 parent_id
	if req.ParentID != "" {
		parentID, err := uuid.Parse(req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("invalid parent_id: %w", err)
		}
		// 验证父角色是否存在
		parent, err := s.roleRepo.GetByID(ctx, parentID)
		if err != nil {
			return nil, fmt.Errorf("get parent role: %w", err)
		}
		if parent == nil {
			return nil, rbac.NewRoleNotFoundError()
		}
		role.ParentID = &parentID
	}

	if req.SortOrder != nil {
		role.SortOrder = *req.SortOrder
	}

	if err := s.roleRepo.Create(ctx, nil, role); err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}

	return toRoleResponse(role), nil
}

// GetByID 获取角色详情（含权限 ID 列表）
func (s *roleService) GetByID(ctx context.Context, id uuid.UUID) (*dto.RoleDetailResponse, error) {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}
	if role == nil {
		return nil, rbac.NewRoleNotFoundError()
	}

	permIDs, err := s.roleRepo.GetPermissionIDsByRoleID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get role permissions: %w", err)
	}

	permIDStrs := make([]string, 0, len(permIDs))
	for _, pid := range permIDs {
		permIDStrs = append(permIDStrs, pid.String())
	}

	return &dto.RoleDetailResponse{
		RoleResponse:  *toRoleResponse(role),
		PermissionIDs: permIDStrs,
	}, nil
}

// List 分页查询角色列表
func (s *roleService) List(ctx context.Context, req *dto.ListRoleRequest) ([]dto.RoleListResponse, int64, error) {
	roles, total, err := s.roleRepo.List(ctx, req.Page, req.PageSize, req.Keyword)
	if err != nil {
		return nil, 0, fmt.Errorf("list roles: %w", err)
	}

	result := make([]dto.RoleListResponse, 0, len(roles))
	for _, role := range roles {
		result = append(result, dto.RoleListResponse{
			ID:        role.ID.String(),
			Name:      role.Name,
			Code:      role.Code,
			Status:    role.Status,
			IsSystem:  role.IsSystem,
			SortOrder: role.SortOrder,
		})
	}

	return result, total, nil
}

// Update 更新角色
// 系统内置角色不可修改状态和编码，防止 super_admin 等关键角色被禁用
func (s *roleService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}
	if role == nil {
		return nil, rbac.NewRoleNotFoundError()
	}

	// 系统内置角色保护：不可修改状态、编码和父级
	if role.IsSystem {
		if req.Status != nil && *req.Status != 0 {
			return nil, rbac.NewSystemRoleNoEditError()
		}
		if req.Code != "" {
			return nil, rbac.NewSystemRoleNoEditError()
		}
		if req.ParentID != nil {
			// 转成字符串比较，处理 nil 情况（空字符串表示设为根节点）
			var currentParent string
			if role.ParentID != nil {
				currentParent = role.ParentID.String()
			}
			if *req.ParentID != currentParent {
				return nil, rbac.NewSystemRoleNoEditError()
			}
		}
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Code != "" {
		// 检查编码是否已存在（排除自身）
		existing, err := s.roleRepo.GetByCode(ctx, req.Code)
		if err != nil {
			return nil, fmt.Errorf("check role code: %w", err)
		}
		if existing != nil && existing.ID != id {
			return nil, rbac.NewRoleCodeExistsError(req.Code)
		}
		role.Code = req.Code
	}
	// ParentID 使用 *string：nil 表示不修改，空字符串表示设为 null（提升为根节点）
	if req.ParentID != nil {
		if *req.ParentID == "" {
			role.ParentID = nil
		} else {
			parentID, err := uuid.Parse(*req.ParentID)
			if err != nil {
				return nil, fmt.Errorf("invalid parent_id: %w", err)
			}
			// 禁止设置自身为父节点
			if parentID == id {
				return nil, fmt.Errorf("parent_id cannot be self")
			}
			role.ParentID = &parentID
		}
	}
	if req.Description != nil {
		role.Description = *req.Description
	}
	if req.Status != nil {
		role.Status = *req.Status
	}

	if err := s.roleRepo.Update(ctx, nil, role); err != nil {
		return nil, fmt.Errorf("update role: %w", err)
	}

	// 角色状态变更后，失效相关用户权限缓存
	if req.Status != nil {
		if err := s.cacheService.InvalidateRolePermissions(ctx, id); err != nil {
			return nil, fmt.Errorf("invalidate cache: %w", err)
		}
	}

	return toRoleResponse(role), nil
}

// Delete 删除角色
// 系统内置角色不可删除，已被用户关联的角色不可删除
// 删除时在事务中清理角色-权限关联和用户-角色关联，避免产生孤儿数据
func (s *roleService) Delete(ctx context.Context, id uuid.UUID) error {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get role: %w", err)
	}
	if role == nil {
		return rbac.NewRoleNotFoundError()
	}

	// 系统内置角色不可删除
	if role.IsSystem {
		return rbac.NewSystemRoleNoDeleteError()
	}

	// 检查角色是否已关联用户
	count, err := s.roleRepo.CountUsersByRoleID(ctx, id)
	if err != nil {
		return fmt.Errorf("count role users: %w", err)
	}
	if count > 0 {
		return rbac.NewRoleInUseError()
	}

	// 在事务中删除角色及其关联数据
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 删除角色-权限关联
		if err := tx.WithContext(ctx).Where("role_id = ?", id).Delete(&model.RolePermission{}).Error; err != nil {
			return fmt.Errorf("delete role permissions: %w", err)
		}
		// 删除用户-角色关联
		if err := tx.WithContext(ctx).Where("role_id = ?", id).Delete(&model.UserRole{}).Error; err != nil {
			return fmt.Errorf("delete user roles: %w", err)
		}
		// 删除角色
		if err := tx.WithContext(ctx).Delete(&model.Role{}, id).Error; err != nil {
			return fmt.Errorf("delete role: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 失效该角色下所有用户的权限缓存
	if err := s.cacheService.InvalidateRolePermissions(ctx, id); err != nil {
		return fmt.Errorf("invalidate cache: %w", err)
	}

	return nil
}

// AssignPermissions 为角色分配权限（事务执行，并失效相关用户权限缓存）
// 系统内置角色不可修改权限，防止 super_admin 等关键角色权限被篡改
//
// 注意：角色存在性校验放在事务内（加行锁时一并判断），省去一次事务外的独立查询。
// 权限存在性校验仍放在事务外，避免持有行锁期间执行 N 次单条查询导致锁等待时间过长。
func (s *roleService) AssignPermissions(ctx context.Context, id uuid.UUID, req *dto.AssignPermissionsRequest) error {
	// 解析权限 ID 列表
	permIDs := make([]uuid.UUID, 0, len(req.PermissionIDs))
	for _, pidStr := range req.PermissionIDs {
		pid, err := uuid.Parse(pidStr)
		if err != nil {
			return fmt.Errorf("invalid permission id: %s", pidStr)
		}
		permIDs = append(permIDs, pid)
	}

	// 批量校验权限是否存在（在事务外执行，减少锁持有时间）
	// 注意：仅做存在性预校验，状态校验放在事务内（加锁后重新查询），避免 TOCTOU 竞态
	if len(permIDs) > 0 {
		perms, err := s.permissionRepo.GetByIDs(ctx, nil, permIDs)
		if err != nil {
			return fmt.Errorf("get permissions by ids: %w", err)
		}
		if len(perms) != len(permIDs) {
			return rbac.NewPermissionNotFoundError()
		}
	}

	// 事务执行权限分配
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 对角色记录加排他锁，防止并发分配权限时出现唯一键冲突
		// 同时通过行锁查询判断角色是否存在，省去一次独立的 GetByID 查询
		var role model.Role
		if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&role, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return rbac.NewRoleNotFoundError()
			}
			return fmt.Errorf("lock role: %w", err)
		}

		// 系统内置角色不可修改权限
		if role.IsSystem {
			return rbac.NewSystemRoleNoEditError()
		}

		// 事务内重新查询权限并校验状态，避免 TOCTOU 竞态
		if len(permIDs) > 0 {
			perms, err := s.permissionRepo.GetByIDs(ctx, tx, permIDs)
			if err != nil {
				return fmt.Errorf("get permissions by ids in tx: %w", err)
			}
			for _, p := range perms {
				if p.Status != 0 {
					return fmt.Errorf("permission %s is disabled", p.Code)
				}
			}
		}

		if err := s.roleRepo.AssignPermissions(ctx, tx, id, permIDs, req.DataScope); err != nil {
			return fmt.Errorf("assign permissions: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 失效该角色下所有用户的权限缓存
	if err := s.cacheService.InvalidateRolePermissions(ctx, id); err != nil {
		return fmt.Errorf("invalidate cache: %w", err)
	}

	return nil
}

// GetRoleUsers 分页查询角色下用户列表
// dataScope 为数据权限过滤条件，由中间件计算后传入，nil 表示不过滤
func (s *roleService) GetRoleUsers(ctx context.Context, id uuid.UUID, page, pageSize int, dataScope *model.DataScopeCondition) ([]dto.RoleUserResponse, int64, error) {
	// 校验角色是否存在
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, 0, fmt.Errorf("get role: %w", err)
	}
	if role == nil {
		return nil, 0, rbac.NewRoleNotFoundError()
	}

	results, total, err := s.roleRepo.GetUsersByRoleID(ctx, id, page, pageSize, dataScope)
	if err != nil {
		return nil, 0, fmt.Errorf("get role users: %w", err)
	}

	users := make([]dto.RoleUserResponse, 0, len(results))
	for _, item := range results {
		users = append(users, dto.RoleUserResponse{
			ID:       item.ID,
			Username: item.Username,
			RealName: item.RealName,
			Status:   item.Status,
		})
	}

	return users, total, nil
}

// toRoleResponse 将角色模型转换为响应 DTO
func toRoleResponse(role *model.Role) *dto.RoleResponse {
	resp := &dto.RoleResponse{
		ID:          role.ID.String(),
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		SortOrder:   role.SortOrder,
		Status:      role.Status,
		IsSystem:    role.IsSystem,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
	if role.ParentID != nil {
		parentID := role.ParentID.String()
		resp.ParentID = &parentID
	}
	return resp
}
