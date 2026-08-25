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

// PositionService 职位服务接口
type PositionService interface {
	Create(ctx context.Context, req *dto.CreatePositionRequest) (*dto.PositionResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.PositionResponse, error)
	List(ctx context.Context, req *dto.ListPositionRequest) ([]dto.PositionResponse, int64, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdatePositionRequest) (*dto.PositionResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type positionService struct {
	db           *gorm.DB
	positionRepo repo.PositionRepo
}

// NewPositionService 创建职位服务
func NewPositionService(db *gorm.DB, positionRepo repo.PositionRepo) PositionService {
	return &positionService{
		db:           db,
		positionRepo: positionRepo,
	}
}

// Create 创建职位
func (s *positionService) Create(ctx context.Context, req *dto.CreatePositionRequest) (*dto.PositionResponse, error) {
	// 检查编码是否已存在
	existing, err := s.positionRepo.GetByCode(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("check position code: %w", err)
	}
	if existing != nil {
		return nil, rbac.NewPositionCodeExistsError(req.Code)
	}

	pos := &model.Position{
		ID:          uuid.New(),
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
	}

	if req.Level != nil {
		pos.Level = *req.Level
	}
	if req.VoteWeight != nil {
		pos.VoteWeight = *req.VoteWeight
	}
	if req.SortOrder != nil {
		pos.SortOrder = *req.SortOrder
	}

	err = s.positionRepo.Create(ctx, nil, pos)
	if err != nil {
		return nil, fmt.Errorf("create position: %w", err)
	}

	return s.GetByID(ctx, pos.ID)
}

// GetByID 获取职位详情
func (s *positionService) GetByID(ctx context.Context, id uuid.UUID) (*dto.PositionResponse, error) {
	pos, err := s.positionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get position: %w", err)
	}
	if pos == nil {
		return nil, rbac.NewPositionNotFoundError()
	}

	return s.toPositionResponse(pos), nil
}

// List 获取职位列表
func (s *positionService) List(ctx context.Context, req *dto.ListPositionRequest) ([]dto.PositionResponse, int64, error) {
	positions, total, err := s.positionRepo.List(ctx, req.Page, req.PageSize, req.Keyword)
	if err != nil {
		return nil, 0, fmt.Errorf("list positions: %w", err)
	}

	result := make([]dto.PositionResponse, 0, len(positions))
	for i := range positions {
		result = append(result, *s.toPositionResponse(&positions[i]))
	}

	return result, total, nil
}

// Update 更新职位
func (s *positionService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdatePositionRequest) (*dto.PositionResponse, error) {
	pos, err := s.positionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get position: %w", err)
	}
	if pos == nil {
		return nil, rbac.NewPositionNotFoundError()
	}

	if req.Name != "" {
		pos.Name = req.Name
	}
	if req.Description != "" {
		pos.Description = req.Description
	}
	if req.Level != nil {
		pos.Level = *req.Level
	}
	if req.VoteWeight != nil {
		pos.VoteWeight = *req.VoteWeight
	}
	if req.SortOrder != nil {
		pos.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		pos.Status = *req.Status
	}

	err = s.positionRepo.Update(ctx, nil, pos)
	if err != nil {
		return nil, fmt.Errorf("update position: %w", err)
	}

	return s.GetByID(ctx, id)
}

// Delete 删除职位
func (s *positionService) Delete(ctx context.Context, id uuid.UUID) error {
	pos, err := s.positionRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get position: %w", err)
	}
	if pos == nil {
		return rbac.NewPositionNotFoundError()
	}

	return s.positionRepo.Delete(ctx, id)
}

// ========== 工具函数 ==========

// toPositionResponse 将职位模型转换为响应 DTO
func (s *positionService) toPositionResponse(pos *model.Position) *dto.PositionResponse {
	return &dto.PositionResponse{
		ID:          pos.ID.String(),
		Name:        pos.Name,
		Code:        pos.Code,
		Level:       pos.Level,
		VoteWeight:  pos.VoteWeight,
		Description: pos.Description,
		SortOrder:   pos.SortOrder,
		Status:      pos.Status,
		CreatedAt:   pos.CreatedAt,
		UpdatedAt:   pos.UpdatedAt,
	}
}
