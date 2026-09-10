package model

import (
	"time"

	"github.com/google/uuid"
)

type TaskTransfer struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TaskID             uuid.UUID  `gorm:"type:uuid;not null"`
	FromUserID         uuid.UUID  `gorm:"type:uuid;not null"`
	ToUserID           uuid.UUID  `gorm:"type:uuid;not null"`
	InitiatorID        uuid.UUID  `gorm:"type:uuid;not null"`
	SourceDepartmentID uuid.UUID  `gorm:"type:uuid;not null"`
	TargetDepartmentID uuid.UUID  `gorm:"type:uuid;not null"`
	SourceCenterID     uuid.UUID  `gorm:"type:uuid;not null"`
	TargetCenterID     uuid.UUID  `gorm:"type:uuid;not null"`
	SupervisorRole     string     `gorm:"type:varchar(32);not null"`
	Kind               string     `gorm:"type:varchar(32);not null"`
	Status             string     `gorm:"type:varchar(32);not null"`
	Revision           int64      `gorm:"not null"`
	Reason             string     `gorm:"type:text;not null"`
	PreviousStatus     int16      `gorm:"type:smallint;not null"`
	WorkflowInstanceID *uuid.UUID `gorm:"type:uuid"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	CompletedAt        *time.Time
}

func (TaskTransfer) TableName() string { return "task_transfers" }

type TransferSignature struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	TransferID  uuid.UUID `gorm:"type:uuid;not null"`
	Requirement string    `gorm:"type:varchar(32);not null"`
	SignerID    uuid.UUID `gorm:"type:uuid;not null"`
	SignerRole  string    `gorm:"type:varchar(32);not null"`
	Waived      bool      `gorm:"not null"`
	Decision    string    `gorm:"type:varchar(16);not null"`
	Comment     string    `gorm:"type:text;not null"`
	CreatedAt   time.Time
}

func (TransferSignature) TableName() string { return "task_transfer_signatures" }

type TransferActor struct {
	ID           uuid.UUID
	DepartmentID *uuid.UUID
	Roles        []string
}
type TransferDepartment struct {
	ID       uuid.UUID
	ParentID *uuid.UUID
	Name     string
	Status   int16
}
