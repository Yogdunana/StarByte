package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AttendeeRepo interface {
	Add(ctx context.Context, items []model.Attendee) error
	Remove(ctx context.Context, meetingID, userID uuid.UUID) error
	Get(ctx context.Context, meetingID, userID uuid.UUID) (*model.Attendee, error)
	Update(ctx context.Context, a *model.Attendee) error
	List(ctx context.Context, meetingID uuid.UUID) ([]model.AttendeeNamed, error)
	IsAttendee(ctx context.Context, meetingID, userID uuid.UUID) (bool, error)
}

type attendeeRepo struct{ db *gorm.DB }

func NewAttendeeRepo(db *gorm.DB) AttendeeRepo {
	return &attendeeRepo{db: db}
}

func (r *attendeeRepo) Add(ctx context.Context, items []model.Attendee) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&items).Error
}

func (r *attendeeRepo) Remove(ctx context.Context, meetingID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("meeting_id = ? AND user_id = ?", meetingID, userID).
		Delete(&model.Attendee{}).Error
}

func (r *attendeeRepo) Get(ctx context.Context, meetingID, userID uuid.UUID) (*model.Attendee, error) {
	var a model.Attendee
	err := r.db.WithContext(ctx).Where("meeting_id = ? AND user_id = ?", meetingID, userID).First(&a).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *attendeeRepo) Update(ctx context.Context, a *model.Attendee) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *attendeeRepo) List(ctx context.Context, meetingID uuid.UUID) ([]model.AttendeeNamed, error) {
	var rows []model.AttendeeNamed
	err := r.db.WithContext(ctx).Table("meeting_attendees AS a").
		Select("a.*, COALESCE(u.real_name, u.username, '') AS real_name, COALESCE(u.username, '') AS username, COALESCE(p.code, '') AS position_code").
		Joins("LEFT JOIN users u ON u.id = a.user_id").
		Joins("LEFT JOIN positions p ON p.id = u.position_id").
		Where("a.meeting_id = ?", meetingID).
		Order("a.created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *attendeeRepo) IsAttendee(ctx context.Context, meetingID, userID uuid.UUID) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Attendee{}).
		Where("meeting_id = ? AND user_id = ?", meetingID, userID).Count(&n).Error
	return n > 0, err
}
