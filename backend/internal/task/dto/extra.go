package dto

import "time"

type CommentRequest struct {
	Content  string   `json:"content" binding:"required,max=2000"`
	Mentions []string `json:"mentions"`
}

type CommentResponse struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	UserID    string    `json:"user_id"`
	User      Person    `json:"user"`
	Content   string    `json:"content"`
	Mentions  []string  `json:"mentions"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LogResponse struct {
	ID         string    `json:"id"`
	ActionType string    `json:"action_type"`
	OldValue   string    `json:"old_value"`
	NewValue   string    `json:"new_value"`
	Operator   Person    `json:"operator"`
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
}

type AttachmentResponse struct {
	ID         string    `json:"id"`
	TaskID     string    `json:"task_id"`
	FileID     string    `json:"file_id"`
	FileName   string    `json:"file_name"`
	FilePath   string    `json:"file_path"`
	FileSize   int64     `json:"file_size"`
	FileType   string    `json:"file_type"`
	UploadedBy string    `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
}

type StatsResponse struct {
	Total      int64            `json:"total"`
	Overdue    int64            `json:"overdue"`
	ByStatus   map[string]int64 `json:"by_status"`
	ByPriority map[string]int64 `json:"by_priority"`
}
