package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/leave/service"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct{ svc service.Service }

func New(svc service.Service) *Handler { return &Handler{svc: svc} }

func withPermission(group *gin.RouterGroup, permCode string, cache rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cache))
	return g
}

// loadPerms 给登录即可访问的路由填入权限列表，供 my vs all / 余额 IDOR 判断。
func loadPerms(cache rbacService.PermissionCacheService) gin.HandlerFunc {
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
		if super {
			c.Set("is_super_admin", true)
			c.Set("user_permissions", []string{"*"})
		} else {
			c.Set("user_permissions", perms)
		}
		c.Next()
	}
}

// RegisterRoutes 注册 /api/v1/leave。静态路径须在 /:id 之前。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, cache rbacService.PermissionCacheService) {
	g := r.Group("/leave")
	g.Use(loadPerms(cache))

	g.GET("/types", h.Types)
	g.GET("/my", h.ListMine)
	g.GET("/balance", h.Balances)
	g.POST("", h.Submit)

	read := withPermission(g, "leave:read", cache)
	read.GET("", h.List)
	read.GET("/stats", h.Stats)

	g.GET("/:id", h.Get)

	approve := withPermission(g, "leave:approve", cache)
	approve.PUT("/:id/approve", h.Approve)
	approve.PUT("/:id/reject", h.Reject)
}
