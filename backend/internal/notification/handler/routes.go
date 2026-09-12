package handler

import (
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册通知模块路由
// protected: 需要鉴权的用户路由
// systemProtected: 需要鉴权+权限校验的管理员路由
// wsHandler: WebSocket 连接处理器（注册在 /ws 路径下，独立于 API 组）
func RegisterRoutes(
	protected *gin.RouterGroup,
	systemProtected *gin.RouterGroup,
	notificationHandler *NotificationHandler,
	templateHandler *TemplateHandler,
	wsHandler *WSHandler,
	emailHandler *EmailHandler,
	cacheService rbacService.PermissionCacheService,
) {
	// 用户通知路由（登录即可访问）
	if protected != nil {
		notifications := protected.Group("/notifications")
		{
			notifications.GET("", notificationHandler.List)
			notifications.GET("/unread/count", notificationHandler.UnreadCount)
			notifications.POST("/:id/read", notificationHandler.MarkAsRead)
			notifications.POST("/read-all", notificationHandler.MarkAllAsRead)
			notifications.DELETE("/:id", notificationHandler.Delete)
		}
		if emailHandler != nil {
			email := withPermission(protected.Group("/notifications/email"), "notification:send", cacheService)
			email.POST("/send", emailHandler.SendEmail)
			email.POST("/batch", emailHandler.SendEmailBatch)
			email.GET("/logs", emailHandler.ListEmailLogs)
		}
	}

	// 管理员发送：须 notification:send（与 /notifications/email/send 对齐）。
	// main.go 把 systemProtected 接到同一 JWT 组，因此这里必须自己挂权限，不能只靠分组名。
	if systemProtected != nil {
		for _, prefix := range []string{"/notifications", "/system/notifications"} {
			send := withPermission(systemProtected.Group(prefix), "notification:send", cacheService)
			send.POST("/send", notificationHandler.Send)
			send.POST("/broadcast", notificationHandler.Broadcast)
		}

		templates := systemProtected.Group("/notification-templates")
		{
			templates.GET("", templateHandler.List)
			templates.POST("", templateHandler.Create)
			templates.GET("/:id", templateHandler.Get)
			templates.PUT("/:id", templateHandler.Update)
			templates.DELETE("/:id", templateHandler.Delete)
			templates.POST("/:id/test", templateHandler.Test)
		}
	}

	// WebSocket 路由（独立于 API 组，需要 JWT 认证）
	// 注意: WebSocket 路由在 main.go 中通过 r.GET("/ws/notifications", ...) 注册
	// 因为它不在 /api/v1 路径下
}

// RegisterWSRoute 在根路由组注册 WebSocket 路由
func RegisterWSRoute(r *gin.Engine, wsHandler *WSHandler) {
	r.GET("/ws/notifications", wsHandler.HandleConnection)
}

func withPermission(group *gin.RouterGroup, permCode string, cacheService rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cacheService))
	return g
}
