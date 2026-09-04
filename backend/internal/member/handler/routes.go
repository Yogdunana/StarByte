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

func withReadScope(
	group *gin.RouterGroup,
	cacheService rbacService.PermissionCacheService,
	db *gorm.DB,
	deptRepo rbacRepo.DepartmentRepo,
) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission("member:read"))
	g.Use(middleware.RequireDataScope("member"))
	g.Use(middleware.PermissionRequired(cacheService))
	g.Use(middleware.DataScopeMiddleware(db, deptRepo, cacheService))
	return g
}

// RegisterRoutes 注册 /api/v1/member
func RegisterRoutes(
	r *gin.RouterGroup,
	h *MemberHandler,
	cacheService rbacService.PermissionCacheService,
	db *gorm.DB,
	deptRepo rbacRepo.DepartmentRepo,
) {
	member := r.Group("/member")

	// 登录即可：提交、我的申请、补充提交、部门下拉
	member.POST("/applications", h.Submit)
	member.GET("/applications/my", h.MyApplications)
	member.POST("/applications/:id/resubmit", h.Resubmit)
	member.GET("/departments", h.ListDepartments)

	read := withReadScope(member, cacheService, db, deptRepo)
	read.GET("/applications", h.ListApplications)
	read.GET("/applications/:id", h.GetApplication)
	read.GET("/applications/:id/history", h.ApplicationHistory)
	read.GET("/profiles", h.ListProfiles)
	read.GET("/profiles/:id", h.GetProfile)
	read.GET("/profiles/:id/history", h.ProfileHistory)
	read.GET("/stats/applications", h.ApplicationStats)
	read.GET("/stats/members", h.MemberStats)

	export := withReadScope(member, cacheService, db, deptRepo)
	export.Use(middleware.RequirePermission("member:export"))
	export.Use(middleware.PermissionRequired(cacheService))
	export.GET("/profiles/export", h.ExportProfiles)

	approve := withPermission(member, "member:approve", cacheService)
	approve.POST("/applications/:id/approve", h.Approve)
	approve.POST("/applications/:id/reject", h.Reject)
	approve.POST("/applications/:id/supplement", h.Supplement)

	update := withReadScope(member, cacheService, db, deptRepo)
	update.Use(middleware.RequirePermission("member:update"))
	update.Use(middleware.PermissionRequired(cacheService))
	update.PUT("/profiles/:id", h.UpdateProfile)

	manage := withPermission(member, "member:manage", cacheService)
	manage.PUT("/profiles/:id/status", h.UpdateProfileStatus)
}
