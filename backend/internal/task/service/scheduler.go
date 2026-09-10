package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
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
	cleanupTicker := time.NewTicker(30 * time.Second)
	defer cleanupTicker.Stop()
	s.cleanupFiles()
	s.tick()
	for {
		select {
		case <-ticker.C:
			s.tick()
		case <-cleanupTicker.C:
			s.cleanupFiles()
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
	for _, task := range soon {
		n, err := s.remindTask(ctx, task.ID, now, false)
		sent += n
		if err != nil {
			return sent, err
		}
	}
	overdue, err := s.tasks.ListOverdue(ctx, now)
	if err != nil {
		return sent, err
	}
	for _, task := range overdue {
		n, err := s.remindTask(ctx, task.ID, now, true)
		sent += n
		if err != nil {
			return sent, err
		}
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

func (s *ReminderScheduler) cleanupFiles() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if _, err := s.svc.ProcessAttachmentDeletions(ctx); err != nil {
		logger.Warn("task attachment cleanup pending retry", zap.Error(err))
	}
}
