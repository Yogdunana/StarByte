package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/metrics"
	"github.com/gin-gonic/gin"
)

// Metrics records Prometheus HTTP counters and latency histograms.
// Unmatched routes use path="unmatched" to avoid unbounded label cardinality.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// Record in defer so a handler panic still increments counters.
		// Register this middleware *before* ErrorHandler so the recovered
		// 500 is already written when the defer runs.
		defer func() {
			path := c.FullPath()
			if path == "" {
				path = "unmatched"
			}
			if path == "/metrics" {
				return
			}
			status := strconv.Itoa(c.Writer.Status())
			metrics.HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
			// WS handlers block until disconnect; session length is not HTTP latency.
			if strings.HasPrefix(path, "/ws") {
				return
			}
			elapsed := time.Since(start).Seconds()
			metrics.HTTPRequestDuration.WithLabelValues(c.Request.Method, path).Observe(elapsed)
			metrics.RecordHTTPLatency(elapsed)
		}()
		c.Next()
	}
}
