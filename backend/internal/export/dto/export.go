package dto

import "time"

// TableExportRequest is the body for Excel / CSV / PDF / JSON table exports.
type TableExportRequest struct {
	Locale    string     `json:"-"`
	Filename  string     `json:"filename"`
	Title     string     `json:"title"`
	Columns   []string   `json:"columns"`
	Rows      [][]string `json:"rows"`
	Sheet     string     `json:"sheet"`
	Delimiter string     `json:"delimiter"`
}

// TemplateExportRequest is the body for built-in HTML template printing.
type TemplateExportRequest struct {
	Locale    string            `json:"-"`
	Filename  string            `json:"filename"`
	Watermark string            `json:"watermark"`
	Vars      map[string]string `json:"vars"`
	Variables map[string]string `json:"variables"`
}

// ExportTaskResponse is returned by start-export and GET task.
type ExportTaskResponse struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	Progress  int    `json:"progress"`
	FileID    string `json:"file_id,omitempty"`
	Error     string `json:"error,omitempty"`
	Filename  string `json:"filename,omitempty"`
	Format    string `json:"format,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// TemplateInfo describes a built-in print template.
type TemplateInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// DownloadInfo is the JSON body when a presigned URL is available.
type DownloadInfo struct {
	FileID      string `json:"file_id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	URL         string `json:"url,omitempty"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
}

// DownloadResult is the service-layer download payload (not serialized as-is).
type DownloadResult struct {
	FileID      string
	Filename    string
	ContentType string
	URL         string
	Bytes       []byte
	ExpiresAt   time.Time
}
