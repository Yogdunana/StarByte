package handler

import (
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func withPerm(group *gin.RouterGroup, code string, cache rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(code))
	g.Use(middleware.PermissionRequired(cache))
	return g
}

// RegisterRoutes 注册 /api/v1/contracts。静态路径须在 /:id 之前。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, cache rbacService.PermissionCacheService) {
	g := r.Group("/contracts")
	read := withPerm(g, "contract:read", cache)
	read.GET("/templates", h.Templates)
	read.GET("/expiring", h.Expiring)
	read.GET("", h.List)
	read.GET("/:id", h.Get)

	withPerm(g, "contract:create", cache).POST("", h.Create)
	manage := withPerm(g, "contract:manage", cache)
	manage.PUT("/:id", h.Update)
	manage.DELETE("/:id", h.Delete)
}
