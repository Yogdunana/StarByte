package middleware

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var writeMethods = map[string]bool{
	"POST":   true,
	"PUT":    true,
	"DELETE": true,
}

var skipAuditPaths = map[string]bool{
	"/api/v1/auth/login":  true,
	"/api/v1/auth/logout": true,
}

const (
	maxRequestBodySize       = 4096
	maxResponseBodySize      = 2048
	maxBodyReadSize          = 32 * 1024 * 1024
	auditLogWorkerBufferSize = 256
)

type auditLogWriter struct {
	db   *gorm.DB
	ch   chan AuditLogEntry
	done chan struct{}
}

var (
	auditWriterMu sync.Mutex
	auditWriter   *auditLogWriter
)

func CloseAuditWriter() {
	auditWriterMu.Lock()
	defer auditWriterMu.Unlock()
	if auditWriter != nil {
		close(auditWriter.ch)
		<-auditWriter.done
		auditWriter = nil
	}
}

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

func (w *auditLogWriter) write(entry AuditLogEntry) {
	if w.db == nil {
		return
	}
	defer func() {
		_ = recover()
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

func (w *auditResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *auditResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter does not implement http.Hijacker")
}

// AuditLog 记录 POST/PUT/DELETE。GET 不记录。登录/登出由 auth 事件写入。
func AuditLog(db *gorm.DB) gin.HandlerFunc {
	writer := getAuditWriter(db)
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path
		if !writeMethods[method] || skipAuditPaths[path] {
			c.Next()
			return
		}

		auditRW := &auditResponseWriter{ResponseWriter: c.Writer}
		c.Writer = auditRW
		start := time.Now()

		reqBody := readAndRestoreBody(c)

		c.Next()

		durationMs := int(time.Since(start).Milliseconds())
		userIDStr := auth.GetUserID(c)
		username := auth.GetUsername(c)
		var userID *uuid.UUID
		if userIDStr != "" {
			if parsed, err := uuid.Parse(userIDStr); err == nil {
				userID = &parsed
			}
		}

		respBody := string(auditRW.body)
		if len(respBody) > maxResponseBodySize {
			respBody = respBody[:maxResponseBodySize] + "...[truncated]"
		}

		entry := AuditLogEntry{
			ID:             uuid.New(),
			UserID:         userID,
			Username:       username,
			Operation:      method + " " + path,
			Method:         method,
			Path:           path,
			Module:         ParseModule(path),
			Action:         ActionFromMethod(method),
			IP:             c.ClientIP(),
			UserAgent:      c.Request.UserAgent(),
			RequestParams:  sanitizeRequestBody(path, reqBody),
			ResponseStatus: c.Writer.Status(),
			ResponseBody:   sanitizeResponseBody(respBody),
			DurationMs:     durationMs,
			RequestID:      c.GetString("request_id"),
			CreatedAt:      time.Now(),
		}
		writer.write(entry)

		logger.Info("audit log",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Int("duration_ms", durationMs),
			zap.String("ip", c.ClientIP()),
			zap.String("user_id", userIDStr),
			zap.String("username", username),
			zap.String("request_id", c.GetString("request_id")),
		)
	}
}

func readAndRestoreBody(c *gin.Context) string {
	if c.Request.Body == nil {
		return ""
	}
	bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodyReadSize))
	if err != nil {
		logger.Warn("audit log: failed to read request body", zap.Error(err))
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	if len(bodyBytes) > maxRequestBodySize {
		return string(bodyBytes[:maxRequestBodySize]) + "...[truncated]"
	}
	return string(bodyBytes)
}
