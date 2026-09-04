package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	CategoryImage    = "image"
	CategoryDocument = "document"
	CategoryVideo    = "video"

	StorageTypeMinIO = "minio"
)

// File 对应 files 表（000001 + 000009）
type File struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Name          string     `gorm:"type:varchar(255);not null" json:"name"`
	OriginalName  string     `gorm:"type:varchar(255)" json:"original_name"`
	Path          string     `gorm:"type:varchar(500);not null" json:"path"`
	Size          int64      `gorm:"type:bigint" json:"size"`
	MimeType      string     `gorm:"type:varchar(100)" json:"mime_type"`
	StorageType   string     `gorm:"type:varchar(20);default:minio" json:"storage_type"`
	Bucket        string     `gorm:"type:varchar(100)" json:"bucket"`
	UploadedBy    *uuid.UUID `gorm:"type:uuid;index" json:"uploaded_by"`
	Category      string     `gorm:"type:varchar(20);default:document;index" json:"category"`
	IsPublic      bool       `gorm:"default:false" json:"is_public"`
	ThumbnailPath string     `gorm:"type:varchar(500)" json:"thumbnail_path"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (File) TableName() string {
	return "files"
}

// FileWithUploader 列表查询时带上传者姓名
type FileWithUploader struct {
	File
	UploaderName string `gorm:"column:uploader_name" json:"uploader_name"`
}
