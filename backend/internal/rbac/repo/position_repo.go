package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PositionRepo 职位数据访问接口
// 定义职位相关的数据库操作，支持事务传入以保证复杂操作的原子性。
type PositionRepo interface {
	// Create 创建职位记录
	Create(ctx context.Context, tx *gorm.DB, pos *model.Position) error
	// GetByID 根据 ID 查询职位，未找到返回 nil, nil
	GetByID(ctx context.Context, id uuid.UUID) (*model.Position, error)
	// GetByCode 根据编码查询职位，未找到返回 nil, nil
	GetByCode(ctx context.Context, code string) (*model.Position, error)
	// List 分页查询职位列表，支持关键字模糊搜索（名称和编码）
	List(ctx context.Context, page, pageSize int, keyword string) ([]model.Position, int64, error)
	// Update 更新职位记录（全量保存）
	Update(ctx context.Context, tx *gorm.DB, pos *model.Position) error
	// Delete 根据 ID 删除职位
	Delete(ctx context.Context, tx *gorm.DB, id uuid.UUID) error
	// CountUsersByPositionID 统计指定职位下的有效用户数（排除已删除用户）
	CountUsersByPositionID(ctx context.Context, tx *gorm.DB, positionID uuid.UUID) (int64, error)
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

	// 防御性校验：限制 pageSize 最大值（调用方应已处理默认值）
	if pageSize > 100 {
		pageSize = 100
	}

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

func (r *positionRepo) Delete(ctx context.Context, tx *gorm.DB, id uuid.UUID) error {
	return r.getDB(tx).WithContext(ctx).Delete(&model.Position{}, id).Error
}

// CountUsersByPositionID 统计指定职位下的有效用户数
// 使用原生 SQL 查询 users 表，避免引入 user 模块依赖
func (r *positionRepo) CountUsersByPositionID(ctx context.Context, tx *gorm.DB, positionID uuid.UUID) (int64, error) {
	var count int64
	err := r.getDB(tx).WithContext(ctx).
		Table("users").
		Where("position_id = ? AND deleted_at IS NULL", positionID).
		Count(&count).Error
	return count, err
}
