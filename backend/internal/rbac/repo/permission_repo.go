package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionRepo 权限数据访问接口
type PermissionRepo interface {
	Create(ctx context.Context, tx *gorm.DB, permission *model.Permission) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Permission, error)
	GetByCode(ctx context.Context, code string) (*model.Permission, error)
	List(ctx context.Context) ([]model.Permission, error)
	Update(ctx context.Context, tx *gorm.DB, permission *model.Permission) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountChildren(ctx context.Context, parentID uuid.UUID) (int64, error)
	GetPermissionIDsByUserID(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

type permissionRepo struct {
	db *gorm.DB
}

// NewPermissionRepo 创建权限 Repo
func NewPermissionRepo(db *gorm.DB) PermissionRepo {
	return &permissionRepo{db: db}
}

func (r *permissionRepo) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

// Create 创建权限
func (r *permissionRepo) Create(ctx context.Context, tx *gorm.DB, permission *model.Permission) error {
	return r.getDB(tx).WithContext(ctx).Create(permission).Error
}

// GetByID 根据 ID 查询权限，未找到返回 nil, nil
func (r *permissionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Permission, error) {
	var perm model.Permission
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&perm).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &perm, err
}

// GetByCode 根据编码查询权限，未找到返回 nil, nil
func (r *permissionRepo) GetByCode(ctx context.Context, code string) (*model.Permission, error) {
	var perm model.Permission
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&perm).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &perm, err
}

// List 查询全部权限（用于构建权限树）
func (r *permissionRepo) List(ctx context.Context) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.WithContext(ctx).
		Order("sort_order ASC, created_at ASC").
		Find(&permissions).Error
	return permissions, err
}

// Update 更新权限
func (r *permissionRepo) Update(ctx context.Context, tx *gorm.DB, permission *model.Permission) error {
	return r.getDB(tx).WithContext(ctx).Save(permission).Error
}

// Delete 删除权限
func (r *permissionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Permission{}, id).Error
}

// CountChildren 统计指定父级权限下的子权限数量
func (r *permissionRepo) CountChildren(ctx context.Context, parentID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Permission{}).
		Where("parent_id = ?", parentID).
		Count(&count).Error
	return count, err
}

// GetPermissionIDsByUserID 查询用户拥有的权限 ID 列表
// 关联 user_roles -> role_permissions -> permissions，
// 过滤已过期角色关联（expired_at IS NULL OR expired_at > NOW()）且权限状态为启用（status=0）
func (r *permissionRepo) GetPermissionIDsByUserID(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var permIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Table("user_roles AS ur").
		Joins("JOIN role_permissions rp ON ur.role_id = rp.role_id").
		Joins("JOIN permissions p ON rp.permission_id = p.id").
		Where("ur.user_id = ?", userID).
		Where("ur.expired_at IS NULL OR ur.expired_at > NOW()").
		Where("p.status = ?", 0).
		Group("rp.permission_id").
		Pluck("rp.permission_id", &permIDs).Error
	return permIDs, err
}
