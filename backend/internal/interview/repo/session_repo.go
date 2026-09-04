package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/internal/interview/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionRepo interface {
	Create(ctx context.Context, s *model.Session) error
	Update(ctx context.Context, s *model.Session) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Session, error)
	GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.SessionWithNames, error)
	List(ctx context.Context, req *dto.ListSessionRequest, scope *rbacModel.DataScopeCondition) ([]model.SessionWithNames, int64, error)
	CountCandidates(ctx context.Context, sessionID uuid.UUID) (int64, error)
}

type sessionRepo struct{ db *gorm.DB }

func NewSessionRepo(db *gorm.DB) SessionRepo {
	return &sessionRepo{db: db}
}

func (r *sessionRepo) Create(ctx context.Context, s *model.Session) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *sessionRepo) Update(ctx context.Context, s *model.Session) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *sessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Session{}, "id = ?", id).Error
}

func (r *sessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Session, error) {
	var s model.Session
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&s).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *sessionRepo) namedQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table("interview_sessions AS s").
		Select(`s.*, COALESCE(d.name, '') AS department_name,
			(SELECT COUNT(*) FROM interviews i WHERE i.session_id = s.id AND i.status <> ?) AS candidate_count`,
			model.InterviewCancelled).
		Joins("LEFT JOIN departments d ON d.id = s.department_id")
}

func (r *sessionRepo) GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.SessionWithNames, error) {
	var row model.SessionWithNames
	err := r.namedQuery(ctx).Where("s.id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *sessionRepo) List(ctx context.Context, req *dto.ListSessionRequest, scope *rbacModel.DataScopeCondition) ([]model.SessionWithNames, int64, error) {
	q := r.namedQuery(ctx)
	q = applyScope(q, scope)
	if req.Status != nil {
		q = q.Where("s.status = ?", *req.Status)
	}
	if req.Round != nil {
		q = q.Where("s.round = ?", *req.Round)
	}
	if req.DepartmentID != "" {
		q = q.Where("s.department_id = ?", req.DepartmentID)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		q = q.Where("s.title ILIKE ? OR s.location ILIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(req.Page, req.PageSize)
	var rows []model.SessionWithNames
	err := q.Order("s.start_time DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}

func (r *sessionRepo) CountCandidates(ctx context.Context, sessionID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Interview{}).
		Where("session_id = ? AND status <> ?", sessionID, model.InterviewCancelled).
		Count(&n).Error
	return n, err
}
