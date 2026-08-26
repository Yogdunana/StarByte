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
// 提供职位的增删改查等业务逻辑。
type PositionService interface {
	// Create 创建新职位，校验编码唯一性
	Create(ctx context.Context, req *dto.CreatePositionRequest) (*dto.PositionResponse, error)
	// GetByID 根据 ID 查询职位详情
	GetByID(ctx context.Context, id uuid.UUID) (*dto.PositionResponse, error)
	// List 分页查询职位列表，支持关键字模糊搜索
	List(ctx context.Context, req *dto.ListPositionRequest) ([]dto.PositionResponse, int64, error)
	// Update 更新职位信息，支持部分字段更新
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdatePositionRequest) (*dto.PositionResponse, error)
	// Delete 删除职位，已关联用户的职位不可删除
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

	if err := s.positionRepo.Create(ctx, nil, pos); err != nil {
		return nil, fmt.Errorf("create position: %w", err)
	}

	return toPositionResponse(pos), nil
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

	return toPositionResponse(pos), nil
}

// List 获取职位列表
func (s *positionService) List(ctx context.Context, req *dto.ListPositionRequest) ([]dto.PositionResponse, int64, error) {
	positions, total, err := s.positionRepo.List(ctx, req.Page, req.PageSize, req.Keyword)
	if err != nil {
		return nil, 0, fmt.Errorf("list positions: %w", err)
	}

	result := make([]dto.PositionResponse, 0, len(positions))
	for i := range positions {
		result = append(result, *toPositionResponse(&positions[i]))
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
	if req.Description != nil {
		pos.Description = *req.Description
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

	if err := s.positionRepo.Update(ctx, nil, pos); err != nil {
		return nil, fmt.Errorf("update position: %w", err)
	}

	return toPositionResponse(pos), nil
}

// Delete 删除职位
// 已关联用户的职位不可删除
// 使用事务确保检查和删除的原子性，避免 TOCTOU 竞态条件
func (s *positionService) Delete(ctx context.Context, id uuid.UUID) error {
	pos, err := s.positionRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get position: %w", err)
	}
	if pos == nil {
		return rbac.NewPositionNotFoundError()
	}

	// 在事务内执行检查和删除，确保原子性
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 检查是否有用户使用该职位
		userCount, err := s.positionRepo.CountUsersByPositionID(ctx, tx, id)
		if err != nil {
			return fmt.Errorf("count position users: %w", err)
		}
		if userCount > 0 {
			return rbac.NewPositionInUseError()
		}

		// 执行删除
		if err := s.positionRepo.Delete(ctx, tx, id); err != nil {
			return fmt.Errorf("delete position: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ========== 工具函数 ==========

// toPositionResponse 将职位模型转换为响应 DTO
func toPositionResponse(pos *model.Position) *dto.PositionResponse {
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
