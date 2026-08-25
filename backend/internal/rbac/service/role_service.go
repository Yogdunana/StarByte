package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Yogdunana/StarByte/backend/internal/rbac"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/dto"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RoleService 角色服务接口
type RoleService interface {
	Create(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.RoleDetailResponse, error)
	List(ctx context.Context, req *dto.ListRoleRequest) ([]dto.RoleListResponse, int64, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AssignPermissions(ctx context.Context, id uuid.UUID, req *dto.AssignPermissionsRequest) error
	GetRoleUsers(ctx context.Context, id uuid.UUID, page, pageSize int) ([]dto.RoleUserResponse, int64, error)
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

	if err := s.roleRepo.Create(ctx, nil, role); err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}

	return s.toRoleResponse(role), nil
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
		RoleResponse:  *s.toRoleResponse(role),
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
func (s *roleService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}
	if role == nil {
		return nil, rbac.NewRoleNotFoundError()
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}

	if err := s.roleRepo.Update(ctx, nil, role); err != nil {
		return nil, fmt.Errorf("update role: %w", err)
	}

	return s.toRoleResponse(role), nil
}

// Delete 删除角色
// 系统内置角色不可删除，已被用户关联的角色不可删除
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

	if err := s.roleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete role: %w", err)
	}

	return nil
}

// AssignPermissions 为角色分配权限（事务执行，并失效相关用户权限缓存）
func (s *roleService) AssignPermissions(ctx context.Context, id uuid.UUID, req *dto.AssignPermissionsRequest) error {
	// 校验角色是否存在
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get role: %w", err)
	}
	if role == nil {
		return rbac.NewRoleNotFoundError()
	}

	// 解析权限 ID 列表
	permIDs := make([]uuid.UUID, 0, len(req.PermissionIDs))
	for _, pidStr := range req.PermissionIDs {
		pid, err := uuid.Parse(pidStr)
		if err != nil {
			return response.NewError(response.CodeBadRequest, "无效的权限ID: "+pidStr)
		}
		permIDs = append(permIDs, pid)
	}

	// 校验权限是否存在
	for _, pid := range permIDs {
		perm, err := s.permissionRepo.GetByID(ctx, pid)
		if err != nil {
			return fmt.Errorf("get permission: %w", err)
		}
		if perm == nil {
			return rbac.NewPermissionNotFoundError()
		}
	}

	// 事务执行权限分配
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.roleRepo.AssignPermissions(ctx, tx, id, permIDs); err != nil {
			return fmt.Errorf("assign permissions: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 失效该角色下所有用户的权限缓存（失败不影响主流程，仅记录日志）
	if err := s.cacheService.InvalidateRolePermissions(ctx, id); err != nil {
		logger.Warn("invalidate role permissions cache failed",
			zap.String("role_id", id.String()),
			zap.Error(err))
	}

	return nil
}

// GetRoleUsers 分页查询角色下用户列表
func (s *roleService) GetRoleUsers(ctx context.Context, id uuid.UUID, page, pageSize int) ([]dto.RoleUserResponse, int64, error) {
	// 校验角色是否存在
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, 0, fmt.Errorf("get role: %w", err)
	}
	if role == nil {
		return nil, 0, rbac.NewRoleNotFoundError()
	}

	results, total, err := s.roleRepo.GetUsersByRoleID(ctx, id, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("get role users: %w", err)
	}

	users := make([]dto.RoleUserResponse, 0, len(results))
	for _, m := range results {
		users = append(users, dto.RoleUserResponse{
			ID:       mapGetString(m, "id"),
			Username: mapGetString(m, "username"),
			RealName: mapGetString(m, "real_name"),
			Status:   mapGetInt(m, "status"),
		})
	}

	return users, total, nil
}

// ========== 工具函数 ==========

// toRoleResponse 将角色模型转换为响应 DTO
func (s *roleService) toRoleResponse(role *model.Role) *dto.RoleResponse {
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
		resp.ParentID = role.ParentID.String()
	}
	return resp
}

// mapGetString 从 map[string]interface{} 中安全提取字符串
func mapGetString(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// mapGetInt 从 map[string]interface{} 中安全提取整型
func mapGetInt(m map[string]interface{}, key string) int {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch val := v.(type) {
	case int:
		return val
	case int16:
		return int(val)
	case int32:
		return int(val)
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		i, _ := strconv.Atoi(val)
		return i
	default:
		return 0
	}
}
