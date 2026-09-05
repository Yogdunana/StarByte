package repo

import (
	"context"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MeetingRepo interface {
	Create(ctx context.Context, m *model.Meeting) error
	Update(ctx context.Context, m *model.Meeting) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Meeting, error)
	GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.MeetingWithNames, error)
	List(ctx context.Context, req *dto.ListMeetingRequest, scope *rbacModel.DataScopeCondition) ([]model.MeetingWithNames, int64, error)
	GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error)
}

type meetingRepo struct{ db *gorm.DB }

func NewMeetingRepo(db *gorm.DB) MeetingRepo {
	return &meetingRepo{db: db}
}

func (r *meetingRepo) Create(ctx context.Context, m *model.Meeting) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *meetingRepo) Update(ctx context.Context, m *model.Meeting) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *meetingRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Meeting{}, "id = ?", id).Error
}

func (r *meetingRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Meeting, error) {
	var m model.Meeting
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *meetingRepo) namedQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table("meetings AS m").
		Select(`m.*, COALESCE(u.real_name, u.username, '') AS organizer_name,
			(SELECT COUNT(*) FROM meeting_attendees a WHERE a.meeting_id = m.id) AS attendee_count,
			(SELECT COUNT(*) FROM meeting_attendees a WHERE a.meeting_id = m.id AND a.attended = true) AS checked_in_count`).
		Joins("LEFT JOIN users u ON u.id = m.organizer_id")
}

func (r *meetingRepo) GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.MeetingWithNames, error) {
	var row model.MeetingWithNames
	err := r.namedQuery(ctx).Where("m.id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *meetingRepo) List(ctx context.Context, req *dto.ListMeetingRequest, scope *rbacModel.DataScopeCondition) ([]model.MeetingWithNames, int64, error) {
	q := r.namedQuery(ctx)
	q = applyScope(q, scope)
	if req.Status != nil {
		q = q.Where("m.status = ?", *req.Status)
	}
	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("m.title ILIKE ? OR m.location ILIKE ?", like, like)
	}
	if req.StartDate != "" {
		q = q.Where("m.start_time >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		q = q.Where("m.start_time <= ?", req.EndDate+" 23:59:59")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(req.Page, req.PageSize)
	var rows []model.MeetingWithNames
	err := q.Order("m.start_time DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}

func (r *meetingRepo) GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error) {
	var u model.NamedUser
	err := r.db.WithContext(ctx).Table("users AS u").
		Select("u.id, u.real_name, u.username, u.position_id, COALESCE(p.code, '') AS position_code").
		Joins("LEFT JOIN positions p ON p.id = u.position_id").
		Where("u.id = ?", id).First(&u).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
