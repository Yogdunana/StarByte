package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/importer"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestImportTimetableCreatesDedicatedLayerAndReplaces(t *testing.T) {
	svc, mem, owner, _, _ := setupSvc(t)
	ctx := context.Background()
	scope := selfScope(owner)
	loc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	start := time.Date(2026, 9, 7, 0, 0, 0, 0, loc)

	raw := mustSampleXLSX(t)
	first, err := svc.ImportTimetable(ctx, owner, "课表.xlsx", raw, start, scope)
	require.NoError(t, err)
	assert.Equal(t, model.SourceTimetable, first.Source)
	assert.Greater(t, first.EventCount, 0)
	assert.False(t, first.Replaced)

	cals, _, _, _, err := svc.ListCalendars(ctx, owner, &dto.ListCalendarRequest{PageSize: 50}, scope)
	require.NoError(t, err)
	var personal, table int
	for _, c := range cals {
		switch c.Source {
		case model.SourcePersonal:
			personal++
		case model.SourceTimetable:
			table++
			assert.NotEqual(t, "", c.Color)
		}
	}
	assert.Equal(t, 1, personal)
	assert.Equal(t, 1, table)

	second, err := svc.ImportTimetable(ctx, owner, "课表.xlsx", raw, start, scope)
	require.NoError(t, err)
	assert.True(t, second.Replaced)
	assert.Equal(t, first.CalendarID, second.CalendarID)
	assert.Equal(t, first.EventCount, second.EventCount)

	generated := 0
	for _, ev := range mem.events {
		if ev.CalendarID.String() == first.CalendarID && ev.Origin == model.OriginGenerated {
			generated++
		}
	}
	assert.Equal(t, first.EventCount, generated)
}

func TestImportTimetableRequiresSemesterStart(t *testing.T) {
	svc, _, owner, _, _ := setupSvc(t)
	_, err := svc.ImportTimetable(context.Background(), owner, "a.xlsx", []byte("xx"), time.Time{}, selfScope(owner))
	require.Error(t, err)
	assert.Equal(t, response.CodeScheduleImportInvalid, err.(*response.AppError).Code)
}

func TestImportICSNewLayer(t *testing.T) {
	svc, _, owner, _, _ := setupSvc(t)
	raw := []byte("BEGIN:VCALENDAR\nBEGIN:VEVENT\nUID:1\nSUMMARY:社团\nDTSTART:20260911T020000Z\nDTEND:20260911T030000Z\nEND:VEVENT\nEND:VCALENDAR\n")
	out, err := svc.ImportICS(context.Background(), owner, "club.ics", raw, "", selfScope(owner))
	require.NoError(t, err)
	assert.Equal(t, model.SourceImport, out.Source)
	assert.Equal(t, 1, out.EventCount)
}

func TestGoogleStatusUnconfigured(t *testing.T) {
	svc, _, owner, _, _ := setupSvc(t)
	svc.google = GoogleSettings{}
	st, err := svc.GoogleStatus(context.Background(), owner)
	require.NoError(t, err)
	assert.False(t, st.Configured)
	_, err = svc.GoogleConnectURL(context.Background(), owner)
	require.Error(t, err)
	assert.Equal(t, response.CodeScheduleGoogleNotReady, err.(*response.AppError).Code)
}

func TestImporterInterfaceKinds(t *testing.T) {
	var _ importer.Importer = importer.TimetableImporter{}
	var _ importer.Importer = importer.ICSImporter{}
	assert.Equal(t, "timetable", importer.TimetableImporter{}.Kind())
	assert.Equal(t, "ics", importer.ICSImporter{}.Kind())
}

func mustSampleXLSX(t *testing.T) []byte {
	t.Helper()
	f := excelize.NewFile()
	sheet := "段茗尧_1120240002"
	require.NoError(t, f.SetSheetName("Sheet1", sheet))
	require.NoError(t, f.SetCellValue(sheet, "A1", "2026-2027学年 第一学期 段茗尧[1120240002] 课表"))
	for i, h := range []string{"节次/星期", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六", "星期日"} {
		cell, err := excelize.CoordinatesToCellName(i+1, 2)
		require.NoError(t, err)
		require.NoError(t, f.SetCellValue(sheet, cell, h))
	}
	require.NoError(t, f.SetCellValue(sheet, "A3", "第一单节08:00-08:45"))
	require.NoError(t, f.SetCellValue(sheet, "B3", "马克思主义基本原理 01\n1-18周 王兰 08:00-10:30 【3-227】"))
	buf, err := f.WriteToBuffer()
	require.NoError(t, err)
	return buf.Bytes()
}
