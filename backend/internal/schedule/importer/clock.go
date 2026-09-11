package importer

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var clockRe = regexp.MustCompile(`(\d{1,2}):(\d{2})\s*[-–~]\s*(\d{1,2}):(\d{2})`)

type clockRange struct {
	StartH, StartM int
	EndH, EndM     int
}

func parseClockRange(raw string) (clockRange, bool) {
	m := clockRe.FindStringSubmatch(raw)
	if m == nil {
		return clockRange{}, false
	}
	sh, _ := strconv.Atoi(m[1])
	sm, _ := strconv.Atoi(m[2])
	eh, _ := strconv.Atoi(m[3])
	em, _ := strconv.Atoi(m[4])
	if sh > 23 || eh > 23 || sm > 59 || em > 59 {
		return clockRange{}, false
	}
	return clockRange{StartH: sh, StartM: sm, EndH: eh, EndM: em}, true
}

func (c clockRange) onDate(day time.Time) (time.Time, time.Time) {
	start := time.Date(day.Year(), day.Month(), day.Day(), c.StartH, c.StartM, 0, 0, day.Location())
	end := time.Date(day.Year(), day.Month(), day.Day(), c.EndH, c.EndM, 0, 0, day.Location())
	if !end.After(start) {
		end = end.Add(24 * time.Hour)
	}
	return start, end
}

var roomRe = regexp.MustCompile(`【([^】]+)】`)

func parseRoom(raw string) string {
	m := roomRe.FindStringSubmatch(raw)
	if m == nil {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func stripDecorations(raw string) string {
	s := weekTokenRe.ReplaceAllString(raw, " ")
	s = clockRe.ReplaceAllString(s, " ")
	s = roomRe.ReplaceAllString(s, " ")
	s = strings.NewReplacer("，", " ", ",", " ", "；", " ", ";", " ").Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

func weekdayIndex(header string) (int, bool) {
	h := strings.TrimSpace(header)
	names := []string{"星期一", "星期二", "星期三", "星期四", "星期五", "星期六", "星期日"}
	alts := []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}
	for i, n := range names {
		if strings.Contains(h, n) || strings.Contains(h, alts[i]) {
			return i, true
		}
	}
	return 0, false
}

func dateOfWeek(semesterMonday time.Time, week, weekday int) time.Time {
	if week < 1 {
		week = 1
	}
	return semesterMonday.AddDate(0, 0, (week-1)*7+weekday)
}

func formatUID(parts ...string) string {
	return strings.Join(parts, ":")
}

func describeClock(c clockRange) string {
	return fmt.Sprintf("%02d:%02d-%02d:%02d", c.StartH, c.StartM, c.EndH, c.EndM)
}
