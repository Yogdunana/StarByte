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

// DepartmentService 部门服务接口
type DepartmentService interface {
	Create(ctx context.Context, req *dto.CreateDepartmentRequest) (*dto.DepartmentResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.DepartmentResponse, error)
	GetTree(ctx context.Context) ([]dto.DepartmentTreeResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateDepartmentRequest) (*dto.DepartmentResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type departmentService struct {
	db       *gorm.DB
	deptRepo repo.DepartmentRepo
}

// NewDepartmentService 创建部门服务
func NewDepartmentService(db *gorm.DB, deptRepo repo.DepartmentRepo) DepartmentService {
	return &departmentService{
		db:       db,
		deptRepo: deptRepo,
	}
}

// Create 创建部门
func (s *departmentService) Create(ctx context.Context, req *dto.CreateDepartmentRequest) (*dto.DepartmentResponse, error) {
	// 检查编码是否已存在
	existing, err := s.deptRepo.GetByCode(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("check department code: %w", err)
	}
	if existing != nil {
		return nil, rbac.NewDeptCodeExistsError(req.Code)
	}

	dept := &model.Department{
		ID:          uuid.New(),
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
	}

	if req.SortOrder != nil {
		dept.SortOrder = *req.SortOrder
	}

	// 解析 parent_id
	if req.ParentID != "" {
		parentID, err := uuid.Parse(req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parse parent_id: %w", err)
		}
		dept.ParentID = &parentID
	}

	// 解析 leader_id
	if req.LeaderID != "" {
		leaderID, err := uuid.Parse(req.LeaderID)
		if err != nil {
			return nil, fmt.Errorf("parse leader_id: %w", err)
		}
		dept.LeaderID = &leaderID
	}

	err = s.deptRepo.Create(ctx, nil, dept)
	if err != nil {
		return nil, fmt.Errorf("create department: %w", err)
	}

	return s.GetByID(ctx, dept.ID)
}

// GetByID 获取部门详情
func (s *departmentService) GetByID(ctx context.Context, id uuid.UUID) (*dto.DepartmentResponse, error) {
	dept, err := s.deptRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get department: %w", err)
	}
	if dept == nil {
		return nil, rbac.NewDeptNotFoundError()
	}

	return s.toDepartmentResponse(dept), nil
}

// GetTree 获取部门树
func (s *departmentService) GetTree(ctx context.Context) ([]dto.DepartmentTreeResponse, error) {
	depts, err := s.deptRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}

	// 按 parent_id 分组子部门
	childrenMap := make(map[uuid.UUID][]*model.Department)
	for i := range depts {
		dept := &depts[i]
		if dept.ParentID != nil {
			childrenMap[*dept.ParentID] = append(childrenMap[*dept.ParentID], dept)
		}
	}

	// 构建根部门树（parent_id 为空即为根节点）
	tree := make([]dto.DepartmentTreeResponse, 0)
	for i := range depts {
		dept := &depts[i]
		if dept.ParentID == nil {
			tree = append(tree, *s.buildDepartmentTree(dept, childrenMap))
		}
	}

	return tree, nil
}

// Update 更新部门
func (s *departmentService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateDepartmentRequest) (*dto.DepartmentResponse, error) {
	dept, err := s.deptRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get department: %w", err)
	}
	if dept == nil {
		return nil, rbac.NewDeptNotFoundError()
	}

	if req.Name != "" {
		dept.Name = req.Name
	}
	if req.Description != "" {
		dept.Description = req.Description
	}
	if req.SortOrder != nil {
		dept.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		dept.Status = *req.Status
	}
	if req.LeaderID != "" {
		leaderID, err := uuid.Parse(req.LeaderID)
		if err != nil {
			return nil, fmt.Errorf("parse leader_id: %w", err)
		}
		dept.LeaderID = &leaderID
	}

	err = s.deptRepo.Update(ctx, nil, dept)
	if err != nil {
		return nil, fmt.Errorf("update department: %w", err)
	}

	return s.GetByID(ctx, id)
}

// Delete 删除部门
func (s *departmentService) Delete(ctx context.Context, id uuid.UUID) error {
	dept, err := s.deptRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get department: %w", err)
	}
	if dept == nil {
		return rbac.NewDeptNotFoundError()
	}

	// 检查是否有子部门
	count, err := s.deptRepo.CountChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("count children: %w", err)
	}
	if count > 0 {
		return rbac.NewDeptHasChildrenError()
	}

	return s.deptRepo.Delete(ctx, id)
}

// ========== 工具函数 ==========

// toDepartmentResponse 将部门模型转换为响应 DTO
func (s *departmentService) toDepartmentResponse(dept *model.Department) *dto.DepartmentResponse {
	resp := &dto.DepartmentResponse{
		ID:          dept.ID.String(),
		Name:        dept.Name,
		Code:        dept.Code,
		Description: dept.Description,
		SortOrder:   dept.SortOrder,
		Status:      dept.Status,
		CreatedAt:   dept.CreatedAt,
		UpdatedAt:   dept.UpdatedAt,
	}
	if dept.ParentID != nil {
		resp.ParentID = dept.ParentID.String()
	}
	if dept.LeaderID != nil {
		resp.LeaderID = dept.LeaderID.String()
	}
	return resp
}

// buildDepartmentTree 递归构建部门树
func (s *departmentService) buildDepartmentTree(dept *model.Department, childrenMap map[uuid.UUID][]*model.Department) *dto.DepartmentTreeResponse {
	resp := s.toDepartmentResponse(dept)

	children := childrenMap[dept.ID]
	if len(children) > 0 {
		resp.Children = make([]dto.DepartmentTreeResponse, 0, len(children))
		for _, child := range children {
			resp.Children = append(resp.Children, *s.buildDepartmentTree(child, childrenMap))
		}
	}

	return resp
}
