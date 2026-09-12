package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/announcement/service"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct{ svc service.Service }

func New(svc service.Service) *Handler { return &Handler{svc: svc} }

func withPermission(group *gin.RouterGroup, permCode string, cache rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cache))
	return g
}

// RegisterRoutes 注册 /api/v1/announcements。静态路径须在 /:id 之前。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, cache rbacService.PermissionCacheService) {
	g := r.Group("/announcements")

	read := withPermission(g, "announcement:read", cache)
	read.GET("", h.List)
	read.GET("/unread-count", h.UnreadCount)
	read.GET("/:id", h.Get)
	read.POST("/:id/read", h.MarkRead)

	create := withPermission(g, "announcement:create", cache)
	create.POST("", h.Create)

	update := withPermission(g, "announcement:update", cache)
	update.PUT("/:id", h.Update)

	del := withPermission(g, "announcement:delete", cache)
	del.DELETE("/:id", h.Delete)

	publish := withPermission(g, "announcement:publish", cache)
	publish.POST("/:id/publish", h.Publish)

	manage := withPermission(g, "announcement:manage", cache)
	manage.GET("/:id/read-status", h.ReadStatus)
	manage.POST("/:id/pin", h.Pin)
	manage.POST("/:id/archive", h.Archive)
}
