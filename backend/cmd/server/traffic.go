package main

import (
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/circuitbreaker"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/ratelimit"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// applyAPITraffic mounts IP/route token buckets and the circuit breaker on /api/v1.
// /health stays outside this group. Global 1000/s and login 5/min stay in main.go.
func applyAPITraffic(api *gin.RouterGroup, rdb *redis.Client, cfg ratelimit.Config, br *circuitbreaker.Breaker) {
	api.Use(ratelimit.Middleware(rdb, cfg))
	api.Use(circuitbreaker.Middleware(br, circuitbreaker.PingDegrade))
}

func applyUserTraffic(protected *gin.RouterGroup, rdb *redis.Client, cfg ratelimit.Config) {
	protected.Use(ratelimit.UserMiddleware(rdb, cfg))
}
