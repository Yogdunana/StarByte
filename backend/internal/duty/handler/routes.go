package handler

import (
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func withPermission(group *gin.RouterGroup, permCode string, cacheService rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cacheService))
	return g
}

// RegisterRoutes 注册 /api/v1/duty 路由
func RegisterRoutes(
	r *gin.RouterGroup,
	h *DutyHandler,
	cacheService rbacService.PermissionCacheService,
) {
	duty := r.Group("/duty")

	// 读取权限
	read := withPermission(duty, "duty:read", cacheService)
	read.GET("/schedule", h.ListSchedules)
	read.GET("/stats", h.GetDutyStats)

	// 创建排班
	create := withPermission(duty, "duty:create", cacheService)
	create.POST("/schedule", h.CreateSchedule)
	create.POST("/schedule/batch", h.BatchCreateSchedules)

	// 调整排班（拖拽）
	update := withPermission(duty, "duty:update", cacheService)
	update.PUT("/schedule/:id", h.UpdateSchedule)

	// 调班申请（登录即可）
	duty.POST("/swap", h.CreateSwapRequest)
	duty.GET("/swap", h.ListSwapRequests)
	// 调班审批需要 update 权限
	approveSwap := withPermission(duty, "duty:update", cacheService)
	approveSwap.PUT("/swap/:id", h.ActionSwap)
}
