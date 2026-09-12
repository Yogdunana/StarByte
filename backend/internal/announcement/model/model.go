package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Announcement 公告
type Announcement struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Title       string         `gorm:"type:varchar(200);not null" json:"title"`
	Content     string         `gorm:"type:text;not null;default:''" json:"content"`
	ContentType string         `gorm:"type:varchar(16);not null;default:markdown" json:"content_type"`
	Category    string         `gorm:"type:varchar(32);not null" json:"category"`
	Pinned      bool           `gorm:"not null;default:false" json:"pinned"`
	Required    bool           `gorm:"not null;default:false" json:"required"`
	Status      int16          `gorm:"type:smallint;not null;default:0" json:"status"`
	ScheduledAt *time.Time     `json:"scheduled_at"`
	PublishedAt *time.Time     `json:"published_at"`
	AuthorID    uuid.UUID      `gorm:"type:uuid;not null" json:"author_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Announcement) TableName() string { return "announcements" }

// AnnouncementNamed 公告 + 作者名 + 当前用户是否已读
type AnnouncementNamed struct {
	Announcement
	AuthorName string `gorm:"column:author_name" json:"author_name"`
	IsRead     bool   `gorm:"column:is_read" json:"is_read"`
}

// AnnouncementRead 阅读回执
type AnnouncementRead struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	AnnouncementID uuid.UUID `gorm:"type:uuid;not null" json:"announcement_id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	ReadAt         time.Time `json:"read_at"`
}

func (AnnouncementRead) TableName() string { return "announcement_reads" }

// ReaderNamed 已读人员
type ReaderNamed struct {
	UserID   uuid.UUID `gorm:"column:user_id"`
	RealName string    `gorm:"column:real_name"`
	Username string    `gorm:"column:username"`
	ReadAt   time.Time `gorm:"column:read_at"`
}

// NamedUser 关联查询用的用户
type NamedUser struct {
	ID       uuid.UUID
	RealName string
	Username string
}
