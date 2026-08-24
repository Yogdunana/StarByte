package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Username     string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	RealName     string         `gorm:"type:varchar(50)" json:"real_name"`
	AvatarURL    string         `gorm:"type:varchar(500)" json:"avatar_url"`
	Email        string         `gorm:"type:varchar(100)" json:"email"`
	Phone        string         `gorm:"type:varchar(20)" json:"phone"`
	Gender       int            `gorm:"type:smallint;default:0" json:"gender"`       // 0=未知 1=男 2=女
	Status       int            `gorm:"type:smallint;default:0;index" json:"status"` // 0=正常 1=禁用 2=锁定
	DepartmentID *uuid.UUID     `gorm:"type:uuid;index" json:"department_id"`
	PositionID   *uuid.UUID     `gorm:"type:uuid" json:"position_id"`
	LastLoginAt  *time.Time     `json:"last_login_at"`
	LastLoginIP  string         `gorm:"type:varchar(50)" json:"last_login_ip"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

// UserIdentity 用户身份关联
type UserIdentity struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	IdentityType  string    `gorm:"type:varchar(20);not null" json:"identity_type"`
	IdentityValue string    `gorm:"type:varchar(100);not null" json:"identity_value"`
	IsPrimary     bool      `gorm:"default:false" json:"is_primary"`
	CreatedAt     time.Time `json:"created_at"`
}

func (UserIdentity) TableName() string {
	return "user_identities"
}

// 注意：Role、Permission、Department、Position 模型已迁移至
// backend/internal/rbac/model/ 包，由 RBAC 模块统一管理
