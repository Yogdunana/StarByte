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
		if !feature.ShouldScheduleOn(&flag, now) && !feature.ShouldScheduleOff(&flag, now) {
			continue
		}
		if s.persistSchedule(ctx, flag.ID, systemActor) {
			changed++
		}
	}
	if changed > 0 {
		s.invalidate(ctx)
	}
	return changed
}

func (s *flagService) persistSchedule(ctx context.Context, id uuid.UUID, actor uuid.UUID) bool {
	latest, err := s.rows.GetByID(ctx, id)
	if err != nil || latest == nil {
		logCacheErr("schedule-get", err)
		return false
	}
	now := s.clock()
	var enabled bool
	var action string
	switch {
	case feature.ShouldScheduleOn(latest, now):
		enabled, action = true, model.ActionScheduleOn
	case feature.ShouldScheduleOff(latest, now):
		enabled, action = false, model.ActionScheduleOff
	default:
		return false
	}
	if latest.Enabled == enabled {
		return false
	}
	before := *latest
	latest.Enabled = enabled
	latest.UpdatedAt = now
	if err := s.rows.Update(ctx, latest); err != nil {
		logCacheErr("schedule-update", err)
		return false
	}
	s.audit(ctx, latest, actor, action, &before, latest, action)
	return true
}
