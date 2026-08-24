package model

import (
	"time"

	"github.com/google/uuid"
)

// Department 部门模型
type Department struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string     `gorm:"type:varchar(100);not null" json:"name"`
	Code        string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index" json:"parent_id"`
	LeaderID    *uuid.UUID `gorm:"type:uuid" json:"leader_id"`
	Description string     `gorm:"type:varchar(255)" json:"description"`
	SortOrder   int        `gorm:"default:0" json:"sort_order"`
	Status      int        `gorm:"type:smallint;default:0" json:"status"` // 0=启用 1=禁用
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Department) TableName() string {
	return "departments"
}

// 部门状态常量
const (
	DepartmentStatusEnabled  = 0 // 启用
	DepartmentStatusDisabled = 1 // 禁用
)
