package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ParentID  *uuid.UUID     `gorm:"type:uuid" json:"parent_id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Slug      string         `gorm:"type:varchar(80);not null" json:"slug"`
	SortOrder int            `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Category) TableName() string { return "knowledge_categories" }

type Doc struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Kind           string         `gorm:"type:varchar(16);not null" json:"kind"`
	Slug           string         `gorm:"type:varchar(80);not null" json:"slug"`
	Title          string         `gorm:"type:varchar(200);not null" json:"title"`
	Summary        string         `gorm:"type:varchar(500);not null;default:''" json:"summary"`
	Content        string         `gorm:"type:text;not null;default:''" json:"content"`
	CategoryID     *uuid.UUID     `gorm:"type:uuid" json:"category_id"`
	Visibility     string         `gorm:"type:varchar(32);not null;default:authenticated" json:"visibility"`
	PermissionCode string         `gorm:"type:varchar(64)" json:"permission_code"`
	Status         int16          `gorm:"type:smallint;not null;default:0" json:"status"`
	Version        int            `gorm:"not null;default:1" json:"version"`
	AuthorID       uuid.UUID      `gorm:"type:uuid;not null" json:"author_id"`
	PublishedAt    *time.Time     `json:"published_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Doc) TableName() string { return "knowledge_docs" }

type DocNamed struct {
	Doc
	AuthorName   string `gorm:"column:author_name" json:"author_name"`
	CategoryName string `gorm:"column:category_name" json:"category_name"`
}

type DocVersion struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DocID     uuid.UUID `gorm:"type:uuid;not null" json:"doc_id"`
	Version   int       `gorm:"not null" json:"version"`
	Title     string    `gorm:"type:varchar(200);not null" json:"title"`
	Summary   string    `gorm:"type:varchar(500);not null;default:''" json:"summary"`
	Content   string    `gorm:"type:text;not null;default:''" json:"content"`
	EditorID  uuid.UUID `gorm:"type:uuid;not null" json:"editor_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (DocVersion) TableName() string { return "knowledge_doc_versions" }

type VersionNamed struct {
	DocVersion
	EditorName string `gorm:"column:editor_name" json:"editor_name"`
}

type Attachment struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DocID     uuid.UUID `gorm:"type:uuid;not null" json:"doc_id"`
	FileID    uuid.UUID `gorm:"type:uuid;not null" json:"file_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (Attachment) TableName() string { return "knowledge_attachments" }

type AttachmentNamed struct {
	Attachment
	FileName string `gorm:"column:file_name" json:"file_name"`
	FileSize int64  `gorm:"column:file_size" json:"file_size"`
}

type NamedUser struct {
	ID       uuid.UUID
	RealName string
	Username string
}
