package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *scheduleService) ListEvents(ctx context.Context, viewer uuid.UUID, req *dto.ListEventRequest, scope *rbacModel.DataScopeCondition) ([]*dto.EventResponse, int64, int, int, error) {
	if req == nil {
		req = &dto.ListEventRequest{}
	}
	if req.CalendarID != "" {
		cid, err := uuid.Parse(req.CalendarID)
		if err != nil {
			return nil, 0, 0, 0, response.NewError(response.CodeBadRequest, "无效的日历ID")
		}
		if _, err := s.requireCalendarView(ctx, viewer, cid, scope); err != nil {
			return nil, 0, 0, 0, err
		}
	}
	rows, total, err := s.rows.ListEvents(ctx, viewer, req, scope)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("list events: %w", err)
	}
	out := make([]*dto.EventResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapEvent(&rows[i], viewer, scope, rows[i].MemberRole))
	}
	page, size := normalizePage(req.Page, req.PageSize)
	return out, total, page, size, nil
}

func (s *scheduleService) RangeEvents(ctx context.Context, viewer uuid.UUID, req *dto.RangeEventRequest, scope *rbacModel.DataScopeCondition) ([]*dto.EventResponse, error) {
	if req.End.Before(req.Start) {
		return nil, response.NewError(response.CodeScheduleInvalidTime, "结束时间不能早于开始时间")
	}
	if req.End.Sub(req.Start) > 400*24*time.Hour {
		return nil, response.NewError(response.CodeBadRequest, "查询窗口不能超过 400 天")
	}
	var rows []*dto.EventResponse
	for page := 1; page <= 50; page++ {
		listReq := &dto.ListEventRequest{CalendarID: req.CalendarID, Start: req.Start, End: req.End, Page: page, PageSize: 200}
		chunk, total, _, _, err := s.ListEvents(ctx, viewer, listReq, scope)
		if err != nil {
			return nil, err
		}
		rows = append(rows, chunk...)
		if int64(len(rows)) >= total || len(chunk) == 0 {
			break
		}
	}
	expanded := make([]*dto.EventResponse, 0, len(rows))
	for _, row := range rows {
		ev := eventFromResponse(row)
		for _, occ := range expandOccurrences(ev, req.Start, req.End) {
			cp := *row
			cp.StartAt = occ.Start
			cp.EndAt = occ.End
			start := occ.Start
			cp.OccurrenceStart = &start
			expanded = append(expanded, &cp)
		}
	}
	return expanded, nil
}

func (s *scheduleService) CreateEvent(ctx context.Context, operator uuid.UUID, req *dto.CreateEventRequest, scope *rbacModel.DataScopeCondition) (*dto.EventResponse, error) {
	if err := validateEventTime(req.StartAt, req.EndAt); err != nil {
		return nil, err
	}
	rule, err := validateRecurrence(req.Recurrence, req.RecurrenceUntil, req.StartAt)
	if err != nil {
		return nil, err
	}
	cal, err := s.resolveEventCalendar(ctx, operator, req.CalendarID, scope)
	if err != nil {
		return nil, err
	}
	if err := s.ensureNoConflict(ctx, cal.ID, uuid.Nil, &model.Event{
		StartAt: req.StartAt, EndAt: req.EndAt, Recurrence: rule, RecurrenceUntil: req.RecurrenceUntil,
	}); err != nil {
		return nil, err
	}
	meetingID, err := parseOptionalUUID(req.MeetingID)
	if err != nil {
		return nil, response.NewError(response.CodeBadRequest, "会议 ID 无效")
	}
	now := time.Now()
	row := &model.Event{
		ID: uuid.New(), CalendarID: cal.ID, Title: strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description), Location: strings.TrimSpace(req.Location),
		StartAt: req.StartAt, EndAt: req.EndAt, AllDay: req.AllDay, Color: strings.TrimSpace(req.Color),
		Status: model.EventConfirmed, Recurrence: rule, RecurrenceUntil: req.RecurrenceUntil,
		MeetingID: meetingID, Origin: model.OriginManual, CreatedBy: operator, CreatedAt: now, UpdatedAt: now,
	}
	atts, err := s.buildAttendees(ctx, row.ID, req.AttendeeIDs)
	if err != nil {
		return nil, err
	}
	rems, err := s.buildReminders(row.ID, req.RemindMinutes)
	if err != nil {
		return nil, err
	}
	if err := s.rows.CreateEventWithDetails(ctx, row, atts, rems); err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}
	return s.GetEvent(ctx, operator, row.ID, scope)
}

func (s *scheduleService) GetEvent(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.EventResponse, error) {
	row, memberRole, isAttendee, err := s.loadVisibleEvent(ctx, viewer, id, scope)
	if err != nil {
		return nil, err
	}
	out := mapEvent(row, viewer, scope, memberRole)
	atts, err := s.rows.ListAttendees(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list attendees: %w", err)
	}
	out.Attendees = make([]dto.MemberResponse, 0, len(atts))
	for i := range atts {
		out.Attendees = append(out.Attendees, mapAttendee(&atts[i]))
	}
	rems, err := s.rows.ListReminders(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list reminders: %w", err)
	}
	out.Reminders = mapReminders(rems)
	_ = isAttendee
	return out, nil
}

func (s *scheduleService) UpdateEvent(ctx context.Context, operator, id uuid.UUID, req *dto.UpdateEventRequest, scope *rbacModel.DataScopeCondition) (*dto.EventResponse, error) {
	row, memberRole, _, err := s.loadVisibleEvent(ctx, operator, id, scope)
	if err != nil {
		return nil, err
	}
	if !canEditEvent(scope, row, operator, memberRole) {
		return nil, response.NewError(response.CodeScheduleNoAccess, "无权修改该日程")
	}
	if req.Title != nil {
		row.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		row.Description = strings.TrimSpace(*req.Description)
	}
	if req.Location != nil {
		row.Location = strings.TrimSpace(*req.Location)
	}
	if req.StartAt != nil || req.EndAt != nil {
		if ignoreOccurrenceTimes(row.Event, req.StartAt, req.EndAt) {
			req.StartAt = nil
			req.EndAt = nil
		}
	}
	if req.StartAt != nil {
		row.StartAt = *req.StartAt
	}
	if req.EndAt != nil {
		row.EndAt = *req.EndAt
	}
	if req.AllDay != nil {
		row.AllDay = *req.AllDay
	}
	if req.Color != nil {
		row.Color = strings.TrimSpace(*req.Color)
	}
	if req.Status != nil {
		row.Status = *req.Status
	}
	if req.Recurrence != nil || req.RecurrenceUntil != nil {
		rule := row.Recurrence
		if req.Recurrence != nil {
			rule = *req.Recurrence
		}
		norm, err := validateRecurrence(rule, firstTime(req.RecurrenceUntil, row.RecurrenceUntil), row.StartAt)
		if err != nil {
			return nil, err
		}
		row.Recurrence = norm
		if req.RecurrenceUntil != nil {
			row.RecurrenceUntil = req.RecurrenceUntil
		}
	}
	if req.MeetingID != nil {
		mid, err := parseOptionalUUID(*req.MeetingID)
		if err != nil {
			return nil, response.NewError(response.CodeBadRequest, "会议 ID 无效")
		}
		row.MeetingID = mid
	}
	if err := validateEventTime(row.StartAt, row.EndAt); err != nil {
		return nil, err
	}
	if err := s.ensureNoConflict(ctx, row.CalendarID, row.ID, &row.Event); err != nil {
		return nil, err
	}
	row.UpdatedAt = time.Now()
	if err := s.rows.UpdateEvent(ctx, &row.Event); err != nil {
		return nil, fmt.Errorf("update event: %w", err)
	}
	return s.GetEvent(ctx, operator, id, scope)
}

func (s *scheduleService) DeleteEvent(ctx context.Context, operator, id uuid.UUID, scope *rbacModel.DataScopeCondition) error {
	row, memberRole, _, err := s.loadVisibleEvent(ctx, operator, id, scope)
	if err != nil {
		return err
	}
	if !canEditEvent(scope, row, operator, memberRole) {
		return response.NewError(response.CodeScheduleNoAccess, "无权删除该日程")
	}
	if err := s.rows.DeleteEvent(ctx, id); err != nil {
		return fmt.Errorf("delete event: %w", err)
	}
	return nil
}

func (s *scheduleService) SetReminders(ctx context.Context, operator, id uuid.UUID, minutes []int, scope *rbacModel.DataScopeCondition) ([]dto.ReminderResponse, error) {
	row, memberRole, _, err := s.loadVisibleEvent(ctx, operator, id, scope)
	if err != nil {
		return nil, err
	}
	if !canEditEvent(scope, row, operator, memberRole) {
		return nil, response.NewError(response.CodeScheduleNoAccess, "无权设置提醒")
	}
	if err := s.replaceReminders(ctx, id, minutes); err != nil {
		return nil, err
	}
	rems, err := s.rows.ListReminders(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list reminders: %w", err)
	}
	return mapReminders(rems), nil
}

func (s *scheduleService) RSVP(ctx context.Context, operator, id uuid.UUID, status int16, scope *rbacModel.DataScopeCondition) error {
	if !model.ValidAttendeeResponse(status) {
		return response.NewError(response.CodeBadRequest, "回复状态不合法")
	}
	if _, _, _, err := s.loadVisibleEvent(ctx, operator, id, scope); err != nil {
		return err
	}
	att, err := s.rows.GetAttendee(ctx, id, operator)
	if err != nil {
		return fmt.Errorf("get attendee: %w", err)
	}
	if att == nil {
		return response.NewError(response.CodeScheduleAttendeeGone, "你不是该日程的参与人")
	}
	att.ResponseStatus = status
	att.UpdatedAt = time.Now()
	if err := s.rows.UpdateAttendee(ctx, att); err != nil {
		return fmt.Errorf("update attendee: %w", err)
	}
	return nil
}

func (s *scheduleService) resolveEventCalendar(ctx context.Context, operator uuid.UUID, calendarID string, scope *rbacModel.DataScopeCondition) (*model.Calendar, error) {
	if strings.TrimSpace(calendarID) == "" {
		if err := s.ensurePersonal(ctx, operator); err != nil {
			return nil, err
		}
		cal, err := s.rows.PersonalCalendar(ctx, operator)
		if err != nil {
			return nil, fmt.Errorf("lookup personal calendar: %w", err)
		}
		if cal == nil {
			return nil, response.NewError(response.CodeCalendarNotFound, "个人日历不存在")
		}
		return cal, nil
	}
	id, err := uuid.Parse(calendarID)
	if err != nil {
		return nil, response.NewError(response.CodeBadRequest, "无效的日历ID")
	}
	named, err := s.requireCalendarEdit(ctx, operator, id, scope)
	if err != nil {
		return nil, err
	}
	return &named.Calendar, nil
}

func (s *scheduleService) loadVisibleEvent(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*model.EventNamed, int16, bool, error) {
	row, err := s.rows.GetEventNamed(ctx, id)
	if err != nil {
		return nil, 0, false, fmt.Errorf("get event: %w", err)
	}
	if row == nil {
		return nil, 0, false, response.NewError(response.CodeScheduleNotFound, "日程事件不存在")
	}
	member, err := s.rows.GetMember(ctx, row.CalendarID, viewer)
	if err != nil {
		return nil, 0, false, err
	}
	var role int16
	if member != nil {
		role = member.Role
	}
	att, err := s.rows.GetAttendee(ctx, id, viewer)
	if err != nil {
		return nil, 0, false, err
	}
	isAttendee := att != nil
	if !canViewEvent(scope, row, viewer, role, isAttendee) {
		return nil, 0, false, response.NewError(response.CodeScheduleNoAccess, "无权查看该日程")
	}
	return row, role, isAttendee, nil
}

func (s *scheduleService) ensureNoConflict(ctx context.Context, calendarID, exclude uuid.UUID, candidate *model.Event) error {
	windowStart, windowEnd := candidate.StartAt, candidate.EndAt
	if model.NormalizeRecurrence(candidate.Recurrence) != model.RecurrenceNone {
		if candidate.RecurrenceUntil != nil && candidate.RecurrenceUntil.After(windowEnd) {
			windowEnd = *candidate.RecurrenceUntil
		} else if candidate.RecurrenceUntil == nil {
			windowEnd = candidate.StartAt.AddDate(1, 0, 0)
		}
	}
	rows, err := s.rows.Overlapping(ctx, calendarID, exclude, windowStart, windowEnd)
	if err != nil {
		return fmt.Errorf("check conflict: %w", err)
	}
	for i := range rows {
		if occurrencesOverlap(&rows[i], candidate, windowStart, windowEnd) {
			return response.NewError(response.CodeScheduleConflict, "同一日历存在时间冲突")
		}
	}
	return nil
}

func occurrencesOverlap(a, b *model.Event, windowStart, windowEnd time.Time) bool {
	left := expandOccurrences(a, windowStart, windowEnd)
	right := expandOccurrences(b, windowStart, windowEnd)
	for _, x := range left {
		for _, y := range right {
			if x.Start.Before(y.End) && x.End.After(y.Start) {
				return true
			}
		}
	}
	return false
}

func ignoreOccurrenceTimes(ev model.Event, start, end *time.Time) bool {
	if model.NormalizeRecurrence(ev.Recurrence) == model.RecurrenceNone || start == nil {
		return false
	}
	if start.Equal(ev.StartAt) {
		return false
	}
	dur := ev.EndAt.Sub(ev.StartAt)
	if end != nil && end.Sub(*start) != dur {
		return false
	}
	windowStart := start.Add(-time.Second)
	windowEnd := start.Add(time.Second)
	for _, occ := range expandOccurrences(&ev, windowStart, windowEnd) {
		if occ.Start.Equal(*start) {
			return true
		}
	}
	return false
}

func (s *scheduleService) replaceAttendees(ctx context.Context, eventID uuid.UUID, ids []string) error {
	rows, err := s.buildAttendees(ctx, eventID, ids)
	if err != nil {
		return err
	}
	return s.rows.ReplaceAttendees(ctx, eventID, rows)
}

func (s *scheduleService) buildAttendees(ctx context.Context, eventID uuid.UUID, ids []string) ([]model.Attendee, error) {
	seen := map[uuid.UUID]struct{}{}
	rows := make([]model.Attendee, 0, len(ids))
	now := time.Now()
	for _, raw := range ids {
		uid, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil {
			return nil, response.NewError(response.CodeBadRequest, "参与人 ID 无效")
		}
		if _, ok := seen[uid]; ok {
			continue
		}
		user, err := s.rows.GetUser(ctx, uid)
		if err != nil {
			return nil, fmt.Errorf("lookup attendee: %w", err)
		}
		if user == nil {
			return nil, response.NewError(response.CodeBadRequest, "参与人不存在")
		}
		seen[uid] = struct{}{}
		rows = append(rows, model.Attendee{ID: uuid.New(), EventID: eventID, UserID: uid, CreatedAt: now, UpdatedAt: now})
	}
	return rows, nil
}

func (s *scheduleService) replaceReminders(ctx context.Context, eventID uuid.UUID, minutes []int) error {
	rows, err := s.buildReminders(eventID, minutes)
	if err != nil {
		return err
	}
	return s.rows.ReplaceReminders(ctx, eventID, rows)
}

func (s *scheduleService) buildReminders(eventID uuid.UUID, minutes []int) ([]model.Reminder, error) {
	seen := map[int]struct{}{}
	rows := make([]model.Reminder, 0, len(minutes))
	now := time.Now()
	for _, m := range minutes {
		if !model.ValidRemindMinutes(m) {
			return nil, response.NewError(response.CodeScheduleReminderInvalid, "提醒仅支持提前 5/15/30/60 分钟")
		}
		if _, ok := seen[m]; ok {
			continue
		}
		seen[m] = struct{}{}
		rows = append(rows, model.Reminder{ID: uuid.New(), EventID: eventID, MinutesBefore: m, Method: model.RemindApp, CreatedAt: now})
	}
	return rows, nil
}

func validateEventTime(start, end time.Time) error {
	if start.IsZero() || end.IsZero() {
		return response.NewError(response.CodeScheduleInvalidTime, "开始和结束时间不能为空")
	}
	if end.Before(start) {
		return response.NewError(response.CodeScheduleInvalidTime, "结束时间不能早于开始时间")
	}
	return nil
}

func parseOptionalUUID(raw string) (*uuid.UUID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func firstTime(a, b *time.Time) *time.Time {
	if a != nil {
		return a
	}
	return b
}

func eventFromResponse(row *dto.EventResponse) *model.Event {
	id, _ := uuid.Parse(row.ID)
	cid, _ := uuid.Parse(row.CalendarID)
	return &model.Event{
		ID: id, CalendarID: cid, StartAt: row.StartAt, EndAt: row.EndAt,
		Recurrence: row.Recurrence, RecurrenceUntil: row.RecurrenceUntil,
	}
}
