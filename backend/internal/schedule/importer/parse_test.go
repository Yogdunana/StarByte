package importer

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestParseWeeks(t *testing.T) {
	cases := []struct {
		in   string
		want []int
	}{
		{"1-18周 王兰 08:00-10:30 【3-227】", seq(1, 18, "")},
		{"1-17周(单) 李雷", []int{1, 3, 5, 7, 9, 11, 13, 15, 17}},
		{"2-18周(双) 韩梅", []int{2, 4, 6, 8, 10, 12, 14, 16, 18}},
		{"1-7周,9-18周 赵强", append(seq(1, 7, ""), seq(9, 18, "")...)},
		{"1-2周,4-18周 钱伟", append(seq(1, 2, ""), seq(4, 18, "")...)},
		{"10周 周敏", []int{10}},
		{"8周 吴芳", []int{8}},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, ParseWeeks(tc.in), tc.in)
	}
}

func TestParseCellSessionsFromFixtureFile(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(file), "testdata", "cells.txt"))
	require.NoError(t, err)
	text := string(raw)
	require.Contains(t, text, "马克思主义基本原理 01")
	require.Contains(t, text, "1-17周(单)")
	require.Contains(t, text, "2-18周(双)")
	require.Contains(t, text, "1-7周,9-18周")
	require.Contains(t, text, "1-2周,4-18周")
	marx := ParseCellSessions("马克思主义基本原理 01\n1-18周 王兰 08:00-10:30 【3-227】")
	require.Len(t, marx, 1)
	assert.Equal(t, "3-227", marx[0].Room)
}

func TestParseCellSessions(t *testing.T) {
	marx := ParseCellSessions("马克思主义基本原理 01\n1-18周 王兰 08:00-10:30 【3-227】")
	require.Len(t, marx, 1)
	assert.Equal(t, "马克思主义基本原理 01", marx[0].Title)
	assert.Equal(t, "王兰", marx[0].Teacher)
	assert.Equal(t, "3-227", marx[0].Room)
	assert.True(t, marx[0].HasClock)
	assert.Equal(t, 8, marx[0].Clock.StartH)
	assert.Equal(t, 10, marx[0].Clock.EndH)
	assert.Equal(t, 18, len(marx[0].Weeks))

	odd := ParseCellSessions("线性代数 02\n1-17周(单) 李雷 08:00-09:35 【1-101】")
	require.Len(t, odd, 1)
	assert.Equal(t, []int{1, 3, 5, 7, 9, 11, 13, 15, 17}, odd[0].Weeks)

	multi := ParseCellSessions("高等数学A 01\n1-7周 王兰 08:00-10:30 【3-227】\n高等数学A 02\n9-18周 李雷 08:00-10:30 【3-228】")
	require.Len(t, multi, 2)
	assert.Equal(t, "高等数学A 01", multi[0].Title)
	assert.Equal(t, "高等数学A 02", multi[1].Title)
	assert.Equal(t, "3-228", multi[1].Room)
	assert.Equal(t, seq(9, 18, ""), multi[1].Weeks)
}

func TestParseTitleRow(t *testing.T) {
	meta := parseTitleRow("2026-2027学年 第一学期 段茗尧[1120240002] 课表")
	assert.Equal(t, "2026-2027", meta.AcademicYear)
	assert.Equal(t, "第一学期", meta.Term)
	assert.Equal(t, "段茗尧", meta.StudentName)
	assert.Equal(t, "1120240002", meta.StudentNo)
	assert.Contains(t, meta.SourceKey, "1120240002")
	assert.Contains(t, meta.CalendarName, "段茗尧")
}

func TestPeriodClock(t *testing.T) {
	c, ok := parseClockRange("第一单节08:00-08:45")
	require.True(t, ok)
	assert.Equal(t, clockRange{8, 0, 8, 45}, c)
	c, ok = parseClockRange("第四双节第1小节14:00-14:45")
	require.True(t, ok)
	assert.Equal(t, 14, c.StartH)
	assert.Equal(t, 45, c.EndM)
}

func TestTimetableImporterGrid(t *testing.T) {
	raw := buildSampleXLSX(t)
	loc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	start := time.Date(2026, 9, 7, 0, 0, 0, 0, loc) // Monday of week 1
	events, meta, err := TimetableImporter{}.Parse(context.Background(), raw, Options{SemesterStart: start, Timezone: loc})
	require.NoError(t, err)
	assert.Equal(t, "1120240002", meta.StudentNo)
	assert.NotEmpty(t, events)

	var marx []DraftEvent
	for _, ev := range events {
		if strings.HasPrefix(ev.Title, "马克思主义基本原理") {
			marx = append(marx, ev)
		}
	}
	require.Len(t, marx, 18)
	assert.Equal(t, "3-227", marx[0].Location)
	assert.Equal(t, time.Date(2026, 9, 7, 8, 0, 0, 0, loc), marx[0].StartAt)
	assert.Equal(t, time.Date(2026, 9, 7, 10, 30, 0, 0, loc), marx[0].EndAt)

	var odd []DraftEvent
	for _, ev := range events {
		if strings.HasPrefix(ev.Title, "线性代数") {
			odd = append(odd, ev)
		}
	}
	require.Len(t, odd, 9)
	assert.Equal(t, time.Tuesday, odd[0].StartAt.Weekday())

	var single []DraftEvent
	for _, ev := range events {
		if strings.HasPrefix(ev.Title, "形势与政策") {
			single = append(single, ev)
		}
	}
	require.Len(t, single, 1)
	assert.Equal(t, 10, isoWeekOffset(start, single[0].StartAt))
}

func TestICSImporter(t *testing.T) {
	raw := []byte(`BEGIN:VCALENDAR
VERSION:2.0
X-WR-CALNAME:社团活动
BEGIN:VEVENT
UID:evt-1
SUMMARY:例会
DTSTART:20260911T100000Z
DTEND:20260911T110000Z
LOCATION:A101
DESCRIPTION:周会
END:VEVENT
END:VCALENDAR
`)
	events, meta, err := ICSImporter{}.Parse(context.Background(), raw, Options{Filename: "club.ics"})
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "社团活动", meta.CalendarName)
	assert.Equal(t, "例会", events[0].Title)
	assert.Equal(t, "A101", events[0].Location)
	assert.Equal(t, "ics:evt-1", events[0].ExternalUID)
}

func seq(from, to int, _ string) []int {
	out := make([]int, 0, to-from+1)
	for i := from; i <= to; i++ {
		out = append(out, i)
	}
	return out
}

func isoWeekOffset(monday, day time.Time) int {
	return int(day.Sub(monday).Hours()/24)/7 + 1
}

func buildSampleXLSX(t *testing.T) []byte {
	t.Helper()
	f := excelize.NewFile()
	sheet := "段茗尧_1120240002"
	require.NoError(t, f.SetSheetName("Sheet1", sheet))
	require.NoError(t, f.SetCellValue(sheet, "A1", "2026-2027学年 第一学期 段茗尧[1120240002] 课表"))
	headers := []string{"节次/星期", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六", "星期日"}
	for i, h := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, 2)
		require.NoError(t, err)
		require.NoError(t, f.SetCellValue(sheet, cell, h))
	}
	require.NoError(t, f.SetCellValue(sheet, "A3", "第一单节08:00-08:45"))
	require.NoError(t, f.SetCellValue(sheet, "B3", "马克思主义基本原理 01\n1-18周 王兰 08:00-10:30 【3-227】"))
	require.NoError(t, f.SetCellValue(sheet, "C3", "线性代数 02\n1-17周(单) 李雷 08:00-09:35 【1-101】"))
	require.NoError(t, f.SetCellValue(sheet, "A4", "第四双节第1小节14:00-14:45"))
	require.NoError(t, f.SetCellValue(sheet, "D4", "程序设计 04\n1-7周,9-18周 赵强 14:00-15:35 【实验楼A-301】"))
	require.NoError(t, f.SetCellValue(sheet, "E4", "形势与政策 06\n10周 周敏 19:00-20:35 【报告厅】"))
	require.NoError(t, f.SetCellValue(sheet, "A6", "未排具体节次课程"))
	require.NoError(t, f.SetCellValue(sheet, "B6", "应忽略"))
	buf, err := f.WriteToBuffer()
	require.NoError(t, err)
	return buf.Bytes()
}
