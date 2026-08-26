package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// HealthCheck returns a liveness probe handler.
// It responds with 200 and {"status": "ok"} as long as the process
// is running. No external dependencies are checked.
func HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.OK(c, gin.H{
			"status": "ok",
		})
	}
}

// ReadinessCheck returns a readiness probe handler.
// It verifies that critical dependencies (database and Redis) are
// reachable before returning 200. If any dependency is unreachable,
// it returns 503 with the failing component identified.
//
// Response body:
//
//	{
//	  "status": "ready",           // or "not ready"
//	  "checks": {
//	    "db": "ok",                // or "fail"
//	    "redis": "ok"              // or "fail"
//	  }
//	}
func ReadinessCheck(db *gorm.DB, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		checks := gin.H{}
		allOK := true

		// Check database connectivity.
		dbStatus := "ok"
		if db == nil {
			dbStatus = "fail"
			allOK = false
		} else {
			sqlDB, err := db.DB()
			if err != nil {
				dbStatus = "fail"
				allOK = false
			} else if err := sqlDB.PingContext(ctx); err != nil {
				dbStatus = "fail"
				allOK = false
			}
		}
		checks["db"] = dbStatus

		// Check Redis connectivity.
		redisStatus := "ok"
		if rdb == nil {
			redisStatus = "fail"
			allOK = false
		} else if err := rdb.Ping(ctx).Err(); err != nil {
			redisStatus = "fail"
			allOK = false
		}
		checks["redis"] = redisStatus

		if !allOK {
			c.JSON(http.StatusServiceUnavailable, response.Response{
				Code:    response.CodeInternalError,
				Message: "not ready",
				Data: gin.H{
					"status": "not ready",
					"checks": checks,
				},
				RequestID: c.GetString("request_id"),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		response.OK(c, gin.H{
			"status": "ready",
			"checks": checks,
		})
	}
}
