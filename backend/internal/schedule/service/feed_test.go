package service

import (
	"context"
	"testing"
	"time"

	activityModel "github.com/Yogdunana/StarByte/backend/internal/activity/model"
	interviewModel "github.com/Yogdunana/StarByte/backend/internal/interview/model"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubFeed struct {
	source string
	id     uuid.UUID
	events []*dto.EventResponse
}

func (f *stubFeed) Source() string        { return f.source }
func (f *stubFeed) CalendarID() uuid.UUID { return f.id }
func (f *stubFeed) Calendar(viewer uuid.UUID) *dto.CalendarResponse {
	return virtualCalendar(f.id, f.source, layerTitle(f.source), "", viewer)
}
func (f *stubFeed) Events(_ context.Context, _ uuid.UUID, start, end time.Time) ([]*dto.EventResponse, error) {
	out := make([]*dto.EventResponse, 0, len(f.events))
	for _, ev := range f.events {
		if overlapsWindow(ev.StartAt, ev.EndAt, start, end) {
			cp := *ev
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (f *stubFeed) Event(_ context.Context, _ uuid.UUID, id uuid.UUID) (*dto.EventResponse, error) {
	for _, ev := range f.events {
		if ev.ID == id.String() {
			cp := *ev
			return &cp, nil
		}
	}
	return nil, nil
}

func setupSvcWithFeeds(t *testing.T, feeds ...LayerFeed) (*scheduleService, uuid.UUID) {
	t.Helper()
	mem := newMem()
	owner := uuid.New()
	dept := uuid.New()
	mem.addUser(owner, "张三", &dept)
	svc := New(mem, nil, feeds...).(*scheduleService)
	return svc, owner
}

func TestVirtualLayersMergeAndStayReadOnly(t *testing.T) {
	actID := uuid.New()
	ivID := uuid.New()
	start := time.Date(2026, 9, 11, 14, 0, 0, 0, time.UTC)
	activity := projectedEvent(actID, ActivityLayerID, model.SourceActivity, "迎新晚会", "", "礼堂", "/activity/"+actID.String(), start, start.Add(2*time.Hour), dto.Person{}, start, start)
	interview := projectedEvent(ivID, InterviewLayerID, model.SourceInterview, "一面 · 李四", "", "B201", "/interview/my", start.Add(3*time.Hour), start.Add(3*time.Hour+30*time.Minute), dto.Person{}, start, start)
	svc, owner := setupSvcWithFeeds(t,
		&stubFeed{source: model.SourceActivity, id: ActivityLayerID, events: []*dto.EventResponse{activity}},
		&stubFeed{source: model.SourceInterview, id: InterviewLayerID, events: []*dto.EventResponse{interview}},
	)
	ctx := context.Background()
	scope := selfScope(owner)

	cals, total, _, _, err := svc.ListCalendars(ctx, owner, &dto.ListCalendarRequest{}, scope)
	require.NoError(t, err)
	require.GreaterOrEqual(t, total, int64(3))
	sources := map[string]bool{}
	for _, c := range cals {
		sources[c.Source] = true
		if c.Source == model.SourceActivity || c.Source == model.SourceInterview {
			assert.False(t, c.CanEdit)
			assert.Equal(t, model.DefaultLayerColor(c.Source), c.Color)
		}
	}
	assert.True(t, sources[model.SourcePersonal])
	assert.True(t, sources[model.SourceActivity])
	assert.True(t, sources[model.SourceInterview])

	got, err := svc.GetCalendar(ctx, owner, ActivityLayerID, scope)
	require.NoError(t, err)
	assert.Equal(t, model.SourceActivity, got.Source)
	assert.False(t, got.CanEdit)

	_, err = svc.UpdateCalendar(ctx, owner, ActivityLayerID, &dto.UpdateCalendarRequest{}, scope)
	require.Error(t, err)
	assert.Equal(t, response.CodeScheduleInvalidState, err.(*response.AppError).Code)

	_, err = svc.CreateEvent(ctx, owner, &dto.CreateEventRequest{
		CalendarID: ActivityLayerID.String(), Title: "写入活动层",
		StartAt: start, EndAt: start.Add(time.Hour),
	}, scope)
	require.Error(t, err)
	assert.Equal(t, response.CodeScheduleInvalidState, err.(*response.AppError).Code)

	window := &dto.RangeEventRequest{Start: start.Add(-time.Hour), End: start.Add(24 * time.Hour)}
	rows, err := svc.RangeEvents(ctx, owner, window, scope)
	require.NoError(t, err)
	var sawActivity, sawInterview bool
	for _, ev := range rows {
		if ev.ID == actID.String() {
			sawActivity = true
			assert.Equal(t, "/activity/"+actID.String(), ev.Link)
			assert.False(t, ev.CanEdit)
			assert.Equal(t, model.SourceActivity, ev.Source)
		}
		if ev.ID == ivID.String() {
			sawInterview = true
			assert.Equal(t, "/interview/my", ev.Link)
			assert.Equal(t, model.SourceInterview, ev.Source)
		}
	}
	assert.True(t, sawActivity)
	assert.True(t, sawInterview)

	onlyActivity, err := svc.RangeEvents(ctx, owner, &dto.RangeEventRequest{
		Start: window.Start, End: window.End, CalendarID: ActivityLayerID.String(),
	}, scope)
	require.NoError(t, err)
	require.Len(t, onlyActivity, 1)
	assert.Equal(t, actID.String(), onlyActivity[0].ID)

	loaded, err := svc.GetEvent(ctx, owner, actID, scope)
	require.NoError(t, err)
	assert.Equal(t, "迎新晚会", loaded.Title)

	err = svc.DeleteEvent(ctx, owner, actID, scope)
	require.Error(t, err)
	assert.Equal(t, response.CodeScheduleInvalidState, err.(*response.AppError).Code)
}

func TestRangeEventsStoredCalendarExcludesFeeds(t *testing.T) {
	start := time.Date(2026, 9, 11, 10, 0, 0, 0, time.UTC)
	feedEv := projectedEvent(uuid.New(), ActivityLayerID, model.SourceActivity, "不该出现", "", "", "/activity/x", start, start.Add(time.Hour), dto.Person{}, start, start)
	svc, owner := setupSvcWithFeeds(t, &stubFeed{source: model.SourceActivity, id: ActivityLayerID, events: []*dto.EventResponse{feedEv}})
	ctx := context.Background()
	scope := selfScope(owner)
	personal, err := svc.CreateEvent(ctx, owner, &dto.CreateEventRequest{
		Title: "个人事", StartAt: start, EndAt: start.Add(time.Hour),
	}, scope)
	require.NoError(t, err)

	rows, err := svc.RangeEvents(ctx, owner, &dto.RangeEventRequest{
		Start: start.Add(-time.Minute), End: start.Add(2 * time.Hour), CalendarID: personal.CalendarID,
	}, scope)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, personal.ID, rows[0].ID)
}

func TestVirtualLayersPaginateWithStored(t *testing.T) {
	svc, owner := setupSvcWithFeeds(t,
		&stubFeed{source: model.SourceActivity, id: ActivityLayerID},
		&stubFeed{source: model.SourceInterview, id: InterviewLayerID},
	)
	ctx := context.Background()
	scope := selfScope(owner)
	for i := 0; i < 3; i++ {
		_, err := svc.CreateCalendar(ctx, owner, &dto.CreateCalendarRequest{
			Name: "共享" + string(rune('A'+i)), CalendarType: model.CalendarProject,
		}, scope)
		require.NoError(t, err)
	}
	page1, total, _, size, err := svc.ListCalendars(ctx, owner, &dto.ListCalendarRequest{Page: 1, PageSize: 2}, scope)
	require.NoError(t, err)
	assert.Equal(t, 2, size)
	assert.GreaterOrEqual(t, total, int64(6))
	require.Len(t, page1, 2)
	assert.Equal(t, model.SourceActivity, page1[0].Source)
	assert.Equal(t, model.SourceInterview, page1[1].Source)

	page2, _, _, _, err := svc.ListCalendars(ctx, owner, &dto.ListCalendarRequest{Page: 2, PageSize: 2}, scope)
	require.NoError(t, err)
	require.NotEmpty(t, page2)
	for _, c := range page2 {
		assert.NotEqual(t, model.SourceActivity, c.Source)
		assert.NotEqual(t, model.SourceInterview, c.Source)
	}
}

func TestMapActivityAndInterviewProjection(t *testing.T) {
	act := uuid.New()
	org := uuid.New()
	start := time.Date(2026, 9, 20, 9, 0, 0, 0, time.UTC)
	ev := mapActivityEvent(&activityProj{
		ID: act, Title: "讲座", Description: "简介", Location: "A1",
		StartTime: start, EndTime: start.Add(90 * time.Minute),
		Status: activityModel.ActivityOpen, OrganizerID: org, OrganizerName: "组织者",
		CreatedAt: start, UpdatedAt: start,
	})
	require.NotNil(t, ev)
	assert.Equal(t, "/activity/"+act.String(), ev.Link)
	assert.False(t, ev.CanEdit)
	assert.Equal(t, model.SourceActivity, ev.Source)
	assert.Equal(t, model.DefaultLayerColor(model.SourceActivity), ev.Color)

	when := start.Add(2 * time.Hour)
	iv := mapInterviewEvent(&interviewProj{
		ID: uuid.New(), ApplicantID: uuid.New(), ApplicantName: "王五",
		SessionTitle: "二面", Location: "线上面试", ScheduledAt: &when, Duration: 0,
		CreatedAt: when, UpdatedAt: when,
	})
	require.NotNil(t, iv)
	assert.Equal(t, "二面 · 王五", iv.Title)
	assert.Equal(t, when.Add(time.Duration(interviewModel.DefaultDuration)*time.Minute), iv.EndAt)
	assert.Equal(t, "/interview/my", iv.Link)
	assert.Equal(t, model.SourceInterview, iv.Source)

	assert.Nil(t, mapActivityEvent(&activityProj{StartTime: start, EndTime: start.Add(-time.Minute)}))
	assert.Nil(t, mapInterviewEvent(&interviewProj{ApplicantName: "无时间"}))
}
