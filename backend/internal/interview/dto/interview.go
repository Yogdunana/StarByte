package dto

import "time"

type CreateInterviewRequest struct {
	SessionID     string     `json:"session_id" binding:"required,uuid"`
	ApplicantID   string     `json:"applicant_id" binding:"omitempty,uuid"`
	ApplicationID string     `json:"application_id" binding:"omitempty,uuid"`
	ScheduledTime *time.Time `json:"scheduled_time"`
	Location      string     `json:"location" binding:"max=200"`
	Duration      int        `json:"duration"`
	Notes         string     `json:"notes"`
}

type ListInterviewRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	SessionID string `form:"session_id"`
	Status    *int16 `form:"status"`
	Result    *int16 `form:"result"`
	Keyword   string `form:"keyword"`
	MyRole    string `form:"role"`
}

type AssignEvaluatorsRequest struct {
	EvaluatorIDs []string `json:"evaluator_ids" binding:"required,min=1"`
}

type CheckinRequest struct {
	Token string `json:"token"`
}

type Person struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type InterviewResponse struct {
	ID              string     `json:"id"`
	SessionID       string     `json:"session_id,omitempty"`
	SessionTitle    string     `json:"session_title,omitempty"`
	Applicant       Person     `json:"applicant"`
	StudentNo       string     `json:"student_no,omitempty"`
	ApplicationID   string     `json:"application_id,omitempty"`
	Status          int16      `json:"status"`
	ScheduledTime   *time.Time `json:"scheduled_time,omitempty"`
	ActualStartTime *time.Time `json:"actual_start_time,omitempty"`
	ActualEndTime   *time.Time `json:"actual_end_time,omitempty"`
	Result          int16      `json:"result"`
	ResultComment   string     `json:"result_comment,omitempty"`
	Location        string     `json:"location,omitempty"`
	Duration        int        `json:"duration,omitempty"`
	Score           *float64   `json:"score,omitempty"`
	Evaluators      []Person   `json:"evaluators"`
	DepartmentName  string     `json:"department_name,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
