package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubSvc struct {
	rangeHit bool
	getHit   bool
}

func (s *stubSvc) ListCalendars(context.Context, uuid.UUID, *dto.ListCalendarRequest, *rbacModel.DataScopeCondition) ([]*dto.CalendarResponse, int64, int, int, error) {
	return nil, 0, 1, 20, nil
}
func (s *stubSvc) CreateCalendar(context.Context, uuid.UUID, *dto.CreateCalendarRequest, *rbacModel.DataScopeCondition) (*dto.CalendarResponse, error) {
	return &dto.CalendarResponse{ID: uuid.New().String()}, nil
}
func (s *stubSvc) GetCalendar(context.Context, uuid.UUID, uuid.UUID, *rbacModel.DataScopeCondition) (*dto.CalendarResponse, error) {
	return &dto.CalendarResponse{ID: uuid.New().String()}, nil
}
func (s *stubSvc) UpdateCalendar(context.Context, uuid.UUID, uuid.UUID, *dto.UpdateCalendarRequest, *rbacModel.DataScopeCondition) (*dto.CalendarResponse, error) {
	return &dto.CalendarResponse{ID: uuid.New().String()}, nil
}
func (s *stubSvc) DeleteCalendar(context.Context, uuid.UUID, uuid.UUID, *rbacModel.DataScopeCondition) error {
	return nil
}
func (s *stubSvc) ListMembers(context.Context, uuid.UUID, uuid.UUID, *rbacModel.DataScopeCondition) ([]dto.MemberResponse, error) {
	return nil, nil
}
func (s *stubSvc) AddMember(context.Context, uuid.UUID, uuid.UUID, *dto.AddMemberRequest, *rbacModel.DataScopeCondition) (*dto.MemberResponse, error) {
	return &dto.MemberResponse{ID: uuid.New().String()}, nil
}
func (s *stubSvc) RemoveMember(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, *rbacModel.DataScopeCondition) error {
	return nil
}
func (s *stubSvc) ListEvents(context.Context, uuid.UUID, *dto.ListEventRequest, *rbacModel.DataScopeCondition) ([]*dto.EventResponse, int64, int, int, error) {
	return nil, 0, 1, 20, nil
}
func (s *stubSvc) RangeEvents(context.Context, uuid.UUID, *dto.RangeEventRequest, *rbacModel.DataScopeCondition) ([]*dto.EventResponse, error) {
	s.rangeHit = true
	return []*dto.EventResponse{}, nil
}
func (s *stubSvc) CreateEvent(context.Context, uuid.UUID, *dto.CreateEventRequest, *rbacModel.DataScopeCondition) (*dto.EventResponse, error) {
	return &dto.EventResponse{ID: uuid.New().String()}, nil
}
func (s *stubSvc) GetEvent(context.Context, uuid.UUID, uuid.UUID, *rbacModel.DataScopeCondition) (*dto.EventResponse, error) {
	s.getHit = true
	return &dto.EventResponse{ID: uuid.New().String()}, nil
}
func (s *stubSvc) UpdateEvent(context.Context, uuid.UUID, uuid.UUID, *dto.UpdateEventRequest, *rbacModel.DataScopeCondition) (*dto.EventResponse, error) {
	return &dto.EventResponse{ID: uuid.New().String()}, nil
}
func (s *stubSvc) DeleteEvent(context.Context, uuid.UUID, uuid.UUID, *rbacModel.DataScopeCondition) error {
	return nil
}
func (s *stubSvc) SetReminders(context.Context, uuid.UUID, uuid.UUID, []int, *rbacModel.DataScopeCondition) ([]dto.ReminderResponse, error) {
	return nil, nil
}
func (s *stubSvc) RSVP(context.Context, uuid.UUID, uuid.UUID, int16, *rbacModel.DataScopeCondition) error {
	return nil
}
func (s *stubSvc) DispatchDueReminders(context.Context, string, func(string)) error { return nil }

func testRouter(svc *stubSvc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")
	api.Use(func(c *gin.Context) {
		c.Set(auth.ContextKeyUserID, uuid.New().String())
		c.Next()
	})
	h := New(svc)
	g := api.Group("/schedules")
	g.GET("/calendars", h.ListCalendars)
	g.GET("/events/range", h.RangeEvents)
	g.GET("/events", h.ListEvents)
	g.GET("/events/:id", h.GetEvent)
	return r
}

func TestEventsRangeIsNotCapturedAsID(t *testing.T) {
	svc := &stubSvc{}
	r := testRouter(svc)
	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/schedules/events/range?start="+start+"&end="+end, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.rangeHit)
	assert.False(t, svc.getHit)
}
