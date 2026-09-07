package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	LevelWarning    int16 = 1
	LevelSevere     int16 = 2
	LevelDemerit    int16 = 3
	LevelProbation  int16 = 4
	LevelExpel      int16 = 5
	StatusPending   int16 = 0
	StatusActive    int16 = 1
	StatusRevoked   int16 = 2
	StatusAppealing int16 = 3
	AppealPending   int16 = 0
	AppealAccepted  int16 = 1
	AppealRejected  int16 = 2
	FlowKey               = "discipline_approve"
)

type Record struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	Title          string     `gorm:"type:varchar(200);not null" json:"title"`
	Description    string     `gorm:"type:text" json:"description"`
	Level          int16      `gorm:"type:smallint;not null;default:1" json:"level"`
	Status         int16      `gorm:"type:smallint;not null;default:0" json:"status"`
	IssuedBy       *uuid.UUID `gorm:"type:uuid" json:"issued_by"`
	IssuedAt       time.Time  `json:"issued_at"`
	FlowInstanceID *uuid.UUID `gorm:"type:uuid" json:"flow_instance_id"`
	ApprovedBy     *uuid.UUID `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt     *time.Time `json:"approved_at"`
	RevokeReason   string     `gorm:"type:text" json:"revoke_reason"`
	RevokedBy      *uuid.UUID `gorm:"type:uuid" json:"revoked_by"`
	RevokedAt      *time.Time `json:"revoked_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (Record) TableName() string { return "discipline_records" }

type Appeal struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	RecordID    uuid.UUID  `gorm:"type:uuid;not null" json:"record_id"`
	ApplicantID uuid.UUID  `gorm:"type:uuid;not null" json:"applicant_id"`
	Reason      string     `gorm:"type:text;not null" json:"reason"`
	Status      int16      `gorm:"type:smallint;not null;default:0" json:"status"`
	ReviewerID  *uuid.UUID `gorm:"type:uuid" json:"reviewer_id"`
	ReviewedAt  *time.Time `json:"reviewed_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (Appeal) TableName() string { return "discipline_appeals" }

type RecordNamed struct {
	Record
	UserName       string     `gorm:"column:user_name"`
	IssuerName     string     `gorm:"column:issuer_name"`
	DepartmentID   *uuid.UUID `gorm:"column:department_id"`
	DepartmentName string     `gorm:"column:department_name"`
}

type NamedUser struct {
	ID           uuid.UUID
	RealName     string
	Username     string
	DepartmentID *uuid.UUID
}

func ValidLevel(level int16) bool {
	return level >= LevelWarning && level <= LevelExpel
}
