package service

import (
	"context"

	"github.com/google/uuid"

	rbac "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
)

func (s *taskService) filterParents(ctx context.Context, viewer uuid.UUID, scope *rbac.DataScopeCondition, rows []*dto.TaskResponse) error {
	ids := []uuid.UUID{}
	for _, row := range rows {
		if row.Parent != nil {
			ids = append(ids, uuid.MustParse(row.Parent.ID))
		}
	}
	parents, err := s.tasks.ListByIDs(ctx, ids)
	if err != nil {
		return err
	}
	allowed := map[string]bool{}
	for _, p := range parents {
		allowed[p.ID.String()] = canViewTask(&p, viewer, scope)
	}
	for _, row := range rows {
		if row.Parent != nil && !allowed[row.Parent.ID] {
			row.Parent = nil
		}
	}
	return nil
}
