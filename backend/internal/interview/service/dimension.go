package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/internal/interview/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *interviewService) ListDimensions(ctx context.Context) ([]dto.DimensionResponse, error) {
	rows, err := s.evals.ListDimensions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list dimensions: %w", err)
	}
	out := make([]dto.DimensionResponse, 0, len(rows))
	for i := range rows {
		out = append(out, *mapDimension(&rows[i]))
	}
	return out, nil
}

func (s *interviewService) CreateDimension(ctx context.Context, req *dto.CreateDimensionRequest) (*dto.DimensionResponse, error) {
	exist, err := s.evals.GetDimensionByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("get dimension: %w", err)
	}
	if exist != nil {
		return nil, response.NewError(response.CodeConflict, "维度名称已存在")
	}
	now := time.Now()
	d := &model.Dimension{
		ID: uuid.New(), Name: req.Name, Weight: req.Weight,
		MaxScore: req.MaxScore, SortOrder: req.SortOrder,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := s.evals.CreateDimension(ctx, d); err != nil {
		return nil, fmt.Errorf("create dimension: %w", err)
	}
	return mapDimension(d), nil
}

func (s *interviewService) UpdateDimension(ctx context.Context, id uuid.UUID, req *dto.UpdateDimensionRequest) (*dto.DimensionResponse, error) {
	d, err := s.evals.GetDimensionByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get dimension: %w", err)
	}
	if d == nil {
		return nil, response.NewError(response.CodeInterviewDimGone, "维度不存在")
	}
	if req.Name != nil {
		d.Name = *req.Name
	}
	if req.Weight != nil {
		d.Weight = *req.Weight
	}
	if req.MaxScore != nil {
		d.MaxScore = *req.MaxScore
	}
	if req.SortOrder != nil {
		d.SortOrder = *req.SortOrder
	}
	d.UpdatedAt = time.Now()
	if err := s.evals.UpdateDimension(ctx, d); err != nil {
		return nil, fmt.Errorf("update dimension: %w", err)
	}
	return mapDimension(d), nil
}

func (s *interviewService) DeleteDimension(ctx context.Context, id uuid.UUID) error {
	d, err := s.evals.GetDimensionByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get dimension: %w", err)
	}
	if d == nil {
		return response.NewError(response.CodeInterviewDimGone, "维度不存在")
	}
	return s.evals.DeleteDimension(ctx, id)
}

func (s *interviewService) Stats(ctx context.Context, q *dto.StatsQuery) (*dto.StatsResponse, error) {
	if q == nil {
		q = &dto.StatsQuery{}
	}
	row, buckets, depts, err := s.evals.Stats(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("stats: %w", err)
	}
	var rate float64
	if row.Total > 0 {
		rate = float64(int(float64(row.PassCount)/float64(row.Total)*1000+0.5)) / 10
	}
	out := &dto.StatsResponse{
		Total: row.Total, PassCount: row.PassCount, FailCount: row.FailCount,
		PendingCount: row.PendingCount, PassRate: rate,
		ScoreBuckets: make([]dto.ScoreBucketVO, 0, len(buckets)),
		ByDepartment: make([]dto.DeptStatVO, 0, len(depts)),
	}
	for _, b := range buckets {
		out.ScoreBuckets = append(out.ScoreBuckets, dto.ScoreBucketVO{Range: b.Range, Count: b.Count})
	}
	for _, d := range depts {
		out.ByDepartment = append(out.ByDepartment, dto.DeptStatVO{
			Department: d.Department, Count: d.Count, PassCount: d.PassCount,
		})
	}
	return out, nil
}
