package model

import (
	"time"

	"github.com/google/uuid"
)

// PermissionType 权限类型
type PermissionType int

// String 返回权限类型的字符串表示
func (t PermissionType) String() string {
	switch t {
	case PermissionTypeMenu:
		return "menu"
	case PermissionTypeButton:
		return "button"
	case PermissionTypeAPI:
		return "api"
	default:
		return "menu"
	}
}

// ParsePermissionType 将字符串解析为权限类型
func ParsePermissionType(s string) PermissionType {
	switch s {
	case "menu":
		return PermissionTypeMenu
	case "button":
		return PermissionTypeButton
	case "api":
		return PermissionTypeAPI
	default:
		return PermissionTypeMenu
	}
}

// Permission 权限模型
type Permission struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Code        string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"code"`
	Resource    string         `gorm:"type:varchar(50);index" json:"resource"`
	Action      string         `gorm:"type:varchar(50)" json:"action"`
	Description string         `gorm:"type:varchar(255)" json:"description"`
	ParentID    *uuid.UUID     `gorm:"type:uuid;index" json:"parent_id"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	Type        PermissionType `gorm:"type:smallint;default:1" json:"type"` // 1=菜单 2=按钮 3=接口
	Path        string         `gorm:"type:varchar(255)" json:"path"`
	Icon        string         `gorm:"type:varchar(50)" json:"icon"`
	APIMethod   string         `gorm:"type:varchar(10)" json:"api_method"`
	APIPath     string         `gorm:"type:varchar(255)" json:"api_path"`
	IsSystem    bool           `gorm:"default:false" json:"is_system"`
	Status      int            `gorm:"type:smallint;default:0" json:"status"` // 0=启用 1=禁用
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func (Permission) TableName() string {
	return "permissions"
}

// 权限类型常量（使用 iota 从 1 开始递增）
const (
	PermissionTypeMenu   PermissionType = iota + 1 // 菜单
	PermissionTypeButton                           // 按钮
	PermissionTypeAPI                              // 接口
)

// 权限状态常量（使用 iota 自动递增，0=启用，1=禁用）
const (
	PermissionStatusEnabled  = iota // 启用
	PermissionStatusDisabled        // 禁用
)
