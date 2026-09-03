package handler

import (
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
)

// withPermission 创建一个带有权限校验的路由组
func withPermission(group *gin.RouterGroup, permCode string, cacheService rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cacheService))
	return g
}

// RegisterRoutes 注册审计日志模块路由
// 需要鉴权 + audit:read 权限
func RegisterRoutes(
	r *gin.RouterGroup,
	auditHandler *AuditHandler,
	cacheService rbacService.PermissionCacheService,
) {
	auditLogs := r.Group("/audit-logs")
	{
		// 查询权限：audit:read
		withPermission(auditLogs, "audit:read", cacheService).GET("", auditHandler.List)
		withPermission(auditLogs, "audit:read", cacheService).GET("/:id", auditHandler.GetByID)
		withPermission(auditLogs, "audit:read", cacheService).GET("/export", auditHandler.Export)

		// 归档权限：audit:archive
		withPermission(auditLogs, "audit:archive", cacheService).POST("/archive", auditHandler.TriggerArchive)
	}
}
