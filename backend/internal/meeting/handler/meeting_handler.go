package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// CreateMeeting 创建会议
// @Summary 创建会议
// @Tags 会议
// @Accept json
// @Produce json
// @Param request body dto.CreateMeetingRequest true "会议"
// @Success 200 {object} response.Response
// @Router /meetings [post]
func (h *MeetingHandler) CreateMeeting(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.CreateMeeting(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// ListMeetings 会议列表
// @Summary 会议列表
// @Tags 会议
// @Produce json
// @Success 200 {object} response.Response
// @Router /meetings [get]
func (h *MeetingHandler) ListMeetings(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ListMeetingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	list, total, err := h.svc.ListMeetings(c.Request.Context(), userID, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	page, size := defaultPage(req.Page, req.PageSize)
	response.Page(c, list, total, page, size)
}

// GetMeeting 会议详情
// @Summary 会议详情
// @Tags 会议
// @Produce json
// @Param id path string true "会议ID"
// @Success 200 {object} response.Response
// @Router /meetings/{id} [get]
func (h *MeetingHandler) GetMeeting(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.GetMeeting(c.Request.Context(), id, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// UpdateMeeting 更新会议
// @Summary 更新会议
// @Tags 会议
// @Router /meetings/{id} [put]
func (h *MeetingHandler) UpdateMeeting(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.UpdateMeeting(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// DeleteMeeting 删除会议
// @Summary 删除会议
// @Tags 会议
// @Router /meetings/{id} [delete]
func (h *MeetingHandler) DeleteMeeting(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteMeeting(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// StartMeeting 开始会议
// @Summary 开始会议
// @Tags 会议
// @Router /meetings/{id}/start [post]
func (h *MeetingHandler) StartMeeting(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.StartMeeting(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// EndMeeting 结束会议
// @Summary 结束会议
// @Tags 会议
// @Router /meetings/{id}/end [post]
func (h *MeetingHandler) EndMeeting(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.EndMeeting(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// CancelMeeting 取消会议
// @Summary 取消会议
// @Tags 会议
// @Router /meetings/{id}/cancel [post]
func (h *MeetingHandler) CancelMeeting(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CancelMeetingRequest
	_ = c.ShouldBindJSON(&req)
	out, err := h.svc.CancelMeeting(c.Request.Context(), id, req.Reason)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// UpdateMinutes 更新纪要
// @Summary 更新纪要
// @Tags 会议
// @Router /meetings/{id}/minutes [put]
func (h *MeetingHandler) UpdateMinutes(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.MinutesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.UpdateMinutes(c.Request.Context(), id, req.Minutes)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// MeetingQRCode 签到二维码
// @Summary 获取签到二维码
// @Tags 会议
// @Router /meetings/{id}/qrcode [get]
func (h *MeetingHandler) MeetingQRCode(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	meta, png, err := h.svc.MeetingQRCode(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	if c.Query("format") == "png" {
		c.Header("Content-Type", "image/png")
		c.Writer.WriteHeader(200)
		_, _ = c.Writer.Write(png)
		return
	}
	response.OK(c, meta)
}
