package model

import (
	"time"

	"github.com/google/uuid"
)

type Report struct {
	ID                  uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID              uuid.UUID    `gorm:"type:uuid;not null" json:"user_id"`
	DepartmentID        *uuid.UUID   `gorm:"type:uuid" json:"department_id"`
	ReportType          ReportType   `gorm:"type:varchar(16);not null" json:"report_type"`
	PeriodStart         time.Time    `gorm:"type:date;not null" json:"period_start"`
	PeriodEnd           time.Time    `gorm:"type:date;not null" json:"period_end"`
	WorkContent         string       `gorm:"type:text;not null;default:''" json:"work_content"`
	NextPlan            string       `gorm:"type:text;not null;default:''" json:"next_plan"`
	ProblemsSuggestions string       `gorm:"type:text;not null;default:''" json:"problems_suggestions"`
	ReviewStatus        ReviewStatus `gorm:"type:varchar(16);not null;default:'pending'" json:"review_status"`
	SubmittedAt         time.Time    `gorm:"not null" json:"submitted_at"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`
}

func (Report) TableName() string { return "work_reports" }
