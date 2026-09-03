package middleware

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
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

// sensitivePaths defines paths whose request bodies should not be captured
// in audit logs to prevent storing sensitive data like passwords.
var sensitivePaths = map[string]bool{
	"/api/v1/auth/login":    true,
	"/api/v1/auth/register": true,
	"/api/v1/auth/refresh":  true,
	"/api/v1/user/password": true,
}

// sensitiveFields defines the JSON field names whose values should be redacted
// in audit logs. This list is shared with the audit service layer to ensure
// consistent desensitization.
var sensitiveFields = []string{
	"password",
	"old_password",
	"new_password",
	"secret",
	"token",
	"access_token",
	"refresh_token",
}

// sensitiveFieldPattern matches JSON fields like "password":"value" and
// replaces the value with "[redacted]".
var sensitiveFieldPattern = regexp.MustCompile(
	`(?i)"(` + strings.Join(sensitiveFields, "|") + `)"\s*:\s*"[^"]*"`,
)

// sanitizeRequestBody redacts sensitive data from the request body.
// For sensitive paths (login, register, password change), the entire
// body is replaced with a placeholder. For other paths, known sensitive
// JSON fields are redacted.
func sanitizeRequestBody(path, body string) string {
	if sensitivePaths[path] {
		return "[redacted: sensitive endpoint]"
	}
	return sensitiveFieldPattern.ReplaceAllString(body, `"$1":"[redacted]"`)
}

// maxBodyReadSize limits how much of the request body is read into memory.
// This prevents memory exhaustion from very large request bodies (e.g., file
// uploads). Bodies larger than this are truncated; the audit log only stores
// up to maxRequestBodySize bytes anyway.
const maxBodyReadSize = 32 * 1024 * 1024 // 32 MB

// auditLogWorkerBufferSize is the buffer size for the audit log worker channel.
// When the channel is full, new entries are dropped to avoid blocking the
// response and to provide backpressure.
const auditLogWorkerBufferSize = 256

// auditLogWriter is a background worker that writes audit log entries to
// the database using a buffered channel for backpressure control. This
// prevents unbounded goroutine creation under high load.
type auditLogWriter struct {
	db   *gorm.DB
	ch   chan AuditLogEntry
	done chan struct{}
}

var (
	auditWriterMu sync.Mutex
	auditWriter   *auditLogWriter
)

// CloseAuditWriter gracefully shuts down the audit log worker. It closes the
// channel, which causes the background goroutine to drain remaining entries
// and exit. This should be called during server graceful shutdown to avoid
// losing pending audit log entries.
//
// It is safe to call multiple times; subsequent calls are no-ops.
func CloseAuditWriter() {
	auditWriterMu.Lock()
	defer auditWriterMu.Unlock()
	if auditWriter != nil {
		close(auditWriter.ch)
		<-auditWriter.done
		auditWriter = nil
	}
}

// getAuditWriter returns a shared auditLogWriter instance. The writer is
// initialized on first use and starts a background worker goroutine.
func getAuditWriter(db *gorm.DB) *auditLogWriter {
	auditWriterMu.Lock()
	defer auditWriterMu.Unlock()
	if auditWriter == nil {
		auditWriter = &auditLogWriter{
			db:   db,
			ch:   make(chan AuditLogEntry, auditLogWorkerBufferSize),
			done: make(chan struct{}),
		}
		go auditWriter.run()
	}
	return auditWriter
}

// run is the background worker loop. It reads entries from the channel
// and writes them to the database. The loop exits when the channel is closed.
func (w *auditLogWriter) run() {
	for entry := range w.ch {
		if err := w.db.Create(&entry).Error; err != nil {
			logger.Error("audit log: failed to write to database",
				zap.Error(err),
				zap.String("operation", entry.Operation),
				zap.String("request_id", entry.RequestID),
			)
		}
	}
	close(w.done)
}

// write sends the audit entry to the worker channel. If the channel is
// full (buffer exhausted), the entry is dropped with a warning log to
// prevent blocking the HTTP response. A recover guards against the
// edge case where the channel is closed during graceful shutdown while
// a request is still in flight.
func (w *auditLogWriter) write(entry AuditLogEntry) {
	if w.db == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			// Channel was closed during shutdown — entry is lost.
		}
	}()
	select {
	case w.ch <- entry:
	default:
		logger.Warn("audit log: channel full, dropping entry",
			zap.String("operation", entry.Operation),
			zap.String("request_id", entry.RequestID),
		)
	}
}

// auditResponseWriter wraps gin.ResponseWriter to capture the response body
// for audit logging. It only captures up to maxResponseBodySize bytes.
// It also proxies Hijack and Flush for WebSocket and streaming support.
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

// Flush proxies the Flush method to the underlying writer for streaming.
func (w *auditResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack proxies the Hijack method to the underlying writer for WebSocket.
func (w *auditResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter does not implement http.Hijacker")
}

// AuditLog returns a gin middleware that records audit log entries for
// all state-changing requests (POST, PUT, PATCH, DELETE). Read requests
// (GET, OPTIONS, HEAD) are not audited.
//
// This middleware should be registered BEFORE JWTAuth so that failed
// authentication attempts are also captured in audit logs. User info
// will be empty for failed auth attempts.
//
// The middleware captures the following information:
//   - User ID and username (from JWT context, if authenticated)
//   - HTTP method, path, and client IP
//   - Request body (truncated to 4 KB, sensitive fields redacted)
//   - Response status code and body (truncated to 2 KB)
//   - Request duration in milliseconds
//   - Request ID (from RequestID middleware)
//
// Audit log entries are written asynchronously via a worker pool to
// avoid adding latency to the response. If the database write fails,
// the error is logged but does not affect the response.
//
// The db parameter should be the GORM database instance. If nil, the
// middleware is a no-op.
func AuditLog(db *gorm.DB) gin.HandlerFunc {
	writer := getAuditWriter(db)
	return func(c *gin.Context) {
		method := c.Request.Method

		// Only audit write operations.
		if !writeMethods[method] {
			c.Next()
			return
		}

		// Wrap the response writer to capture the body.
		auditRW := &auditResponseWriter{
			ResponseWriter: c.Writer,
		}
		c.Writer = auditRW

		// Capture start time.
		start := time.Now()

		// Capture request body (truncated). We read the body if it exists,
		// regardless of ContentLength, to support chunked transfer encoding.
		// A size limit is applied to prevent memory exhaustion from very
		// large bodies.
		var reqBody string
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodyReadSize))
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
			// Redact sensitive data.
			reqBody = sanitizeRequestBody(c.Request.URL.Path, reqBody)
		}

		// Process the request.
		c.Next()

		// Calculate duration.
		duration := time.Since(start)
		durationMs := int(duration.Milliseconds())

		// Extract user info from context (set by JWTAuth).
		// Will be empty for failed auth attempts.
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
		respBody := string(auditRW.body)
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

		// Write asynchronously via worker pool.
		writer.write(entry)

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
