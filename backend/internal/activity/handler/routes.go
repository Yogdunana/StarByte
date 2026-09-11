package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/activity/service"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type ActivityHandler struct {
	svc service.ActivityService
}

func NewActivityHandler(svc service.ActivityService) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}

func withPermission(group *gin.RouterGroup, permCode string, cacheService rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cacheService))
	return g
}

// RegisterRoutes 注册 /api/v1/activities 路由
func RegisterRoutes(
	r *gin.RouterGroup,
	h *ActivityHandler,
	cacheService rbacService.PermissionCacheService,
) {
	g := r.Group("/activities")

	read := withPermission(g, "activity:read", cacheService)
	read.GET("", h.ListActivities)
	read.GET("/:id", h.GetActivity)

	// 登录用户：本人报名 / 签到 / 评价 / 查看自己的报名
	g.GET("/:id/register", h.GetMyRegistration)
	g.POST("/:id/register", h.Register)
	g.DELETE("/:id/register", h.CancelRegistration)
	g.POST("/:id/checkin", h.Checkin)
	g.POST("/:id/survey", h.SubmitSurvey)

	create := withPermission(g, "activity:create", cacheService)
	create.POST("", h.CreateActivity)

	update := withPermission(g, "activity:update", cacheService)
	update.PUT("/:id", h.UpdateActivity)
	update.POST("/:id/start", h.StartActivity)
	update.POST("/:id/end", h.EndActivity)
	update.POST("/:id/cancel", h.CancelActivity)

	del := withPermission(g, "activity:delete", cacheService)
	del.DELETE("/:id", h.DeleteActivity)

	manage := withPermission(g, "activity:manage", cacheService)
	manage.GET("/:id/registrations", h.ListRegistrations)
	manage.POST("/:id/registrations/:uid/approve", h.ApproveRegistration)
	manage.GET("/:id/stats", h.GetStats)
	manage.GET("/:id/qrcode", h.CheckinQRCode)
}
