package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/leave/service"
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

func withDataScope(group *gin.RouterGroup, resource string, cache rbacService.PermissionCacheService, db *gorm.DB, depts rbacRepo.DepartmentRepo) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequireDataScope(resource))
	g.Use(middleware.DataScopeMiddleware(db, depts, cache))
	return g
}

// RegisterRoutes 注册 /api/v1/leave。静态路径须在 /:id 之前。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, cache rbacService.PermissionCacheService, db *gorm.DB, depts rbacRepo.DepartmentRepo) {
	g := r.Group("/leave")
	g.Use(loadPerms(cache))

	g.GET("/types", h.Types)
	g.GET("/my", h.ListMine)
	g.POST("", h.Submit)

	readScope := withDataScope(g, "leave:read", cache, db, depts)
	readScope.GET("/balance", h.Balances)

	read := withPermission(g, "leave:read", cache)
	read.Use(middleware.RequireDataScope("leave:read"))
	read.Use(middleware.DataScopeMiddleware(db, depts, cache))
	read.GET("", h.List)
	read.GET("/stats", h.Stats)

	idScope := withDataScope(g, "leave:read", cache, db, depts)
	idScope.GET("/:id", h.Get)

	approve := withPermission(g, "leave:approve", cache)
	approve.Use(middleware.RequireDataScope("leave:approve"))
	approve.Use(middleware.DataScopeMiddleware(db, depts, cache))
	approve.PUT("/:id/approve", h.Approve)
	approve.PUT("/:id/reject", h.Reject)
}
