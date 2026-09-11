package service

import (
	"context"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/model"
	"github.com/google/uuid"
)

// 合成图层 ID：稳定、可切换，且不写入 calendars / schedule_events。
var (
	ActivityLayerID  = uuid.MustParse("aaaaaaaa-0000-4000-8000-0000000000a1")
	InterviewLayerID = uuid.MustParse("bbbbbbbb-0000-4000-8000-0000000000b2")
)

// LayerFeed 把其它模块的时间窗投影成只读日历层，在 ListCalendars / RangeEvents 合并。
type LayerFeed interface {
	Source() string
	CalendarID() uuid.UUID
	Calendar(viewer uuid.UUID) *dto.CalendarResponse
	Events(ctx context.Context, viewer uuid.UUID, start, end time.Time) ([]*dto.EventResponse, error)
	Event(ctx context.Context, viewer, id uuid.UUID) (*dto.EventResponse, error)
}

func virtualCalendar(id uuid.UUID, source, name, desc string, viewer uuid.UUID) *dto.CalendarResponse {
	now := time.Unix(0, 0).UTC()
	return &dto.CalendarResponse{
		ID:           id.String(),
		Name:         name,
		Description:  desc,
		CalendarType: model.CalendarPersonal,
		Source:       source,
		Color:        model.DefaultLayerColor(source),
		Owner:        dto.Person{ID: viewer.String()},
		MemberRole:   model.MemberViewer,
		CanEdit:      false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func projectedEvent(id, calendarID uuid.UUID, source, title, desc, location, link string, start, end time.Time, creator dto.Person, created, updated time.Time) *dto.EventResponse {
	color := model.DefaultLayerColor(source)
	return &dto.EventResponse{
		ID:            id.String(),
		CalendarID:    calendarID.String(),
		CalendarName:  layerTitle(source),
		CalendarColor: color,
		Title:         title,
		Description:   desc,
		Location:      location,
		StartAt:       start,
		EndAt:         end,
		Color:         color,
		Status:        model.EventConfirmed,
		Recurrence:    model.RecurrenceNone,
		Creator:       creator,
		CanEdit:       false,
		Source:        source,
		Link:          link,
		CreatedAt:     created,
		UpdatedAt:     updated,
	}
}

func layerTitle(source string) string {
	switch source {
	case model.SourceActivity:
		return "活动"
	case model.SourceInterview:
		return "面试"
	default:
		return source
	}
}

func (s *scheduleService) feedByCalendar(id uuid.UUID) LayerFeed {
	for _, f := range s.feeds {
		if f != nil && f.CalendarID() == id {
			return f
		}
	}
	return nil
}

func (s *scheduleService) virtualCalendars(viewer uuid.UUID, req *dto.ListCalendarRequest) []*dto.CalendarResponse {
	out := make([]*dto.CalendarResponse, 0, len(s.feeds))
	kw := ""
	if req != nil {
		kw = strings.ToLower(strings.TrimSpace(req.Keyword))
	}
	for _, f := range s.feeds {
		if f == nil {
			continue
		}
		cal := f.Calendar(viewer)
		if cal == nil {
			continue
		}
		if req != nil && req.CalendarType != nil && *req.CalendarType != cal.CalendarType {
			continue
		}
		if kw != "" && !strings.Contains(strings.ToLower(cal.Name), kw) {
			continue
		}
		out = append(out, cal)
	}
	return out
}

func (s *scheduleService) collectFeedEvents(ctx context.Context, viewer uuid.UUID, start, end time.Time) ([]*dto.EventResponse, error) {
	out := make([]*dto.EventResponse, 0)
	for _, f := range s.feeds {
		if f == nil {
			continue
		}
		rows, err := f.Events(ctx, viewer, start, end)
		if err != nil {
			return nil, err
		}
		out = append(out, rows...)
	}
	return out, nil
}

func (s *scheduleService) feedEvent(ctx context.Context, viewer, id uuid.UUID) (*dto.EventResponse, error) {
	for _, f := range s.feeds {
		if f == nil {
			continue
		}
		row, err := f.Event(ctx, viewer, id)
		if err != nil {
			return nil, err
		}
		if row != nil {
			return row, nil
		}
	}
	return nil, nil
}

func overlapsWindow(start, end, windowStart, windowEnd time.Time) bool {
	return start.Before(windowEnd) && end.After(windowStart)
}
