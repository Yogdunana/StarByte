package handler

import (
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func withScope(
	group *gin.RouterGroup,
	code string,
	cache rbacService.PermissionCacheService,
	db *gorm.DB,
	deptRepo rbacRepo.DepartmentRepo,
) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(code))
	g.Use(middleware.RequireDataScope("contract"))
	g.Use(middleware.PermissionRequired(cache))
	g.Use(middleware.DataScopeMiddleware(db, deptRepo, cache))
	return g
}

// RegisterRoutes 注册 /api/v1/contracts。静态路径须在 /:id 之前。
func RegisterRoutes(
	r *gin.RouterGroup,
	h *Handler,
	cache rbacService.PermissionCacheService,
	db *gorm.DB,
	deptRepo rbacRepo.DepartmentRepo,
) {
	g := r.Group("/contracts")
	read := withScope(g, "contract:read", cache, db, deptRepo)
	read.GET("/templates", h.Templates)
	read.GET("/expiring", h.Expiring)
	read.GET("", h.List)
	read.GET("/:id", h.Get)

	withScope(g, "contract:create", cache, db, deptRepo).POST("", h.Create)
	manage := withScope(g, "contract:manage", cache, db, deptRepo)
	manage.PUT("/:id", h.Update)
	manage.DELETE("/:id", h.Delete)
}
