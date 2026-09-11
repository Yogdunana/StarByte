// Package handler 提醒 HTTP 路由处理
package handler

import (
	"schedule-service/internal/service"

	"github.com/gin-gonic/gin"
)

// ============================================================================
// ReminderHTTPHandler 提醒路由处理
// ============================================================================

// ReminderHTTPHandler 提醒 HTTP 处理器
type ReminderHTTPHandler struct {
	reminderSvc service.ScheduleReminderService
}

// NewReminderHTTPHandler 构造提醒处理器
func NewReminderHTTPHandler(reminderSvc service.ScheduleReminderService) *ReminderHTTPHandler {
	return &ReminderHTTPHandler{reminderSvc: reminderSvc}
}

// Register 注册提醒相关路由
//   group 应为已挂载 JWT 中间件的 /api/v1 组
func (h *ReminderHTTPHandler) Register(rg *gin.RouterGroup) {
	reminders := rg.Group("/reminders")
	{
		reminders.POST("", h.Create)           // 创建提醒
		reminders.GET("", h.List)              // 列表查询
		reminders.GET("/:id", h.GetByID)       // 详情
		reminders.PUT("/:id", h.Update)        // 更新
		reminders.DELETE("/:id", h.Delete)     // 删除
		reminders.POST("/:id/snooze", h.Snooze) // 推迟
		reminders.POST("/:id/cancel", h.Cancel) // 取消
	}

	// 按事件维度查询提醒
	rg.GET("/events/:event_id/reminders", h.ListByEvent)
}

// Create 创建提醒
// POST /api/v1/reminders
func (h *ReminderHTTPHandler) Create(c *gin.Context) {
	var req service.CreateReminderRequest
	if !ShouldBindJSON(c, &req) {
		return
	}
	actorID := ActorIDFromContext(c)
	resp, err := h.reminderSvc.Create(c.Request.Context(), actorID, &req)
	if err != nil {
		Fail(c, err)
		return
	}
	OKCreated(c, resp)
}

// List 列表查询
// GET /api/v1/reminders?page=1&page_size=20&event_id=10&status=pending
func (h *ReminderHTTPHandler) List(c *gin.Context) {
	var req service.ReminderListRequest
	if !ShouldBindQuery(c, &req) {
		return
	}
	actorID := ActorIDFromContext(c)
	resp, err := h.reminderSvc.ListByUser(c.Request.Context(), actorID, &req)
	if err != nil {
		Fail(c, err)
		return
	}
	OKList(c, resp.List, resp.Total, resp.Page, resp.PageSize)
}

// ListByEvent 查询某事件的全部提醒
// GET /api/v1/events/:event_id/reminders
func (h *ReminderHTTPHandler) ListByEvent(c *gin.Context) {
	eventID, ok := parsePathUint(c, "event_id")
	if !ok {
		return
	}
	actorID := ActorIDFromContext(c)
	list, err := h.reminderSvc.ListByEvent(c.Request.Context(), actorID, eventID)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, list)
}

// GetByID 查询提醒详情
// GET /api/v1/reminders/:id
func (h *ReminderHTTPHandler) GetByID(c *gin.Context) {
	id, ok := parsePathUint(c, "id")
	if !ok {
		return
	}
	actorID := ActorIDFromContext(c)
	resp, err := h.reminderSvc.GetByID(c.Request.Context(), actorID, id)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, resp)
}

// Update 更新提醒
// PUT /api/v1/reminders/:id
func (h *ReminderHTTPHandler) Update(c *gin.Context) {
	id, ok := parsePathUint(c, "id")
	if !ok {
		return
	}
	var req service.UpdateReminderRequest
	if !ShouldBindJSON(c, &req) {
		return
	}
	actorID := ActorIDFromContext(c)
	resp, err := h.reminderSvc.Update(c.Request.Context(), actorID, id, &req)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, resp)
}

// Delete 删除提醒
// DELETE /api/v1/reminders/:id
func (h *ReminderHTTPHandler) Delete(c *gin.Context) {
	id, ok := parsePathUint(c, "id")
	if !ok {
		return
	}
	actorID := ActorIDFromContext(c)
	if err := h.reminderSvc.Delete(c.Request.Context(), actorID, id); err != nil {
		Fail(c, err)
		return
	}
	OK(c, nil)
}

// Snooze 推迟提醒
// POST /api/v1/reminders/:id/snooze  body: {"snooze_minutes":10}
func (h *ReminderHTTPHandler) Snooze(c *gin.Context) {
	id, ok := parsePathUint(c, "id")
	if !ok {
		return
	}
	var req service.SnoozeRequest
	if !ShouldBindJSON(c, &req) {
		return
	}
	actorID := ActorIDFromContext(c)
	resp, err := h.reminderSvc.Snooze(c.Request.Context(), actorID, id, &req)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, resp)
}

// Cancel 取消提醒
// POST /api/v1/reminders/:id/cancel
func (h *ReminderHTTPHandler) Cancel(c *gin.Context) {
	id, ok := parsePathUint(c, "id")
	if !ok {
		return
	}
	actorID := ActorIDFromContext(c)
	if err := h.reminderSvc.Cancel(c.Request.Context(), actorID, id); err != nil {
		Fail(c, err)
		return
	}
	OK(c, nil)
}
