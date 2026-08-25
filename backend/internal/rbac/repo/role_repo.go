package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleRepo 角色数据访问接口
type RoleRepo interface {
	Create(ctx context.Context, tx *gorm.DB, role *model.Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Role, error)
	GetByCode(ctx context.Context, code string) (*model.Role, error)
	List(ctx context.Context, page, pageSize int, keyword string) ([]model.Role, int64, error)
	Update(ctx context.Context, tx *gorm.DB, role *model.Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	AssignPermissions(ctx context.Context, tx *gorm.DB, roleID uuid.UUID, permissionIDs []uuid.UUID, dataScope string) error
	GetPermissionIDsByRoleID(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error)
	GetUsersByRoleID(ctx context.Context, roleID uuid.UUID, page, pageSize int, dataScope *model.DataScopeCondition) ([]RoleUserListItem, int64, error)
	CountUsersByRoleID(ctx context.Context, roleID uuid.UUID) (int64, error)
	GetRoleIDsByUserID(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

// RoleUserListItem 角色下用户列表项
type RoleUserListItem struct {
	ID       string `gorm:"column:id"`
	Username string `gorm:"column:username"`
	RealName string `gorm:"column:real_name"`
	Status   int    `gorm:"column:status"`
}

type roleRepo struct {
	db *gorm.DB
}

// NewRoleRepo 创建角色 Repo
func NewRoleRepo(db *gorm.DB) RoleRepo {
	return &roleRepo{db: db}
}

func (r *roleRepo) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

// Create 创建角色
func (r *roleRepo) Create(ctx context.Context, tx *gorm.DB, role *model.Role) error {
	return r.getDB(tx).WithContext(ctx).Create(role).Error
}

// GetByID 根据 ID 查询角色，未找到返回 nil, nil
func (r *roleRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&role).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &role, err
}

// GetByCode 根据编码查询角色，未找到返回 nil, nil
func (r *roleRepo) GetByCode(ctx context.Context, code string) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&role).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &role, err
}

// List 分页查询角色列表
func (r *roleRepo) List(ctx context.Context, page, pageSize int, keyword string) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Role{})

	if keyword != "" {
		query = query.Where("(name LIKE ? OR code LIKE ?)",
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
		Find(&roles).Error

	return roles, total, err
}

// Update 更新角色
func (r *roleRepo) Update(ctx context.Context, tx *gorm.DB, role *model.Role) error {
	return r.getDB(tx).WithContext(ctx).Save(role).Error
}

// Delete 删除角色
func (r *roleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Role{}, id).Error
}

// AssignPermissions 分配角色权限：先删除已有关联，再批量插入新关联
// dataScope 指定数据权限范围，为空时默认使用 department
func (r *roleRepo) AssignPermissions(ctx context.Context, tx *gorm.DB, roleID uuid.UUID, permissionIDs []uuid.UUID, dataScope string) error {
	db := r.getDB(tx).WithContext(ctx)

	// 删除现有角色-权限关联
	if err := db.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
		return err
	}

	// 没有新权限则直接返回
	if len(permissionIDs) == 0 {
		return nil
	}

	// 默认使用 department 数据范围
	if dataScope == "" {
		dataScope = model.DataScopeDepartment
	}

	// 批量插入新的关联
	rolePerms := make([]model.RolePermission, 0, len(permissionIDs))
	for _, permID := range permissionIDs {
		rolePerms = append(rolePerms, model.RolePermission{
			ID:           uuid.New(),
			RoleID:       roleID,
			PermissionID: permID,
			DataScope:    dataScope,
		})
	}

	return db.Create(&rolePerms).Error
}

// GetPermissionIDsByRoleID 查询角色拥有的权限 ID 列表
func (r *roleRepo) GetPermissionIDsByRoleID(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error) {
	var permIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&model.RolePermission{}).
		Where("role_id = ?", roleID).
		Pluck("permission_id", &permIDs).Error
	return permIDs, err
}

// GetUsersByRoleID 分页查询角色关联的用户列表（join user_roles 与 users）
// dataScope 为数据权限过滤条件，为空时不过滤
func (r *roleRepo) GetUsersByRoleID(ctx context.Context, roleID uuid.UUID, page, pageSize int, dataScope *model.DataScopeCondition) ([]RoleUserListItem, int64, error) {
	var total int64

	// 统计角色下有效用户数
	countQuery := r.db.WithContext(ctx).
		Table("user_roles AS ur").
		Joins("JOIN users u ON ur.user_id = u.id").
		Where("ur.role_id = ? AND u.deleted_at IS NULL AND (ur.expired_at IS NULL OR ur.expired_at > NOW())", roleID)

	// 应用数据权限过滤
	if dataScope != nil && dataScope.Query != "" {
		countQuery = countQuery.Where(dataScope.Query, dataScope.Args...)
	}

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []RoleUserListItem{}, 0, nil
	}

	// 分页查询用户信息，id 使用 ::text 避免驱动返回值类型不一致
	var results []RoleUserListItem
	offset := (page - 1) * pageSize
	listQuery := r.db.WithContext(ctx).
		Table("user_roles AS ur").
		Select("u.id::text AS id, u.username, u.real_name, u.status").
		Joins("JOIN users u ON ur.user_id = u.id").
		Where("ur.role_id = ? AND u.deleted_at IS NULL AND (ur.expired_at IS NULL OR ur.expired_at > NOW())", roleID).
		Order("ur.created_at DESC").
		Offset(offset).
		Limit(pageSize)

	// 应用数据权限过滤
	if dataScope != nil && dataScope.Query != "" {
		listQuery = listQuery.Where(dataScope.Query, dataScope.Args...)
	}

	err := listQuery.Scan(&results).Error

	return results, total, err
}

// CountUsersByRoleID 统计角色下有效用户数
func (r *roleRepo) CountUsersByRoleID(ctx context.Context, roleID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("user_roles AS ur").
		Joins("JOIN users u ON ur.user_id = u.id").
		Where("ur.role_id = ? AND u.deleted_at IS NULL AND (ur.expired_at IS NULL OR ur.expired_at > NOW())", roleID).
		Count(&count).Error
	return count, err
}

// GetRoleIDsByUserID 查询用户拥有的角色 ID 列表（过滤已过期关联和已禁用角色）
func (r *roleRepo) GetRoleIDsByUserID(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var roleIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&model.UserRole{}).
		Joins("JOIN roles r ON user_roles.role_id = r.id").
		Where("user_roles.user_id = ? AND (user_roles.expired_at IS NULL OR user_roles.expired_at > NOW())", userID).
		Where("r.status = ?", 0).
		Pluck("user_roles.role_id", &roleIDs).Error
	return roleIDs, err
}
