package service

import (
	"strings"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/model"
	"github.com/google/uuid"
)

func mapCalendar(row *model.CalendarNamed, viewer uuid.UUID, scope *rbacModel.DataScopeCondition) *dto.CalendarResponse {
	dept := ""
	if row.DepartmentID != nil {
		dept = row.DepartmentID.String()
	}
	return &dto.CalendarResponse{
		ID: row.ID.String(), Name: row.Name, Description: row.Description,
		CalendarType: row.CalendarType, Color: row.Color,
		Owner:        dto.Person{ID: row.OwnerID.String(), Name: row.OwnerName},
		DepartmentID: dept, DepartmentName: row.DepartmentName, MemberRole: row.MemberRole,
		CanEdit:   canEditCalendar(scope, &row.Calendar, viewer, row.MemberRole),
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func mapMember(row *model.CalendarMemberNamed) dto.MemberResponse {
	return dto.MemberResponse{
		ID:   row.ID.String(),
		User: dto.Person{ID: row.UserID.String(), Name: displayName(row.RealName, row.Username)},
		Role: row.Role, CreatedAt: row.CreatedAt,
	}
}

func mapAttendee(row *model.AttendeeNamed) dto.MemberResponse {
	return dto.MemberResponse{
		ID:   row.ID.String(),
		User: dto.Person{ID: row.UserID.String(), Name: displayName(row.RealName, row.Username)},
		Role: row.ResponseStatus, CreatedAt: row.CreatedAt,
	}
}

func mapEvent(row *model.EventNamed, viewer uuid.UUID, scope *rbacModel.DataScopeCondition, memberRole int16) *dto.EventResponse {
	meeting := ""
	if row.MeetingID != nil {
		meeting = row.MeetingID.String()
	}
	color := row.Color
	if color == "" {
		color = row.CalendarColor
	}
	return &dto.EventResponse{
		ID: row.ID.String(), CalendarID: row.CalendarID.String(), CalendarName: row.CalendarName,
		CalendarColor: row.CalendarColor, Title: row.Title, Description: row.Description,
		Location: row.Location, StartAt: row.StartAt, EndAt: row.EndAt, AllDay: row.AllDay,
		Color: color, Status: row.Status, Recurrence: row.Recurrence, RecurrenceUntil: row.RecurrenceUntil,
		MeetingID: meeting, Creator: dto.Person{ID: row.CreatedBy.String(), Name: row.CreatorName},
		AttendeeCount: row.AttendeeCount, CanEdit: canEditEvent(scope, row, viewer, memberRole),
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func mapReminders(rows []model.Reminder) []dto.ReminderResponse {
	out := make([]dto.ReminderResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.ReminderResponse{
			ID: r.ID.String(), MinutesBefore: r.MinutesBefore, Method: r.Method, TriggeredAt: r.TriggeredAt,
		})
	}
	return out
}

func displayName(realName, username string) string {
	if s := strings.TrimSpace(realName); s != "" {
		return s
	}
	return username
}

func defaultColor(c string) string {
	c = strings.TrimSpace(c)
	if c == "" {
		return "#2563eb"
	}
	return c
}

func normalizePage(page, size int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 200 {
		size = 200
	}
	return page, size
}
