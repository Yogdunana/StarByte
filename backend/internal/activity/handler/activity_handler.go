package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/activity/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateActivity 创建活动
func (h *ActivityHandler) CreateActivity(c *gin.Context) {
	operator, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.CreateActivity(c.Request.Context(), operator, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// UpdateActivity 更新活动
func (h *ActivityHandler) UpdateActivity(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.UpdateActivity(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// DeleteActivity 删除活动
func (h *ActivityHandler) DeleteActivity(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteActivity(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithoutData(c)
}

// GetActivity 活动详情
func (h *ActivityHandler) GetActivity(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.svc.GetActivity(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// ListActivities 活动列表
func (h *ActivityHandler) ListActivities(c *gin.Context) {
	var req dto.ListActivityRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	req.Page, req.PageSize = defaultPage(req.Page, req.PageSize)
	list, total, err := h.svc.ListActivities(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, req.Page, req.PageSize)
}

// StartActivity 开始活动
func (h *ActivityHandler) StartActivity(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.svc.StartActivity(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// EndActivity 结束活动
func (h *ActivityHandler) EndActivity(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.svc.EndActivity(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// CancelActivity 取消活动
func (h *ActivityHandler) CancelActivity(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var body struct {
		Reason string `json:"reason" binding:"max=500"`
	}
	_ = c.ShouldBindJSON(&body)
	result, err := h.svc.CancelActivity(c.Request.Context(), id, body.Reason)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// Register 报名
func (h *ActivityHandler) Register(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.svc.Register(c.Request.Context(), id, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// CancelRegistration 取消报名
func (h *ActivityHandler) CancelRegistration(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.CancelRegistration(c.Request.Context(), id, userID); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithoutData(c)
}

// ListRegistrations 报名列表
func (h *ActivityHandler) ListRegistrations(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	list, err := h.svc.ListRegistrations(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

// ApproveRegistration 审批报名
func (h *ActivityHandler) ApproveRegistration(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	uid, err := uuid.Parse(c.Param("uid"))
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}
	var req dto.ApproveRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.ApproveRegistration(c.Request.Context(), id, uid, req.Approve, req.Reason)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// Checkin 签到
func (h *ActivityHandler) Checkin(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CheckinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.Checkin(c.Request.Context(), id, userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// GetStats 活动统计
func (h *ActivityHandler) GetStats(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.svc.GetStats(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// SubmitSurvey 提交满意度调查
func (h *ActivityHandler) SubmitSurvey(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.SurveyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.SubmitSurvey(c.Request.Context(), id, userID, &req); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithoutData(c)
}


