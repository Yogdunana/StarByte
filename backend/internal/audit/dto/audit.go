package dto

import "time"

// ========== 请求 DTO ==========

// ListAuditLogRequest 审计日志列表请求
type ListAuditLogRequest struct {
	Page      int        `form:"page,default=1" binding:"min=1"`
	PageSize  int        `form:"page_size,default=20" binding:"min=1,max=100"`
	Username  string     `form:"username"`
	Operation string     `form:"operation"` // 操作类型，如 "POST /api/v1/users"
	Method    string     `form:"method"`    // HTTP 方法
	Path      string     `form:"path"`      // 请求路径（模糊匹配）
	IP        string     `form:"ip"`
	RequestID string     `form:"request_id"`
	StatusMin *int       `form:"status_min"` // 响应状态码范围：最小值
	StatusMax *int       `form:"status_max"` // 响应状态码范围：最大值
	StartTime *time.Time `form:"start_time" time_format:"2006-01-02T15:04:05Z07:00"`
	EndTime   *time.Time `form:"end_time" time_format:"2006-01-02T15:04:05Z07:00"`
}

// ExportAuditLogRequest 导出审计日志请求
type ExportAuditLogRequest struct {
	Format    string     `form:"format" binding:"required,oneof=csv json"` // csv 或 json
	Username  string     `form:"username"`
	Operation string     `form:"operation"`
	Method    string     `form:"method"`
	Path      string     `form:"path"`
	IP        string     `form:"ip"`
	StatusMin *int       `form:"status_min"`
	StatusMax *int       `form:"status_max"`
	StartTime *time.Time `form:"start_time" time_format:"2006-01-02T15:04:05Z07:00"`
	EndTime   *time.Time `form:"end_time" time_format:"2006-01-02T15:04:05Z07:00"`
}

// ArchiveRequest 归档请求
type ArchiveRequest struct {
	BeforeDays int `json:"before_days"` // 归档多少天前的日志，默认 90
}

// ========== 响应 DTO ==========

// AuditLogResponse 审计日志响应
type AuditLogResponse struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	Operation      string `json:"operation"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	IP             string `json:"ip"`
	UserAgent      string `json:"user_agent"`
	RequestParams  string `json:"request_params"`
	ResponseStatus int    `json:"response_status"`
	ResponseBody   string `json:"response_body"`
	DurationMs     int    `json:"duration_ms"`
	RequestID      string `json:"request_id"`
	IsArchived     bool   `json:"is_archived"`
	CreatedAt      string `json:"created_at"`
}

// AuditLogListResponse 审计日志列表项响应（简化版，不含大字段）
type AuditLogListResponse struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	Operation      string `json:"operation"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	IP             string `json:"ip"`
	ResponseStatus int    `json:"response_status"`
	DurationMs     int    `json:"duration_ms"`
	RequestID      string `json:"request_id"`
	CreatedAt      string `json:"created_at"`
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
