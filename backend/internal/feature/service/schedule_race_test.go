package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/feature/dto"
	"github.com/Yogdunana/StarByte/backend/internal/feature/model"
	"github.com/google/uuid"
)

type afterGetScheduleRepo struct {
	*memRepo
	afterGet func()
}

func (r *afterGetScheduleRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Flag, error) {
	row, err := r.memRepo.GetByID(ctx, id)
	if fn := r.afterGet; fn != nil {
		r.afterGet = nil
		fn()
	}
	return row, err
}

func TestSchedulePreservesEditAfterRead(t *testing.T) {
	for _, initiallyEnabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "schedule_on", true: "schedule_off"}[initiallyEnabled], func(t *testing.T) {
			ctx := context.Background()
			rows := &afterGetScheduleRepo{memRepo: newMemRepo()}
			svc := New(rows, NewMemorySnapshot(), NewMemoryBus(), stubRoles{}, stubUsers{}, stubPerms{}).(*flagService)
			now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
			boundary := now.Add(time.Hour)
			svc.now = func() time.Time { return now }
			rules := model.Rules{StartsAt: &boundary}
			if initiallyEnabled {
				rules = model.Rules{EndsAt: &boundary}
			}
			created, err := svc.Create(ctx, uuid.New(), &dto.CreateFlagRequest{
				FlagKey: "race.after_read", Name: "Race", FlagType: model.TypeBoolean,
				Enabled: initiallyEnabled, Rules: rules,
			})
			if err != nil {
				t.Fatal(err)
			}
			id := uuid.MustParse(created.ID)
			now = boundary.Add(time.Second)
			rows.afterGet = func() {
				_, err := svc.Update(ctx, uuid.New(), id, &dto.UpdateFlagRequest{Enabled: &initiallyEnabled})
				if err != nil {
					t.Fatal(err)
				}
			}
			if n := svc.applySchedules(ctx); n != 0 {
				t.Fatalf("stale scheduler reported %d updates", n)
			}
			got, err := rows.GetByID(ctx, id)
			if err != nil || got.Enabled != initiallyEnabled || !got.UpdatedAt.Equal(now) {
				t.Fatalf("manual choice lost: %+v, %v", got, err)
			}
			for _, audit := range rows.audits {
				if audit.Action == model.ActionScheduleOn || audit.Action == model.ActionScheduleOff {
					t.Fatal("skipped scheduler wrote an audit")
				}
			}
		})
	}
}
