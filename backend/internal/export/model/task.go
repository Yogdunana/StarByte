package model

import "time"

const (
	StatusPending = "pending"
	StatusRunning = "running"
	StatusDone    = "done"
	StatusFailed  = "failed"

	TaskTTL = 2 * time.Hour
	FileTTL = 30 * time.Minute

	KeyTaskPrefix = "export:task:"
	KeyFilePrefix = "export:file:"
	KeyBlobPrefix = "export:blob:"
)

// ExportTask is stored as JSON in Redis under export:task:{id}.
type ExportTask struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id,omitempty"`
	Status    string    `json:"status"`
	Progress  int       `json:"progress"`
	FileID    string    `json:"file_id,omitempty"`
	Error     string    `json:"error,omitempty"`
	Filename  string    `json:"filename,omitempty"`
	Format    string    `json:"format,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// FileMeta maps a downloadable export file. Bytes live in export:blob:{id}
// when object storage is unavailable.
type FileMeta struct {
	FileID      string    `json:"file_id"`
	UserID      string    `json:"user_id,omitempty"`
	ObjectKey   string    `json:"object_key,omitempty"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Ext         string    `json:"ext"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func TaskKey(id string) string { return KeyTaskPrefix + id }
func FileKey(id string) string { return KeyFilePrefix + id }
func BlobKey(id string) string { return KeyBlobPrefix + id }
