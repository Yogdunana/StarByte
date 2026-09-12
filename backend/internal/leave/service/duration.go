package service

import (
	"math"
	"time"
)

// CalculateDurationDays 按工作日计算请假时长（排除周六周日）。
// 算法原样保留自 SMB-Star PR #162，phase-1 不改审批/余额算术。
func CalculateDurationDays(startTime, endTime time.Time) float64 {
	if startTime.Truncate(24 * time.Hour).Equal(endTime.Truncate(24 * time.Hour)) {
		return 1
	}

	totalDays := 0.0
	current := startTime.Truncate(24 * time.Hour)
	end := endTime.Truncate(24 * time.Hour)

	for current.Before(end) || current.Equal(end) {
		weekday := current.Weekday()
		if weekday != time.Saturday && weekday != time.Sunday {
			totalDays++
		}
		current = current.Add(24 * time.Hour)
	}

	if endTime.Sub(current.Add(-24*time.Hour)) < 12*time.Hour && totalDays > 0 {
		totalDays = totalDays - 0.5
	}

	return math.Max(totalDays, 1)
}
