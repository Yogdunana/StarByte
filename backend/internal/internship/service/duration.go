package service

import "time"

// CalculateDuration 按 Issue 计算实习时长（天）。end 为零值时取当前时间。
func CalculateDuration(start, end time.Time) int {
	if end.IsZero() {
		end = time.Now()
	}
	if !end.After(start) {
		return 0
	}
	return int(end.Sub(start).Hours() / 24)
}

// CalculateMonthlyDuration 按月统计重叠天数，与 CalculateDuration 同一套截断规则。
func CalculateMonthlyDuration(start, end time.Time) map[string]int {
	if end.IsZero() {
		end = time.Now()
	}
	out := map[string]int{}
	total := CalculateDuration(start, end)
	if total <= 0 {
		return out
	}
	cur := dateOnly(start)
	last := dateOnly(end)
	for cur.Before(last) {
		out[cur.Format("2006-01")]++
		cur = cur.AddDate(0, 0, 1)
	}
	return out
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func durationOf(start time.Time, end *time.Time) int {
	finish := time.Now()
	if end != nil && !end.IsZero() {
		finish = *end
	}
	return CalculateDuration(start, finish)
}
