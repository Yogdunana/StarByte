package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
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

// Logger logs each request's method, path, status, latency, request id,
// client IP, and user id (if authenticated).
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		requestID := c.GetString("request_id")
		clientIP := c.ClientIP()
		userID := auth.GetUserID(c)

		logger.Info("request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("request_id", requestID),
			zap.String("client_ip", clientIP),
			zap.String("user_id", userID),
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
					"code":       response.CodeInternalError,
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

// CORSWithConfig returns a CORS middleware that uses the provided
// configuration. It supports configurable origins, methods, headers,
// and credentials. If the config has empty slices, safe defaults are used.
func CORSWithConfig(cfg config.CORSConfig) gin.HandlerFunc {
	origins := cfg.AllowedOrigins
	if len(origins) == 0 {
		origins = []string{"*"}
	}
	methods := cfg.AllowedMethods
	if len(methods) == 0 {
		methods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	}
	headers := cfg.AllowedHeaders
	if len(headers) == 0 {
		headers = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-Id"}
	}
	exposeHeaders := cfg.ExposeHeaders
	if len(exposeHeaders) == 0 {
		exposeHeaders = []string{"X-Request-Id"}
	}
	allowCredentials := cfg.AllowCredentials

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Determine the allowed origin.
		allowedOrigin := ""
		for _, o := range origins {
			if o == "*" {
				allowedOrigin = "*"
				break
			}
			if o == origin {
				allowedOrigin = o
				break
			}
		}

		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
		}
		if allowCredentials && allowedOrigin != "*" {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Allow-Methods", strings.Join(methods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(headers, ", "))
		c.Header("Access-Control-Expose-Headers", strings.Join(exposeHeaders, ", "))
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// CORS returns a CORS middleware with default permissive settings.
// Prefer CORSWithConfig() for production use.
func CORS() gin.HandlerFunc {
	return CORSWithConfig(config.CORSConfig{})
}
