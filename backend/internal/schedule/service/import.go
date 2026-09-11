package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/importer"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *scheduleService) ImportTimetable(ctx context.Context, operator uuid.UUID, filename string, raw []byte, semesterStart time.Time, scope *rbacModel.DataScopeCondition) (*dto.ImportResult, error) {
	if err := s.ensurePersonal(ctx, operator); err != nil {
		return nil, err
	}
	if semesterStart.IsZero() {
		return nil, response.NewError(response.CodeScheduleImportInvalid, "semester_start 必填，应为第 1 周星期一的 ISO 日期（YYYY-MM-DD）")
	}
	if len(raw) == 0 {
		return nil, response.NewError(response.CodeScheduleImportInvalid, "未上传课表文件")
	}
	if !strings.HasSuffix(strings.ToLower(filename), ".xlsx") && !strings.HasSuffix(strings.ToLower(filename), ".xls") {
		return nil, response.NewError(response.CodeScheduleImportInvalid, "课表导入仅支持 xlsx")
	}
	events, meta, err := importer.TimetableImporter{}.Parse(ctx, raw, importer.Options{SemesterStart: semesterStart, Filename: filename})
	if err != nil {
		return nil, response.NewError(response.CodeScheduleImportInvalid, "课表解析失败: "+err.Error())
	}
	if len(events) == 0 {
		return nil, response.NewError(response.CodeScheduleImportEmpty, "课表中没有可导入的节次（页脚「未排具体节次课程」已跳过）")
	}
	cal, replaced, err := s.ensureLayer(ctx, operator, model.SourceTimetable, meta.SourceKey, meta.CalendarName, meta.Title)
	if err != nil {
		return nil, err
	}
	if err := s.writeDrafts(ctx, operator, cal.ID, model.OriginGenerated, events); err != nil {
		return nil, err
	}
	out, err := s.GetCalendar(ctx, operator, cal.ID, scope)
	if err != nil {
		return nil, err
	}
	return &dto.ImportResult{
		CalendarID: cal.ID.String(), Calendar: out, EventCount: len(events),
		Replaced: replaced, Source: model.SourceTimetable,
	}, nil
}

func (s *scheduleService) ImportICS(ctx context.Context, operator uuid.UUID, filename string, raw []byte, calendarID string, scope *rbacModel.DataScopeCondition) (*dto.ImportResult, error) {
	if err := s.ensurePersonal(ctx, operator); err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, response.NewError(response.CodeScheduleImportInvalid, "未上传 ICS 文件")
	}
	events, meta, err := importer.ICSImporter{}.Parse(ctx, raw, importer.Options{Filename: filename})
	if err != nil {
		return nil, response.NewError(response.CodeScheduleImportInvalid, "ICS 解析失败: "+err.Error())
	}
	if len(events) == 0 {
		return nil, response.NewError(response.CodeScheduleImportEmpty, "ICS 中没有可导入的事件")
	}
	var cal *model.Calendar
	replaced := false
	if strings.TrimSpace(calendarID) != "" {
		id, err := uuid.Parse(calendarID)
		if err != nil {
			return nil, response.NewError(response.CodeBadRequest, "无效的日历ID")
		}
		named, err := s.requireCalendarEdit(ctx, operator, id, scope)
		if err != nil {
			return nil, err
		}
		if named.OwnerID != operator && !isAllScope(scope) {
			return nil, response.NewError(response.CodeCalendarNoAccess, "只能导入到自己的日历图层")
		}
		cal = &named.Calendar
		replaced = true
	} else {
		var err error
		cal, replaced, err = s.ensureLayer(ctx, operator, model.SourceImport, meta.SourceKey, meta.CalendarName, meta.Title)
		if err != nil {
			return nil, err
		}
	}
	if err := s.writeDrafts(ctx, operator, cal.ID, model.OriginICS, events); err != nil {
		return nil, err
	}
	out, err := s.GetCalendar(ctx, operator, cal.ID, scope)
	if err != nil {
		return nil, err
	}
	return &dto.ImportResult{
		CalendarID: cal.ID.String(), Calendar: out, EventCount: len(events),
		Replaced: replaced, Source: model.NormalizeSource(cal.Source),
	}, nil
}

func (s *scheduleService) ensureLayer(ctx context.Context, owner uuid.UUID, source, sourceKey, name, desc string) (*model.Calendar, bool, error) {
	existing, err := s.rows.CalendarBySource(ctx, owner, source, sourceKey)
	if err != nil {
		return nil, false, fmt.Errorf("lookup layer: %w", err)
	}
	if existing != nil {
		existing.SourceKey = sourceKey
		existing.Name = firstNonEmpty(name, existing.Name)
		existing.Description = firstNonEmpty(desc, existing.Description)
		existing.UpdatedAt = time.Now()
		if err := s.rows.UpdateCalendar(ctx, existing); err != nil {
			return nil, false, fmt.Errorf("update layer: %w", err)
		}
		return existing, true, nil
	}
	now := time.Now()
	row := &model.Calendar{
		ID: uuid.New(), Name: firstNonEmpty(name, layerDefaultName(source)), Description: strings.TrimSpace(desc),
		CalendarType: model.CalendarPersonal, Source: source, SourceKey: sourceKey,
		Color: model.DefaultLayerColor(source), OwnerID: owner, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.rows.CreateCalendar(ctx, row); err != nil {
		return nil, false, fmt.Errorf("create layer: %w", err)
	}
	return row, false, nil
}

func (s *scheduleService) writeDrafts(ctx context.Context, operator, calendarID uuid.UUID, origin string, drafts []importer.DraftEvent) error {
	now := time.Now()
	rows := make([]model.Event, 0, len(drafts))
	for _, d := range drafts {
		if d.EndAt.Before(d.StartAt) {
			continue
		}
		rows = append(rows, model.Event{
			ID: uuid.New(), CalendarID: calendarID, Title: clip(d.Title, 200),
			Description: d.Description, Location: clip(d.Location, 200),
			StartAt: d.StartAt, EndAt: d.EndAt, AllDay: d.AllDay,
			Status: model.EventConfirmed, Recurrence: model.RecurrenceNone,
			Origin: origin, ExternalUID: clip(d.ExternalUID, 200),
			CreatedBy: operator, CreatedAt: now, UpdatedAt: now,
		})
	}
	if err := s.rows.ReplaceOriginEvents(ctx, calendarID, origin, rows); err != nil {
		return fmt.Errorf("replace imported events: %w", err)
	}
	return nil
}

func layerDefaultName(source string) string {
	switch source {
	case model.SourceTimetable:
		return "学校课表"
	case model.SourceGoogle:
		return "Google 日历"
	case model.SourceImport:
		return "导入日历"
	default:
		return "我的日历"
	}
}

func firstNonEmpty(v, fallback string) string {
	if s := strings.TrimSpace(v); s != "" {
		return s
	}
	return fallback
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if n > 0 && len([]rune(s)) > n {
		return string([]rune(s)[:n])
	}
	return s
}
