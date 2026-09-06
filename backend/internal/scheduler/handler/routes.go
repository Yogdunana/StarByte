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

// RegisterRoutes 注册 /api/v1/system/scheduler。
func RegisterRoutes(r *gin.RouterGroup, h *SchedulerHandler, cacheService rbacService.PermissionCacheService) {
	g := r.Group("/system/scheduler")
	read := withPermission(g, "scheduler:read", cacheService)
	read.GET("/handlers", h.Handlers)
	read.GET("/tasks", h.List)
	read.GET("/tasks/:id", h.Get)
	read.GET("/tasks/:id/logs", h.Logs)

	withPermission(g, "scheduler:create", cacheService).POST("/tasks", h.Create)
	withPermission(g, "scheduler:update", cacheService).PUT("/tasks/:id", h.Update)
	withPermission(g, "scheduler:delete", cacheService).DELETE("/tasks/:id", h.Delete)
	withPermission(g, "scheduler:run", cacheService).POST("/tasks/:id/run", h.Run)
	withPermission(g, "scheduler:manage", cacheService).POST("/tasks/:id/pause", h.Pause)
	withPermission(g, "scheduler:manage", cacheService).POST("/tasks/:id/resume", h.Resume)
}
