package dto

import (
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/form/model"
)

type ListQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Keyword  string `form:"keyword"`
	Status   *int16 `form:"status"`
}

type SubmissionQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type CreateFormRequest struct {
	Name        string            `json:"name" binding:"required,max=100"`
	Description string            `json:"description" binding:"max=500"`
	Fields      []model.FormField `json:"fields"`
	Status      *int16            `json:"status"`
}

type UpdateFormRequest struct {
	Name        *string            `json:"name" binding:"omitempty,max=100"`
	Description *string            `json:"description" binding:"omitempty,max=500"`
	Fields      *[]model.FormField `json:"fields"`
	Status      *int16             `json:"status"`
}

type FormListItem struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Status          int16     `json:"status"`
	SubmissionCount int64     `json:"submission_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type FormDetail struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Status          int16             `json:"status"`
	Fields          []model.FormField `json:"fields"`
	SubmissionCount int64             `json:"submission_count"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type SubmitResponse struct {
	SubmissionID string    `json:"submission_id"`
	SubmittedAt  time.Time `json:"submitted_at"`
}

type UserBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SubmissionItem struct {
	ID          string                 `json:"id"`
	Data        map[string]interface{} `json:"data"`
	SubmittedBy *UserBrief             `json:"submitted_by"`
	SubmittedAt time.Time              `json:"submitted_at"`
}
