package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DepartmentRepo 部门数据访问接口
type DepartmentRepo interface {
	Create(ctx context.Context, tx *gorm.DB, dept *model.Department) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Department, error)
	GetByCode(ctx context.Context, code string) (*model.Department, error)
	List(ctx context.Context) ([]model.Department, error)
	Update(ctx context.Context, tx *gorm.DB, dept *model.Department) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountChildren(ctx context.Context, parentID uuid.UUID) (int64, error)
	GetDepartmentAndSubIDs(ctx context.Context, parentID uuid.UUID) ([]uuid.UUID, error)
}

type departmentRepo struct {
	db *gorm.DB
}

// NewDepartmentRepo 创建部门 Repo
func NewDepartmentRepo(db *gorm.DB) DepartmentRepo {
	return &departmentRepo{db: db}
}

func (r *departmentRepo) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *departmentRepo) Create(ctx context.Context, tx *gorm.DB, dept *model.Department) error {
	return r.getDB(tx).WithContext(ctx).Create(dept).Error
}

func (r *departmentRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Department, error) {
	var dept model.Department
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&dept).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &dept, err
}

func (r *departmentRepo) GetByCode(ctx context.Context, code string) (*model.Department, error) {
	var dept model.Department
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&dept).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &dept, err
}

func (r *departmentRepo) List(ctx context.Context) ([]model.Department, error) {
	var depts []model.Department
	err := r.db.WithContext(ctx).
		Order("sort_order ASC, created_at ASC").
		Find(&depts).Error
	return depts, err
}

func (r *departmentRepo) Update(ctx context.Context, tx *gorm.DB, dept *model.Department) error {
	return r.getDB(tx).WithContext(ctx).Save(dept).Error
}

func (r *departmentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Department{}, id).Error
}

func (r *departmentRepo) CountChildren(ctx context.Context, parentID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Department{}).
		Where("parent_id = ?", parentID).
		Count(&count).Error
	return count, err
}

// GetDepartmentAndSubIDs 递归获取部门及其所有子部门ID
// 采用迭代方式逐层查询子部门，避免递归 CTE 在部分 PostgreSQL 场景下的兼容性问题
func (r *departmentRepo) GetDepartmentAndSubIDs(ctx context.Context, parentID uuid.UUID) ([]uuid.UUID, error) {
	result := []uuid.UUID{parentID}
	currentLevel := []uuid.UUID{parentID}

	for len(currentLevel) > 0 {
		var nextLevel []uuid.UUID
		err := r.db.WithContext(ctx).
			Model(&model.Department{}).
			Where("parent_id IN ?", currentLevel).
			Pluck("id", &nextLevel).Error
		if err != nil {
			return nil, err
		}
		result = append(result, nextLevel...)
		currentLevel = nextLevel
	}

	return result, nil
}
