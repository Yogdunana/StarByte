package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandDailyInWindow(t *testing.T) {
	start := time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
	ev := &model.Event{
		StartAt: start, EndAt: start.Add(time.Hour), Recurrence: model.RecurrenceDaily,
	}
	until := time.Date(2026, 9, 3, 9, 0, 0, 0, time.UTC)
	ev.RecurrenceUntil = &until
	out := expandOccurrences(ev, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC))
	require.Len(t, out, 3)
	assert.Equal(t, start, out[0].Start)
	assert.Equal(t, start.AddDate(0, 0, 2), out[2].Start)
}

func TestExpandWeeklySkipsOutsideWindow(t *testing.T) {
	start := time.Date(2026, 9, 1, 14, 0, 0, 0, time.UTC)
	ev := &model.Event{StartAt: start, EndAt: start.Add(2 * time.Hour), Recurrence: model.RecurrenceWeekly}
	out := expandOccurrences(ev, time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC))
	require.Len(t, out, 1)
	assert.Equal(t, start.AddDate(0, 0, 14), out[0].Start)
}

func TestExpandNoneOutsideWindow(t *testing.T) {
	start := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	ev := &model.Event{StartAt: start, EndAt: start.Add(time.Hour), Recurrence: model.RecurrenceNone}
	out := expandOccurrences(ev, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	assert.Empty(t, out)
}

func TestValidateRecurrenceRejectsRRULE(t *testing.T) {
	_, err := validateRecurrence("FREQ=WEEKLY;BYDAY=MO", nil, time.Now())
	require.Error(t, err)
}

func TestRangeEventsExpandsWeekly(t *testing.T) {
	svc, _, owner, _, _ := setupSvc(t)
	ctx := context.Background()
	scope := selfScope(owner)
	start := time.Date(2026, 9, 7, 10, 0, 0, 0, time.UTC) // Monday
	_, err := svc.CreateEvent(ctx, owner, dtoCreate(start, model.RecurrenceWeekly), scope)
	require.NoError(t, err)
	list, err := svc.RangeEvents(ctx, owner, &dto.RangeEventRequest{
		Start: time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC),
	}, scope)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 3)
}

func dtoCreate(start time.Time, rule string) *dto.CreateEventRequest {
	return &dto.CreateEventRequest{
		Title: "周会", StartAt: start, EndAt: start.Add(time.Hour), Recurrence: rule,
	}
}
