package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLog 审计日志模型，对应 audit_logs 表。
// 由 AuditLog 中间件在每次写操作（POST/PUT/PATCH/DELETE）时异步写入。
type AuditLog struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID         *uuid.UUID     `gorm:"type:uuid;index" json:"user_id"`
	Username       string         `gorm:"type:varchar(50)" json:"username"`
	Operation      string         `gorm:"type:varchar(100);not null;index" json:"operation"`
	Method         string         `gorm:"type:varchar(10)" json:"method"`
	Path           string         `gorm:"type:varchar(500)" json:"path"`
	IP             string         `gorm:"type:varchar(50);index" json:"ip"`
	UserAgent      string         `gorm:"type:varchar(500)" json:"user_agent"`
	RequestParams  string         `gorm:"type:text" json:"request_params"`
	ResponseStatus int            `gorm:"type:int" json:"response_status"`
	ResponseBody   string         `gorm:"type:text" json:"response_body"`
	DurationMs     int            `gorm:"type:int" json:"duration_ms"`
	RequestID      string         `gorm:"type:varchar(100)" json:"request_id"`
	IsArchived     bool           `gorm:"type:boolean;default:false;index" json:"is_archived"`
	ArchivedAt     *time.Time     `gorm:"type:timestamp" json:"archived_at,omitempty"`
	CreatedAt      time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;index" json:"created_at"`
}

// TableName 表名
func (AuditLog) TableName() string {
	return "audit_logs"
}

// AuditLogArchive 归档日志模型，对应 audit_log_archives 表。
// 由归档定时任务将超过保留期限的日志归档到此表（并可导出至 MinIO）。
type AuditLogArchive struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ArchiveDate   string         `gorm:"type:varchar(10);index" json:"archive_date"` // YYYY-MM-DD
	RecordCount   int64          `gorm:"type:bigint" json:"record_count"`
	MinIOObject   string         `gorm:"type:varchar(500)" json:"minio_object"` // MinIO 对象路径
	Status        int            `gorm:"type:smallint;default:0" json:"status"` // 0=进行中 1=成功 2=失败
	CreatedAt     time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName 表名
func (AuditLogArchive) TableName() string {
	return "audit_log_archives"
}

// 确保 AuditLogArchive 实现 gorm.Tabler 接口
var _ gorm.Tabler = (*AuditLogArchive)(nil)
