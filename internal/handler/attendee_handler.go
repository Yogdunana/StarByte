// Package handler 参与人 HTTP 路由处理
package handler

import (
	"schedule-service/internal/service"

	"github.com/gin-gonic/gin"
)

// ============================================================================
// AttendeeHTTPHandler 参与人路由处理
// ============================================================================

// AttendeeHTTPHandler 参与人 HTTP 处理器
type AttendeeHTTPHandler struct {
	attendeeSvc service.ScheduleAttendeeService
}

// NewAttendeeHTTPHandler 构造参与人处理器
func NewAttendeeHTTPHandler(attendeeSvc service.ScheduleAttendeeService) *AttendeeHTTPHandler {
	return &AttendeeHTTPHandler{attendeeSvc: attendeeSvc}
}

// Register 注册参与人相关路由
//   依附在 /events/:event_id/attendees 下（RESTful 嵌套资源）
//   group 应为已挂载 JWT 中间件的 /api/v1 组
func (h *AttendeeHTTPHandler) Register(rg *gin.RouterGroup) {
	// 参与人作为事件的子资源
	attendees := rg.Group("/events/:event_id/attendees")
	{
		attendees.POST("", h.Add)                 // 批量添加
		attendees.GET("", h.ListByEvent)          // 事件参与人列表
		attendees.GET("/:id", h.GetByID)           // 参与人详情
		attendees.PATCH("/:id", h.Update)          // 更新角色
		attendees.DELETE("/:id", h.Remove)        // 移除参与人
		attendees.POST("/:id/respond", h.Respond) // 响应邀请
	}

	// 当前用户的参与人视图（独立路径）
	rg.GET("/me/attendees", h.ListByUser)
}

// Add 添加参与人（支持单条或批量）
// POST /api/v1/events/:event_id/attendees
//   body: 单个 {"user_id":1,"role":"required"}
//      或批量 [{"user_id":1,"role":"required"}, {...}]
//   判断策略：尝试解析为数组，否则解析为单对象
func (h *AttendeeHTTPHandler) Add(c *gin.Context) {
	eventID, ok := parsePathUint(c, "event_id")
	if !ok {
		return
	}
	actorID := ActorIDFromContext(c)

	// 优先尝试批量解析
	var batchReq []service.CreateAttendeeRequest
	if err := c.ShouldBindJSON(&batchReq); err == nil && len(batchReq) > 0 {
		resp, err := h.attendeeSvc.BatchAdd(c.Request.Context(), actorID, eventID, batchReq)
		if err != nil {
			Fail(c, err)
			return
		}
		OKCreated(c, resp)
		return
	}

	// 回退到单个解析
	var req service.CreateAttendeeRequest
	if !ShouldBindJSON(c, &req) {
		return
	}
	resp, err := h.attendeeSvc.Add(c.Request.Context(), actorID, eventID, &req)
	if err != nil {
		Fail(c, err)
		return
	}
	OKCreated(c, resp)
}

// ListByEvent 查询事件参与人列表
// GET /api/v1/events/:event_id/attendees?page=1&page_size=20&role=required
func (h *AttendeeHTTPHandler) ListByEvent(c *gin.Context) {
	eventID, ok := parsePathUint(c, "event_id")
	if !ok {
		return
	}
	var req service.PageRequest
	if !ShouldBindQuery(c, &req) {
		return
	}
	actorID := ActorIDFromContext(c)
	resp, err := h.attendeeSvc.ListByEvent(c.Request.Context(), actorID, eventID, &req)
	if err != nil {
		Fail(c, err)
		return
	}
	OKList(c, resp.List, resp.Total, resp.Page, resp.PageSize)
}

// GetByID 查询参与人记录详情
// GET /api/v1/events/:event_id/attendees/:id
func (h *AttendeeHTTPHandler) GetByID(c *gin.Context) {
	id, ok := parsePathUint(c, "id")
	if !ok {
		return
	}
	actorID := ActorIDFromContext(c)
	resp, err := h.attendeeSvc.GetByID(c.Request.Context(), actorID, id)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, resp)
}

// Update 更新参与人角色
// PATCH /api/v1/events/:event_id/attendees/:id  body: {"role":"optional"}
func (h *AttendeeHTTPHandler) Update(c *gin.Context) {
	id, ok := parsePathUint(c, "id")
	if !ok {
		return
	}
	var body struct {
		Role string `json:"role" binding:"required"`
	}
	if !ShouldBindJSON(c, &body) {
		return
	}
	actorID := ActorIDFromContext(c)
	resp, err := h.attendeeSvc.Update(c.Request.Context(), actorID, id, body.Role)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, resp)
}

// Remove 移除参与人
// DELETE /api/v1/events/:event_id/attendees/:id
func (h *AttendeeHTTPHandler) Remove(c *gin.Context) {
	id, ok := parsePathUint(c, "id")
	if !ok {
		return
	}
	actorID := ActorIDFromContext(c)
	if err := h.attendeeSvc.Remove(c.Request.Context(), actorID, id); err != nil {
		Fail(c, err)
		return
	}
	OK(c, nil)
}

// Respond 响应事件邀请
// POST /api/v1/events/:event_id/attendees/:id/respond  body: {"status":"accepted"}
func (h *AttendeeHTTPHandler) Respond(c *gin.Context) {
	id, ok := parsePathUint(c, "id")
	if !ok {
		return
	}
	var req service.RespondInviteRequest
	if !ShouldBindJSON(c, &req) {
		return
	}
	actorID := ActorIDFromContext(c)
	resp, err := h.attendeeSvc.Respond(c.Request.Context(), actorID, id, &req)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, resp)
}

// ListByUser 查询当前用户的参与人视图
// GET /api/v1/me/attendees?page=1&page_size=20&status=pending
func (h *AttendeeHTTPHandler) ListByUser(c *gin.Context) {
	var req service.AttendeeListRequest
	if !ShouldBindQuery(c, &req) {
		return
	}
	actorID := ActorIDFromContext(c)
	resp, err := h.attendeeSvc.ListByUser(c.Request.Context(), actorID, &req)
	if err != nil {
		Fail(c, err)
		return
	}
	OKList(c, resp.List, resp.Total, resp.Page, resp.PageSize)
}
