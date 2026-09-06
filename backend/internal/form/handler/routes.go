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

// RegisterRoutes 注册 /api/v1/forms。静态路径须在 /:id 之前。
func RegisterRoutes(r *gin.RouterGroup, h *FormHandler, cacheService rbacService.PermissionCacheService) {
	g := r.Group("/forms")
	g.GET("", h.List)
	g.GET("/:id", h.Get)

	withPermission(g, "form:write", cacheService).POST("", h.Create)
	withPermission(g, "form:write", cacheService).PUT("/:id", h.Update)
	withPermission(g, "form:submit", cacheService).POST("/:id/submit", h.Submit)
	withPermission(g, "form:read", cacheService).GET("/:id/submissions", h.ListSubmissions)
}
