package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Username      string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	PasswordHash  string         `gorm:"type:varchar(255);not null" json:"-"`
	RealName      string         `gorm:"type:varchar(50)" json:"real_name"`
	AvatarURL     string         `gorm:"type:varchar(500)" json:"avatar_url"`
	Email         string         `gorm:"type:varchar(100)" json:"email"`
	Phone         string         `gorm:"type:varchar(20)" json:"phone"`
	Gender        int            `gorm:"type:smallint;default:0" json:"gender"` // 0=未知 1=男 2=女
	Status        int            `gorm:"type:smallint;default:0;index" json:"status"` // 0=正常 1=禁用 2=锁定
	DepartmentID  *uuid.UUID     `gorm:"type:uuid;index" json:"department_id"`
	PositionID    *uuid.UUID     `gorm:"type:uuid" json:"position_id"`
	LastLoginAt   *time.Time     `json:"last_login_at"`
	LastLoginIP   string         `gorm:"type:varchar(50)" json:"last_login_ip"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

// UserIdentity 用户身份关联
type UserIdentity struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID         uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	IdentityType   string    `gorm:"type:varchar(20);not null" json:"identity_type"`
	IdentityValue  string    `gorm:"type:varchar(100);not null" json:"identity_value"`
	IsPrimary      bool      `gorm:"default:false" json:"is_primary"`
	CreatedAt      time.Time `json:"created_at"`
}

func (UserIdentity) TableName() string {
	return "user_identities"
}

// Role 角色模型
type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(50);not null" json:"name"`
	Code        string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Description string    `gorm:"type:varchar(255)" json:"description"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index" json:"parent_id"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	Status      int       `gorm:"type:smallint;default:0" json:"status"`
	IsSystem    bool      `gorm:"default:false" json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Role) TableName() string {
	return "roles"
}

// Permission 权限模型
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Code        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"code"`
	Resource    string    `gorm:"type:varchar(50);index" json:"resource"`
	Action      string    `gorm:"type:varchar(50)" json:"action"`
	Description string    `gorm:"type:varchar(255)" json:"description"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index" json:"parent_id"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	Type        int       `gorm:"type:smallint;default:1" json:"type"` // 1=菜单 2=按钮 3=接口
	Path        string    `gorm:"type:varchar(255)" json:"path"`
	Icon        string    `gorm:"type:varchar(50)" json:"icon"`
	APIMethod   string    `gorm:"type:varchar(10)" json:"api_method"`
	APIPath     string    `gorm:"type:varchar(255)" json:"api_path"`
	IsSystem    bool      `gorm:"default:false" json:"is_system"`
	Status      int       `gorm:"type:smallint;default:0" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Permission) TableName() string {
	return "permissions"
}

// Department 部门模型
type Department struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Code        string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index" json:"parent_id"`
	LeaderID    *uuid.UUID `gorm:"type:uuid" json:"leader_id"`
	Description string    `gorm:"type:varchar(255)" json:"description"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	Status      int       `gorm:"type:smallint;default:0" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Department) TableName() string {
	return "departments"
}

// Position 职务模型
type Position struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(50);not null" json:"name"`
	Code        string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Level       int       `gorm:"default:0" json:"level"`
	VoteWeight  float64   `gorm:"type:decimal(5,2);default:1.00" json:"vote_weight"`
	Description string    `gorm:"type:varchar(255)" json:"description"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	Status      int       `gorm:"type:smallint;default:0" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Position) TableName() string {
	return "positions"
}
