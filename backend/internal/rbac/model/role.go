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
type RolePermission struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RoleID       uuid.UUID `gorm:"type:uuid;index;not null" json:"role_id"`
	PermissionID uuid.UUID `gorm:"type:uuid;index;not null" json:"permission_id"`
	DataScope    string    `gorm:"type:varchar(20);default:department" json:"data_scope"` // all, department, department_and_sub, self, custom
	CreatedAt    time.Time `json:"created_at"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

// UserRole 用户-角色关联模型
type UserRole struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	RoleID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"role_id"`
	ExpiredAt *time.Time `json:"expired_at"`
	CreatedAt time.Time  `json:"created_at"`
}

func (UserRole) TableName() string {
	return "user_roles"
}

// RoleDataScope 角色数据权限-自定义部门关联模型
type RoleDataScope struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RolePermissionID uuid.UUID `gorm:"type:uuid;index;not null" json:"role_permission_id"`
	DepartmentID     uuid.UUID `gorm:"type:uuid;index;not null" json:"department_id"`
	CreatedAt        time.Time `json:"created_at"`
}

func (RoleDataScope) TableName() string {
	return "role_data_scopes"
}
