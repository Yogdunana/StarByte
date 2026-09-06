package main

import (
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/circuitbreaker"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/ratelimit"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// applyAPITraffic mounts IP/route token buckets on /api/v1.
// Circuit breaker is not applied here: unauthenticated login/register must not
// share a trip with other clients (slow-body / P99 DoS).
func applyAPITraffic(api *gin.RouterGroup, rdb *redis.Client, cfg ratelimit.Config) {
	api.Use(ratelimit.Middleware(rdb, cfg))
}

// applyProtectedTraffic runs after JWT: user token bucket then per-route breaker.
func applyProtectedTraffic(protected *gin.RouterGroup, rdb *redis.Client, cfg ratelimit.Config, br *circuitbreaker.Breaker) {
	protected.Use(ratelimit.UserMiddleware(rdb, cfg))
	protected.Use(circuitbreaker.Middleware(br, nil))
}
