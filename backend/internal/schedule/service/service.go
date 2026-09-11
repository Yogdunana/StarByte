package service

import (
	"context"
	"net/http"
	"time"

	notifdto "github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	notifsvc "github.com/Yogdunana/StarByte/backend/internal/notification/service"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/repo"
	"github.com/google/uuid"
)

type Notifier interface {
	Send(ctx context.Context, userIDs []uuid.UUID, template string, vars map[string]interface{}) error
}

type notificationAdapter struct{ inner notifsvc.NotificationService }

func NewNotifier(inner notifsvc.NotificationService) Notifier {
	if inner == nil {
		return nil
	}
	return &notificationAdapter{inner: inner}
}

func (a *notificationAdapter) Send(ctx context.Context, userIDs []uuid.UUID, template string, vars map[string]interface{}) error {
	if len(userIDs) == 0 {
		return nil
	}
	return a.inner.Send(ctx, &notifdto.SendNotificationRequest{
		UserIDs: userIDs, TemplateCode: template, Variables: vars, Channels: []string{"in_app", "websocket"},
	})
}

type Service interface {
	ListCalendars(ctx context.Context, viewer uuid.UUID, req *dto.ListCalendarRequest, scope *rbacModel.DataScopeCondition) ([]*dto.CalendarResponse, int64, int, int, error)
	CreateCalendar(ctx context.Context, operator uuid.UUID, req *dto.CreateCalendarRequest, scope *rbacModel.DataScopeCondition) (*dto.CalendarResponse, error)
	GetCalendar(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.CalendarResponse, error)
	UpdateCalendar(ctx context.Context, operator, id uuid.UUID, req *dto.UpdateCalendarRequest, scope *rbacModel.DataScopeCondition) (*dto.CalendarResponse, error)
	DeleteCalendar(ctx context.Context, operator, id uuid.UUID, scope *rbacModel.DataScopeCondition) error
	ListMembers(ctx context.Context, viewer, calendarID uuid.UUID, scope *rbacModel.DataScopeCondition) ([]dto.MemberResponse, error)
	AddMember(ctx context.Context, operator, calendarID uuid.UUID, req *dto.AddMemberRequest, scope *rbacModel.DataScopeCondition) (*dto.MemberResponse, error)
	RemoveMember(ctx context.Context, operator, calendarID, userID uuid.UUID, scope *rbacModel.DataScopeCondition) error

	ListEvents(ctx context.Context, viewer uuid.UUID, req *dto.ListEventRequest, scope *rbacModel.DataScopeCondition) ([]*dto.EventResponse, int64, int, int, error)
	RangeEvents(ctx context.Context, viewer uuid.UUID, req *dto.RangeEventRequest, scope *rbacModel.DataScopeCondition) ([]*dto.EventResponse, error)
	CreateEvent(ctx context.Context, operator uuid.UUID, req *dto.CreateEventRequest, scope *rbacModel.DataScopeCondition) (*dto.EventResponse, error)
	GetEvent(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.EventResponse, error)
	UpdateEvent(ctx context.Context, operator, id uuid.UUID, req *dto.UpdateEventRequest, scope *rbacModel.DataScopeCondition) (*dto.EventResponse, error)
	DeleteEvent(ctx context.Context, operator, id uuid.UUID, scope *rbacModel.DataScopeCondition) error
	SetReminders(ctx context.Context, operator, id uuid.UUID, minutes []int, scope *rbacModel.DataScopeCondition) ([]dto.ReminderResponse, error)
	RSVP(ctx context.Context, operator, id uuid.UUID, response int16, scope *rbacModel.DataScopeCondition) error
	DispatchDueReminders(ctx context.Context, payload string, logf func(string)) error

	ImportTimetable(ctx context.Context, operator uuid.UUID, filename string, raw []byte, semesterStart time.Time, scope *rbacModel.DataScopeCondition) (*dto.ImportResult, error)
	ImportICS(ctx context.Context, operator uuid.UUID, filename string, raw []byte, calendarID string, scope *rbacModel.DataScopeCondition) (*dto.ImportResult, error)

	GoogleStatus(ctx context.Context, operator uuid.UUID) (*dto.GoogleStatusResponse, error)
	GoogleConnectURL(ctx context.Context, operator uuid.UUID) (*dto.GoogleConnectResponse, error)
	GoogleCallback(ctx context.Context, operator uuid.UUID, code, state string, scope *rbacModel.DataScopeCondition) (*dto.GoogleStatusResponse, error)
	GoogleDisconnect(ctx context.Context, operator uuid.UUID) error
	GoogleSync(ctx context.Context, operator uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.ImportResult, error)
	DispatchGoogleSync(ctx context.Context, payload string, logf func(string)) error
	ParseGoogleState(state string) (uuid.UUID, error)
	FrontendRedirect() string
}

type scheduleService struct {
	rows   repo.Repository
	notify Notifier
	google GoogleSettings
	httpDo func(*http.Request) (*http.Response, error)
}

func New(rows repo.Repository, notify Notifier) Service {
	return &scheduleService{rows: rows, notify: notify, google: LoadGoogleSettings()}
}
