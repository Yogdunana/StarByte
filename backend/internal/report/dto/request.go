package dto

import "time"

type CreateReportRequest struct {
	ReportType          string     `json:"report_type" binding:"required,oneof=daily weekly monthly"`
	PeriodStart         *time.Time `json:"period_start" binding:"required,ltefield=PeriodEnd"`
	PeriodEnd           *time.Time `json:"period_end" binding:"required"`
	WorkContent         string     `json:"work_content"`
	NextPlan            string     `json:"next_plan"`
	ProblemsSuggestions string     `json:"problems_suggestions"`
}

type UpdateReportRequest struct {
	WorkContent         *string `json:"work_content" binding:"required_without_all=NextPlan ProblemsSuggestions"`
	NextPlan            *string `json:"next_plan" binding:"required_without_all=WorkContent ProblemsSuggestions"`
	ProblemsSuggestions *string `json:"problems_suggestions" binding:"required_without_all=WorkContent NextPlan"`
}

type ReviewReportRequest struct {
	ReviewStatus string `json:"review_status" binding:"required,oneof=approved rejected"`
}

type ListReportRequest struct {
	Page         int        `form:"page" binding:"omitempty,min=1"`
	PageSize     int        `form:"page_size" binding:"omitempty,min=1,max=100"`
	ReportType   string     `form:"report_type" binding:"omitempty,oneof=daily weekly monthly"`
	ReviewStatus string     `form:"review_status" binding:"omitempty,oneof=pending approved rejected"`
	UserID       string     `form:"user_id" binding:"omitempty,uuid"`
	DepartmentID string     `form:"department_id" binding:"omitempty,uuid"`
	PeriodStart  *time.Time `form:"period_start" time_format:"2006-01-02"`
	PeriodEnd    *time.Time `form:"period_end" time_format:"2006-01-02"`
}
