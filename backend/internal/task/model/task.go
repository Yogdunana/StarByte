package model

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Title             string     `gorm:"type:varchar(200);not null" json:"title"`
	Description       string     `gorm:"type:text" json:"description"`
	Status            int16      `gorm:"type:smallint;not null;default:0" json:"status"`
	Priority          int16      `gorm:"type:smallint;not null;default:1" json:"priority"`
	CreatorID         uuid.UUID  `gorm:"type:uuid;not null" json:"creator_id"`
	AssigneeID        *uuid.UUID `gorm:"type:uuid" json:"assignee_id"`
	DepartmentID      *uuid.UUID `gorm:"type:uuid" json:"department_id"`
	ParentID          *uuid.UUID `gorm:"type:uuid" json:"parent_id"`
	DueDate           *time.Time `json:"due_date"`
	Progress          int16      `gorm:"type:smallint;not null;default:0" json:"progress"`
	Tags              string     `gorm:"type:varchar(500)" json:"tags"`
	RelatedType       string     `gorm:"type:varchar(50)" json:"related_type"`
	RelatedID         *uuid.UUID `gorm:"type:uuid" json:"related_id"`
	SortOrder         int        `gorm:"not null;default:0" json:"sort_order"`
	CompletedAt       *time.Time `json:"completed_at"`
	DueRemindedAt     *time.Time `json:"due_reminded_at"`
	OverdueRemindedAt *time.Time `json:"overdue_reminded_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (Task) TableName() string { return "tasks" }

type TaskWithNames struct {
	Task
	CreatorName     string `gorm:"column:creator_name"`
	AssigneeName    string `gorm:"column:assignee_name"`
	AssigneeAvatar  string `gorm:"column:assignee_avatar"`
	DepartmentName  string `gorm:"column:department_name"`
	ParentTitle     string `gorm:"column:parent_title"`
	CommentCount    int64  `gorm:"column:comment_count"`
	AttachmentCount int64  `gorm:"column:attachment_count"`
}

type NamedUser struct {
	ID       uuid.UUID
	RealName string
	Username string
	Avatar   string
}

type TaskLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TaskID     uuid.UUID `gorm:"type:uuid;not null" json:"task_id"`
	ActionType string    `gorm:"type:varchar(32);not null" json:"action_type"`
	OldValue   string    `gorm:"type:varchar(500);not null;default:''" json:"old_value"`
	NewValue   string    `gorm:"type:varchar(500);not null;default:''" json:"new_value"`
	OperatorID uuid.UUID `gorm:"type:uuid;not null" json:"operator_id"`
	Comment    string    `gorm:"type:text;not null;default:''" json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
}

func (TaskLog) TableName() string { return "task_logs" }

type TaskLogNamed struct {
	TaskLog
	OperatorName string `gorm:"column:operator_name"`
}

type TaskComment struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TaskID    uuid.UUID `gorm:"type:uuid;not null" json:"task_id"`
	AuthorID  uuid.UUID `gorm:"type:uuid;not null" json:"author_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Mentions  string    `gorm:"type:text;not null;default:'[]'" json:"mentions"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (TaskComment) TableName() string { return "task_comments" }

type TaskCommentNamed struct {
	TaskComment
	AuthorName string `gorm:"column:author_name"`
}

type TaskAttachment struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	TaskID     uuid.UUID  `gorm:"type:uuid;not null" json:"task_id"`
	FileID     uuid.UUID  `gorm:"type:uuid;not null" json:"file_id"`
	UploadedBy *uuid.UUID `gorm:"type:uuid" json:"uploaded_by"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (TaskAttachment) TableName() string { return "task_attachments" }

type TaskAttachmentNamed struct {
	TaskAttachment
	FileName   string `gorm:"column:file_name"`
	FilePath   string `gorm:"column:file_path"`
	FileSize   int64  `gorm:"column:file_size"`
	FileType   string `gorm:"column:file_type"`
	UploaderID string `gorm:"column:uploader_id"`
}
