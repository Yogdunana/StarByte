package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/task/model"
)

func (m *memTasks) ListByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Task, error) {
	rows := []model.Task{}
	for _, id := range ids {
		row, err := m.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if row != nil {
			rows = append(rows, *row)
		}
	}
	return rows, nil
}

func (m *memTasks) SearchUsers(ctx context.Context, keyword string) ([]model.NamedUser, error) {
	out := []model.NamedUser{}
	for _, u := range m.users {
		out = append(out, *u)
	}
	return out, nil
}
