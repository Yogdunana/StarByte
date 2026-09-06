package middleware

import (
	"context"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var processStart = time.Now()

func appVersion() string {
	if v := os.Getenv("APP_VERSION"); v != "" {
		return v
	}
	return "1.0.0"
}

// HealthCheck returns a liveness probe handler.
// Extra fields (uptime / version / go_version) are for operators; K8s still keys off HTTP 200.
func HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "ok",
			"uptime":     time.Since(processStart).Truncate(time.Second).String(),
			"version":    appVersion(),
			"go_version": runtime.Version(),
		})
	}
}

// MinioPinger probes object storage. Nil means MinIO is not configured.
type MinioPinger func(ctx context.Context) error

// ReadinessCheck verifies database, Redis, and MinIO.
// Returns 503 when any dependency is down.
func ReadinessCheck(db *gorm.DB, rdb *redis.Client, pingMinio MinioPinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		allOK := true
		checks := gin.H{
			"database": probeDep(&allOK, func() error { return pingDB(ctx, db) }),
			"redis":    probeDep(&allOK, func() error { return pingRedis(ctx, rdb) }),
			"minio":    probeDep(&allOK, func() error { return pingMinioDep(ctx, pingMinio) }),
		}

		status := "ready"
		code := http.StatusOK
		if !allOK {
			status = "not_ready"
			code = http.StatusServiceUnavailable
		}
		c.JSON(code, gin.H{"status": status, "checks": checks})
	}
}

func probeDep(allOK *bool, fn func() error) gin.H {
	start := time.Now()
	if err := fn(); err != nil {
		*allOK = false
		return gin.H{
			"status":     "fail",
			"error":      err.Error(),
			"latency_ms": time.Since(start).Milliseconds(),
		}
	}
	return gin.H{"status": "ok", "latency_ms": time.Since(start).Milliseconds()}
}

func pingDB(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return errDepNotConfigured("database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func pingRedis(ctx context.Context, rdb *redis.Client) error {
	if rdb == nil {
		return errDepNotConfigured("redis")
	}
	return rdb.Ping(ctx).Err()
}

func pingMinioDep(ctx context.Context, ping MinioPinger) error {
	if ping == nil {
		return errDepNotConfigured("minio")
	}
	return ping(ctx)
}

type depError string

func (e depError) Error() string { return string(e) }

func errDepNotConfigured(name string) error {
	return depError(name + " not configured")
}
