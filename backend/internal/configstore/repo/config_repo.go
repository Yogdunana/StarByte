package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/configstore/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConfigRepo interface {
	Create(ctx context.Context, row *model.Config) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Config, error)
	GetByKey(ctx context.Context, key string) (*model.Config, error)
	List(ctx context.Context, category, keyword string) ([]model.Config, error)
	Update(ctx context.Context, row *model.Config) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type configRepo struct {
	db *gorm.DB
}

func NewConfigRepo(db *gorm.DB) ConfigRepo {
	return &configRepo{db: db}
}

func (r *configRepo) Create(ctx context.Context, row *model.Config) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *configRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Config, error) {
	var row model.Config
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *configRepo) GetByKey(ctx context.Context, key string) (*model.Config, error) {
	var row model.Config
	err := r.db.WithContext(ctx).Where("config_key = ?", key).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *configRepo) List(ctx context.Context, category, keyword string) ([]model.Config, error) {
	q := r.db.WithContext(ctx).Model(&model.Config{})
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("config_key ILIKE ? OR description ILIKE ?", like, like)
	}
	var rows []model.Config
	err := q.Order("category ASC, config_key ASC").Find(&rows).Error
	return rows, err
}

func (r *configRepo) Update(ctx context.Context, row *model.Config) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *configRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Config{}, id).Error
}
