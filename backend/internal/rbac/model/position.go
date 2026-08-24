package model

import (
	"time"

	"github.com/google/uuid"
)

// Position 职位模型
type Position struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(50);not null" json:"name"`
	Code        string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Level       int       `gorm:"default:0" json:"level"`
	VoteWeight  float64   `gorm:"type:decimal(5,2);default:1.00" json:"vote_weight"`
	Description string    `gorm:"type:varchar(255)" json:"description"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	Status      int       `gorm:"type:smallint;default:0" json:"status"` // 0=启用 1=禁用
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Position) TableName() string {
	return "positions"
}

// 职位状态常量
const (
	PositionStatusEnabled  = 0 // 启用
	PositionStatusDisabled = 1 // 禁用
)
