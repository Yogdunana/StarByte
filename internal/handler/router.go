// Package handler 路由注册入口
// 将三个 handler 的路由统一挂载到 /api/v1 下，并应用 JWT 中间件
package handler

import (
	"schedule-service/internal/service"

	"github.com/gin-gonic/gin"
)

// ScheduleHandlers 聚合三个 handler 便于统一注册
type ScheduleHandlers struct {
	Event    *EventHTTPHandler
	Reminder *ReminderHTTPHandler
	Attendee *AttendeeHTTPHandler
}

// NewScheduleHandlers 构造聚合 handler
func NewScheduleHandlers(
	eventSvc service.ScheduleEventService,
	reminderSvc service.ScheduleReminderService,
	attendeeSvc service.ScheduleAttendeeService,
) *ScheduleHandlers {
	return &ScheduleHandlers{
		Event:    NewEventHTTPHandler(eventSvc),
		Reminder: NewReminderHTTPHandler(reminderSvc),
		Attendee: NewAttendeeHTTPHandler(attendeeSvc),
	}
}

// RegisterRoutes 注册日程管理模块全部路由
//   router: gin.Engine 实例
//   jwtSecret: JWT 签名密钥
//   路由前缀: /api/v1
//   所有路由均需 JWT 认证
//
// 完整路由表:
//
//	事件 (EventHTTPHandler):
//	  POST   /api/v1/events                       创建事件
//	  GET    /api/v1/events                       列表查询
//	  GET    /api/v1/events/visible               我可见的全部
//	  GET    /api/v1/events/attended              我参与的
//	  GET    /api/v1/events/range                 日/周/月视图
//	  GET    /api/v1/events/:id                   详情
//	  PUT    /api/v1/events/:id                   更新
//	  DELETE /api/v1/events/:id                   删除
//	  PATCH  /api/v1/events/:id/status            状态变更
//	  POST   /api/v1/events/:id/meeting           关联会议
//	  DELETE /api/v1/events/:id/meeting           取消关联
//
//	提醒 (ReminderHTTPHandler):
//	  POST   /api/v1/reminders                    创建提醒
//	  GET    /api/v1/reminders                     列表查询
//	  GET    /api/v1/reminders/:id                 详情
//	  PUT    /api/v1/reminders/:id                 更新
//	  DELETE /api/v1/reminders/:id                 删除
//	  POST   /api/v1/reminders/:id/snooze         推迟
//	  POST   /api/v1/reminders/:id/cancel         取消
//	  GET    /api/v1/events/:event_id/reminders   事件下提醒列表
//
//	参与人 (AttendeeHTTPHandler):
//	  POST   /api/v1/events/:event_id/attendees           添加(单/批量)
//	  GET    /api/v1/events/:event_id/attendees           事件参与人列表
//	  GET    /api/v1/events/:event_id/attendees/:id       参与人详情
//	  PATCH  /api/v1/events/:event_id/attendees/:id       更新角色
//	  DELETE /api/v1/events/:event_id/attendees/:id        移除参与人
//	  POST   /api/v1/events/:event_id/attendees/:id/respond  响应邀请
//	  GET    /api/v1/me/attendees                          我的参与视图
func (h *ScheduleHandlers) RegisterRoutes(router *gin.Engine, jwtSecret string) {
	// 1. 健康检查（无需鉴权）
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 2. 日程管理 API 组（挂载 JWT 中间件）
	v1 := router.Group("/api/v1")
	v1.Use(JWTAuthMiddleware(jwtSecret))
	{
		h.Event.Register(v1)
		h.Reminder.Register(v1)
		h.Attendee.Register(v1)
	}
}
