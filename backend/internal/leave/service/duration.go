package service

import (
	"math"
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

// CalculateDurationDays 按工作日计算请假时长（排除周六周日）。
// 规则与 SMB-Star PR #162 相同；日历日按 Asia/Shanghai，避免 UTC Truncate 把东八区当天判成昨天。
func CalculateDurationDays(startTime, endTime time.Time) float64 {
	startDay := calendarDate(startTime)
	endDay := calendarDate(endTime)
	if startDay.Equal(endDay) {
		return 1
	}

	totalDays := 0.0
	current := startDay
	for !current.After(endDay) {
		weekday := current.Weekday()
		if weekday != time.Saturday && weekday != time.Sunday {
			totalDays++
		}
		current = current.AddDate(0, 0, 1)
	}

	endLocal := endTime.In(bizLocation())
	if endLocal.Sub(endDay) < 12*time.Hour && totalDays > 0 {
		totalDays = totalDays - 0.5
	}

	return math.Max(totalDays, 1)
}
