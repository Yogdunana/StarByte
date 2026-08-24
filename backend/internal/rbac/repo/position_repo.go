package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PositionRepo 职位数据访问接口
type PositionRepo interface {
	Create(ctx context.Context, tx *gorm.DB, pos *model.Position) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Position, error)
	GetByCode(ctx context.Context, code string) (*model.Position, error)
	List(ctx context.Context, page, pageSize int, keyword string) ([]model.Position, int64, error)
	Update(ctx context.Context, tx *gorm.DB, pos *model.Position) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type positionRepo struct {
	db *gorm.DB
}

// NewPositionRepo 创建职位 Repo
func NewPositionRepo(db *gorm.DB) PositionRepo {
	return &positionRepo{db: db}
}

func (r *positionRepo) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *positionRepo) Create(ctx context.Context, tx *gorm.DB, pos *model.Position) error {
	return r.getDB(tx).WithContext(ctx).Create(pos).Error
}

func (r *positionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Position, error) {
	var pos model.Position
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&pos).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &pos, err
}

func (r *positionRepo) GetByCode(ctx context.Context, code string) (*model.Position, error) {
	var pos model.Position
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&pos).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &pos, err
}

func (r *positionRepo) List(ctx context.Context, page, pageSize int, keyword string) ([]model.Position, int64, error) {
	var positions []model.Position
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Position{})

	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.
		Order("sort_order ASC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&positions).Error

	return positions, total, err
}

func (r *positionRepo) Update(ctx context.Context, tx *gorm.DB, pos *model.Position) error {
	return r.getDB(tx).WithContext(ctx).Save(pos).Error
}

func (r *positionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Position{}, id).Error
}
