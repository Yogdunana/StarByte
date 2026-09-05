package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/internship/service"
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InternshipHandler struct {
	svc service.InternshipService
}

func NewInternshipHandler(svc service.InternshipService) *InternshipHandler {
	return &InternshipHandler{svc: svc}
}

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
	g.Use(middleware.RequirePermission("internship:read"))
	g.Use(middleware.RequireDataScope("internship"))
	g.Use(middleware.PermissionRequired(cacheService))
	g.Use(middleware.DataScopeMiddleware(db, deptRepo, cacheService))
	return g
}

func withWriteScope(
	group *gin.RouterGroup,
	permCode string,
	cacheService rbacService.PermissionCacheService,
	db *gorm.DB,
	deptRepo rbacRepo.DepartmentRepo,
) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.RequireDataScope("internship"))
	g.Use(middleware.PermissionRequired(cacheService))
	g.Use(middleware.DataScopeMiddleware(db, deptRepo, cacheService))
	return g
}

// RegisterRoutes 注册 /api/v1/internships 与 /system/internship-config。静态路径必须在 /:id 之前。
func RegisterRoutes(
	r *gin.RouterGroup,
	h *InternshipHandler,
	cacheService rbacService.PermissionCacheService,
	db *gorm.DB,
	deptRepo rbacRepo.DepartmentRepo,
) {
	g := r.Group("/internships")
	g.GET("/my", h.ListMine)

	read := withReadScope(g, cacheService, db, deptRepo)
	read.GET("/stats/duration", h.DurationStats)
	read.GET("/stats/ranking", h.Ranking)
	read.GET("/stats/department", h.DepartmentStats)
	read.GET("", h.ListInternships)
	read.GET("/:id", h.GetInternship)

	create := withWriteScope(g, "internship:create", cacheService, db, deptRepo)
	create.POST("", h.CreateInternship)

	update := withWriteScope(g, "internship:update", cacheService, db, deptRepo)
	update.PUT("/:id", h.UpdateInternship)
	update.POST("/:id/complete", h.CompleteInternship)
	update.POST("/:id/report", h.SubmitReport)

	del := withWriteScope(g, "internship:delete", cacheService, db, deptRepo)
	del.DELETE("/:id", h.DeleteInternship)

	eval := withWriteScope(g, "internship:evaluate", cacheService, db, deptRepo)
	eval.POST("/:id/mentor-comment", h.MentorComment)

	sysRead := withPermission(r, "internship:read", cacheService)
	sysRead.GET("/system/internship-config", h.GetConfig)
	sysWrite := withPermission(r, "system:config", cacheService)
	sysWrite.PUT("/system/internship-config", h.UpdateConfig)
}
