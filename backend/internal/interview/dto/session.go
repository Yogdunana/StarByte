package dto

import "time"

type CreateSessionRequest struct {
	Title         string    `json:"title" binding:"required,max=200"`
	Round         int       `json:"round" binding:"required,min=1,max=20"`
	DepartmentID  string    `json:"department_id" binding:"omitempty,uuid"`
	StartTime     time.Time `json:"start_time" binding:"required"`
	EndTime       time.Time `json:"end_time" binding:"required"`
	Location      string    `json:"location" binding:"max=200"`
	OnlineLink    string    `json:"online_link" binding:"max=500"`
	MaxCandidates int       `json:"max_candidates" binding:"required,min=1,max=500"`
	Description   string    `json:"description"`
}

type UpdateSessionRequest struct {
	Title         *string    `json:"title"`
	Round         *int       `json:"round"`
	DepartmentID  *string    `json:"department_id"`
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	Location      *string    `json:"location"`
	OnlineLink    *string    `json:"online_link"`
	MaxCandidates *int       `json:"max_candidates"`
	Description   *string    `json:"description"`
}

type ListSessionRequest struct {
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
	Status       *int16 `form:"status"`
	DepartmentID string `form:"department_id"`
	Round        *int16 `form:"round"`
	Keyword      string `form:"keyword"`
}

type SessionResponse struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Round          int16     `json:"round"`
	DepartmentID   string    `json:"department_id,omitempty"`
	DepartmentName string    `json:"department_name,omitempty"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Location       string    `json:"location"`
	OnlineLink     string    `json:"online_link,omitempty"`
	Status         int16     `json:"status"`
	MaxCandidates  int       `json:"max_candidates"`
	Description    string    `json:"description"`
	CandidateCount int64     `json:"candidate_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type QRCodeResponse struct {
	SessionID   string `json:"session_id"`
	Token       string `json:"token"`
	CheckinPath string `json:"checkin_path"`
	PNGBase64   string `json:"png_base64"`
}
