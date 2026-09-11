package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	interviewModel "github.com/Yogdunana/StarByte/backend/internal/interview/model"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type interviewFeed struct{ db *gorm.DB }

func NewInterviewFeed(db *gorm.DB) LayerFeed {
	return &interviewFeed{db: db}
}

func (f *interviewFeed) Source() string        { return model.SourceInterview }
func (f *interviewFeed) CalendarID() uuid.UUID { return InterviewLayerID }
func (f *interviewFeed) Calendar(viewer uuid.UUID) *dto.CalendarResponse {
	return virtualCalendar(InterviewLayerID, model.SourceInterview, "面试", "你作为面试者、面试官或场次创建人的安排（只读投影）", viewer)
}

type interviewProj struct {
	ID            uuid.UUID
	ApplicantID   uuid.UUID
	ApplicantName string
	SessionTitle  string
	Location      string
	ScheduledAt   *time.Time
	Duration      int
	SessionStart  *time.Time
	SessionEnd    *time.Time
	Status        int16
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (f *interviewFeed) Events(ctx context.Context, viewer uuid.UUID, start, end time.Time) ([]*dto.EventResponse, error) {
	rows, err := f.query(ctx, viewer, uuid.Nil)
	if err != nil {
		return nil, err
	}
	out := make([]*dto.EventResponse, 0, len(rows))
	for i := range rows {
		ev := mapInterviewEvent(&rows[i])
		if ev == nil || !overlapsWindow(ev.StartAt, ev.EndAt, start, end) {
			continue
		}
		out = append(out, ev)
	}
	return out, nil
}

func (f *interviewFeed) Event(ctx context.Context, viewer, id uuid.UUID) (*dto.EventResponse, error) {
	rows, err := f.query(ctx, viewer, id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return mapInterviewEvent(&rows[0]), nil
}

func (f *interviewFeed) query(ctx context.Context, viewer, id uuid.UUID) ([]interviewProj, error) {
	if f.db == nil {
		return nil, nil
	}
	q := f.db.WithContext(ctx).Table("interviews AS i").
		Select(`i.id, i.applicant_id, i.location, i.scheduled_at, i.duration, i.status,
			i.created_at, i.updated_at,
			COALESCE(u.real_name, u.username, '') AS applicant_name,
			COALESCE(s.title, '') AS session_title,
			s.start_time AS session_start, s.end_time AS session_end`).
		Joins("LEFT JOIN interview_sessions s ON s.id = i.session_id").
		Joins("LEFT JOIN users u ON u.id = i.applicant_id").
		Where("i.status <> ?", interviewModel.InterviewCancelled).
		Where(`
			i.applicant_id = ?
			OR EXISTS (SELECT 1 FROM interview_interviewers ev WHERE ev.interview_id = i.id AND ev.interviewer_id = ?)
			OR s.created_by = ?`,
			viewer, viewer, viewer,
		)
	if id != uuid.Nil {
		q = q.Where("i.id = ?", id)
	}
	var rows []interviewProj
	if err := q.Order("COALESCE(i.scheduled_at, s.start_time)").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list interview layer: %w", err)
	}
	return rows, nil
}

func mapInterviewEvent(row *interviewProj) *dto.EventResponse {
	if row == nil {
		return nil
	}
	start, end, ok := interviewWindow(row)
	if !ok {
		return nil
	}
	title := strings.TrimSpace(row.SessionTitle)
	if title == "" {
		title = "面试"
	}
	if name := strings.TrimSpace(row.ApplicantName); name != "" {
		title = title + " · " + name
	}
	return projectedEvent(
		row.ID, InterviewLayerID, model.SourceInterview,
		title, "", row.Location,
		"/interview/my",
		start, end,
		dto.Person{ID: row.ApplicantID.String(), Name: row.ApplicantName},
		row.CreatedAt, row.UpdatedAt,
	)
}

func interviewWindow(row *interviewProj) (time.Time, time.Time, bool) {
	if row.ScheduledAt != nil && !row.ScheduledAt.IsZero() {
		d := row.Duration
		if d <= 0 {
			d = interviewModel.DefaultDuration
		}
		start := *row.ScheduledAt
		return start, start.Add(time.Duration(d) * time.Minute), true
	}
	if row.SessionStart != nil && row.SessionEnd != nil && !row.SessionStart.IsZero() && !row.SessionEnd.Before(*row.SessionStart) {
		return *row.SessionStart, *row.SessionEnd, true
	}
	return time.Time{}, time.Time{}, false
}
