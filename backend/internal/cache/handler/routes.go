package handler

import (
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func withPermission(group *gin.RouterGroup, permCode string, cacheService rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cacheService))
	return g
}

// RegisterRoutes 注册 /api/v1/system/cache。/pattern/:pattern 必须在 /:key 之前。
func RegisterRoutes(r *gin.RouterGroup, h *CacheHandler, cacheService rbacService.PermissionCacheService) {
	g := r.Group("/system/cache")
	withPermission(g, "cache:read", cacheService).GET("/stats", h.Stats)
	withPermission(g, "cache:delete", cacheService).DELETE("/pattern/:pattern", h.DeletePattern)
	withPermission(g, "cache:delete", cacheService).DELETE("/:key", h.DeleteKey)
	withPermission(g, "cache:manage", cacheService).POST("/warmup", h.Warmup)
}
