package dto

import "time"

type CreateRecordRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	Title       string `json:"title" binding:"required,max=200"`
	Description string `json:"description"`
	Level       int16  `json:"level" binding:"required,oneof=1 2 3 4 5"`
}

type UpdateRecordRequest struct {
	Title       *string `json:"title" binding:"omitempty,max=200"`
	Description *string `json:"description"`
	Level       *int16  `json:"level" binding:"omitempty,oneof=1 2 3 4 5"`
}

type ListRecordRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Keyword  string `form:"keyword"`
	Status   *int16 `form:"status"`
	Level    *int16 `form:"level"`
	UserID   string `form:"user_id"`
}

type ApproveRequest struct {
	Comment string `json:"comment"`
}

type RevokeRequest struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

type AppealRequest struct {
	Reason string `json:"reason" binding:"required,max=2000"`
}

type Person struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AppealResponse struct {
	ID         string     `json:"id"`
	Reason     string     `json:"reason"`
	Status     int16      `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
}

type RecordResponse struct {
	ID             string           `json:"id"`
	User           Person           `json:"user"`
	Title          string           `json:"title"`
	Description    string           `json:"description"`
	Level          int16            `json:"level"`
	Status         int16            `json:"status"`
	IssuedBy       *Person          `json:"issued_by,omitempty"`
	IssuedAt       time.Time        `json:"issued_at"`
	FlowInstanceID string           `json:"flow_instance_id,omitempty"`
	ApprovedAt     *time.Time       `json:"approved_at,omitempty"`
	RevokeReason   string           `json:"revoke_reason,omitempty"`
	RevokedAt      *time.Time       `json:"revoked_at,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	Appeals        []AppealResponse `json:"appeals,omitempty"`
}
