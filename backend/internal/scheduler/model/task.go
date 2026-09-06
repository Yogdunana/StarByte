package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusActive   int16 = 0
	StatusPaused   int16 = 1
	StatusDeleted  int16 = 2
	StatusFinished int16 = 3
)

type Task struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Name       string     `gorm:"type:varchar(100);not null"`
	Code       string     `gorm:"type:varchar(64);uniqueIndex;not null"`
	CronExpr   string     `gorm:"type:varchar(64);not null;default:''"`
	RunAt      *time.Time `gorm:"type:timestamptz"`
	Timezone   string     `gorm:"type:varchar(64);not null;default:Asia/Shanghai"`
	HandlerKey string     `gorm:"type:varchar(64);not null"`
	Payload    string     `gorm:"type:text;not null;default:''"`
	DependsOn  string     `gorm:"type:text;not null;default:'[]'"`
	ShardKey   string     `gorm:"type:varchar(64);not null;default:''"`
	Status     int16      `gorm:"type:smallint;not null;default:0"`
	MaxRetries int        `gorm:"not null;default:3"`
	TimeoutSec int        `gorm:"not null;default:60"`
	RetryCount int        `gorm:"not null;default:0"`
	NextRunAt  *time.Time `gorm:"type:timestamptz"`
	LastRunAt  *time.Time `gorm:"type:timestamptz"`
	LastStatus string     `gorm:"type:varchar(20);not null;default:''"`
	CreatedBy  *uuid.UUID `gorm:"type:uuid"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (Task) TableName() string { return "scheduler_tasks" }

func (t *Task) IsCron() bool { return t.CronExpr != "" }
