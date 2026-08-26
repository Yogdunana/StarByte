package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DepartmentRepo 部门数据访问接口
// 定义部门相关的数据库操作，支持事务传入以保证复杂操作的原子性。
type DepartmentRepo interface {
	// Create 创建部门记录
	// tx 为事务对象，为 nil 时使用默认数据库连接
	Create(ctx context.Context, tx *gorm.DB, dept *model.Department) error
	// GetByID 根据 ID 查询部门，未找到返回 nil, nil
	GetByID(ctx context.Context, id uuid.UUID) (*model.Department, error)
	// GetByCode 根据编码查询部门，未找到返回 nil, nil
	GetByCode(ctx context.Context, code string) (*model.Department, error)
	// List 查询所有部门，按 sort_order 和 id 排序
	List(ctx context.Context) ([]model.Department, error)
	// Update 更新部门记录（全量保存）
	Update(ctx context.Context, tx *gorm.DB, dept *model.Department) error
	// Delete 根据 ID 删除部门
	Delete(ctx context.Context, tx *gorm.DB, id uuid.UUID) error
	// CountChildren 统计指定父部门下的子部门数量
	CountChildren(ctx context.Context, tx *gorm.DB, parentID uuid.UUID) (int64, error)
	// CountUsersByDeptID 统计指定部门下的有效用户数（排除已删除用户）
	CountUsersByDeptID(ctx context.Context, tx *gorm.DB, deptID uuid.UUID) (int64, error)
	// GetDepartmentAndSubIDs 获取部门及其所有子部门 ID 列表
	// 采用迭代逐层查询，限制最大深度防止环路导致死循环
	GetDepartmentAndSubIDs(ctx context.Context, parentID uuid.UUID) ([]uuid.UUID, error)
}

type departmentRepo struct {
	db *gorm.DB
}

// maxDeptDepth 部门递归查询最大深度，防止脏数据形成环导致死循环
const maxDeptDepth = 20

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
		Order("sort_order ASC, id ASC").
		Find(&depts).Error
	return depts, err
}

func (r *departmentRepo) Update(ctx context.Context, tx *gorm.DB, dept *model.Department) error {
	return r.getDB(tx).WithContext(ctx).Save(dept).Error
}

func (r *departmentRepo) Delete(ctx context.Context, tx *gorm.DB, id uuid.UUID) error {
	return r.getDB(tx).WithContext(ctx).Delete(&model.Department{}, id).Error
}

func (r *departmentRepo) CountChildren(ctx context.Context, tx *gorm.DB, parentID uuid.UUID) (int64, error) {
	var count int64
	err := r.getDB(tx).WithContext(ctx).
		Model(&model.Department{}).
		Where("parent_id = ?", parentID).
		Count(&count).Error
	return count, err
}

// CountUsersByDeptID 统计指定部门下的有效用户数
// 使用原生 SQL 查询 users 表，避免引入 user 模块依赖
func (r *departmentRepo) CountUsersByDeptID(ctx context.Context, tx *gorm.DB, deptID uuid.UUID) (int64, error) {
	var count int64
	err := r.getDB(tx).WithContext(ctx).
		Table("users").
		Where("department_id = ? AND deleted_at IS NULL", deptID).
		Count(&count).Error
	return count, err
}

// GetDepartmentAndSubIDs 获取部门及其所有子部门ID
// 采用迭代方式逐层查询，使用 visited 集合去重并限制最大深度，防止脏数据形成环导致死循环
func (r *departmentRepo) GetDepartmentAndSubIDs(ctx context.Context, parentID uuid.UUID) ([]uuid.UUID, error) {
	result := []uuid.UUID{parentID}
	visited := map[uuid.UUID]bool{parentID: true}
	currentLevel := []uuid.UUID{parentID}
	depth := 0

	for len(currentLevel) > 0 && depth < maxDeptDepth {
		var nextLevel []uuid.UUID
		err := r.db.WithContext(ctx).
			Model(&model.Department{}).
			Where("parent_id IN ?", currentLevel).
			Pluck("id", &nextLevel).Error
		if err != nil {
			return nil, err
		}

		// 过滤已访问节点，防止环路导致死循环
		filtered := make([]uuid.UUID, 0, len(nextLevel))
		for _, id := range nextLevel {
			if !visited[id] {
				visited[id] = true
				filtered = append(filtered, id)
			}
		}

		if len(filtered) == 0 {
			break
		}

		result = append(result, filtered...)
		currentLevel = filtered
		depth++
	}

	return result, nil
}
