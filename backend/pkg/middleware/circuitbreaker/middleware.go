package circuitbreaker

import (
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// Degrade writes a fallback body when the circuit is open. If nil, 21002 is returned.
type Degrade func(c *gin.Context) bool

// Middleware fail-fasts when open, then records status/latency.
func Middleware(b *Breaker, degrade Degrade) gin.HandlerFunc {
	if b == nil {
		b = New(DefaultSettings())
	}
	return func(c *gin.Context) {
		name := c.Request.Method + ":" + c.FullPath()
		if c.FullPath() == "" {
			name = c.Request.Method + ":" + c.Request.URL.Path
		}
		if !b.Allow(name) {
			if degrade != nil && degrade(c) {
				c.Header("X-Degraded", "1")
				c.Abort()
				return
			}
			response.Error(c, response.NewError(response.CodeCircuitOpen, "服务繁忙，请稍后重试"))
			c.Abort()
			return
		}
		start := time.Now()
		c.Next()
		failed := c.Writer.Status() >= 500
		b.Record(name, failed, time.Since(start))
	}
}

// CacheDegrade returns HTTP 200 + 21004 with fallback payload (cache / default).
func CacheDegrade(body any) Degrade {
	return func(c *gin.Context) bool {
		response.Degraded(c, body)
		return true
	}
}

// PingDegrade serves a default pong when /ping is tripped; other routes stay 21002.
func PingDegrade(c *gin.Context) bool {
	path := ""
	if c.Request != nil && c.Request.URL != nil {
		path = c.Request.URL.Path
	}
	if path != "/api/v1/ping" && path != "/ping" {
		return false
	}
	response.Degraded(c, "pong")
	return true
}
