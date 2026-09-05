package dto

import "time"

type Person struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar,omitempty"`
}

type CreateInternshipRequest struct {
	Title        string     `json:"title" binding:"required,max=200"`
	Organization string     `json:"organization" binding:"required,max=200"`
	Description  string     `json:"description"`
	StartDate    time.Time  `json:"start_date" binding:"required"`
	EndDate      *time.Time `json:"end_date"`
	Type         int16      `json:"type" binding:"oneof=0 1 2"`
	UserID       string     `json:"user_id"`
	MentorID     string     `json:"mentor_id"`
	Skills       []string   `json:"skills"`
	Achievements string     `json:"achievements"`
}

type UpdateInternshipRequest struct {
	Title        *string    `json:"title"`
	Organization *string    `json:"organization"`
	Description  *string    `json:"description"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	ClearEndDate bool       `json:"clear_end_date"`
	Type         *int16     `json:"type" binding:"omitempty,oneof=0 1 2"`
	Skills       []string   `json:"skills"`
	Achievements *string    `json:"achievements"`
	MentorID     *string    `json:"mentor_id"`
}

type ListInternshipRequest struct {
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
	Status       *int16 `form:"status"`
	Type         *int16 `form:"type"`
	DepartmentID string `form:"department_id"`
	UserID       string `form:"user_id"`
	Keyword      string `form:"keyword"`
}

type MyInternshipRequest struct {
	Status *int16 `form:"status"`
}

type CompleteRequest struct {
	Report       string `json:"report"`
	Achievements string `json:"achievements"`
}

type ReportRequest struct {
	Report string `json:"report" binding:"required"`
}

type MentorCommentRequest struct {
	MentorComment string `json:"mentor_comment" binding:"required"`
}

type InternshipConfigRequest struct {
	AllowStudentEdit  *bool `json:"allow_student_edit"`
	AllowMinisterEdit *bool `json:"allow_minister_edit"`
	RankingVisible    *bool `json:"ranking_visible"`
}

type InternshipResponse struct {
	ID            string     `json:"id"`
	User          Person     `json:"user"`
	Title         string     `json:"title"`
	Organization  string     `json:"organization"`
	Description   string     `json:"description"`
	StartDate     time.Time  `json:"start_date"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	Status        int16      `json:"status"`
	Type          int16      `json:"type"`
	Mentor        *Person    `json:"mentor,omitempty"`
	Supervisor    *Person    `json:"supervisor,omitempty"`
	Department    *Person    `json:"department,omitempty"`
	Skills        []string   `json:"skills"`
	Achievements  string     `json:"achievements"`
	Report        string     `json:"report,omitempty"`
	MentorComment string     `json:"mentor_comment,omitempty"`
	DurationDays  int        `json:"duration_days"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type InternshipConfigResponse struct {
	AllowStudentEdit  bool `json:"allow_student_edit"`
	AllowMinisterEdit bool `json:"allow_minister_edit"`
	RankingVisible    bool `json:"ranking_visible"`
}
