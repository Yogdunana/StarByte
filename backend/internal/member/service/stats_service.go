package service

import (
	"context"
	"fmt"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
)

func (s *memberService) ApplicationStats(ctx context.Context, q *dto.StatsQuery) (*dto.StatsResponse, error) {
	groupBy := q.GroupBy
	if groupBy == "" {
		groupBy = "date"
	}
	if groupBy != "date" && groupBy != "department" && groupBy != "type" {
		groupBy = "date"
	}
	rows, err := s.apps.Stats(ctx, q.StartDate, q.EndDate, groupBy)
	if err != nil {
		return nil, fmt.Errorf("application stats: %w", err)
	}
	return toStats(groupBy, rows), nil
}

func (s *memberService) MemberStats(ctx context.Context, q *dto.StatsQuery) (*dto.StatsResponse, error) {
	groupBy := q.GroupBy
	if groupBy == "" {
		groupBy = "department"
	}
	if groupBy != "department" && groupBy != "grade" && groupBy != "type" && groupBy != "status" {
		groupBy = "department"
	}
	rows, err := s.profs.Stats(ctx, groupBy)
	if err != nil {
		return nil, fmt.Errorf("member stats: %w", err)
	}
	return toStats(groupBy, rows), nil
}

func toStats(groupBy string, rows []model.StatBucket) *dto.StatsResponse {
	items := make([]dto.StatItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, dto.StatItem{Key: r.Key, Label: r.Label, Count: r.Count})
	}
	return &dto.StatsResponse{GroupBy: groupBy, Items: items}
}
