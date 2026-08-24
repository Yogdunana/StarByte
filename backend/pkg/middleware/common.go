package middleware

import (
	"net/http"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestID generates a unique request id (or reuses the one supplied via the
// X-Request-Id header) and stores it in the gin context under "request_id".
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-Id")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-Id", requestID)
		c.Next()
	}
}

// Logger logs each request's method, path, status, latency and request id.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		requestID := c.GetString("request_id")

		logger.Info("request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("request_id", requestID),
		)
	}
}

// Recovery recovers from panics in subsequent handlers, logs the error and
// returns a 500 response.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					zap.Any("error", r),
					zap.String("request_id", c.GetString("request_id")),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":       5000,
					"message":    "Internal Server Error",
					"data":       nil,
					"request_id": c.GetString("request_id"),
					"timestamp":  time.Now().Unix(),
				})
			}
		}()
		c.Next()
	}
}

// CORS sets the Cross-Origin Resource Sharing headers and short-circuits
// preflight OPTIONS requests.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-Id")
		c.Header("Access-Control-Expose-Headers", "X-Request-Id")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
