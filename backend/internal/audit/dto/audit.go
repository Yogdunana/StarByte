package dto

import "time"

// ListAuditLogRequest 审计日志列表请求
type ListAuditLogRequest struct {
	Page      int        `form:"page,default=1" binding:"min=1"`
	PageSize  int        `form:"page_size,default=20" binding:"min=1,max=100"`
	StartTime *time.Time `form:"start_time" time_format:"2006-01-02T15:04:05Z07:00"`
	EndTime   *time.Time `form:"end_time" time_format:"2006-01-02T15:04:05Z07:00"`
	UserID    string     `form:"user_id"`
	Username  string     `form:"username"`
	Action    string     `form:"action"`
	Module    string     `form:"module"`
	Keyword   string     `form:"keyword"`
	IPAddress string     `form:"ip_address"`
	Method    string     `form:"method"`
}

// ExportAuditLogRequest 导出请求（筛选条件同列表）
type ExportAuditLogRequest struct {
	Format    string     `form:"format" binding:"required,oneof=csv excel"`
	StartTime *time.Time `form:"start_time" time_format:"2006-01-02T15:04:05Z07:00"`
	EndTime   *time.Time `form:"end_time" time_format:"2006-01-02T15:04:05Z07:00"`
	UserID    string     `form:"user_id"`
	Username  string     `form:"username"`
	Action    string     `form:"action"`
	Module    string     `form:"module"`
	Keyword   string     `form:"keyword"`
	IPAddress string     `form:"ip_address"`
	Method    string     `form:"method"`
}

// ArchiveRequest 归档请求
type ArchiveRequest struct {
	BeforeDays int `json:"before_days"`
}

// AuditUser 操作人
type AuditUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	RealName string `json:"real_name"`
}

// AuditLogResponse 审计日志详情
type AuditLogResponse struct {
	ID           string    `json:"id"`
	User         AuditUser `json:"user"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	Module       string    `json:"module"`
	Action       string    `json:"action"`
	RequestBody  string    `json:"request_body"`
	ResponseCode int       `json:"response_code"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	DurationMs   int       `json:"duration_ms"`
	Timestamp    string    `json:"timestamp"`
}

// AuditLogListResponse 列表项
type AuditLogListResponse struct {
	ID           string    `json:"id"`
	User         AuditUser `json:"user"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	Module       string    `json:"module"`
	Action       string    `json:"action"`
	RequestBody  string    `json:"request_body"`
	ResponseCode int       `json:"response_code"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	DurationMs   int       `json:"duration_ms"`
	Timestamp    string    `json:"timestamp"`
}

// ArchiveResponse 归档响应
type ArchiveResponse struct {
	ArchiveID   string `json:"archive_id"`
	RecordCount int64  `json:"record_count"`
	ArchiveDate string `json:"archive_date"`
	MinIOObject string `json:"minio_object"`
	Status      int    `json:"status"`
	Message     string `json:"message"`
}
