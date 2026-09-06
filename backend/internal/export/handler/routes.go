package handler

import (
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func withPermission(group *gin.RouterGroup, permCode string, cacheService rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cacheService))
	return g
}

func withAnyPermission(group *gin.RouterGroup, cacheService rbacService.PermissionCacheService, codes ...string) *gin.RouterGroup {
	g := group.Group("")
	g.Use(func(c *gin.Context) {
		userIDStr := auth.GetUserID(c)
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			response.Error(c, response.NewUnauthorizedError("无效的用户身份"))
			c.Abort()
			return
		}
		perms, isSuper, err := cacheService.GetUserPermissionsAndSuperAdmin(c.Request.Context(), userID)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		if isSuper {
			c.Set("is_super_admin", true)
			c.Set("user_permissions", []string{"*"})
			c.Next()
			return
		}
		have := map[string]struct{}{}
		for _, p := range perms {
			have[p] = struct{}{}
		}
		for _, code := range codes {
			if _, ok := have[code]; ok {
				c.Set("user_permissions", perms)
				c.Next()
				return
			}
		}
		response.Error(c, response.NewForbiddenError("权限不足"))
		c.Abort()
	})
	return g
}

// RegisterRoutes 注册 /api/v1/export。
func RegisterRoutes(r *gin.RouterGroup, h *ExportHandler, cacheService rbacService.PermissionCacheService) {
	g := r.Group("/export")
	withPermission(g, "export:excel", cacheService).POST("/excel", h.ExportExcel)
	withPermission(g, "export:csv", cacheService).POST("/csv", h.ExportCSV)
	withPermission(g, "export:pdf", cacheService).POST("/pdf", h.ExportPDF)
	withPermission(g, "export:json", cacheService).POST("/json", h.ExportJSON)
	withPermission(g, "export:template", cacheService).POST("/template/:template_id", h.ExportTemplate)
	withPermission(g, "export:read", cacheService).GET("/tasks/:task_id", h.GetTask)
	withPermission(g, "export:download", cacheService).GET("/download/:file_id", h.Download)
	withAnyPermission(g, cacheService, "export:read", "export:template").GET("/templates", h.ListTemplates)
}
