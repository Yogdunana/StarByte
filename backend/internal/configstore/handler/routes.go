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

// RegisterRoutes 注册 /api/v1/system/configs。/key/:key 必须在 /:id 之前。
func RegisterRoutes(
	r *gin.RouterGroup,
	h *ConfigHandler,
	cacheService rbacService.PermissionCacheService,
) {
	g := r.Group("/system/configs")
	withPermission(g, "config:read", cacheService).GET("", h.List)
	withPermission(g, "config:read", cacheService).GET("/key/:key", h.GetByKey)
	withPermission(g, "config:create", cacheService).POST("", h.Create)
	withPermission(g, "config:update", cacheService).PUT("/:id", h.Update)
	withPermission(g, "config:delete", cacheService).DELETE("/:id", h.Delete)

	smtp := r.Group("/system/smtp")
	withPermission(smtp, "config:read", cacheService).GET("", h.GetSMTP)
	withPermission(smtp, "config:update", cacheService).PUT("", h.UpdateSMTP)
	withPermission(smtp, "config:update", cacheService).POST("/test", h.TestSMTP)
}
