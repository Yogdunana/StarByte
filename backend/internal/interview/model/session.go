package model

import (
	"time"

	"github.com/google/uuid"
)

// Session 面试场次。
type Session struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Title         string     `gorm:"type:varchar(200);not null" json:"title"`
	Round         int16      `gorm:"type:smallint;not null;default:1" json:"round"`
	DepartmentID  *uuid.UUID `gorm:"type:uuid" json:"department_id"`
	StartTime     time.Time  `json:"start_time"`
	EndTime       time.Time  `json:"end_time"`
	Location      string     `gorm:"type:varchar(200);not null;default:''" json:"location"`
	OnlineLink    string     `gorm:"type:varchar(500);not null;default:''" json:"online_link"`
	Status        int16      `gorm:"type:smallint;not null;default:0" json:"status"`
	MaxCandidates int        `gorm:"not null;default:20" json:"max_candidates"`
	Description   string     `gorm:"type:text;not null;default:''" json:"description"`
	CreatedBy     *uuid.UUID `gorm:"type:uuid" json:"created_by"`
	QRToken       string     `gorm:"column:qr_token;type:varchar(64);not null;default:''" json:"-"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (Session) TableName() string { return "interview_sessions" }

// SessionWithNames 场次联表。
type SessionWithNames struct {
	Session
	DepartmentName string `gorm:"column:department_name"`
	CandidateCount int64  `gorm:"column:candidate_count"`
}

// NamedUser 用户摘要。
type NamedUser struct {
	ID       uuid.UUID
	RealName string
	Username string
}

// ApplicationBrief 入会申请摘要，用于导入面试者。
type ApplicationBrief struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	RealName     string
	StudentNo    string
	DepartmentID *uuid.UUID
	Status       int16
}
