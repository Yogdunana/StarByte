package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListCalendars 日历列表
// @Summary 日历列表
// @Tags 日程
// @Produce json
// @Router /schedules/calendars [get]
// @Security BearerAuth
func (h *Handler) ListCalendars(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ListCalendarRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	list, total, page, size, err := h.svc.ListCalendars(c.Request.Context(), userID, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, page, size)
}

// CreateCalendar 创建日历
// @Summary 创建日历
// @Tags 日程
// @Accept json
// @Produce json
// @Param request body dto.CreateCalendarRequest true "日历"
// @Router /schedules/calendars [post]
// @Security BearerAuth
func (h *Handler) CreateCalendar(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.CreateCalendar(c.Request.Context(), userID, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// GetCalendar 日历详情
// @Summary 日历详情
// @Tags 日程
// @Produce json
// @Param id path string true "日历ID"
// @Router /schedules/calendars/{id} [get]
// @Security BearerAuth
func (h *Handler) GetCalendar(c *gin.Context) {
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
	out, err := h.svc.GetCalendar(c.Request.Context(), userID, id, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// UpdateCalendar 更新日历
// @Summary 更新日历
// @Tags 日程
// @Router /schedules/calendars/{id} [put]
// @Security BearerAuth
func (h *Handler) UpdateCalendar(c *gin.Context) {
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
	var req dto.UpdateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.UpdateCalendar(c.Request.Context(), userID, id, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// DeleteCalendar 删除日历
// @Summary 删除日历
// @Tags 日程
// @Router /schedules/calendars/{id} [delete]
// @Security BearerAuth
func (h *Handler) DeleteCalendar(c *gin.Context) {
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
	if err := h.svc.DeleteCalendar(c.Request.Context(), userID, id, dataScope(c)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// ListMembers 日历成员
// @Summary 日历成员
// @Tags 日程
// @Router /schedules/calendars/{id}/members [get]
// @Security BearerAuth
func (h *Handler) ListMembers(c *gin.Context) {
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
	out, err := h.svc.ListMembers(c.Request.Context(), userID, id, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// AddMember 添加日历成员
// @Summary 添加日历成员
// @Tags 日程
// @Router /schedules/calendars/{id}/members [post]
// @Security BearerAuth
func (h *Handler) AddMember(c *gin.Context) {
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
	var req dto.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.AddMember(c.Request.Context(), userID, id, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// RemoveMember 移除日历成员
// @Summary 移除日历成员
// @Tags 日程
// @Router /schedules/calendars/{id}/members/{uid} [delete]
// @Security BearerAuth
func (h *Handler) RemoveMember(c *gin.Context) {
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
	uid, err := parseNamedID(c, "uid")
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.RemoveMember(c.Request.Context(), userID, id, uid, dataScope(c)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
