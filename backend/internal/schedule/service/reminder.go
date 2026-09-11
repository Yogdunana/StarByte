package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/schedule/model"
	"github.com/google/uuid"
)

func (s *scheduleService) DispatchDueReminders(ctx context.Context, _ string, logf func(string)) error {
	if logf == nil {
		logf = func(string) {}
	}
	now := time.Now()
	rows, err := s.rows.ListDueReminders(ctx, now, 200)
	if err != nil {
		return fmt.Errorf("list due reminders: %w", err)
	}
	sent := 0
	for i := range rows {
		row := &rows[i]
		occ := dueOccurrence(row, now)
		if occ == nil {
			continue
		}
		targets := uniqueUsers(row.OwnerID, row.CreatedBy)
		atts, err := s.rows.ListAttendees(ctx, row.EventID)
		if err != nil {
			return fmt.Errorf("list reminder attendees: %w", err)
		}
		for _, a := range atts {
			targets = append(targets, a.UserID)
		}
		targets = uniqueUsers(targets...)
		if s.notify != nil {
			if err := s.notify.Send(ctx, targets, "schedule_reminder", map[string]interface{}{
				"title":    row.Title,
				"start_at": occ.Format(time.RFC3339),
				"minutes":  fmt.Sprintf("%d", row.MinutesBefore),
			}); err != nil {
				logf("notify failed: " + err.Error())
				continue
			}
		}
		if err := s.rows.MarkReminderTriggered(ctx, row.ID, *occ); err != nil {
			return fmt.Errorf("mark reminder: %w", err)
		}
		sent++
	}
	logf(fmt.Sprintf("dispatched %d reminders", sent))
	return nil
}

func dueOccurrence(row *model.DueReminder, now time.Time) *time.Time {
	ev := &model.Event{
		StartAt: row.StartAt, EndAt: row.EndAt, Recurrence: row.Recurrence, RecurrenceUntil: row.RecurrenceUntil,
	}
	if ev.EndAt.IsZero() {
		ev.EndAt = ev.StartAt.Add(time.Hour)
	}
	windowStart := now.Add(-14 * 24 * time.Hour)
	windowEnd := now.Add(time.Duration(row.MinutesBefore) * time.Minute)
	for _, occ := range expandOccurrences(ev, windowStart, windowEnd) {
		fireAt := occ.Start.Add(-time.Duration(row.MinutesBefore) * time.Minute)
		if fireAt.After(now) {
			continue
		}
		if row.TriggeredAt != nil && !occ.Start.After(*row.TriggeredAt) {
			continue
		}
		start := occ.Start
		return &start
	}
	return nil
}

// reminderIsDueCandidate 与 repo.ListDueReminders 的窗口条件对齐：只保留现在该发的一条。
func reminderIsDueCandidate(ev *model.Event, rem *model.Reminder, now time.Time) bool {
	if ev == nil || rem == nil {
		return false
	}
	mins := time.Duration(rem.MinutesBefore) * time.Minute
	rec := model.NormalizeRecurrence(ev.Recurrence)
	if rem.TriggeredAt == nil {
		if ev.StartAt.After(now.Add(mins)) {
			return false
		}
		return rec == model.RecurrenceNone || ev.RecurrenceUntil == nil || !ev.RecurrenceUntil.Before(ev.StartAt)
	}
	var next time.Time
	switch rec {
	case model.RecurrenceDaily:
		next = rem.TriggeredAt.Add(24 * time.Hour)
	case model.RecurrenceWeekly:
		next = rem.TriggeredAt.Add(7 * 24 * time.Hour)
	case model.RecurrenceMonthly:
		next = rem.TriggeredAt.AddDate(0, 1, 0)
	default:
		return false
	}
	if next.After(now.Add(mins)) {
		return false
	}
	return ev.RecurrenceUntil == nil || !ev.RecurrenceUntil.Before(next)
}

func uniqueUsers(ids ...uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]struct{}{}
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
