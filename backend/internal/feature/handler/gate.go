package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/feature/service"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// BypassFunc lets staff keep managing a surface while members are in grayscale.
type BypassFunc func(*gin.Context) bool

// RequireFlag aborts with 34007 when the current user misses the flag.
func RequireFlag(svc service.Service, key string, bypass BypassFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if bypass != nil && bypass(c) {
			c.Next()
			return
		}
		uid, err := getUserID(c)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		sub, err := svc.Resolve(c.Request.Context(), uid)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		if !svc.Enabled(c.Request.Context(), key, sub) {
			response.Error(c, response.NewError(response.CodeFeatureDisabled, "该功能尚未对你开放"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// StaffBypass returns true when the request already loaded write permissions.
func StaffBypass(codes ...string) BypassFunc {
	return func(c *gin.Context) bool {
		raw, ok := c.Get("user_permissions")
		if !ok {
			return false
		}
		perms, _ := raw.([]string)
		allow := map[string]struct{}{}
		for _, code := range codes {
			allow[code] = struct{}{}
		}
		for _, p := range perms {
			if p == "*" {
				return true
			}
			if _, hit := allow[p]; hit {
				return true
			}
		}
		return false
	}
}
