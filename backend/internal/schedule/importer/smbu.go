package importer

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// TimetableImporter 解析深北莫（SMBU）学期课表 XLSX。
type TimetableImporter struct{}

func (TimetableImporter) Kind() string { return "timetable" }

func (TimetableImporter) Parse(_ context.Context, raw []byte, opts Options) ([]DraftEvent, Meta, error) {
	if len(raw) == 0 {
		return nil, Meta{}, fmt.Errorf("empty xlsx")
	}
	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return nil, Meta{}, fmt.Errorf("open xlsx: %w", err)
	}
	defer func() { _ = f.Close() }()

	sheet := pickSheet(f)
	if sheet == "" {
		return nil, Meta{}, fmt.Errorf("no sheet")
	}
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, Meta{}, fmt.Errorf("read sheet: %w", err)
	}
	if len(rows) < 2 {
		return nil, Meta{}, fmt.Errorf("sheet too short")
	}

	meta := parseTitleRow(firstNonEmpty(rows[0]))
	weekdays := parseWeekdayHeaders(rows[1])
	if len(weekdays) == 0 {
		return nil, Meta{}, fmt.Errorf("missing weekday headers")
	}

	loc := locationOrShanghai(opts.Timezone)
	start := mondayOf(opts.SemesterStart.In(loc))
	var events []DraftEvent
	for r := 2; r < len(rows); r++ {
		row := rows[r]
		periodLabel := cellAt(row, 0)
		if strings.Contains(periodLabel, "未排具体节次") {
			break
		}
		if strings.TrimSpace(periodLabel) == "" && emptyRow(row) {
			continue
		}
		periodClock, hasPeriod := parseClockRange(periodLabel)
		for col, weekday := range weekdays {
			text := cellAt(row, col)
			if strings.TrimSpace(text) == "" {
				continue
			}
			if strings.Contains(text, "未排具体节次") {
				continue
			}
			sessions := ParseCellSessions(text)
			if len(sessions) == 0 {
				continue
			}
			for _, sess := range sessions {
				clock := periodClock
				hasClock := hasPeriod
				if sess.HasClock {
					clock = sess.Clock
					hasClock = true
				}
				if !hasClock || len(sess.Weeks) == 0 || strings.TrimSpace(sess.Title) == "" {
					continue
				}
				for _, week := range sess.Weeks {
					day := dateOfWeek(start, week, weekday)
					begin, end := clock.onDate(day)
					desc := sess.Teacher
					if weekHint := strings.TrimSpace(sess.Raw); weekHint != "" && !strings.Contains(desc, weekHint) {
						if desc != "" {
							desc = desc + "\n" + weekHint
						} else {
							desc = weekHint
						}
					}
					events = append(events, DraftEvent{
						Title:       sess.Title,
						Description: desc,
						Location:    sess.Room,
						StartAt:     begin,
						EndAt:       end,
						ExternalUID: formatUID("smbu", meta.StudentNo, sess.Title, fmt.Sprintf("w%d-d%d", week, weekday), describeClock(clock)),
					})
				}
			}
		}
	}
	meta.Timezone = loc.String()
	return events, meta, nil
}

func pickSheet(f *excelize.File) string {
	for _, name := range f.GetSheetList() {
		if strings.TrimSpace(name) != "" {
			return name
		}
	}
	return ""
}

func parseWeekdayHeaders(row []string) map[int]int {
	out := map[int]int{}
	for i, cell := range row {
		if i == 0 {
			continue
		}
		if wd, ok := weekdayIndex(cell); ok {
			out[i] = wd
		}
	}
	return out
}

func cellAt(row []string, i int) string {
	if i < 0 || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func firstNonEmpty(row []string) string {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return strings.TrimSpace(c)
		}
	}
	return ""
}

func emptyRow(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}
