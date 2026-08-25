package middleware

import (
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// PermissionRequired returns a gin middleware that enforces permission-based
// access control. The required permission code is expected to be stored in the
// request context by RequirePermission during route registration.
//
// Super administrators bypass the permission check entirely. For ordinary users
// the middleware fetches the cached permission list and verifies that it
// contains the required permission code. On success the user's permission list
// is stored in the context under "user_permissions" for downstream use.
func PermissionRequired(cacheService rbacService.PermissionCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := auth.GetUserID(c)
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			response.Error(c, response.NewUnauthorizedError("无效的用户身份"))
			c.Abort()
			return
		}

		// 超级管理员跳过权限校验
		isSuper, err := cacheService.IsSuperAdmin(c.Request.Context(), userID)
		if err != nil {
			logger.Error("check super admin failed", zap.Error(err))
			response.Error(c, err)
			c.Abort()
			return
		}
		if isSuper {
			c.Set("is_super_admin", true)
			c.Next()
			return
		}

		// 获取路由注册时设置的所需权限码
		requiredPerm, exists := c.Get("required_permission")
		if !exists {
			c.Next()
			return
		}

		permCode, ok := requiredPerm.(string)
		if !ok {
			c.Next()
			return
		}

		// 从缓存获取用户权限列表
		perms, err := cacheService.GetUserPermissions(c.Request.Context(), userID)
		if err != nil {
			logger.Error("get user permissions failed", zap.Error(err))
			response.Error(c, err)
			c.Abort()
			return
		}

		// 校验用户是否拥有所需权限（O(1) map 查找）
		permMap := make(map[string]struct{}, len(perms))
		for _, p := range perms {
			permMap[p] = struct{}{}
		}
		if _, ok := permMap[permCode]; ok {
			c.Set("user_permissions", perms)
			c.Next()
			return
		}

		response.Error(c, response.NewForbiddenError("权限不足"))
		c.Abort()
	}
}

// RequirePermission returns a gin middleware that records the permission code
// required to access a route. It does not perform any enforcement itself; the
// actual check is performed by PermissionRequired, which reads the stored code
// from the context. This split allows route groups to declare their permission
// requirement at registration time while a single PermissionRequired middleware
// (registered on the parent group) performs the verification.
func RequirePermission(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("required_permission", code)
		c.Next()
	}
}
