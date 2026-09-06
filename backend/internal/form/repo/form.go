package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/form/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, rec *model.Form) error
	Update(ctx context.Context, rec *model.Form) error
	Get(ctx context.Context, id uuid.UUID) (*model.Form, error)
	GetByName(ctx context.Context, name string) (*model.Form, error)
	List(ctx context.Context, keyword string, status *int16, offset, limit int) ([]model.Form, int64, error)
	CountSubmissions(ctx context.Context, formID uuid.UUID) (int64, error)
	CreateSubmission(ctx context.Context, rec *model.Submission) error
	ListSubmissions(ctx context.Context, formID uuid.UUID, offset, limit int) ([]model.Submission, int64, error)
	UserDisplayNames(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error)
}

type repository struct{ db *gorm.DB }

func New(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, rec *model.Form) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *repository) Update(ctx context.Context, rec *model.Form) error {
	return r.db.WithContext(ctx).Save(rec).Error
}

func (r *repository) Get(ctx context.Context, id uuid.UUID) (*model.Form, error) {
	var rec model.Form
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *repository) GetByName(ctx context.Context, name string) (*model.Form, error) {
	var rec model.Form
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *repository) List(ctx context.Context, keyword string, status *int16, offset, limit int) ([]model.Form, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Form{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("(name ILIKE ? OR description ILIKE ?)", like, like)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Form
	err := q.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *repository) CountSubmissions(ctx context.Context, formID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Submission{}).Where("form_id = ?", formID).Count(&n).Error
	return n, err
}

func (r *repository) CreateSubmission(ctx context.Context, rec *model.Submission) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *repository) ListSubmissions(ctx context.Context, formID uuid.UUID, offset, limit int) ([]model.Submission, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Submission{}).Where("form_id = ?", formID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Submission
	err := q.Order("submitted_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *repository) UserDisplayNames(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	out := make(map[uuid.UUID]string, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	type row struct {
		ID   uuid.UUID
		Name string
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("users").
		Select("id, COALESCE(NULLIF(real_name, ''), username) AS name").
		Where("id IN ? AND deleted_at IS NULL", ids).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, rec := range rows {
		out[rec.ID] = rec.Name
	}
	return out, nil
}
