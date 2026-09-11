package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/task/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *taskService) transaction(ctx context.Context, id uuid.UUID, fn func(*taskService) error) error {
	if s.db == nil {
		return fn(s)
	}
	pending := []func(){}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if id != uuid.Nil {
			if err := repo.LockTask(ctx, tx, id); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return response.NewError(response.CodeTaskNotFound, "任务不存在")
				}
				return err
			}
		}
		bound := *s
		bound.db = nil
		bound.afterCommit = &pending
		bound.tasks = repo.NewTaskRepo(tx)
		bound.logs = repo.NewLogRepo(tx)
		bound.comments = repo.NewCommentRepo(tx)
		bound.files = repo.NewAttachmentRepo(tx)
		bound.cleanup = repo.NewCleanupRepo(tx)
		bound.assignments = repo.NewAssignmentRepo(tx)
		bound.transfers = repo.NewTransferRepo(tx)
		if s.engine != nil {
			flow, deliver, err := s.engine.BindTransaction(tx)
			if err != nil {
				return err
			}
			bound.flow = flow
			pending = append(pending, func() { deliver(context.WithoutCancel(ctx)) })
		}
		return fn(&bound)
	})
	if err == nil {
		for _, effect := range pending {
			effect()
		}
	}
	return err
}
func taskMutation[T any](ctx context.Context, s *taskService, id uuid.UUID, fn func(*taskService) (T, error)) (T, error) {
	var out T
	err := s.transaction(ctx, id, func(bound *taskService) error { var err error; out, err = fn(bound); return err })
	return out, err
}
