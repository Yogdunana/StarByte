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

// RegisterRoutes 注册 /api/v1/system/audit-logs
func RegisterRoutes(
	r *gin.RouterGroup,
	auditHandler *AuditHandler,
	cacheService rbacService.PermissionCacheService,
) {
	auditLogs := r.Group("/system/audit-logs")

	withPermission(auditLogs, "audit:export", cacheService).GET("/export", auditHandler.Export)
	withPermission(auditLogs, "audit:archive", cacheService).POST("/archive", auditHandler.TriggerArchive)

	read := withPermission(auditLogs, "audit:read", cacheService)
	read.GET("", auditHandler.List)
	read.GET("/:id", auditHandler.GetByID)
}
