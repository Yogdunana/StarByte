package dto

import (
	"time"

	"github.com/google/uuid"
)

type SendEmailRequest struct {
	To             []string `json:"to" binding:"required,min=1,dive,email"`
	CC             []string `json:"cc" binding:"omitempty,dive,email"`
	Subject        string   `json:"subject" binding:"required,max=200"`
	Body           string   `json:"body" binding:"required"`
	IsHTML         bool     `json:"is_html"`
	AttachmentIDs  []string `json:"attachment_ids"`
	TemplateCode   string   `json:"template_code"`
	NotificationID *string  `json:"notification_id"`
}

type SendEmailResponse struct {
	MessageID      string `json:"message_id"`
	Status         string `json:"status"`
	RecipientCount int    `json:"recipient_count"`
}

type BatchRecipient struct {
	Email string            `json:"email" binding:"required,email"`
	Data  map[string]string `json:"data"`
}

type BatchSendEmailRequest struct {
	Recipients   []BatchRecipient `json:"recipients" binding:"required,min=1,max=50"`
	TemplateCode string           `json:"template_code"`
	Subject      string           `json:"subject" binding:"omitempty,max=200"`
	Body         string           `json:"body"`
	IsHTML       bool             `json:"is_html"`
}

type BatchSendEmailResponse struct {
	Total  int `json:"total"`
	Queued int `json:"queued"`
	Failed int `json:"failed"`
}

type ListEmailLogsRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	Status    string `form:"status"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

type EmailLogResponse struct {
	ID           string     `json:"id"`
	To           string     `json:"to"`
	CC           string     `json:"cc,omitempty"`
	Subject      string     `json:"subject"`
	TemplateCode string     `json:"template_code,omitempty"`
	Status       string     `json:"status"`
	ErrorMessage string     `json:"error_message,omitempty"`
	RetryCount   int16      `json:"retry_count"`
	SentAt       *time.Time `json:"sent_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

func ParseLogStatus(name string) (int16, bool) {
	switch name {
	case "", "all":
		return -1, true
	case "queued":
		return 0, true
	case "sent":
		return 1, true
	case "failed":
		return 2, true
	case "retrying":
		return 3, true
	default:
		return 0, false
	}
}

func ParseOptionalUUID(raw string) (*uuid.UUID, error) {
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
