package handler

import (
	"github.com/gin-gonic/gin"

	rbac "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func taskViewer() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := getUserID(c)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		c.Request = c.Request.WithContext(model.WithViewer(c.Request.Context(), model.Viewer{ID: id, Scope: dataScope(c)}))
		c.Next()
	}
}

func taskCapability(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, _ := model.ViewerFromContext(c.Request.Context())
		if v.Allowed == nil {
			v.Allowed = map[string]bool{}
			v.Scopes = map[string]*rbac.DataScopeCondition{}
		}
		allowed := c.GetBool("is_super_admin")
		if raw, ok := c.Get("user_permissions"); ok {
			if perms, ok := raw.([]string); ok {
				for _, p := range perms {
					if p == permission {
						allowed = true
					}
				}
			}
		}
		v.Allowed[permission] = allowed
		v.Scopes[permission] = middleware.GetDataScopeFromContext(c)
		c.Request = c.Request.WithContext(model.WithViewer(c.Request.Context(), v))
		c.Next()
	}
}

func personalTaskViewer(cache rbacService.PermissionCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := getUserID(c)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		perms, super, err := cache.GetUserPermissionsAndSuperAdmin(c.Request.Context(), id)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		c.Set("user_permissions", perms)
		c.Set("is_super_admin", super)
		c.Request = c.Request.WithContext(model.WithViewer(c.Request.Context(), model.Viewer{ID: id, Scope: &rbac.DataScopeCondition{Query: "1 = 0", IsSelf: true}}))
		c.Next()
	}
}
