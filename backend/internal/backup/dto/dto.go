package dto

// Record is a backup job as returned by the API.
type Record struct {
	ID             string  `json:"id"`
	TriggerSource  string  `json:"trigger_source"`
	Status         int16   `json:"status"`
	Storage        string  `json:"storage"`
	ObjectKey      string  `json:"object_key"`
	Filename       string  `json:"filename"`
	ChecksumSHA256 string  `json:"checksum_sha256"`
	SizeBytes      int64   `json:"size_bytes"`
	Encrypted      bool    `json:"encrypted"`
	StartedAt      *string `json:"started_at"`
	FinishedAt     *string `json:"finished_at"`
	ErrorMessage   string  `json:"error_message"`
	CreatedBy      string  `json:"created_by,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// ListRequest is GET /system/backups query.
type ListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   *int16 `form:"status"`
}

// CreateRequest is POST /system/backups. Phase-1 always runs async.
type CreateRequest struct {
	Async *bool `json:"async"`
}

// RestoreRequest is POST /system/backups/:id/restore.
// confirmation must be the literal RESTORE (typed confirmation).
type RestoreRequest struct {
	Confirm      bool   `json:"confirm"`
	Confirmation string `json:"confirmation"`
}

// Policy is GET/PUT /system/backups/policies.
type Policy struct {
	Enabled       bool   `json:"enabled"`
	RetentionDays int    `json:"retention_days"`
	CronExpr      string `json:"cron_expr"`
	Timezone      string `json:"timezone"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

// UpdatePolicyRequest is PUT /system/backups/policies.
type UpdatePolicyRequest struct {
	Enabled       *bool   `json:"enabled"`
	RetentionDays *int    `json:"retention_days"`
	CronExpr      *string `json:"cron_expr"`
	Timezone      *string `json:"timezone"`
}

// StorageStats is GET /system/backups/storage.
type StorageStats struct {
	Count              int64  `json:"count"`
	SizeBytes          int64  `json:"size_bytes"`
	Prefix             string `json:"prefix"`
	Bucket             string `json:"bucket"`
	LocalPath          string `json:"local_path,omitempty"`
	Compression        string `json:"compression"`
	EncryptionEnabled  bool   `json:"encryption_enabled"`
	IncrementalEnabled bool   `json:"incremental_enabled"`
	PITREnabled        bool   `json:"pitr_enabled"`
}

// Preview is GET /system/backups/:id/preview — integrity check, no restore.
type Preview struct {
	ID                   string `json:"id"`
	Filename             string `json:"filename"`
	SizeBytes            int64  `json:"size_bytes"`
	ChecksumOK           bool   `json:"checksum_ok"`
	Encrypted            bool   `json:"encrypted"`
	DecryptOK            bool   `json:"decrypt_ok"`
	GzipOK               bool   `json:"gzip_ok"`
	TOCValid             bool   `json:"toc_valid"`
	TOC                  string `json:"toc,omitempty"`
	Ready                bool   `json:"ready"`
	Compression          string `json:"compression"`
	EncryptionConfigured bool   `json:"encryption_configured"`
	Error                string `json:"error,omitempty"`
}
