package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/backup/service"
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

// RegisterRoutes 注册 /api/v1/system/backups。静态路径须在 /:id 之前。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, cache rbacService.PermissionCacheService) {
	g := r.Group("/system/backups")

	read := withPermission(g, "backup:read", cache)
	read.GET("", h.List)
	read.GET("/policies", h.GetPolicy)
	read.GET("/storage", h.Storage)
	read.GET("/:id/preview", h.Preview)
	read.GET("/:id", h.Get)

	create := withPermission(g, "backup:create", cache)
	create.POST("", h.Create)

	del := withPermission(g, "backup:delete", cache)
	del.DELETE("/:id", h.Delete)

	restore := withPermission(g, "backup:restore", cache)
	restore.POST("/:id/restore", h.Restore)
	restore.POST("/:id/restore-drill", h.DrillRestore)

	manage := withPermission(g, "backup:manage", cache)
	manage.PUT("/policies", h.PutPolicy)
}
