package service

import (
	"context"
	"fmt"
	"time"

	activityModel "github.com/Yogdunana/StarByte/backend/internal/activity/model"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type activityFeed struct{ db *gorm.DB }

func NewActivityFeed(db *gorm.DB) LayerFeed {
	return &activityFeed{db: db}
}

func (f *activityFeed) Source() string        { return model.SourceActivity }
func (f *activityFeed) CalendarID() uuid.UUID { return ActivityLayerID }
func (f *activityFeed) Calendar(viewer uuid.UUID) *dto.CalendarResponse {
	return virtualCalendar(ActivityLayerID, model.SourceActivity, "活动", "已发布/进行中的活动，以及你报名的活动（只读投影）", viewer)
}

type activityProj struct {
	ID            uuid.UUID
	Title         string
	Description   string
	Location      string
	StartTime     time.Time
	EndTime       time.Time
	Status        int16
	OrganizerID   uuid.UUID
	OrganizerName string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (f *activityFeed) Events(ctx context.Context, viewer uuid.UUID, start, end time.Time) ([]*dto.EventResponse, error) {
	rows, err := f.query(ctx, viewer, &start, &end, uuid.Nil)
	if err != nil {
		return nil, err
	}
	out := make([]*dto.EventResponse, 0, len(rows))
	for i := range rows {
		if ev := mapActivityEvent(&rows[i]); ev != nil {
			out = append(out, ev)
		}
	}
	return out, nil
}

func (f *activityFeed) Event(ctx context.Context, viewer, id uuid.UUID) (*dto.EventResponse, error) {
	rows, err := f.query(ctx, viewer, nil, nil, id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return mapActivityEvent(&rows[0]), nil
}

func (f *activityFeed) query(ctx context.Context, viewer uuid.UUID, start, end *time.Time, id uuid.UUID) ([]activityProj, error) {
	if f.db == nil {
		return nil, nil
	}
	q := f.db.WithContext(ctx).Table("activities AS a").
		Select(`a.id, a.title, a.description, a.location, a.start_time, a.end_time,
			a.status, a.organizer_id, a.created_at, a.updated_at,
			COALESCE(u.real_name, u.username, '') AS organizer_name`).
		Joins("LEFT JOIN users u ON u.id = a.organizer_id").
		Where("a.deleted_at IS NULL").
		Where("a.status <> ?", activityModel.ActivityCancelled).
		Where(`
			a.status IN ? OR a.organizer_id = ? OR EXISTS (
				SELECT 1 FROM activity_registrations r
				WHERE r.activity_id = a.id AND r.user_id = ? AND r.status IN ?
			)`,
			[]int16{activityModel.ActivityOpen, activityModel.ActivityOngoing, activityModel.ActivityEnded},
			viewer, viewer,
			[]int16{activityModel.RegPending, activityModel.RegApproved, activityModel.RegWaitlist},
		)
	if id != uuid.Nil {
		q = q.Where("a.id = ?", id)
	}
	if start != nil && end != nil {
		q = q.Where("a.start_time < ? AND a.end_time > ?", *end, *start)
	}
	var rows []activityProj
	if err := q.Order("a.start_time").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list activity layer: %w", err)
	}
	return rows, nil
}

func mapActivityEvent(row *activityProj) *dto.EventResponse {
	if row == nil || row.EndTime.Before(row.StartTime) {
		return nil
	}
	return projectedEvent(
		row.ID, ActivityLayerID, model.SourceActivity,
		row.Title, row.Description, row.Location,
		"/activity/"+row.ID.String(),
		row.StartTime, row.EndTime,
		dto.Person{ID: row.OrganizerID.String(), Name: row.OrganizerName},
		row.CreatedAt, row.UpdatedAt,
	)
}
