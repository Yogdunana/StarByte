package repo

import (
	"context"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/task/model"
)

func (r *taskRepo) ListByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Task, error) {
	var rows []model.Task
	if len(ids) == 0 {
		return rows, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error
	return rows, err
}
