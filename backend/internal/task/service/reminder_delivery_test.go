package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/task/model"
)

type failedReminder struct{}

func (failedReminder) Send(context.Context, []uuid.UUID, string, map[string]interface{}) error {
	return errors.New("delivery failed")
}
func TestReminderRechecksStateAndRetriesFailure(t *testing.T) {
	svc, tasks, _, notify := newTestSvc()
	ctx := context.Background()
	now := time.Now()
	due := now.Add(time.Hour)
	task := &model.Task{ID: uuid.New(), CreatorID: uuid.New(), Title: "Reminder", Status: model.StatusDone, DueDate: &due}
	_ = tasks.Create(ctx, task)
	n, err := svc.remindTask(ctx, task.ID, now, false)
	if err != nil || n != 0 || notify.calls != 0 {
		t.Fatal("completed task received stale reminder")
	}
	task.Status = model.StatusDoing
	_ = tasks.Update(ctx, task)
	svc.notify = failedReminder{}
	n, err = svc.remindTask(ctx, task.ID, now, false)
	if err == nil || n != 0 {
		t.Fatal("failed delivery acknowledged")
	}
	stored, _ := tasks.GetByID(ctx, task.ID)
	if stored.DueRemindedAt != nil {
		t.Fatal("failed reminder marked delivered")
	}
	svc.notify = notify
	n, err = svc.remindTask(ctx, task.ID, now, false)
	if err != nil || n != 1 {
		t.Fatalf("retry: %d %v", n, err)
	}
	n, err = svc.remindTask(ctx, task.ID, now, false)
	if err != nil || n != 0 {
		t.Fatal("duplicate reminder")
	}
}
