package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionRepo 权限数据访问接口
// 定义权限相关的数据库操作，支持事务传入以保证复杂操作的原子性。
type PermissionRepo interface {
	// Create 创建权限记录
	Create(ctx context.Context, tx *gorm.DB, permission *model.Permission) error
	// GetByID 根据 ID 查询权限，未找到返回 nil, nil
	GetByID(ctx context.Context, id uuid.UUID) (*model.Permission, error)
	// GetByIDs 根据 ID 列表批量查询权限
	GetByIDs(ctx context.Context, tx *gorm.DB, ids []uuid.UUID) ([]model.Permission, error)
	// GetByCode 根据编码查询权限，未找到返回 nil, nil
	GetByCode(ctx context.Context, code string) (*model.Permission, error)
	// List 查询全部权限（按 sort_order 和 id 排序，用于构建权限树）
	List(ctx context.Context) ([]model.Permission, error)
	// Update 更新权限记录（全量保存）
	Update(ctx context.Context, tx *gorm.DB, permission *model.Permission) error
	// Delete 根据 ID 删除权限
	Delete(ctx context.Context, id uuid.UUID) error
	// CountChildren 统计指定父权限下的子权限数量
	CountChildren(ctx context.Context, parentID uuid.UUID) (int64, error)
	// GetPermissionIDsByUserID 查询用户拥有的权限 ID 列表
	// 关联 user_roles -> roles -> role_permissions -> permissions
	// 过滤条件：角色未过期、角色已启用、权限已启用
	GetPermissionIDsByUserID(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	// GetPermissionCodesByUserID 查询用户拥有的权限编码列表
	// 关联 user_roles -> roles -> role_permissions -> permissions
	// 过滤条件：角色未过期、角色已启用、权限已启用
	GetPermissionCodesByUserID(ctx context.Context, userID uuid.UUID) ([]string, error)
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

// GetByIDs 根据 ID 列表批量查询权限
func (r *permissionRepo) GetByIDs(ctx context.Context, tx *gorm.DB, ids []uuid.UUID) ([]model.Permission, error) {
	if len(ids) == 0 {
		return []model.Permission{}, nil
	}
	var perms []model.Permission
	err := r.getDB(tx).WithContext(ctx).Where("id IN ?", ids).Find(&perms).Error
	return perms, err
}

// List 查询全部权限（用于构建权限树）
func (r *permissionRepo) List(ctx context.Context) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.WithContext(ctx).
		Order("sort_order ASC, id ASC").
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

// baseUserPermissionQuery 构建用户权限查询的基础 JOIN 和过滤条件
// 关联 user_roles -> roles -> role_permissions -> permissions
// 过滤条件：角色未过期、角色已启用、权限已启用
func (r *permissionRepo) baseUserPermissionQuery(ctx context.Context, userID uuid.UUID) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("user_roles AS ur").
		Joins("JOIN roles r ON ur.role_id = r.id").
		Joins("JOIN role_permissions rp ON ur.role_id = rp.role_id").
		Joins("JOIN permissions p ON rp.permission_id = p.id").
		Where("ur.user_id = ? AND (ur.expired_at IS NULL OR ur.expired_at > NOW())", userID).
		Where("r.status = ?", 0). // 角色需启用
		Where("p.status = ?", 0)  // 权限需启用
}

// GetPermissionIDsByUserID 查询用户拥有的权限 ID 列表
func (r *permissionRepo) GetPermissionIDsByUserID(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var permIDs []uuid.UUID
	err := r.baseUserPermissionQuery(ctx, userID).
		Group("rp.permission_id").
		Pluck("rp.permission_id", &permIDs).Error
	return permIDs, err
}

// GetPermissionCodesByUserID 查询用户拥有的权限编码列表
// 通过一次 JOIN 查询直接返回权限编码，减少一次 DB 往返
func (r *permissionRepo) GetPermissionCodesByUserID(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var codes []string
	err := r.baseUserPermissionQuery(ctx, userID).
		Distinct("p.code").
		Pluck("p.code", &codes).Error
	return codes, err
}
