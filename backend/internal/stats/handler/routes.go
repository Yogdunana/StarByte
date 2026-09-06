package handler

import (
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func withPermission(group *gin.RouterGroup, permCode string, cacheService rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cacheService))
	return g
}

func withStatsScope(
	group *gin.RouterGroup,
	permCode string,
	cacheService rbacService.PermissionCacheService,
	db *gorm.DB,
	deptRepo rbacRepo.DepartmentRepo,
) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.RequireDataScope("stats"))
	g.Use(middleware.PermissionRequired(cacheService))
	g.Use(middleware.DataScopeMiddleware(db, deptRepo, cacheService))
	return g
}

// RegisterRoutes 注册 /api/v1/stats。静态路径必须在 /:provider 之前。
func RegisterRoutes(
	r *gin.RouterGroup,
	h *StatsHandler,
	cacheService rbacService.PermissionCacheService,
	db *gorm.DB,
	deptRepo rbacRepo.DepartmentRepo,
) {
	g := r.Group("/stats")
	g.GET("/providers", h.Providers)
	g.GET("/overview", h.Overview)
	withStatsScope(g, "stats:export", cacheService, db, deptRepo).GET("/export/:provider", h.Export)

	read := withStatsScope(g, "stats:read", cacheService, db, deptRepo)
	read.GET("/member-distribution", h.serveNamed("member-distribution"))
	read.GET("/interview-data", h.serveNamed("interview-data"))
	read.GET("/meeting-attendance", h.serveNamed("meeting-attendance"))
	read.GET("/task-progress", h.serveNamed("task-progress"))
	read.GET("/internship-duration", h.serveNamed("internship-duration"))
	read.GET("/:provider", h.Get)
}
