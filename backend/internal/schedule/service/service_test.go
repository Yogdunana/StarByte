package service

import (
	"context"
	"testing"
	"time"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func selfScope(uid uuid.UUID) *rbacModel.DataScopeCondition {
	return &rbacModel.DataScopeCondition{Query: "1 = 0", IsSelf: true, Args: []interface{}{uid}}
}

func deptScope(dept uuid.UUID) *rbacModel.DataScopeCondition {
	return &rbacModel.DataScopeCondition{Query: "department_id = ?", Args: []interface{}{dept}}
}

func setupSvc(t *testing.T) (*scheduleService, *memRepo, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	mem := newMem()
	owner := uuid.New()
	other := uuid.New()
	dept := uuid.New()
	mem.addUser(owner, "张三", &dept)
	mem.addUser(other, "李四", &dept)
	svc := New(mem, nil).(*scheduleService)
	return svc, mem, owner, other, dept
}

func TestPersonalCalendarCRUD(t *testing.T) {
	svc, _, owner, other, _ := setupSvc(t)
	ctx := context.Background()
	scope := selfScope(owner)

	cals, total, _, _, err := svc.ListCalendars(ctx, owner, &dto.ListCalendarRequest{}, scope)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, cals, 1)
	assert.Equal(t, model.CalendarPersonal, cals[0].CalendarType)
	assert.True(t, cals[0].CanEdit)

	start := time.Date(2026, 9, 11, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	ev, err := svc.CreateEvent(ctx, owner, &dto.CreateEventRequest{
		Title: "例会准备", StartAt: start, EndAt: end, Location: "A101",
		RemindMinutes: []int{15},
	}, scope)
	require.NoError(t, err)
	assert.Equal(t, "例会准备", ev.Title)
	assert.Len(t, ev.Reminders, 1)

	got, err := svc.GetEvent(ctx, owner, uuid.MustParse(ev.ID), scope)
	require.NoError(t, err)
	assert.Equal(t, "A101", got.Location)

	title := "改期例会"
	got, err = svc.UpdateEvent(ctx, owner, uuid.MustParse(ev.ID), &dto.UpdateEventRequest{Title: &title}, scope)
	require.NoError(t, err)
	assert.Equal(t, "改期例会", got.Title)

	_, err = svc.GetEvent(ctx, other, uuid.MustParse(ev.ID), selfScope(other))
	require.Error(t, err)
	assert.Equal(t, response.CodeScheduleNoAccess, err.(*response.AppError).Code)

	require.NoError(t, svc.DeleteEvent(ctx, owner, uuid.MustParse(ev.ID), scope))
	_, err = svc.GetEvent(ctx, owner, uuid.MustParse(ev.ID), scope)
	require.Error(t, err)
	assert.Equal(t, response.CodeScheduleNotFound, err.(*response.AppError).Code)
}

func TestListDoesNotHonorClientOwnerID(t *testing.T) {
	svc, mem, owner, other, dept := setupSvc(t)
	ctx := context.Background()
	now := time.Now()
	otherCal := &model.Calendar{
		ID: uuid.New(), Name: "别人的日历", CalendarType: model.CalendarPersonal,
		OwnerID: other, Color: "#111", CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, mem.CreateCalendar(ctx, otherCal))
	require.NoError(t, mem.CreateEvent(ctx, &model.Event{
		ID: uuid.New(), CalendarID: otherCal.ID, Title: "秘密会议",
		StartAt: now, EndAt: now.Add(time.Hour), CreatedBy: other, Recurrence: model.RecurrenceNone,
	}))

	cals, _, _, _, err := svc.ListCalendars(ctx, owner, &dto.ListCalendarRequest{}, selfScope(owner))
	require.NoError(t, err)
	for _, c := range cals {
		assert.NotEqual(t, other.String(), c.Owner.ID, "self scope must not list others' personal calendars")
	}

	events, _, _, _, err := svc.ListEvents(ctx, owner, &dto.ListEventRequest{CalendarID: otherCal.ID.String()}, selfScope(owner))
	require.Error(t, err)
	assert.Equal(t, response.CodeCalendarNoAccess, err.(*response.AppError).Code)
	assert.Nil(t, events)

	_, err = svc.CreateCalendar(ctx, owner, &dto.CreateCalendarRequest{
		Name: "窥探", CalendarType: model.CalendarDepartment, DepartmentID: dept.String(),
	}, selfScope(owner))
	require.NoError(t, err)
}

func TestSharedCalendarMemberAccess(t *testing.T) {
	svc, _, owner, other, _ := setupSvc(t)
	ctx := context.Background()

	cal, err := svc.CreateCalendar(ctx, owner, &dto.CreateCalendarRequest{
		Name: "项目日历", CalendarType: model.CalendarProject,
	}, nil)
	require.NoError(t, err)

	_, err = svc.GetCalendar(ctx, other, uuid.MustParse(cal.ID), selfScope(other))
	require.Error(t, err)
	assert.Equal(t, response.CodeCalendarNoAccess, err.(*response.AppError).Code)

	_, err = svc.AddMember(ctx, owner, uuid.MustParse(cal.ID), &dto.AddMemberRequest{UserID: other.String(), Role: model.MemberEditor}, nil)
	require.NoError(t, err)

	got, err := svc.GetCalendar(ctx, other, uuid.MustParse(cal.ID), selfScope(other))
	require.NoError(t, err)
	assert.True(t, got.CanEdit)

	start := time.Now().Add(2 * time.Hour)
	ev, err := svc.CreateEvent(ctx, other, &dto.CreateEventRequest{
		CalendarID: cal.ID, Title: "联调", StartAt: start, EndAt: start.Add(30 * time.Minute),
	}, selfScope(other))
	require.NoError(t, err)
	assert.Equal(t, "联调", ev.Title)
}

func TestDepartmentScopeHidesPersonal(t *testing.T) {
	svc, mem, owner, other, dept := setupSvc(t)
	ctx := context.Background()
	now := time.Now()
	require.NoError(t, mem.CreateCalendar(ctx, &model.Calendar{
		ID: uuid.New(), Name: "李四个人", CalendarType: model.CalendarPersonal,
		OwnerID: other, CreatedAt: now, UpdatedAt: now,
	}))
	deptCal := &model.Calendar{
		ID: uuid.New(), Name: "技术部", CalendarType: model.CalendarDepartment,
		OwnerID: other, DepartmentID: &dept, CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, mem.CreateCalendar(ctx, deptCal))

	cals, _, _, _, err := svc.ListCalendars(ctx, owner, &dto.ListCalendarRequest{}, deptScope(dept))
	require.NoError(t, err)
	names := map[string]bool{}
	for _, c := range cals {
		names[c.Name] = true
	}
	assert.True(t, names["技术部"])
	assert.False(t, names["李四个人"], "department scope must not leak personal calendars")
}

func TestConflictAndInvalidTime(t *testing.T) {
	svc, _, owner, _, _ := setupSvc(t)
	ctx := context.Background()
	scope := selfScope(owner)
	start := time.Date(2026, 9, 12, 9, 0, 0, 0, time.UTC)
	_, err := svc.CreateEvent(ctx, owner, &dto.CreateEventRequest{
		Title: "A", StartAt: start, EndAt: start.Add(-time.Minute),
	}, scope)
	require.Error(t, err)
	assert.Equal(t, response.CodeScheduleInvalidTime, err.(*response.AppError).Code)

	_, err = svc.CreateEvent(ctx, owner, &dto.CreateEventRequest{
		Title: "A", StartAt: start, EndAt: start.Add(time.Hour),
	}, scope)
	require.NoError(t, err)
	_, err = svc.CreateEvent(ctx, owner, &dto.CreateEventRequest{
		Title: "B", StartAt: start.Add(30 * time.Minute), EndAt: start.Add(90 * time.Minute),
	}, scope)
	require.Error(t, err)
	assert.Equal(t, response.CodeScheduleConflict, err.(*response.AppError).Code)
}

func TestRecurrenceRejectedAndRemindInvalid(t *testing.T) {
	svc, _, owner, _, _ := setupSvc(t)
	ctx := context.Background()
	scope := selfScope(owner)
	start := time.Now().Add(time.Hour)
	_, err := svc.CreateEvent(ctx, owner, &dto.CreateEventRequest{
		Title: "坏规则", StartAt: start, EndAt: start.Add(time.Hour), Recurrence: "FREQ=DAILY;INTERVAL=2",
	}, scope)
	require.Error(t, err)
	assert.Equal(t, response.CodeScheduleRecurrenceLimited, err.(*response.AppError).Code)

	_, err = svc.CreateEvent(ctx, owner, &dto.CreateEventRequest{
		Title: "坏提醒", StartAt: start.Add(2 * time.Hour), EndAt: start.Add(3 * time.Hour), RemindMinutes: []int{7},
	}, scope)
	require.Error(t, err)
	assert.Equal(t, response.CodeScheduleReminderInvalid, err.(*response.AppError).Code)
}

func TestAttendeeRSVPAndReminderDispatch(t *testing.T) {
	svc, mem, owner, other, _ := setupSvc(t)
	n := &captureNotify{}
	svc.notify = n
	ctx := context.Background()
	scope := selfScope(owner)
	start := time.Now().Add(10 * time.Minute)
	ev, err := svc.CreateEvent(ctx, owner, &dto.CreateEventRequest{
		Title: "答辩", StartAt: start, EndAt: start.Add(time.Hour),
		AttendeeIDs: []string{other.String()}, RemindMinutes: []int{15},
	}, scope)
	require.NoError(t, err)

	require.NoError(t, svc.RSVP(ctx, other, uuid.MustParse(ev.ID), model.AttendeeAccepted, selfScope(other)))
	att, err := mem.GetAttendee(ctx, uuid.MustParse(ev.ID), other)
	require.NoError(t, err)
	require.NotNil(t, att)
	assert.Equal(t, model.AttendeeAccepted, att.ResponseStatus)

	require.NoError(t, svc.DispatchDueReminders(ctx, "", func(string) {}))
	assert.Equal(t, 1, n.n)
	require.NoError(t, svc.DispatchDueReminders(ctx, "", func(string) {}))
	assert.Equal(t, 1, n.n)
}

func TestCannotDeleteOwnPersonalCalendar(t *testing.T) {
	svc, _, owner, _, _ := setupSvc(t)
	ctx := context.Background()
	cals, _, _, _, err := svc.ListCalendars(ctx, owner, &dto.ListCalendarRequest{}, selfScope(owner))
	require.NoError(t, err)
	require.NotEmpty(t, cals)
	err = svc.DeleteCalendar(ctx, owner, uuid.MustParse(cals[0].ID), selfScope(owner))
	require.Error(t, err)
	assert.Equal(t, response.CodeScheduleInvalidState, err.(*response.AppError).Code)
}

type captureNotify struct{ n int }

func (c *captureNotify) Send(_ context.Context, _ []uuid.UUID, _ string, _ map[string]interface{}) error {
	c.n++
	return nil
}
