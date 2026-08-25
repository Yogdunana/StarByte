package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/Yogdunana/StarByte/backend/internal/rbac"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/dto"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PermissionService 权限服务接口
// 提供权限的增删改查及树结构构建等业务逻辑，含缓存失效处理。
type PermissionService interface {
	// Create 创建新权限，校验编码唯一性和父权限存在性
	Create(ctx context.Context, req *dto.CreatePermissionRequest) (*dto.PermissionResponse, error)
	// GetByID 根据 ID 查询权限详情
	GetByID(ctx context.Context, id uuid.UUID) (*dto.PermissionResponse, error)
	// GetTree 获取完整的权限树结构（按 sort_order 排序）
	GetTree(ctx context.Context) ([]dto.PermissionTreeResponse, error)
	// Update 更新权限信息，系统内置权限不可修改状态
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdatePermissionRequest) (*dto.PermissionResponse, error)
	// Delete 删除权限，系统内置权限和有子节点的权限不可删除
	Delete(ctx context.Context, id uuid.UUID) error
}

type permissionService struct {
	db             *gorm.DB
	permissionRepo repo.PermissionRepo
	cacheService   PermissionCacheService
}

// NewPermissionService 创建权限服务
func NewPermissionService(db *gorm.DB, permissionRepo repo.PermissionRepo, cacheService PermissionCacheService) PermissionService {
	return &permissionService{
		db:             db,
		permissionRepo: permissionRepo,
		cacheService:   cacheService,
	}
}

// Create 创建权限
func (s *permissionService) Create(ctx context.Context, req *dto.CreatePermissionRequest) (*dto.PermissionResponse, error) {
	// 检查权限编码是否已存在
	existing, err := s.permissionRepo.GetByCode(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("check permission code: %w", err)
	}
	if existing != nil {
		return nil, rbac.NewPermissionCodeExistsError(req.Code)
	}

	permission := &model.Permission{
		ID:          uuid.New(),
		Name:        req.Name,
		Code:        req.Code,
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
		Path:        req.Path,
		Icon:        req.Icon,
		APIMethod:   req.APIMethod,
		APIPath:     req.APIPath,
		Type:        model.ParsePermissionType(req.Type),
	}

	// 解析 parent_id
	if req.ParentID != "" {
		parentID, err := uuid.Parse(req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("invalid parent_id: %w", err)
		}
		// 验证父权限是否存在
		parent, err := s.permissionRepo.GetByID(ctx, parentID)
		if err != nil {
			return nil, fmt.Errorf("get parent permission: %w", err)
		}
		if parent == nil {
			return nil, rbac.NewPermissionNotFoundError()
		}
		permission.ParentID = &parentID
	}

	if err := s.permissionRepo.Create(ctx, nil, permission); err != nil {
		return nil, fmt.Errorf("create permission: %w", err)
	}

	return toPermissionResponse(permission), nil
}

// GetByID 获取权限详情
func (s *permissionService) GetByID(ctx context.Context, id uuid.UUID) (*dto.PermissionResponse, error) {
	perm, err := s.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get permission: %w", err)
	}
	if perm == nil {
		return nil, rbac.NewPermissionNotFoundError()
	}

	return toPermissionResponse(perm), nil
}

// GetTree 获取权限树
func (s *permissionService) GetTree(ctx context.Context) ([]dto.PermissionTreeResponse, error) {
	permissions, err := s.permissionRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}

	// 按 parent_id 分组子权限
	childrenMap := make(map[uuid.UUID][]*model.Permission)
	for i := range permissions {
		perm := &permissions[i]
		if perm.ParentID != nil {
			childrenMap[*perm.ParentID] = append(childrenMap[*perm.ParentID], perm)
		}
	}

	// 对每个父节点的子列表按 sort_order 排序，确保子节点顺序确定
	for _, children := range childrenMap {
		sort.Slice(children, func(i, j int) bool {
			if children[i].SortOrder != children[j].SortOrder {
				return children[i].SortOrder < children[j].SortOrder
			}
			return children[i].ID.String() < children[j].ID.String()
		})
	}

	// 构建根权限树（parent_id 为空即为根节点）
	tree := make([]dto.PermissionTreeResponse, 0)
	for i := range permissions {
		perm := &permissions[i]
		if perm.ParentID == nil {
			tree = append(tree, *buildPermissionTree(perm, childrenMap))
		}
	}

	return tree, nil
}

// Update 更新权限
// 系统内置权限不可修改状态和编码，防止关键权限被禁用
func (s *permissionService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdatePermissionRequest) (*dto.PermissionResponse, error) {
	perm, err := s.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get permission: %w", err)
	}
	if perm == nil {
		return nil, rbac.NewPermissionNotFoundError()
	}

	// 系统内置权限保护：不可修改状态
	if perm.IsSystem && req.Status != nil && *req.Status != 0 {
		return nil, rbac.NewSystemPermissionNoEditError()
	}

	if req.Name != "" {
		perm.Name = req.Name
	}
	if req.Description != nil {
		perm.Description = *req.Description
	}
	if req.Path != "" {
		perm.Path = req.Path
	}
	if req.Icon != "" {
		perm.Icon = req.Icon
	}
	if req.SortOrder != nil {
		perm.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		perm.Status = *req.Status
	}

	if err := s.permissionRepo.Update(ctx, nil, perm); err != nil {
		return nil, fmt.Errorf("update permission: %w", err)
	}

	return toPermissionResponse(perm), nil
}

// Delete 删除权限（系统内置权限不可删除，有子节点的权限不可删除）
// 删除时在事务中清理 role_data_scopes、role_permissions 关联，避免产生孤儿数据
// 删除后失效所有关联角色下用户的权限缓存
func (s *permissionService) Delete(ctx context.Context, id uuid.UUID) error {
	perm, err := s.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get permission: %w", err)
	}
	if perm == nil {
		return rbac.NewPermissionNotFoundError()
	}

	// 系统内置权限不可删除
	if perm.IsSystem {
		return rbac.NewSystemPermissionNoDeleteError()
	}

	// 有子节点则不可删除
	childCount, err := s.permissionRepo.CountChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("count children: %w", err)
	}
	if childCount > 0 {
		return rbac.NewPermissionHasChildrenError()
	}

	// 事务前查询所有拥有该权限的角色 ID，用于后续缓存失效
	var roleIDs []uuid.UUID
	err = s.db.WithContext(ctx).
		Model(&model.RolePermission{}).
		Where("permission_id = ?", id).
		Pluck("role_id", &roleIDs).Error
	if err != nil {
		return fmt.Errorf("get role ids by permission: %w", err)
	}

	// 在事务中删除权限及其关联数据
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 先删除 role_data_scopes 中关联的记录（通过 role_permission_id 子查询）
		if err := tx.WithContext(ctx).
			Where("role_permission_id IN (SELECT id FROM role_permissions WHERE permission_id = ?)", id).
			Delete(&model.RoleDataScope{}).Error; err != nil {
			return fmt.Errorf("delete role data scopes: %w", err)
		}
		// 删除角色-权限关联
		if err := tx.WithContext(ctx).Where("permission_id = ?", id).Delete(&model.RolePermission{}).Error; err != nil {
			return fmt.Errorf("delete role permissions: %w", err)
		}
		// 删除权限
		if err := tx.WithContext(ctx).Delete(&model.Permission{}, id).Error; err != nil {
			return fmt.Errorf("delete permission: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 失效所有关联角色下用户的权限缓存（失败不影响主流程，仅记录日志）
	for _, roleID := range roleIDs {
		if invErr := s.cacheService.InvalidateRolePermissions(ctx, roleID); invErr != nil {
			logger.Warn("invalidate role permissions cache failed after permission delete",
				zap.String("permission_id", id.String()),
				zap.String("role_id", roleID.String()),
				zap.Error(invErr))
		}
	}

	return nil
}

// ========== 工具函数 ==========

// toPermissionResponse 将权限模型转换为响应 DTO
func toPermissionResponse(perm *model.Permission) *dto.PermissionResponse {
	resp := &dto.PermissionResponse{
		ID:          perm.ID.String(),
		Name:        perm.Name,
		Code:        perm.Code,
		Resource:    perm.Resource,
		Action:      perm.Action,
		Description: perm.Description,
		SortOrder:   perm.SortOrder,
		Type:        perm.Type.String(),
		Path:        perm.Path,
		Icon:        perm.Icon,
		APIMethod:   perm.APIMethod,
		APIPath:     perm.APIPath,
		IsSystem:    perm.IsSystem,
		Status:      perm.Status,
		CreatedAt:   perm.CreatedAt,
		UpdatedAt:   perm.UpdatedAt,
	}
	if perm.ParentID != nil {
		parentID := perm.ParentID.String()
		resp.ParentID = &parentID
	}
	return resp
}

// buildPermissionTree 递归构建权限树
func buildPermissionTree(perm *model.Permission, childrenMap map[uuid.UUID][]*model.Permission) *dto.PermissionResponse {
	resp := toPermissionResponse(perm)

	children := childrenMap[perm.ID]
	if len(children) > 0 {
		resp.Children = make([]dto.PermissionTreeResponse, 0, len(children))
		for _, child := range children {
			resp.Children = append(resp.Children, *buildPermissionTree(child, childrenMap))
		}
	}

	return resp
}
