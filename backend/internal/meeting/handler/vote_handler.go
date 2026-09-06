package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// CreateVote 创建投票
// @Summary 创建投票
// @Description 创建投票
// @Tags 会议
// @Router /meetings/{id}/votes [post]
// @Security BearerAuth
func (h *MeetingHandler) CreateVote(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.CreateVote(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// ListVotes 投票列表
// @Summary 投票列表
// @Description 投票列表
// @Tags 会议
// @Router /meetings/{id}/votes [get]
// @Security BearerAuth
func (h *MeetingHandler) ListVotes(c *gin.Context) {
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
	out, err := h.svc.ListVotes(c.Request.Context(), id, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// GetVote 投票详情
// @Summary 投票详情
// @Description 投票详情
// @Tags 会议
// @Router /votes/{id} [get]
// @Security BearerAuth
func (h *MeetingHandler) GetVote(c *gin.Context) {
	userID, _ := getUserID(c)
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.GetVote(c.Request.Context(), id, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// CastVote 投票
// @Summary 投票
// @Description 投票
// @Tags 会议
// @Router /votes/{id}/cast [post]
// @Security BearerAuth
func (h *MeetingHandler) CastVote(c *gin.Context) {
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
	var req dto.CastVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.CastVote(c.Request.Context(), id, userID, req.OptionKey); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// VoteResult 投票结果
// @Summary 投票结果
// @Description 投票结果
// @Tags 会议
// @Router /votes/{id}/result [get]
// @Security BearerAuth
func (h *MeetingHandler) VoteResult(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.VoteResult(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// CloseVote 关闭投票
// @Summary 关闭投票
// @Description 关闭投票
// @Tags 会议
// @Router /votes/{id}/close [post]
// @Security BearerAuth
func (h *MeetingHandler) CloseVote(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.CloseVote(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// MyVote 我的投票记录
// @Summary 我的投票记录
// @Description 我的投票记录
// @Tags 会议
// @Router /votes/{id}/my [get]
// @Security BearerAuth
func (h *MeetingHandler) MyVote(c *gin.Context) {
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
	out, err := h.svc.MyVote(c.Request.Context(), id, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// GetWeightConfig 获取权重配置
// @Summary 获取权重配置
// @Description 获取权重配置
// @Tags 会议
// @Router /system/vote-weight-config [get]
// @Security BearerAuth
func (h *MeetingHandler) GetWeightConfig(c *gin.Context) {
	out, err := h.svc.GetWeightConfig(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// UpdateWeightConfig 更新权重配置
// @Summary 更新权重配置
// @Description 更新权重配置
// @Tags 会议
// @Router /system/vote-weight-config [put]
// @Security BearerAuth
func (h *MeetingHandler) UpdateWeightConfig(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.VoteWeightConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.UpdateWeightConfig(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
