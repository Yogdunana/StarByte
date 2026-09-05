package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListAttendees 参会人列表
// @Summary 参会人列表
// @Tags 会议
// @Router /meetings/{id}/attendees [get]
func (h *MeetingHandler) ListAttendees(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.ListAttendees(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// AddAttendees 添加参会人
// @Summary 添加参会人
// @Tags 会议
// @Router /meetings/{id}/attendees [post]
func (h *MeetingHandler) AddAttendees(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.AddAttendeesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	uids := make([]uuid.UUID, 0, len(req.UserIDs))
	for _, raw := range req.UserIDs {
		uid, err := uuid.Parse(raw)
		if err != nil {
			response.BadRequest(c, "无效的用户ID")
			return
		}
		uids = append(uids, uid)
	}
	out, err := h.svc.AddAttendees(c.Request.Context(), id, uids)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// RemoveAttendee 移除参会人
// @Summary 移除参会人
// @Tags 会议
// @Router /meetings/{id}/attendees/{uid} [delete]
func (h *MeetingHandler) RemoveAttendee(c *gin.Context) {
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
	if err := h.svc.RemoveAttendee(c.Request.Context(), id, uid); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// Checkin 会议签到
// @Summary 会议签到
// @Tags 会议
// @Router /meetings/{id}/checkin [post]
func (h *MeetingHandler) Checkin(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var body struct {
		Token string `json:"token"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.Token == "" {
		body.Token = c.Query("token")
	}
	out, err := h.svc.Checkin(c.Request.Context(), id, userID, body.Token)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
