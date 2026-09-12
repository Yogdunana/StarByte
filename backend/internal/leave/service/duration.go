package service

import (
	"sync"
	"time"
)

const bizTimezone = "Asia/Shanghai"

var (
	bizLocOnce sync.Once
	bizLoc     *time.Location
)

func bizLocation() *time.Location {
	bizLocOnce.Do(func() {
		loc, err := time.LoadLocation(bizTimezone)
		if err != nil {
			loc = time.FixedZone("CST", 8*3600)
		}
		bizLoc = loc
	})
	return bizLoc
}

func calendarDate(t time.Time) time.Time {
	t = t.In(bizLocation())
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, bizLocation())
}

func bizYear(t time.Time) int {
	return t.In(bizLocation()).Year()
}

func isWeekend(t time.Time) bool {
	w := t.Weekday()
	return w == time.Saturday || w == time.Sunday
}

// CalculateDurationDays 按工作日计算请假时长（排除周六周日）。
// 同一工作日计 1 天；结束日若是工作日且未满 12:00 则减 0.5。
// 周末结束日不参与半天扣减；纯周末区间为 0，由提交校验拒绝。
func CalculateDurationDays(startTime, endTime time.Time) float64 {
	startDay := calendarDate(startTime)
	endDay := calendarDate(endTime)
	if endDay.Before(startDay) {
		return 0
	}
	if startDay.Equal(endDay) {
		if isWeekend(startDay) {
			return 0
		}
		return 1
	}

	totalDays := 0.0
	current := startDay
	for !current.After(endDay) {
		if !isWeekend(current) {
			totalDays++
		}
		current = current.AddDate(0, 0, 1)
	}

	endLocal := endTime.In(bizLocation())
	if !isWeekend(endDay) && endLocal.Sub(endDay) < 12*time.Hour && totalDays > 0 {
		totalDays -= 0.5
	}
	return totalDays
}
