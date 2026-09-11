package dto

import "time"

type ReportResponse struct {
	ID                  string    `json:"id"`
	UserID              string    `json:"user_id"`
	DepartmentID        *string   `json:"department_id,omitempty"`
	ReportType          string    `json:"report_type"`
	PeriodStart         time.Time `json:"period_start"`
	PeriodEnd           time.Time `json:"period_end"`
	WorkContent         string    `json:"work_content"`
	NextPlan            string    `json:"next_plan"`
	ProblemsSuggestions string    `json:"problems_suggestions"`
	ReviewStatus        string    `json:"review_status"`
	SubmittedAt         time.Time `json:"submitted_at"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
