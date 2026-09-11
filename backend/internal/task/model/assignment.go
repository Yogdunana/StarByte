package model

import (
	"time"

	"github.com/google/uuid"
)

type AssignmentPolicy struct {
	Mode         string     `json:"mode"`
	DepartmentID uuid.UUID  `json:"department_id"`
	RoleID       *uuid.UUID `json:"role_id,omitempty"`
}
type AssignmentCursor struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	RuleKey    string     `gorm:"type:varchar(160);not null;uniqueIndex"`
	LastUserID *uuid.UUID `gorm:"type:uuid"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (AssignmentCursor) TableName() string { return "task_assignment_cursors" }

type AssignmentRole struct {
	ID   uuid.UUID
	Name string
}
