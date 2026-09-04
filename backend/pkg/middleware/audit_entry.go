package middleware

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// AuditLogEntry 是中间件写入 audit_logs 的表映射。
// 放在 pkg 内，避免 pkg/middleware 依赖 internal 业务模块。
type AuditLogEntry struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID         *uuid.UUID `gorm:"type:uuid;index"`
	Username       string     `gorm:"type:varchar(50)"`
	RealName       string     `gorm:"type:varchar(50)"`
	Operation      string     `gorm:"type:varchar(100);not null;index"`
	Method         string     `gorm:"type:varchar(10)"`
	Path           string     `gorm:"type:varchar(500)"`
	Module         string     `gorm:"type:varchar(50);index"`
	Action         string     `gorm:"type:varchar(20);index"`
	IP             string     `gorm:"type:varchar(50);index"`
	UserAgent      string     `gorm:"type:varchar(500)"`
	RequestParams  string     `gorm:"type:text"`
	ResponseStatus int        `gorm:"type:int"`
	ResponseBody   string     `gorm:"type:text"`
	DurationMs     int        `gorm:"type:int"`
	RequestID      string     `gorm:"type:varchar(100)"`
	CreatedAt      time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;index"`
}

func (AuditLogEntry) TableName() string {
	return "audit_logs"
}

// ParseModule 从 /api/v1/{module}/... 解析模块名。
func ParseModule(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" {
		return parts[2]
	}
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// ActionFromMethod 将 HTTP 方法映射为审计动作。
func ActionFromMethod(method string) string {
	switch strings.ToUpper(method) {
	case "POST":
		return "CREATE"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	default:
		return strings.ToUpper(method)
	}
}
