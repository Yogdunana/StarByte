package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	DirectionExpense int16 = 1
	DirectionIncome  int16 = 2
	CategoryActive   int16 = 0
)

type Category struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Code        string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"code"`
	Direction   int16     `gorm:"type:smallint;not null;default:1" json:"direction"`
	Description string    `gorm:"type:varchar(255)" json:"description"`
	Status      int16     `gorm:"type:smallint;not null;default:0" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Category) TableName() string { return "finance_categories" }

type Record struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	CategoryID   uuid.UUID  `gorm:"type:uuid;not null" json:"category_id"`
	DepartmentID *uuid.UUID `gorm:"type:uuid" json:"department_id"`
	Amount       float64    `gorm:"type:decimal(12,2);not null" json:"amount"`
	Direction    int16      `gorm:"type:smallint;not null;default:1" json:"direction"`
	OccurredAt   time.Time  `gorm:"type:date;not null" json:"occurred_at"`
	Title        string     `gorm:"type:varchar(200);not null" json:"title"`
	Remark       string     `gorm:"type:text" json:"remark"`
	CreatedBy    *uuid.UUID `gorm:"type:uuid" json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (Record) TableName() string { return "finance_records" }

type RecordNamed struct {
	Record
	CategoryName   string `gorm:"column:category_name"`
	DepartmentName string `gorm:"column:department_name"`
	CreatorName    string `gorm:"column:creator_name"`
}

type SummaryRow struct {
	Direction int16   `gorm:"column:direction"`
	Total     float64 `gorm:"column:total"`
	Count     int64   `gorm:"column:item_count"`
}

type CategorySumRow struct {
	CategoryID   uuid.UUID `gorm:"column:category_id"`
	CategoryName string    `gorm:"column:category_name"`
	Direction    int16     `gorm:"column:direction"`
	Total        float64   `gorm:"column:total"`
	Count        int64     `gorm:"column:item_count"`
}
