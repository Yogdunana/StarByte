package model

import (
	"time"

	"github.com/google/uuid"
)

type FileCleanup struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	AttachmentID  uuid.UUID `gorm:"type:uuid"`
	TaskID        uuid.UUID `gorm:"type:uuid"`
	FileID        uuid.UUID `gorm:"type:uuid"`
	UploadedBy    uuid.UUID `gorm:"type:uuid"`
	RequestedBy   uuid.UUID `gorm:"type:uuid"`
	Status        string
	Attempts      int
	LastError     string
	NextAttemptAt time.Time
	CompletedAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (FileCleanup) TableName() string { return "task_file_deletions" }
