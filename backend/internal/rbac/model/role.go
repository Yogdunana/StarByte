package model

import (
	"time"

	"github.com/google/uuid"
)

// Role 角色模型
type Role struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string     `gorm:"type:varchar(50);not null" json:"name"`
	Code        string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Description string     `gorm:"type:varchar(255)" json:"description"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index" json:"parent_id"`
	SortOrder   int        `gorm:"default:0" json:"sort_order"`
	Status      int        `gorm:"type:smallint;default:0" json:"status"` // 0=启用 1=禁用
	IsSystem    bool       `gorm:"default:false" json:"is_system"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Role) TableName() string {
	return "roles"
}

// RolePermission 角色-权限关联模型
// 联合唯一索引 idx_role_perm 确保同一角色下同一权限只关联一次
type RolePermission struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RoleID       uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_role_perm,priority:1;index;not null" json:"role_id"`
	PermissionID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_role_perm,priority:2;index;not null" json:"permission_id"`
	DataScope    string    `gorm:"type:varchar(20);default:department" json:"data_scope"` // all, department, department_and_sub, self, custom
	CreatedAt    time.Time `json:"created_at"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

// UserRole 用户-角色关联模型
// 联合唯一索引 idx_user_role 确保同一用户下同一角色只关联一次
type UserRole struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;uniqueIndex:idx_user_role,priority:1;index;not null" json:"user_id"`
	RoleID    uuid.UUID  `gorm:"type:uuid;uniqueIndex:idx_user_role,priority:2;index;not null" json:"role_id"`
	ExpiredAt *time.Time `json:"expired_at"`
	CreatedAt time.Time  `json:"created_at"`
}

func (UserRole) TableName() string {
	return "user_roles"
}

// RoleDataScope 角色数据权限-自定义部门关联模型
type RoleDataScope struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RolePermissionID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_role_data_scope,priority:1;index;not null" json:"role_permission_id"`
	DepartmentID     uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_role_data_scope,priority:2;index;not null" json:"department_id"`
	CreatedAt        time.Time `json:"created_at"`
}

func (RoleDataScope) TableName() string {
	return "role_data_scopes"
}

// 角色状态常量（使用 iota 自动递增，0=启用，1=禁用）
const (
	RoleStatusEnabled  = iota // 启用
	RoleStatusDisabled        // 禁用
)

// RoleCodeSuperAdmin 是唯一被焊死的角色编码。
// 它的权限集合不允许改动：一旦被改窄，包括操作者自己在内没有人能再授权回去，
// 等于把自己锁在系统外。其余系统角色（advisor / honorary / president …）
// 只保留「不可删除、不可改编码与状态」的保护，权限允许按需调整。
const RoleCodeSuperAdmin = "super_admin"
