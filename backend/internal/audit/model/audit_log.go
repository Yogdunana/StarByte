package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	ActionCreate = "CREATE"
	ActionUpdate = "UPDATE"
	ActionDelete = "DELETE"
	ActionLogin  = "LOGIN"
	ActionLogout = "LOGOUT"

	MaxExportRows       = 10000
	DefaultArchiveDays  = 90
	DefaultIterateBatch = 500
	DefaultListPageSize = 20
)

// AuditLog 审计日志模型，对应 audit_logs 表。
type AuditLog struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID         *uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	Username       string     `gorm:"type:varchar(50)" json:"username"`
	RealName       string     `gorm:"type:varchar(50)" json:"real_name"`
	Operation      string     `gorm:"type:varchar(100);not null;index" json:"operation"`
	Method         string     `gorm:"type:varchar(10)" json:"method"`
	Path           string     `gorm:"type:varchar(500)" json:"path"`
	Module         string     `gorm:"type:varchar(50);index" json:"module"`
	Action         string     `gorm:"type:varchar(20);index" json:"action"`
	IP             string     `gorm:"type:varchar(50);index" json:"ip"`
	UserAgent      string     `gorm:"type:varchar(500)" json:"user_agent"`
	RequestParams  string     `gorm:"type:text" json:"request_params"`
	ResponseStatus int        `gorm:"type:int" json:"response_status"`
	ResponseBody   string     `gorm:"type:text" json:"response_body"`
	DurationMs     int        `gorm:"type:int" json:"duration_ms"`
	RequestID      string     `gorm:"type:varchar(100)" json:"request_id"`
	CreatedAt      time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;index" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

// AuditEntry 写入审计日志的领域条目。
type AuditEntry struct {
	UserID       uuid.UUID
	Username     string
	RealName     string
	Method       string
	Path         string
	Module       string
	Action       string
	RequestBody  []byte
	ResponseCode int
	IPAddress    string
	UserAgent    string
	Duration     int64
	RequestID    string
	Timestamp    time.Time
}

// AuditLogArchive 归档记录，对应 audit_log_archives 表。
type AuditLogArchive struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ArchiveDate string    `gorm:"type:varchar(10);index" json:"archive_date"`
	RecordCount int64     `gorm:"type:bigint" json:"record_count"`
	MinIOObject string    `gorm:"type:varchar(500)" json:"minio_object"`
	Status      int       `gorm:"type:smallint;default:0" json:"status"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (AuditLogArchive) TableName() string {
	return "audit_log_archives"
}
