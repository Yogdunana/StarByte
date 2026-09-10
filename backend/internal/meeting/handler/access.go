package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func meetingViewer(manage bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := getUserID(c)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		c.Request = c.Request.WithContext(model.WithViewer(c.Request.Context(), model.Viewer{ID: id, Scope: middleware.GetDataScopeFromContext(c), Manage: manage}))
		c.Next()
	}
}
func withMeetingScope(group *gin.RouterGroup, permission string, cache rbacService.PermissionCacheService, db *gorm.DB, depts rbacRepo.DepartmentRepo) *gin.RouterGroup {
	return group.Group("", middleware.RequireDataScope(permission), middleware.DataScopeMiddleware(db, depts, cache), meetingViewer(permission != "meeting:read"))
}

// Capture management scope separately from the previously captured read scope.
func meetingCapabilities(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		viewer := model.ViewerFromContext(c.Request.Context())
		scope := middleware.GetDataScopeFromContext(c)
		allowed := c.GetBool("is_super_admin")
		if raw, ok := c.Get("user_permissions"); ok {
			if permissions, ok := raw.([]string); ok {
				for _, p := range permissions {
					if p == permission {
						allowed = true
					}
				}
			}
		}
		switch permission {
		case "meeting:manage":
			viewer.ManageScope = scope
			viewer.CanManage = allowed
		case "meeting:update":
			viewer.UpdateScope = scope
			viewer.CanUpdate = allowed
		case "meeting:delete":
			viewer.DeleteScope = scope
			viewer.CanDelete = allowed
		}
		c.Request = c.Request.WithContext(model.WithViewer(c.Request.Context(), viewer))
		c.Next()
	}
}
