package middleware

import (
	"strings"

	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AttachCurrentRoles resolves active role assignments; JWT role claims may be stale.
func AttachCurrentRoles(cache rbacService.PermissionCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := c.Get("current_roles"); ok {
			c.Next()
			return
		}
		raw := auth.GetUserID(c)
		if raw == "" {
			c.Next()
			return
		}
		uid, err := uuid.Parse(raw)
		if err != nil {
			response.Unauthorized(c, "用户未认证")
			c.Abort()
			return
		}
		roles, err := cache.GetUserRoleCodes(c.Request.Context(), uid)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		if roles == nil {
			roles = []string{}
		}
		c.Set("current_roles", roles)
		c.Next()
	}
}

// RegisteredUserAccess limits accounts without an association/staff role.
// Existing members and custom staff roles keep their normal permission checks.
func RegisteredUserAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, _ := c.Get("current_roles")
		roles, _ := raw.([]string)
		for _, role := range roles {
			if role != "user" {
				c.Next()
				return
			}
		}
		if registeredRouteAllowed(c.Request.Method, c.FullPath()) {
			c.Next()
			return
		}
		response.Forbidden(c, "没有权限执行此操作")
		c.Abort()
	}
}

func registeredRouteAllowed(method, path string) bool {
	path = strings.TrimPrefix(path, "/api/v1")
	switch method + " " + path {
	case "GET /auth/me", "POST /auth/logout", "PUT /auth/password", "POST /auth/change-password",
		"GET /user/me", "PUT /user/profile", "PUT /user/password",
		"POST /member/applications", "GET /member/applications/my", "GET /member/departments",
		"POST /member/applications/:id/resubmit", "GET /member/applications/:id/progress", "GET /member/applications/:id/admission",
		"GET /announcements", "GET /announcements/unread-count", "GET /announcements/:id", "POST /announcements/:id/read":
		return true
	}
	return false
}

// Fresh base-role status takes precedence over stale cached grants.
func baseAccountPermissions(c *gin.Context, perms []string, super bool) ([]string, bool) {
	raw, present := c.Get("current_roles")
	if !present {
		return perms, super
	}
	roles, _ := raw.([]string)
	for _, role := range roles {
		if role != "user" {
			return perms, super
		}
	}
	return []string{"announcement:read"}, false
}
