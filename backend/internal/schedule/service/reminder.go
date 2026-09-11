package service

import (
	"context"
	"fmt"
	"time"

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
				"start_at": row.StartAt.Format(time.RFC3339),
				"minutes":  fmt.Sprintf("%d", row.MinutesBefore),
			}); err != nil {
				logf("notify failed: " + err.Error())
				continue
			}
		}
		if err := s.rows.MarkReminderTriggered(ctx, row.ID, now); err != nil {
			return fmt.Errorf("mark reminder: %w", err)
		}
		sent++
	}
	logf(fmt.Sprintf("dispatched %d reminders", sent))
	return nil
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
