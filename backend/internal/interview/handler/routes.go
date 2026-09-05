package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/interview/service"
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InterviewHandler struct {
	svc service.InterviewService
}

func NewInterviewHandler(svc service.InterviewService) *InterviewHandler {
	return &InterviewHandler{svc: svc}
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
	g.Use(middleware.RequirePermission("interview:read"))
	g.Use(middleware.RequireDataScope("interview"))
	g.Use(middleware.PermissionRequired(cacheService))
	g.Use(middleware.DataScopeMiddleware(db, deptRepo, cacheService))
	return g
}

// RegisterRoutes 注册 /api/v1/interviews。静态路径必须在 /:id 之前。
func RegisterRoutes(
	r *gin.RouterGroup,
	h *InterviewHandler,
	cacheService rbacService.PermissionCacheService,
	db *gorm.DB,
	deptRepo rbacRepo.DepartmentRepo,
) {
	g := r.Group("/interviews")

	g.GET("/my", h.MyInterviews)
	g.POST("/:id/checkin", h.Checkin)

	read := withReadScope(g, cacheService, db, deptRepo)
	read.GET("/sessions", h.ListSessions)
	read.GET("/sessions/:id", h.GetSession)
	read.GET("/dimensions", h.ListDimensions)
	read.GET("/stats", h.Stats)
	read.GET("", h.ListInterviews)
	read.GET("/:id", h.GetInterview)
	read.GET("/:id/evaluations", h.GetEvaluations)

	manage := withPermission(g, "interview:manage", cacheService)
	manage.POST("/sessions", h.CreateSession)
	manage.PUT("/sessions/:id", h.UpdateSession)
	manage.DELETE("/sessions/:id", h.DeleteSession)
	manage.POST("/sessions/:id/start", h.StartSession)
	manage.POST("/sessions/:id/end", h.EndSession)
	manage.GET("/sessions/:id/qrcode", h.SessionQRCode)
	manage.POST("/dimensions", h.CreateDimension)
	manage.PUT("/dimensions/:id", h.UpdateDimension)
	manage.DELETE("/dimensions/:id", h.DeleteDimension)
	manage.POST("", h.CreateInterview)
	manage.POST("/:id/assign", h.AssignEvaluators)
	manage.POST("/:id/result", h.SubmitResult)

	evaluate := withPermission(g, "interview:evaluate", cacheService)
	evaluate.POST("/:id/start", h.StartInterview)
	evaluate.POST("/:id/end", h.EndInterview)
	evaluate.POST("/:id/evaluations", h.SubmitEvaluations)
	evaluate.PUT("/:id/evaluations/:eid", h.UpdateEvaluation)
}
