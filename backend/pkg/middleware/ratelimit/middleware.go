package ratelimit

import (
	"strconv"

	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Middleware applies IP + route token buckets. User bucket is UserMiddleware (after JWT).
func Middleware(rdb *redis.Client, cfg Config) gin.HandlerFunc {
	store := NewStore(rdb)
	return func(c *gin.Context) {
		paintTraffic(c, cfg)
		if deniedByACL(c, cfg) {
			response.Error(c, response.NewError(response.CodeBlacklisted, "访问已被限制"))
			c.Abort()
			return
		}
		if skipLimit(c, cfg) {
			c.Next()
			return
		}
		if !allowOrAbort(c, store, "rl:ip:"+c.ClientIP(), cfg.IP, int(cfg.IP.Burst)) {
			return
		}
		rk := routeKey(c)
		if rk != "unknown" && rk != c.Request.Method+":" {
			if !allowOrAbort(c, store, "rl:rt:"+rk, cfg.Route, int(cfg.Route.Burst)) {
				return
			}
		}
		c.Next()
	}
}

// UserMiddleware applies the per-user bucket after JWT.
func UserMiddleware(rdb *redis.Client, cfg Config) gin.HandlerFunc {
	store := NewStore(rdb)
	return func(c *gin.Context) {
		if skipLimit(c, cfg) {
			c.Next()
			return
		}
		uid := viewerID(c)
		if uid == "" {
			c.Next()
			return
		}
		if inSet(cfg.UserBlacklist, uid) {
			response.Error(c, response.NewError(response.CodeBlacklisted, "访问已被限制"))
			c.Abort()
			return
		}
		if !allowOrAbort(c, store, "rl:uid:"+uid, cfg.User, int(cfg.User.Burst)) {
			return
		}
		c.Next()
	}
}

func allowOrAbort(c *gin.Context, store *Store, key string, b Bucket, limitHdr int) bool {
	res, err := store.Allow(c.Request.Context(), key, b)
	if err != nil {
		logger.Warn("ratelimit: redis error, failing open", zap.String("key", key), zap.Error(err))
		return true
	}
	if limitHdr > 0 {
		c.Header("X-RateLimit-Limit", strconv.Itoa(limitHdr))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(res.Remaining, 10))
	}
	if res.Allowed {
		return true
	}
	retry := res.RetryAfter
	if retry < 1 {
		retry = 1
	}
	c.Header("Retry-After", strconv.Itoa(retry))
	response.Error(c, response.NewError(response.CodeRateLimited, "请求过于频繁，请稍后重试"))
	c.Abort()
	return false
}
