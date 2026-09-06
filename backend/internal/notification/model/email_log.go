package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	EmailQueued   int16 = 0
	EmailSent     int16 = 1
	EmailFailed   int16 = 2
	EmailRetrying int16 = 3
)

// EmailLog 对应 email_logs（000016 + 000031）。
type EmailLog struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID         *uuid.UUID `gorm:"type:uuid"`
	ToAddress      string     `gorm:"type:text;not null"`
	Subject        string     `gorm:"type:varchar(500);not null"`
	TemplateCode   string     `gorm:"type:varchar(100)"`
	Status         int16      `gorm:"type:smallint;not null;default:0"`
	ErrorMessage   string     `gorm:"type:text"`
	SentAt         *time.Time
	CreatedAt      time.Time
	NotificationID *uuid.UUID `gorm:"type:uuid"`
	RetryCount     int16      `gorm:"type:smallint;not null;default:0"`
	CC             string     `gorm:"type:text"`
	IsHTML         bool       `gorm:"not null;default:false"`
}

func (EmailLog) TableName() string { return "email_logs" }

func EmailStatusName(s int16) string {
	switch s {
	case EmailSent:
		return "sent"
	case EmailFailed:
		return "failed"
	case EmailRetrying:
		return "retrying"
	default:
		return "queued"
	}
}
