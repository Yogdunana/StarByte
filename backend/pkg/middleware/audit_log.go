package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuditLogEntry represents a row in the audit_logs table. It is written
// by the AuditLog middleware for every state-changing HTTP request
// (POST, PUT, PATCH, DELETE).
type AuditLogEntry struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID         *uuid.UUID `gorm:"type:uuid;index"`
	Username       string     `gorm:"type:varchar(50)"`
	Operation      string     `gorm:"type:varchar(100);not null;index"`
	Method         string     `gorm:"type:varchar(10)"`
	Path           string     `gorm:"type:varchar(500)"`
	IP             string     `gorm:"type:varchar(50);index"`
	UserAgent      string     `gorm:"type:varchar(500)"`
	RequestParams  string     `gorm:"type:text"`
	ResponseStatus int        `gorm:"type:int"`
	ResponseBody   string     `gorm:"type:text"`
	DurationMs     int        `gorm:"type:int"`
	RequestID      string     `gorm:"type:varchar(100)"`
	CreatedAt      time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;index"`
}

// TableName overrides the default GORM table name.
func (AuditLogEntry) TableName() string {
	return "audit_logs"
}

// writeMethods defines the HTTP methods that trigger audit logging.
// GET and OPTIONS requests are read-only and do not produce audit entries.
var writeMethods = map[string]bool{
	"POST":   true,
	"PUT":    true,
	"PATCH":  true,
	"DELETE": true,
}

// maxRequestBodySize limits the size of the request body captured for audit.
// Large payloads are truncated to avoid excessive storage.
const maxRequestBodySize = 4096

// maxResponseBodySize limits the size of the response body captured for audit.
const maxResponseBodySize = 2048

// auditResponseWriter wraps gin.ResponseWriter to capture the response body
// for audit logging. It only captures up to maxResponseBodySize bytes.
type auditResponseWriter struct {
	gin.ResponseWriter
	body []byte
}

func (w *auditResponseWriter) Write(b []byte) (int, error) {
	if len(w.body) < maxResponseBodySize {
		remaining := maxResponseBodySize - len(w.body)
		if len(b) <= remaining {
			w.body = append(w.body, b...)
		} else {
			w.body = append(w.body, b[:remaining]...)
		}
	}
	return w.ResponseWriter.Write(b)
}

// AuditLog returns a gin middleware that records audit log entries for
// all state-changing requests (POST, PUT, PATCH, DELETE). Read requests
// (GET, OPTIONS, HEAD) are not audited.
//
// The middleware captures the following information:
//   - User ID and username (from JWT context)
//   - HTTP method, path, and client IP
//   - Request body (truncated to 4 KB)
//   - Response status code and body (truncated to 2 KB)
//   - Request duration in milliseconds
//   - Request ID (from RequestID middleware)
//
// Audit log entries are written asynchronously to avoid adding latency
// to the response. If the database write fails, the error is logged but
// does not affect the response.
//
// The db parameter should be the GORM database instance. If nil, the
// middleware is a no-op.
func AuditLog(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		// Only audit write operations.
		if !writeMethods[method] {
			c.Next()
			return
		}

		// Wrap the response writer to capture the body.
		auditWriter := &auditResponseWriter{
			ResponseWriter: c.Writer,
		}
		c.Writer = auditWriter

		// Capture start time.
		start := time.Now()

		// Capture request body (truncated).
		var reqBody string
		if c.Request.Body != nil && c.Request.ContentLength != 0 {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				logger.Warn("audit log: failed to read request body",
					zap.Error(err),
				)
			}
			// Restore the body for downstream handlers.
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			if len(bodyBytes) > maxRequestBodySize {
				reqBody = string(bodyBytes[:maxRequestBodySize]) + "...[truncated]"
			} else {
				reqBody = string(bodyBytes)
			}
		}

		// Process the request.
		c.Next()

		// Calculate duration.
		duration := time.Since(start)
		durationMs := int(duration.Milliseconds())

		// Extract user info from context (set by JWTAuth).
		userIDStr := auth.GetUserID(c)
		username := auth.GetUsername(c)

		var userID *uuid.UUID
		if userIDStr != "" {
			if parsed, err := uuid.Parse(userIDStr); err == nil {
				userID = &parsed
			}
		}

		// Build the operation string from method + path.
		operation := method + " " + c.Request.URL.Path

		// Determine the response body.
		respBody := string(auditWriter.body)
		if len(respBody) > maxResponseBodySize {
			respBody = respBody[:maxResponseBodySize] + "...[truncated]"
		}

		// Build audit entry.
		entry := AuditLogEntry{
			ID:             uuid.New(),
			UserID:         userID,
			Username:       username,
			Operation:      operation,
			Method:         method,
			Path:           c.Request.URL.Path,
			IP:             c.ClientIP(),
			UserAgent:      c.Request.UserAgent(),
			RequestParams:  reqBody,
			ResponseStatus: c.Writer.Status(),
			ResponseBody:   respBody,
			DurationMs:     durationMs,
			RequestID:      c.GetString("request_id"),
			CreatedAt:      time.Now(),
		}

		// Write asynchronously to avoid blocking the response.
		go writeAuditLog(db, entry)

		// Also log structurally for immediate visibility.
		logger.Info("audit log",
			zap.String("method", method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Int("duration_ms", durationMs),
			zap.String("ip", c.ClientIP()),
			zap.String("user_id", userIDStr),
			zap.String("username", username),
			zap.String("request_id", c.GetString("request_id")),
		)
	}
}

// writeAuditLog inserts the audit entry into the database. It is called
// in a goroutine to avoid adding latency to the HTTP response. If the
// insert fails, the error is logged but not surfaced to the client.
func writeAuditLog(db *gorm.DB, entry AuditLogEntry) {
	if db == nil {
		return
	}
	if err := db.Create(&entry).Error; err != nil {
		logger.Error("audit log: failed to write to database",
			zap.Error(err),
			zap.String("operation", entry.Operation),
			zap.String("request_id", entry.RequestID),
		)
	}
}
