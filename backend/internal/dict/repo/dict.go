package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/dict/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DictRepository interface {
	ListTypes(ctx context.Context) ([]model.DictType, error)
	GetTypeByID(ctx context.Context, id uuid.UUID) (*model.DictType, error)
	GetTypeByCode(ctx context.Context, code string) (*model.DictType, error)
	CreateType(ctx context.Context, rec *model.DictType) error
	UpdateType(ctx context.Context, rec *model.DictType) error
	DeleteType(ctx context.Context, id uuid.UUID) error

	ListItemsByTypeID(ctx context.Context, typeID uuid.UUID, enabledOnly bool) ([]model.DictItem, error)
	GetItemByID(ctx context.Context, id uuid.UUID) (*model.DictItem, error)
	CreateItem(ctx context.Context, rec *model.DictItem) error
	UpdateItem(ctx context.Context, rec *model.DictItem) error
	DeleteItem(ctx context.Context, id uuid.UUID) error
}

type dictRepository struct {
	db *gorm.DB
}

func NewDictRepository(db *gorm.DB) DictRepository {
	return &dictRepository{db: db}
}

func (r *dictRepository) ListTypes(ctx context.Context) ([]model.DictType, error) {
	var list []model.DictType
	err := r.db.WithContext(ctx).Order("sort_order ASC, created_at ASC").Find(&list).Error
	return list, err
}

func (r *dictRepository) GetTypeByID(ctx context.Context, id uuid.UUID) (*model.DictType, error) {
	var rec model.DictType
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *dictRepository) GetTypeByCode(ctx context.Context, code string) (*model.DictType, error) {
	var rec model.DictType
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *dictRepository) CreateType(ctx context.Context, rec *model.DictType) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *dictRepository) UpdateType(ctx context.Context, rec *model.DictType) error {
	return r.db.WithContext(ctx).Save(rec).Error
}

func (r *dictRepository) DeleteType(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.DictType{}, "id = ?", id).Error
}

func (r *dictRepository) ListItemsByTypeID(ctx context.Context, typeID uuid.UUID, enabledOnly bool) ([]model.DictItem, error) {
	q := r.db.WithContext(ctx).Where("type_id = ?", typeID)
	if enabledOnly {
		q = q.Where("status = ?", model.StatusEnabled)
	}
	var list []model.DictItem
	err := q.Order("sort_order ASC, created_at ASC").Find(&list).Error
	return list, err
}

func (r *dictRepository) GetItemByID(ctx context.Context, id uuid.UUID) (*model.DictItem, error) {
	var rec model.DictItem
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *dictRepository) CreateItem(ctx context.Context, rec *model.DictItem) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *dictRepository) UpdateItem(ctx context.Context, rec *model.DictItem) error {
	return r.db.WithContext(ctx).Save(rec).Error
}

func (r *dictRepository) DeleteItem(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.DictItem{}, "id = ?", id).Error
}
