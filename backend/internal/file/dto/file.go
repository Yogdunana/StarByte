package dto

import "time"

// UploaderInfo 上传者摘要
type UploaderInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// FileUploadResponse 单文件上传响应（Issue #18 + 前端 name/size 别名）
type FileUploadResponse struct {
	ID           string    `json:"id"`
	Filename     string    `json:"filename"`
	Name         string    `json:"name"`
	OriginalName string    `json:"original_name"`
	FileSize     int64     `json:"file_size"`
	Size         int64     `json:"size"`
	MimeType     string    `json:"mime_type"`
	Category     string    `json:"category"`
	URL          string    `json:"url"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

// FileListItem 文件列表项
type FileListItem struct {
	ID           string       `json:"id"`
	Filename     string       `json:"filename"`
	Name         string       `json:"name"`
	OriginalName string       `json:"original_name"`
	FileSize     int64        `json:"file_size"`
	Size         int64        `json:"size"`
	MimeType     string       `json:"mime_type"`
	Category     string       `json:"category"`
	URL          string       `json:"url"`
	ThumbnailURL string       `json:"thumbnail_url,omitempty"`
	Uploader     UploaderInfo `json:"uploader"`
	UploaderID   string       `json:"uploader_id"`
	UploaderName string       `json:"uploader_name"`
	CreatedAt    time.Time    `json:"created_at"`
}

// FileDetailResponse 文件详情
type FileDetailResponse struct {
	ID           string       `json:"id"`
	Filename     string       `json:"filename"`
	Name         string       `json:"name"`
	OriginalName string       `json:"original_name"`
	FileSize     int64        `json:"file_size"`
	Size         int64        `json:"size"`
	MimeType     string       `json:"mime_type"`
	Category     string       `json:"category"`
	StorageType  string       `json:"storage_type"`
	Bucket       string       `json:"bucket"`
	Path         string       `json:"path"`
	URL          string       `json:"url"`
	ThumbnailURL string       `json:"thumbnail_url,omitempty"`
	IsPublic     bool         `json:"is_public"`
	Uploader     UploaderInfo `json:"uploader"`
	UploaderID   string       `json:"uploader_id"`
	UploaderName string       `json:"uploader_name"`
	CreatedAt    time.Time    `json:"created_at"`
	UploadedAt   time.Time    `json:"uploaded_at"`
}

// ListFilesRequest 文件列表查询
type ListFilesRequest struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	Category   string `form:"category"`
	Keyword    string `form:"keyword"`
	UploaderID string `form:"uploader_id"`
	MimeType   string `form:"mime_type"`
}
