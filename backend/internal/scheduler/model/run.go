package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	RunPending  = "pending"
	RunRunning  = "running"
	RunSuccess  = "success"
	RunFailed   = "failed"
	RunRetrying = "retrying"
	RunDead     = "dead"
)

type Run struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TaskID      uuid.UUID  `gorm:"type:uuid;index;not null"`
	ScheduledAt time.Time  `gorm:"type:timestamptz;not null"`
	StartedAt   *time.Time `gorm:"type:timestamptz"`
	FinishedAt  *time.Time `gorm:"type:timestamptz"`
	Status      string     `gorm:"type:varchar(20);not null;default:pending"`
	Attempt     int        `gorm:"not null;default:1"`
	WorkerID    string     `gorm:"type:varchar(128);not null;default:''"`
	ErrorText   string     `gorm:"type:text;not null;default:''"`
	Output      string     `gorm:"type:text;not null;default:''"`
	CreatedAt   time.Time
}

func (Run) TableName() string { return "scheduler_runs" }

type RunLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	RunID     uuid.UUID `gorm:"type:uuid;index;not null"`
	Level     string    `gorm:"type:varchar(16);not null;default:info"`
	Line      string    `gorm:"type:text;not null"`
	CreatedAt time.Time
}

func (RunLog) TableName() string { return "scheduler_run_logs" }
