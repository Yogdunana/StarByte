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
)

// PermissionService 权限服务接口
type PermissionService interface {
	Create(ctx context.Context, req *dto.CreatePermissionRequest) (*dto.PermissionResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.PermissionResponse, error)
	GetTree(ctx context.Context) ([]dto.PermissionTreeResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdatePermissionRequest) (*dto.PermissionResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type permissionService struct {
	db             *gorm.DB
	permissionRepo repo.PermissionRepo
}

// NewPermissionService 创建权限服务
func NewPermissionService(db *gorm.DB, permissionRepo repo.PermissionRepo) PermissionService {
	return &permissionService{
		db:             db,
		permissionRepo: permissionRepo,
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
		Type:        parsePermissionType(req.Type),
	}

	// 解析 parent_id
	if req.ParentID != "" {
		parentID, err := uuid.Parse(req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("invalid parent_id: %w", err)
		}
		permission.ParentID = &parentID
	}

	if err := s.permissionRepo.Create(ctx, nil, permission); err != nil {
		return nil, fmt.Errorf("create permission: %w", err)
	}

	return s.GetByID(ctx, permission.ID)
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

	return s.toPermissionResponse(perm), nil
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

	// 构建根权限树（parent_id 为空即为根节点）
	tree := make([]dto.PermissionTreeResponse, 0)
	for i := range permissions {
		perm := &permissions[i]
		if perm.ParentID == nil {
			tree = append(tree, *s.buildPermissionTree(perm, childrenMap))
		}
	}

	return tree, nil
}

// Update 更新权限
func (s *permissionService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdatePermissionRequest) (*dto.PermissionResponse, error) {
	perm, err := s.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get permission: %w", err)
	}
	if perm == nil {
		return nil, rbac.NewPermissionNotFoundError()
	}

	if req.Name != "" {
		perm.Name = req.Name
	}
	if req.Description != "" {
		perm.Description = req.Description
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

	return s.GetByID(ctx, id)
}

// Delete 删除权限（系统内置权限不可删除，有子节点的权限不可删除）
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

	if err := s.permissionRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete permission: %w", err)
	}

	return nil
}

// ========== 工具函数 ==========

// parsePermissionType 将权限类型字符串转换为模型类型常量
func parsePermissionType(t string) int {
	switch t {
	case "menu":
		return model.PermissionTypeMenu
	case "button":
		return model.PermissionTypeButton
	case "api":
		return model.PermissionTypeAPI
	default:
		return model.PermissionTypeMenu
	}
}

// permissionTypeToString 将模型类型常量转换为权限类型字符串
func permissionTypeToString(t int) string {
	switch t {
	case model.PermissionTypeMenu:
		return "menu"
	case model.PermissionTypeButton:
		return "button"
	case model.PermissionTypeAPI:
		return "api"
	default:
		return "menu"
	}
}

// toPermissionResponse 将权限模型转换为响应 DTO
func (s *permissionService) toPermissionResponse(perm *model.Permission) *dto.PermissionResponse {
	resp := &dto.PermissionResponse{
		ID:          perm.ID.String(),
		Name:        perm.Name,
		Code:        perm.Code,
		Resource:    perm.Resource,
		Action:      perm.Action,
		Description: perm.Description,
		SortOrder:   perm.SortOrder,
		Type:        permissionTypeToString(perm.Type),
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
		resp.ParentID = perm.ParentID.String()
	}
	return resp
}

// buildPermissionTree 递归构建权限树
func (s *permissionService) buildPermissionTree(perm *model.Permission, childrenMap map[uuid.UUID][]*model.Permission) *dto.PermissionResponse {
	resp := s.toPermissionResponse(perm)

	children := childrenMap[perm.ID]
	if len(children) > 0 {
		resp.Children = make([]dto.PermissionTreeResponse, 0, len(children))
		for _, child := range children {
			resp.Children = append(resp.Children, *s.buildPermissionTree(child, childrenMap))
		}
	}

	return resp
}
