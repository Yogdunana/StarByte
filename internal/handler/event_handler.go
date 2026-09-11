// Package handler 日程事件 HTTP 路由处理
package handler

import (
	"net/http"

	"schedule-service/internal/service"

	"github.com/gin-gonic/gin"
)

// ============================================================================
// EventHTTPHandler 日程事件路由处理
// ============================================================================

// EventHTTPHandler 日程事件 HTTP 处理器
//   依赖 service.ScheduleEventService，仅做协议转换
type EventHTTPHandler struct {
	eventSvc service.ScheduleEventService
}

// NewEventHTTPHandler 构造事件处理器
func NewEventHTTPHandler(eventSvc service.ScheduleEventService) *EventHTTPHandler {
	return &EventHTTPHandler{eventSvc: eventSvc}
}

// Register 注册事件相关路由到指定的 router group
//   group 应为已挂载 JWT 中间件的 /api/v1 组
func (h *EventHTTPHandler) Register(rg *gin.RouterGroup) {
	events := rg.Group("/events")
	{
		events.POST("", h.Create)                  // 创建事件
		events.GET("", h.List)                     // 列表查询
		events.GET("/visible", h.ListVisible)      // 我可见的全部
		events.GET("/attended", h.ListAttended)    // 我参与的
		events.GET("/range", h.ListByTimeRange)    // 日/周/月视图
		events.GET("/:id", h.GetByID)              // 详情
		events.PUT("/:id", h.Update)               // 更新
		events.DELETE("/:id", h.Delete)            // 删除
		events.PATCH("/:id/status", h.UpdateStatus) // 状态变更
		events.POST("/:id/meeting", h.LinkMeeting)    // 关联会议
		events.DELETE("/:id/meeting", h.UnlinkMeeting) // 取消关联
	}
}

// ----------------------------------------------------------------------------
// 路由处理方法
// ----------------------------------------------------------------------------

// Create 创建日程事件
// POST /api/v1/events
func (h *EventHTTPHandler) Create(c *gin.Context) {
	var req service.CreateEventRequest
	if !ShouldBindJSON(c, &req) {
		return
	}
	actorID := ActorIDFromContext(c)
	if actorID == 0 {
		Fail(c, service.ErrUnauthorized)
		return
	}
	resp, err := h.eventSvc.Create(c.Request.Context(), actorID, &req)
	if err != nil {
		Fail(c, err)
		return
	}
	OKCreated(c, resp)
}

// List 列表查询
// GET /api/v1/events?page=1&page_size=20&status=confirmed&...
func (h *EventHTTPHandler) List(c *gin.Context) {
	var req service.EventListRequest
	if !ShouldBindQuery(c, &req) {
		return
	}
	actorID := ActorIDFromContext(c)
	resp, err := h.eventSvc.List(c.Request.Context(), actorID, &req)
	if err != nil {
		Fail(c, err)
		return
	}
	OKList(c, resp.List, resp.Total, resp.Page, resp.PageSize)
}

// ListVisible 查询当前用户可见的全部日程
// GET /api/v1/events/visible?page=1&page_size=20
func (h *EventHTTPHandler) ListVisible(c *gin.Context) {
	var req service.PageRequest
	if !ShouldBindQuery(c, &req) {
		return
	}
	actorID := ActorIDFromContext(c)
	resp, err := h.eventSvc.ListVisibleByUser(c.Request.Context(), actorID, &req)
	if err != nil {
		Fail(c, err)
		return
	}
	OKList(c, resp.List, resp.Total, resp.Page, resp.PageSize)
}

// ListAttended 查询当前用户作为参与人的日程
// GET /api/v1/events/attended?page=1&page_size=20
func (h *EventHTTPHandler) ListAttended(c *gin.Context) {
	var req service.PageRequest
	if !ShouldBindQuery(c, &req) {
		return
	}
	actorID := ActorIDFromContext(c)
	resp, err := h.eventSvc.ListByAttendee(c.Request.Context(), actorID, &req)
	if err != nil {
		Fail(c, err)
		return
	}
	OKList(c, resp.List, resp.Total, resp.Page, resp.PageSize)
}

// ListByTimeRange 日/周/月视图查询
// GET /api/v1/events/range?from=2026-09-01T00:00:00Z&to=2026-09-30T23:59:59Z
//   from/to 使用 RFC3339 格式（含时区）
func (h *EventHTTPHandler) ListByTimeRange(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:    int(service.ErrCodeInvalidParam),
			Message: "from 和 to 参数必填，格式 RFC3339（如 2026-09-01T00:00:00Z）",
		})
		return
	}
	// 解析 RFC3339 时间
	from, err := parseTimeRFC3339(fromStr)
	if err != nil {
		Fail(c, service.NewBizError(service.ErrCodeInvalidParam, "from 时间格式错误，应为 RFC3339", err))
		return
	}
	to, err := parseTimeRFC3339(toStr)
	if err != nil {
		Fail(c, service.NewBizError(service.ErrCodeInvalidParam, "to 时间格式错误，应为 RFC3339", err))
		return
	}
	// 区间合法性校验
	if !to.After(from) {
		Fail(c, service.NewBizError(service.ErrCodeInvalidParam, "to 必须晚于 from", nil))
		return
	}
	actorID := ActorIDFromContext(c)
	list, err := h.eventSvc.ListByTimeRange(c.Request.Context(), actorID, from, to)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, list)
}

// GetByID 查询事件详情
// GET /api/v1/events/:id
func (h *EventHTTPHandler) GetByID(c *gin.Context) {
	id, ok := parsePathUint(c, "id")
	if !ok {
		return
	}
	actorID := ActorIDFromContext(c)
	resp, err := h.eventSvc.GetByID(c.Request.Context(), actorID, id)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, resp)
}

// Update 更新事件
// PUT /api/v1/events/:id
func (h *EventHTTPHandler) Update(c *gin.Context) {
	id, ok := parsePathUint(c, "id")
	if !ok {
		return
	}
	var req service.UpdateEventRequest
	if !ShouldBindJSON(c, &req) {
		return
	}
	actorID := ActorIDFromContext(c)
	resp, err := h.eventSvc.Update(c.Request.Context(), actorID, id, &req)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, resp)
}

// Delete 删除事件
// DELETE /api/v1/events/:id
func (h *EventHTTPHandler) Delete(c *gin.Context) {
	id, ok := parsePathUint(c, "id")
	if !ok {
		return
	}
	actorID := ActorIDFromContext(c)
	if err := h.eventSvc.Delete(c.Request.Context(), actorID, id); err != nil {
		Fail(c, err)
		return
	}
	OK(c, nil)
}

// UpdateStatus 更新事件状态
// PATCH /api/v1/events/:id/status  body: {"status":"confirmed"}
func (h *EventHTTPHandler) UpdateStatus(c *gin.Context) {
	id, ok := parsePathUint(c, "id")
	if !ok {
		return
	}
	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if !ShouldBindJSON(c, &body) {
		return
	}
	actorID := ActorIDFromContext(c)
	if err := h.eventSvc.UpdateStatus(c.Request.Context(), actorID, id, body.Status); err != nil {
		Fail(c, err)
		return
	}
	OK(c, nil)
}

// LinkMeeting 关联会议
// POST /api/v1/events/:id/meeting  body: {"meeting_id":123}
func (h *EventHTTPHandler) LinkMeeting(c *gin.Context) {
	id, ok := parsePathUint(c, "id")
	if !ok {
		return
	}
	var body struct {
		MeetingID uint `json:"meeting_id" binding:"required"`
	}
	if !ShouldBindJSON(c, &body) {
		return
	}
	actorID := ActorIDFromContext(c)
	if err := h.eventSvc.LinkMeeting(c.Request.Context(), actorID, id, body.MeetingID); err != nil {
		Fail(c, err)
		return
	}
	OK(c, nil)
}

// UnlinkMeeting 取消会议关联
// DELETE /api/v1/events/:id/meeting
func (h *EventHTTPHandler) UnlinkMeeting(c *gin.Context) {
	id, ok := parsePathUint(c, "id")
	if !ok {
		return
	}
	actorID := ActorIDFromContext(c)
	if err := h.eventSvc.UnlinkMeeting(c.Request.Context(), actorID, id); err != nil {
		Fail(c, err)
		return
	}
	OK(c, nil)
}

