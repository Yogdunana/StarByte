package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/Yogdunana/StarByte/backend/internal/rbac"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/dto"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DepartmentService 部门服务接口
// 提供部门的增删改查及树结构构建等业务逻辑。
type DepartmentService interface {
	// Create 创建新部门，校验编码唯一性和父部门存在性
	Create(ctx context.Context, req *dto.CreateDepartmentRequest) (*dto.DepartmentResponse, error)
	// GetByID 根据 ID 查询部门详情
	GetByID(ctx context.Context, id uuid.UUID) (*dto.DepartmentResponse, error)
	// GetTree 获取完整的部门树结构（按 sort_order 排序）
	GetTree(ctx context.Context) ([]dto.DepartmentTreeResponse, error)
	// Update 更新部门信息，支持部分字段更新
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateDepartmentRequest) (*dto.DepartmentResponse, error)
	// Delete 删除部门，有子部门或已关联用户时不可删除
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
		// 验证父部门是否存在
		parent, err := s.deptRepo.GetByID(ctx, parentID)
		if err != nil {
			return nil, fmt.Errorf("get parent department: %w", err)
		}
		if parent == nil {
			return nil, rbac.NewDeptNotFoundError()
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

	if err := s.deptRepo.Create(ctx, nil, dept); err != nil {
		return nil, fmt.Errorf("create department: %w", err)
	}

	return toDepartmentResponse(dept), nil
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

	return toDepartmentResponse(dept), nil
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

	// 对每个父节点的子列表按 sort_order 排序，确保子节点顺序确定
	for _, children := range childrenMap {
		sort.Slice(children, func(i, j int) bool {
			if children[i].SortOrder != children[j].SortOrder {
				return children[i].SortOrder < children[j].SortOrder
			}
			return children[i].ID.String() < children[j].ID.String()
		})
	}

	// 构建根部门树（parent_id 为空即为根节点）
	tree := make([]dto.DepartmentTreeResponse, 0)
	for i := range depts {
		dept := &depts[i]
		if dept.ParentID == nil {
			tree = append(tree, *buildDepartmentTree(dept, childrenMap))
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
	if req.Description != nil {
		dept.Description = *req.Description
	}
	if req.SortOrder != nil {
		dept.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		dept.Status = *req.Status
	}
	if req.LeaderID != nil {
		if *req.LeaderID == "" {
			dept.LeaderID = nil
		} else {
			leaderID, err := uuid.Parse(*req.LeaderID)
			if err != nil {
				return nil, fmt.Errorf("parse leader_id: %w", err)
			}
			dept.LeaderID = &leaderID
		}
	}

	if err := s.deptRepo.Update(ctx, nil, dept); err != nil {
		return nil, fmt.Errorf("update department: %w", err)
	}

	return toDepartmentResponse(dept), nil
}

// Delete 删除部门
// 有子部门或已关联用户的部门不可删除
// 使用事务确保检查和删除的原子性，避免 TOCTOU 竞态条件
func (s *departmentService) Delete(ctx context.Context, id uuid.UUID) error {
	dept, err := s.deptRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get department: %w", err)
	}
	if dept == nil {
		return rbac.NewDeptNotFoundError()
	}

	// 在事务内执行检查和删除，确保原子性
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 检查是否有子部门
		childCount, err := s.deptRepo.CountChildren(ctx, tx, id)
		if err != nil {
			return fmt.Errorf("count children: %w", err)
		}
		if childCount > 0 {
			return rbac.NewDeptHasChildrenError()
		}

		// 检查是否有用户属于该部门
		userCount, err := s.deptRepo.CountUsersByDeptID(ctx, tx, id)
		if err != nil {
			return fmt.Errorf("count department users: %w", err)
		}
		if userCount > 0 {
			return rbac.NewDeptInUseError()
		}

		// 执行删除
		if err := s.deptRepo.Delete(ctx, tx, id); err != nil {
			return fmt.Errorf("delete department: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ========== 工具函数 ==========

// toDepartmentResponse 将部门模型转换为响应 DTO
func toDepartmentResponse(dept *model.Department) *dto.DepartmentResponse {
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
		parentID := dept.ParentID.String()
		resp.ParentID = &parentID
	}
	if dept.LeaderID != nil {
		leaderID := dept.LeaderID.String()
		resp.LeaderID = &leaderID
	}
	return resp
}

// buildDepartmentTree 递归构建部门树
func buildDepartmentTree(dept *model.Department, childrenMap map[uuid.UUID][]*model.Department) *dto.DepartmentTreeResponse {
	resp := toDepartmentResponse(dept)

	children := childrenMap[dept.ID]
	if len(children) > 0 {
		resp.Children = make([]dto.DepartmentTreeResponse, 0, len(children))
		for _, child := range children {
			resp.Children = append(resp.Children, *buildDepartmentTree(child, childrenMap))
		}
	}

	return resp
}
