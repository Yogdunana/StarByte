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
	g.Use(middleware.RequireDataScope("discipline"))
	g.Use(middleware.PermissionRequired(cache))
	g.Use(middleware.DataScopeMiddleware(db, deptRepo, cache))
	return g
}

// RegisterRoutes 注册 /api/v1/discipline。静态动作路径须在 /:id 同层用完整路径。
func RegisterRoutes(
	r *gin.RouterGroup,
	h *Handler,
	cache rbacService.PermissionCacheService,
	db *gorm.DB,
	deptRepo rbacRepo.DepartmentRepo,
) {
	g := r.Group("/discipline")
	read := withScope(g, "discipline:read", cache, db, deptRepo)
	read.GET("/records", h.ListRecords)
	read.GET("/records/:id", h.GetRecord)

	create := withScope(g, "discipline:create", cache, db, deptRepo)
	create.POST("/records", h.CreateRecord)
	create.PUT("/records/:id", h.UpdateRecord)

	withScope(g, "discipline:approve", cache, db, deptRepo).POST("/records/:id/approve", h.Approve)
	withScope(g, "discipline:revoke", cache, db, deptRepo).POST("/records/:id/revoke", h.Revoke)

	// 申诉：登录即可，服务层限制只能申诉本人记录
	g.POST("/records/:id/appeal", h.Appeal)
}
