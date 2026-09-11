package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListEvents 事件列表
// @Summary 事件列表
// @Tags 日程
// @Produce json
// @Router /schedules/events [get]
// @Security BearerAuth
func (h *Handler) ListEvents(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ListEventRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	list, total, page, size, err := h.svc.ListEvents(c.Request.Context(), userID, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, page, size)
}

// RangeEvents 范围查询（展开有限重复）
// @Summary 范围查询
// @Tags 日程
// @Produce json
// @Router /schedules/events/range [get]
// @Security BearerAuth
func (h *Handler) RangeEvents(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.RangeEventRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: 需要 start/end")
		return
	}
	list, err := h.svc.RangeEvents(c.Request.Context(), userID, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

// CreateEvent 创建事件
// @Summary 创建事件
// @Tags 日程
// @Accept json
// @Produce json
// @Param request body dto.CreateEventRequest true "事件"
// @Router /schedules/events [post]
// @Security BearerAuth
func (h *Handler) CreateEvent(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.CreateEvent(c.Request.Context(), userID, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// GetEvent 事件详情
// @Summary 事件详情
// @Tags 日程
// @Router /schedules/events/{id} [get]
// @Security BearerAuth
func (h *Handler) GetEvent(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.GetEvent(c.Request.Context(), userID, id, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// UpdateEvent 更新事件
// @Summary 更新事件
// @Tags 日程
// @Router /schedules/events/{id} [put]
// @Security BearerAuth
func (h *Handler) UpdateEvent(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.UpdateEvent(c.Request.Context(), userID, id, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// DeleteEvent 删除事件
// @Summary 删除事件
// @Tags 日程
// @Router /schedules/events/{id} [delete]
// @Security BearerAuth
func (h *Handler) DeleteEvent(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteEvent(c.Request.Context(), userID, id, dataScope(c)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// SetReminders 设置提醒
// @Summary 设置提醒
// @Tags 日程
// @Router /schedules/events/{id}/remind [post]
// @Security BearerAuth
func (h *Handler) SetReminders(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.RemindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.SetReminders(c.Request.Context(), userID, id, req.Minutes, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// RSVP 回复邀请
// @Summary 回复邀请
// @Tags 日程
// @Router /schedules/events/{id}/rsvp [post]
// @Security BearerAuth
func (h *Handler) RSVP(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.RSVPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.RSVP(c.Request.Context(), userID, id, req.Response, dataScope(c)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
