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

// RegisterRoutes 注册 /api/v1/finance。静态路径须在 /records/:id 之前。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, cache rbacService.PermissionCacheService) {
	g := r.Group("/finance")
	read := withPerm(g, "finance:read", cache)
	read.GET("/records", h.ListRecords)
	read.GET("/records/:id", h.GetRecord)
	read.GET("/categories", h.Categories)
	read.GET("/summary", h.Summary)
	read.GET("/export", h.Export)

	withPerm(g, "finance:create", cache).POST("/records", h.CreateRecord)
	manage := withPerm(g, "finance:manage", cache)
	manage.PUT("/records/:id", h.UpdateRecord)
	manage.DELETE("/records/:id", h.DeleteRecord)
}
