package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending       int16 = 0
	StatusRunning       int16 = 1
	StatusSuccess       int16 = 2
	StatusFailed        int16 = 3 // dump failed; no usable artifact
	StatusRestoring     int16 = 4
	StatusRestored      int16 = 5
	StatusRestoreFailed int16 = 6 // artifact intact; restore may be retried

	TriggerManual    = "manual"
	TriggerScheduled = "scheduled"

	StorageMinIO = "minio"
	StorageLocal = "local"

	DefaultRetentionDays = 30
	DefaultCronExpr      = "0 30 2 * * *"
	DefaultTimezone      = "Asia/Shanghai"
	RestoreConfirmToken  = "RESTORE"

	ScheduledTaskCode = "backup_scheduled_full"
	CleanupTaskCode   = "backup_retention_cleanup"
)

// Record is one backup / restore job.
type Record struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TriggerSource  string     `gorm:"type:varchar(16);not null;default:manual"`
	Status         int16      `gorm:"type:smallint;not null;default:0"`
	Storage        string     `gorm:"type:varchar(16);not null;default:''"`
	ObjectKey      string     `gorm:"type:text;not null;default:''"`
	Filename       string     `gorm:"type:varchar(255);not null;default:''"`
	ChecksumSHA256 string     `gorm:"column:checksum_sha256;type:varchar(64);not null;default:''"`
	SizeBytes      int64      `gorm:"not null;default:0"`
	StartedAt      *time.Time `gorm:"type:timestamptz"`
	FinishedAt     *time.Time `gorm:"type:timestamptz"`
	ErrorMessage   string     `gorm:"type:text;not null;default:''"`
	CreatedBy      *uuid.UUID `gorm:"type:uuid"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (Record) TableName() string { return "backup_records" }

// Policy is the singleton retention + schedule row (non-secret).
type Policy struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	Enabled       bool      `gorm:"not null;default:false"`
	RetentionDays int       `gorm:"not null;default:30"`
	CronExpr      string    `gorm:"type:varchar(64);not null"`
	Timezone      string    `gorm:"type:varchar(64);not null;default:Asia/Shanghai"`
	UpdatedBy     *uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (Policy) TableName() string { return "backup_policies" }
