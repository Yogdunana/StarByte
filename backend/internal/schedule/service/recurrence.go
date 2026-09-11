package service

import (
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/schedule/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// Recurrence is limited to none/daily/weekly/monthly (no RFC 5545 RRULE / INTERVAL / BYDAY).
// Range queries expand occurrences inside the window, capped at maxOccurrences.
const maxOccurrences = 200

func validateRecurrence(rule string, until *time.Time, start time.Time) (string, error) {
	norm := model.NormalizeRecurrence(rule)
	if !model.ValidRecurrence(norm) {
		return "", response.NewError(response.CodeScheduleRecurrenceLimited, "仅支持 none/daily/weekly/monthly，自定义 RRULE 未实现")
	}
	if norm != model.RecurrenceNone && until != nil && until.Before(start) {
		return "", response.NewError(response.CodeScheduleInvalidTime, "重复截止不能早于开始时间")
	}
	return norm, nil
}

type occurrence struct {
	Start time.Time
	End   time.Time
}

func expandOccurrences(ev *model.Event, windowStart, windowEnd time.Time) []occurrence {
	dur := ev.EndAt.Sub(ev.StartAt)
	if dur < 0 {
		dur = 0
	}
	rule := model.NormalizeRecurrence(ev.Recurrence)
	if rule == model.RecurrenceNone {
		if ev.EndAt.Before(windowStart) || ev.StartAt.After(windowEnd) {
			return nil
		}
		return []occurrence{{Start: ev.StartAt, End: ev.EndAt}}
	}
	limit := ev.RecurrenceUntil
	out := make([]occurrence, 0, 8)
	cur := ev.StartAt
	for i := 0; i < maxOccurrences; i++ {
		if limit != nil && cur.After(*limit) {
			break
		}
		if cur.After(windowEnd) {
			break
		}
		end := cur.Add(dur)
		if !end.Before(windowStart) {
			out = append(out, occurrence{Start: cur, End: end})
		}
		next := nextOccurrence(cur, rule)
		if !next.After(cur) {
			break
		}
		cur = next
	}
	return out
}

func nextOccurrence(t time.Time, rule string) time.Time {
	switch rule {
	case model.RecurrenceDaily:
		return t.AddDate(0, 0, 1)
	case model.RecurrenceWeekly:
		return t.AddDate(0, 0, 7)
	case model.RecurrenceMonthly:
		return t.AddDate(0, 1, 0)
	default:
		return t
	}
}
