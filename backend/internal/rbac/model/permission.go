package model

import (
	"time"

	"github.com/google/uuid"
)

// Permission 权限模型
type Permission struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string     `gorm:"type:varchar(100);not null" json:"name"`
	Code        string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"code"`
	Resource    string     `gorm:"type:varchar(50);index" json:"resource"`
	Action      string     `gorm:"type:varchar(50)" json:"action"`
	Description string     `gorm:"type:varchar(255)" json:"description"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index" json:"parent_id"`
	SortOrder   int        `gorm:"default:0" json:"sort_order"`
	Type        int        `gorm:"type:smallint;default:1" json:"type"` // 1=菜单 2=按钮 3=接口
	Path        string     `gorm:"type:varchar(255)" json:"path"`
	Icon        string     `gorm:"type:varchar(50)" json:"icon"`
	APIMethod   string     `gorm:"type:varchar(10)" json:"api_method"`
	APIPath     string     `gorm:"type:varchar(255)" json:"api_path"`
	IsSystem    bool       `gorm:"default:false" json:"is_system"`
	Status      int        `gorm:"type:smallint;default:0" json:"status"` // 0=启用 1=禁用
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Permission) TableName() string {
	return "permissions"
}

// 权限类型常量
const (
	PermissionTypeMenu   = 1 // 菜单
	PermissionTypeButton = 2 // 按钮
	PermissionTypeAPI    = 3 // 接口
)

// 权限状态常量
const (
	PermissionStatusEnabled  = 0 // 启用
	PermissionStatusDisabled = 1 // 禁用
)
