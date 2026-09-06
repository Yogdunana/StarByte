package repo

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/notification/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmailLogRepo interface {
	Create(ctx context.Context, row *model.EmailLog) error
	Update(ctx context.Context, row *model.EmailLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.EmailLog, error)
	List(ctx context.Context, status int16, start, end *time.Time, page, pageSize int) ([]*model.EmailLog, int64, error)
}

type emailLogRepo struct{ db *gorm.DB }

func NewEmailLogRepo(db *gorm.DB) EmailLogRepo {
	return &emailLogRepo{db: db}
}

func (r *emailLogRepo) Create(ctx context.Context, row *model.EmailLog) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *emailLogRepo) Update(ctx context.Context, row *model.EmailLog) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *emailLogRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.EmailLog, error) {
	var row model.EmailLog
	if err := r.db.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *emailLogRepo) List(ctx context.Context, status int16, start, end *time.Time, page, pageSize int) ([]*model.EmailLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	q := r.db.WithContext(ctx).Model(&model.EmailLog{})
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	if start != nil {
		q = q.Where("created_at >= ?", *start)
	}
	if end != nil {
		q = q.Where("created_at <= ?", *end)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.EmailLog
	err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
