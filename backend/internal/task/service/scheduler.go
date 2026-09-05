package service

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ReminderScheduler struct {
	svc    TaskService
	stopCh chan struct{}
	done   chan struct{}
}

func NewReminderScheduler(svc TaskService) *ReminderScheduler {
	return &ReminderScheduler{svc: svc, stopCh: make(chan struct{}), done: make(chan struct{})}
}

func (s *ReminderScheduler) Start() {
	go s.run()
	logger.Info("task due reminder scheduler started")
}

func (s *ReminderScheduler) Stop() {
	close(s.stopCh)
	<-s.done
	logger.Info("task due reminder scheduler stopped")
}

func (s *ReminderScheduler) run() {
	defer close(s.done)
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	s.tick()
	for {
		select {
		case <-ticker.C:
			s.tick()
		case <-s.stopCh:
			return
		}
	}
}

func (s *ReminderScheduler) tick() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	n, err := s.svc.RemindDueAndOverdue(ctx)
	if err != nil {
		logger.Warn("task reminder job failed", zap.Error(err))
		return
	}
	if n > 0 {
		logger.Info("task reminder job sent", zap.Int("count", n))
	}
}

func (s *taskService) RemindDueAndOverdue(ctx context.Context) (int, error) {
	now := time.Now()
	sent := 0
	soon, err := s.tasks.ListDueSoon(ctx, now, now.Add(24*time.Hour))
	if err != nil {
		return 0, err
	}
	for i := range soon {
		t := &soon[i]
		s.notifyUsers(ctx, reminderTargets(t), tplTaskDueSoon, t, "")
		mark := now
		t.DueRemindedAt = &mark
		t.UpdatedAt = now
		if err := s.tasks.Update(ctx, t); err != nil {
			logger.Warn("mark due reminder failed", zap.Error(err), zap.String("id", t.ID.String()))
			continue
		}
		sent++
	}
	overdue, err := s.tasks.ListOverdue(ctx, now)
	if err != nil {
		return sent, err
	}
	for i := range overdue {
		t := &overdue[i]
		s.notifyUsers(ctx, reminderTargets(t), tplTaskOverdue, t, "")
		mark := now
		t.OverdueRemindedAt = &mark
		t.UpdatedAt = now
		if err := s.tasks.Update(ctx, t); err != nil {
			logger.Warn("mark overdue reminder failed", zap.Error(err), zap.String("id", t.ID.String()))
			continue
		}
		sent++
	}
	return sent, nil
}

func reminderTargets(t *model.Task) []uuid.UUID {
	seen := map[uuid.UUID]struct{}{}
	out := make([]uuid.UUID, 0, 2)
	add := func(id uuid.UUID) {
		if id == uuid.Nil {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	add(t.CreatorID)
	if t.AssigneeID != nil {
		add(*t.AssigneeID)
	}
	return out
}
