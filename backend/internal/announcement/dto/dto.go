package dto

import "time"

type Person struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Attachment struct {
	FileID string `json:"file_id"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
}

type CreateAnnouncementRequest struct {
	Title        string       `json:"title" binding:"required,max=200"`
	Content      string       `json:"content"`
	ContentType  string       `json:"content_type" binding:"omitempty,oneof=markdown html"`
	Category     string       `json:"category" binding:"required"`
	Pinned       bool         `json:"pinned"`
	Required     bool         `json:"required"`
	SortOrder    *int         `json:"sort_order"`
	ScheduledAt  *time.Time   `json:"scheduled_at"`
	ExpiresAt    *time.Time   `json:"expires_at"`
	AudienceType string       `json:"audience_type" binding:"omitempty,oneof=all role department users"`
	AudienceIDs  []string     `json:"audience_ids"`
	Attachments  []Attachment `json:"attachments"`
}

type UpdateAnnouncementRequest struct {
	Title        *string      `json:"title"`
	Content      *string      `json:"content"`
	ContentType  *string      `json:"content_type"`
	Category     *string      `json:"category"`
	Required     *bool        `json:"required"`
	SortOrder    *int         `json:"sort_order"`
	ScheduledAt  *time.Time   `json:"scheduled_at"`
	ClearSched   bool         `json:"clear_scheduled_at"`
	ExpiresAt    *time.Time   `json:"expires_at"`
	ClearExpires bool         `json:"clear_expires_at"`
	AudienceType *string      `json:"audience_type"`
	AudienceIDs  []string     `json:"audience_ids"`
	Attachments  []Attachment `json:"attachments"`
}

type PinRequest struct {
	Pinned    *bool `json:"pinned"`
	SortOrder *int  `json:"sort_order"`
}

type MarkReadRequest struct {
	DurationSeconds *int `json:"duration_seconds"`
}

type ListAnnouncementRequest struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	Status     *int16 `form:"status"`
	Category   string `form:"category"`
	Keyword    string `form:"keyword"`
	UnreadOnly bool   `form:"unread_only"`
	PinnedOnly bool   `form:"pinned_only"`
}

type AnnouncementResponse struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Content      string       `json:"content"`
	ContentType  string       `json:"content_type"`
	Category     string       `json:"category"`
	Pinned       bool         `json:"pinned"`
	Required     bool         `json:"required"`
	SortOrder    int          `json:"sort_order"`
	Status       int16        `json:"status"`
	ScheduledAt  string       `json:"scheduled_at,omitempty"`
	ExpiresAt    string       `json:"expires_at,omitempty"`
	PublishedAt  string       `json:"published_at,omitempty"`
	AudienceType string       `json:"audience_type"`
	AudienceIDs  []string     `json:"audience_ids"`
	Attachments  []Attachment `json:"attachments"`
	Author       Person       `json:"author"`
	IsRead       bool         `json:"is_read"`
	CreatedAt    string       `json:"created_at"`
	UpdatedAt    string       `json:"updated_at"`
}

type UnreadCountResponse struct {
	Count int64 `json:"count"`
}

type ReaderResponse struct {
	User            Person `json:"user"`
	ReadAt          string `json:"read_at"`
	DurationSeconds int    `json:"duration_seconds"`
}

type ReadStatusResponse struct {
	AnnouncementID     string           `json:"announcement_id"`
	ReadCount          int64            `json:"read_count"`
	UnreadCount        int64            `json:"unread_count"`
	ReadRate           float64          `json:"read_rate"`
	AvgDurationSeconds float64          `json:"avg_duration_seconds"`
	Readers            []ReaderResponse `json:"readers"`
	UnreadUsers        []Person         `json:"unread_users"`
}
