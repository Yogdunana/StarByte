package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

type RouteSecurity struct {
	DB          *gorm.DB
	Cache       rbacService.PermissionCacheService
	Departments rbacRepo.DepartmentRepo
}

func workflowPermission(group *gin.RouterGroup, permission string, security RouteSecurity) *gin.RouterGroup {
	return group.Group("", middleware.RequirePermission(permission), middleware.PermissionRequired(security.Cache))
}
func workflowViewer(group *gin.RouterGroup, permission string, security RouteSecurity) *gin.RouterGroup {
	return group.Group("", middleware.RequireDataScope(permission), middleware.DataScopeMiddleware(security.DB, security.Departments, security.Cache), func(c *gin.Context) {
		id, err := getUserID(c)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		viewer := model.Viewer{ID: id, Scope: middleware.GetDataScopeFromContext(c)}
		c.Request = c.Request.WithContext(model.WithViewer(c.Request.Context(), viewer))
		c.Next()
	})
}
