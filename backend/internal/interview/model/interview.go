package model

import (
	"time"

	"github.com/google/uuid"
)

// Interview 面试记录。scheduled_at / type / score / result 沿用 000011。
type Interview struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SessionID       *uuid.UUID `gorm:"type:uuid" json:"session_id"`
	ApplicationID   *uuid.UUID `gorm:"type:uuid" json:"application_id"`
	ApplicantID     uuid.UUID  `gorm:"type:uuid;not null" json:"applicant_id"`
	Round           int16      `gorm:"type:smallint;not null;default:1" json:"round"`
	Type            int16      `gorm:"type:smallint;not null;default:1" json:"type"`
	Status          int16      `gorm:"type:smallint;not null;default:0" json:"status"`
	ScheduledAt     *time.Time `gorm:"column:scheduled_at" json:"scheduled_time"`
	Location        string     `gorm:"type:varchar(200)" json:"location"`
	Duration        int        `json:"duration"`
	ActualStartTime *time.Time `json:"actual_start_time"`
	ActualEndTime   *time.Time `json:"actual_end_time"`
	Score           *float64   `json:"score"`
	Result          string     `gorm:"type:varchar(50)" json:"result_label"`
	ResultCode      int16      `gorm:"column:result_code;type:smallint;not null;default:0" json:"result"`
	ResultComment   string     `gorm:"type:text;not null;default:''" json:"result_comment"`
	Notes           string     `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (Interview) TableName() string { return "interviews" }

// Interviewer 面试官关联。列名 interviewer_id 沿用 000011。
type Interviewer struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	InterviewID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_interview_interviewer" json:"interview_id"`
	InterviewerID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_interview_interviewer" json:"evaluator_id"`
	CreatedAt     time.Time `json:"created_at"`
}

func (Interviewer) TableName() string { return "interview_interviewers" }

// InterviewWithNames 面试联表。
type InterviewWithNames struct {
	Interview
	ApplicantName  string     `gorm:"column:applicant_name"`
	StudentNo      string     `gorm:"column:student_no"`
	SessionTitle   string     `gorm:"column:session_title"`
	DepartmentID   *uuid.UUID `gorm:"column:department_id"`
	DepartmentName string     `gorm:"column:department_name"`
}

// InterviewerNamed 面试官带姓名。
type InterviewerNamed struct {
	InterviewID   uuid.UUID
	InterviewerID uuid.UUID
	Name          string
}
