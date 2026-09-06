package dto

import "time"

// TraceQueryRequest 按实体查询变更历史
type TraceQueryRequest struct {
	Page     int `form:"page,default=1" binding:"min=1"`
	PageSize int `form:"page_size,default=20" binding:"min=1,max=100"`
}

// ReportRequest 合规报告
type ReportRequest struct {
	StartTime *time.Time `form:"start_time" time_format:"2006-01-02T15:04:05Z07:00"`
	EndTime   *time.Time `form:"end_time" time_format:"2006-01-02T15:04:05Z07:00"`
	Format    string     `form:"format"`
	Module    string     `form:"module"`
}

// ArchiveListRequest 归档列表
type ArchiveListRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=20" binding:"min=1,max=100"`
	ID       string `form:"id"`
	Keyword  string `form:"keyword"`
}

// FieldChange 字段级 Diff
type FieldChange struct {
	Path   string `json:"path"`
	Before any    `json:"before"`
	After  any    `json:"after"`
}

// AuditTraceItem 实体变更历史项
type AuditTraceItem struct {
	ID              string        `json:"id"`
	User            AuditUser     `json:"user"`
	Action          string        `json:"action"`
	Method          string        `json:"method"`
	Path            string        `json:"path"`
	Module          string        `json:"module"`
	EntityType      string        `json:"entity_type"`
	EntityID        string        `json:"entity_id"`
	BeforeJSON      string        `json:"before_json"`
	AfterJSON       string        `json:"after_json"`
	Diff            []FieldChange `json:"diff"`
	ComplianceFlags []string      `json:"compliance_flags"`
	Timestamp       string        `json:"timestamp"`
}

// CountItem 聚合计数
type CountItem struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

// ReportResponse 合规报告 JSON
type ReportResponse struct {
	StartTime    string      `json:"start_time"`
	EndTime      string      `json:"end_time"`
	Total        int64       `json:"total"`
	ByAction     []CountItem `json:"by_action"`
	ByModule     []CountItem `json:"by_module"`
	ByCompliance []CountItem `json:"by_compliance"`
	TopOperators []CountItem `json:"top_operators"`
	Note         string      `json:"note"`
}

// ArchiveListItem 归档列表项
type ArchiveListItem struct {
	ID          string `json:"id"`
	ArchiveDate string `json:"archive_date"`
	RecordCount int64  `json:"record_count"`
	MinIOObject string `json:"minio_object"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// ArchivePullResponse 归档拉取
type ArchivePullResponse struct {
	Archive   ArchiveListItem        `json:"archive"`
	List      []AuditLogListResponse `json:"list"`
	Total     int64                  `json:"total"`
	Page      int                    `json:"page"`
	PageSize  int                    `json:"page_size"`
	Truncated bool                   `json:"truncated"`
}
