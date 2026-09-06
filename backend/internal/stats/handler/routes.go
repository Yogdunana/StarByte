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

// RegisterRoutes 注册 /api/v1/stats。静态路径必须在 /:provider 之前。
func RegisterRoutes(r *gin.RouterGroup, h *StatsHandler, cacheService rbacService.PermissionCacheService) {
	g := r.Group("/stats")
	g.GET("/providers", h.Providers)
	g.GET("/overview", h.Overview)
	withPermission(g, "stats:export", cacheService).GET("/export/:provider", h.Export)

	read := withPermission(g, "stats:read", cacheService)
	read.GET("/member-distribution", h.serveNamed("member-distribution"))
	read.GET("/interview-data", h.serveNamed("interview-data"))
	read.GET("/meeting-attendance", h.serveNamed("meeting-attendance"))
	read.GET("/task-progress", h.serveNamed("task-progress"))
	read.GET("/internship-duration", h.serveNamed("internship-duration"))
	read.GET("/:provider", h.Get)
}
