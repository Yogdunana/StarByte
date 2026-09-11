package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// The scheduler must re-read after locking; its original candidate may already
// have been completed, reassigned or given a new deadline by a user.
func (s *taskService) remindTask(ctx context.Context, id uuid.UUID, now time.Time, overdue bool) (int, error) {
	sent := 0
	err := s.transaction(ctx, id, func(b *taskService) error {
		t, err := b.tasks.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if t == nil || model.IsClosed(t.Status) || t.DueDate == nil {
			return nil
		}
		template := tplTaskDueSoon
		if overdue {
			if !t.DueDate.Before(now) || t.OverdueRemindedAt != nil {
				return nil
			}
			template = tplTaskOverdue
		} else {
			if !t.DueDate.After(now) || t.DueDate.After(now.Add(24*time.Hour)) || t.DueRemindedAt != nil {
				return nil
			}
		}
		// Failed delivery leaves the marker unset for a later retry. Delivery is
		// at-least-once: a crash between sending and commit may repeat a reminder.
		if b.notify == nil {
			return nil
		}
		for _, uid := range reminderTargets(t) {
			name := ""
			u, err := b.tasks.GetUser(ctx, uid)
			if err != nil {
				return err
			}
			if u != nil {
				name = displayName(u)
			}
			if err := b.notify.Send(ctx, []uuid.UUID{uid}, template, map[string]interface{}{"title": t.Title, "real_name": name, "message": "", "due_date": t.DueDate.Format("2006-01-02 15:04")}); err != nil {
				return err
			}
		}
		if overdue {
			t.OverdueRemindedAt = &now
		} else {
			t.DueRemindedAt = &now
		}
		if err := b.tasks.Update(ctx, t); err != nil {
			return err
		}
		sent = 1
		return nil
	})
	var app *response.AppError
	if errors.As(err, &app) && app.Code == response.CodeTaskNotFound {
		return 0, nil
	}
	return sent, err
}
