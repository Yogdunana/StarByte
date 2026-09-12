package dto

import "time"

type Person struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateAnnouncementRequest struct {
	Title       string     `json:"title" binding:"required,max=200"`
	Content     string     `json:"content"`
	ContentType string     `json:"content_type" binding:"omitempty,oneof=markdown html"`
	Category    string     `json:"category" binding:"required"`
	Pinned      bool       `json:"pinned"`
	Required    bool       `json:"required"`
	ScheduledAt *time.Time `json:"scheduled_at"`
}

type UpdateAnnouncementRequest struct {
	Title       *string    `json:"title"`
	Content     *string    `json:"content"`
	ContentType *string    `json:"content_type"`
	Category    *string    `json:"category"`
	Required    *bool      `json:"required"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	ClearSched  bool       `json:"clear_scheduled_at"`
}

type PinRequest struct {
	Pinned *bool `json:"pinned"`
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
	ID          string `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	Category    string `json:"category"`
	Pinned      bool   `json:"pinned"`
	Required    bool   `json:"required"`
	Status      int16  `json:"status"`
	ScheduledAt string `json:"scheduled_at,omitempty"`
	PublishedAt string `json:"published_at,omitempty"`
	Author      Person `json:"author"`
	IsRead      bool   `json:"is_read"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type UnreadCountResponse struct {
	Count int64 `json:"count"`
}

type ReaderResponse struct {
	User   Person `json:"user"`
	ReadAt string `json:"read_at"`
}

type ReadStatusResponse struct {
	AnnouncementID string           `json:"announcement_id"`
	ReadCount      int64            `json:"read_count"`
	UnreadCount    int64            `json:"unread_count"`
	Readers        []ReaderResponse `json:"readers"`
}
