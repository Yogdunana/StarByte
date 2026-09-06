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

// RegisterRoutes 注册 /api/v1/system/search。
func RegisterRoutes(r *gin.RouterGroup, h *SearchHandler, cacheService rbacService.PermissionCacheService) {
	g := r.Group("/system/search")
	read := withPermission(g, "search:read", cacheService)
	read.GET("/resources", h.Resources)
	read.POST("/query", h.Query)
}
