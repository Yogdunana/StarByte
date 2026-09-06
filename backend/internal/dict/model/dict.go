package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusEnabled  int16 = 0
	StatusDisabled int16 = 1
)

type DictType struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Code        string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Description string    `gorm:"type:varchar(255);not null;default:''"`
	SortOrder   int       `gorm:"not null;default:0"`
	Status      int16     `gorm:"type:smallint;not null;default:0"`
	IsSystem    bool      `gorm:"not null;default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (DictType) TableName() string { return "dict_types" }

type DictItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	TypeID    uuid.UUID `gorm:"type:uuid;index;not null"`
	ItemValue string    `gorm:"type:varchar(50);not null"`
	ItemLabel string    `gorm:"type:varchar(100);not null"`
	SortOrder int       `gorm:"not null;default:0"`
	Status    int16     `gorm:"type:smallint;not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (DictItem) TableName() string { return "dict_items" }
