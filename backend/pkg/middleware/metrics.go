package middleware

import (
	"strconv"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/metrics"
	"github.com/gin-gonic/gin"
)

// Metrics records Prometheus HTTP counters and latency histograms.
// Unmatched routes use path="unmatched" to avoid unbounded label cardinality.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}
		if path == "/metrics" {
			return
		}
		status := strconv.Itoa(c.Writer.Status())
		metrics.HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(c.Request.Method, path).Observe(time.Since(start).Seconds())
	}
}
