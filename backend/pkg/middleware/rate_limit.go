package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimitConfig defines the parameters for a fixed-window rate limiter.
//
//   - Rate:    maximum number of requests allowed within the window
//   - Window:  time window in seconds
//   - KeyFunc: function that generates the Redis key for rate limiting
type RateLimitConfig struct {
	Rate    int
	Window  int
	KeyFunc func(c *gin.Context) string
}

// Predefined rate limit configurations per Issue #14 requirements.
var (
	// GlobalRateLimit: 1000 req/s across the entire application.
	GlobalRateLimit = RateLimitConfig{
		Rate:   1000,
		Window: 1,
		KeyFunc: func(c *gin.Context) string {
			return "ratelimit:global"
		},
	}

	// PerIPRateLimit: 100 req/min per client IP.
	PerIPRateLimit = RateLimitConfig{
		Rate:   100,
		Window: 60,
		KeyFunc: func(c *gin.Context) string {
			return "ratelimit:ip:" + c.ClientIP()
		},
	}

	// LoginRateLimit: 5 req/min for login endpoint (brute-force protection).
	LoginRateLimit = RateLimitConfig{
		Rate:   5,
		Window: 60,
		KeyFunc: func(c *gin.Context) string {
			return "ratelimit:login:" + c.ClientIP()
		},
	}
)

// fixedWindowScript is a Redis Lua script that implements a fixed-window
// counter rate limiter. It atomically increments a counter and returns the
// current count. If the counter exceeds the limit, it returns 0 (denied).
//
// Keys:
//
//	KEYS[1] = the rate limit key
//
// Args:
//
//	ARGV[1] = window in seconds
//	ARGV[2] = max requests in the window
//
// Returns:
//
//	{1, remaining}  if allowed (remaining = max - count)
//	{0, 0}          if denied
var fixedWindowScript = redis.NewScript(`
	local key = KEYS[1]
	local window = tonumber(ARGV[1])
	local max_req = tonumber(ARGV[2])

	local count = redis.call('INCR', key)
	if count == 1 then
		redis.call('EXPIRE', key, window)
	end

	if count > max_req then
		return {0, 0}
	end

	return {1, max_req - count}
`)

// RateLimit returns a gin middleware that enforces rate limiting using
// Redis. If Redis is unavailable or returns an error, the middleware
// fails open (allows the request) to avoid blocking the entire service.
//
// The middleware sets the following response headers:
//   - X-RateLimit-Limit:     maximum requests in the window
//   - X-RateLimit-Remaining: remaining requests in the current window
//
// When the rate limit is exceeded, the middleware returns HTTP 429 with
// error code CodeTooManyReq (1006).
func RateLimit(rdb *redis.Client, cfg RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil {
			// Fail open: no Redis, no rate limiting.
			c.Next()
			return
		}

		key := cfg.KeyFunc(c)

		result, err := fixedWindowScript.Run(context.Background(), rdb,
			[]string{key},
			cfg.Window, cfg.Rate,
		).Result()

		if err != nil {
			// Redis error — fail open to avoid blocking the service.
			logger.Warn("rate limit: redis error, failing open",
				zap.String("key", key),
				zap.Error(err),
			)
			c.Next()
			return
		}

		vals, ok := result.([]interface{})
		if !ok || len(vals) < 2 {
			logger.Warn("rate limit: unexpected redis script result",
				zap.Any("result", result),
			)
			c.Next()
			return
		}

		allowed, _ := vals[0].(int64)
		remaining, _ := vals[1].(int64)

		// Set rate limit headers.
		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.Rate))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))

		if allowed == 0 {
			retryAfter := cfg.Window
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, response.Response{
				Code:      response.CodeTooManyReq,
				Message:   fmt.Sprintf("请求过于频繁，请 %d 秒后重试", retryAfter),
				Data:      nil,
				RequestID: c.GetString("request_id"),
				Timestamp: time.Now().Unix(),
			})
			return
		}

		c.Next()
	}
}

// RateLimitWithFallback returns a middleware that applies rate limiting
// with a fallback to fail open. This is a convenience wrapper that
// recovers from panics in the rate limiter.
func RateLimitWithFallback(rdb *redis.Client, cfg RateLimitConfig) gin.HandlerFunc {
	limiter := RateLimit(rdb, cfg)
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Warn("rate limit: panic recovered, failing open",
					zap.Any("error", r),
				)
				c.Next()
			}
		}()
		limiter(c)
	}
}
