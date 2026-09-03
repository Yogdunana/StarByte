package handler

import (
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
	}

	// 管理员通知路由（需要权限）
	if systemProtected != nil {
		systemNotifications := systemProtected.Group("/notifications")
		{
			systemNotifications.POST("/send", notificationHandler.Send)
			systemNotifications.POST("/broadcast", notificationHandler.Broadcast)
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
