package service

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/feature"
	"github.com/Yogdunana/StarByte/backend/internal/feature/model"
	"github.com/google/uuid"
)

func (s *flagService) startScheduleLoop(ctx context.Context) {
	tick := s.tickEvery
	if tick <= 0 {
		tick = 30 * time.Second
	}
	go func() {
		s.applySchedules(ctx)
		t := time.NewTicker(tick)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s.applySchedules(context.Background())
			}
		}
	}()
}

// applySchedules persists enabled=true/false when starts_at/ends_at cross now.
// Evaluation already applies the window without waiting for this ticker.
func (s *flagService) applySchedules(ctx context.Context) int {
	flags, err := s.rows.ListAll(ctx)
	if err != nil {
		logCacheErr("schedule-list", err)
		return 0
	}
	now := s.clock()
	changed := 0
	systemActor := uuid.Nil
	for i := range flags {
		flag := flags[i]
		switch {
		case feature.ShouldScheduleOn(&flag, now):
			if s.persistSchedule(ctx, &flag, true, model.ActionScheduleOn, systemActor) {
				changed++
			}
		case feature.ShouldScheduleOff(&flag, now):
			if s.persistSchedule(ctx, &flag, false, model.ActionScheduleOff, systemActor) {
				changed++
			}
		}
	}
	if changed > 0 {
		s.invalidate(ctx)
	}
	return changed
}

func (s *flagService) persistSchedule(ctx context.Context, flag *model.Flag, enabled bool, action string, actor uuid.UUID) bool {
	latest, err := s.rows.GetByID(ctx, flag.ID)
	if err != nil || latest == nil {
		logCacheErr("schedule-get", err)
		return false
	}
	if latest.Enabled == enabled {
		return false
	}
	before := *latest
	latest.Enabled = enabled
	latest.UpdatedAt = s.clock()
	if err := s.rows.Update(ctx, latest); err != nil {
		logCacheErr("schedule-update", err)
		return false
	}
	s.audit(ctx, latest, actor, action, &before, latest, action)
	return true
}
