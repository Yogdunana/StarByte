package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// CreateInterview 创建面试记录
// @Summary 创建面试记录
// @Tags 面试
// @Accept json
// @Produce json
// @Param request body dto.CreateInterviewRequest true "面试"
// @Success 200 {object} response.Response
// @Router /interviews [post]
func (h *InterviewHandler) CreateInterview(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateInterviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.CreateInterview(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// ListInterviews 面试列表
// @Summary 面试列表
// @Tags 面试
// @Produce json
// @Success 200 {object} response.Response
// @Router /interviews [get]
func (h *InterviewHandler) ListInterviews(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ListInterviewRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	list, total, err := h.svc.ListInterviews(c.Request.Context(), userID, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	page, size := defaultPage(req.Page, req.PageSize)
	response.Page(c, list, total, page, size)
}

// GetInterview 面试详情
// @Summary 面试详情
// @Tags 面试
// @Param id path string true "面试ID"
// @Success 200 {object} response.Response
// @Router /interviews/{id} [get]
func (h *InterviewHandler) GetInterview(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.GetInterview(c.Request.Context(), id, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// AssignEvaluators 分配面试官
// @Summary 分配面试官
// @Tags 面试
// @Param id path string true "面试ID"
// @Success 200 {object} response.Response
// @Router /interviews/{id}/assign [post]
func (h *InterviewHandler) AssignEvaluators(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.AssignEvaluatorsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.AssignEvaluators(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// Checkin 签到
// @Summary 面试签到
// @Tags 面试
// @Param id path string true "面试ID"
// @Success 200 {object} response.Response
// @Router /interviews/{id}/checkin [post]
func (h *InterviewHandler) Checkin(c *gin.Context) {
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
	var req dto.CheckinRequest
	_ = c.ShouldBindJSON(&req)
	out, err := h.svc.Checkin(c.Request.Context(), userID, id, req.Token)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// StartInterview 开始面试
// @Summary 开始面试
// @Tags 面试
// @Param id path string true "面试ID"
// @Success 200 {object} response.Response
// @Router /interviews/{id}/start [post]
func (h *InterviewHandler) StartInterview(c *gin.Context) {
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
	out, err := h.svc.StartInterview(c.Request.Context(), userID, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// EndInterview 结束面试
// @Summary 结束面试
// @Tags 面试
// @Param id path string true "面试ID"
// @Success 200 {object} response.Response
// @Router /interviews/{id}/end [post]
func (h *InterviewHandler) EndInterview(c *gin.Context) {
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
	out, err := h.svc.EndInterview(c.Request.Context(), userID, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// MyInterviews 我的面试
// @Summary 我的面试
// @Tags 面试
// @Produce json
// @Success 200 {object} response.Response
// @Router /interviews/my [get]
func (h *InterviewHandler) MyInterviews(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var status *int16
	var req struct {
		Status *int16 `form:"status"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	status = req.Status
	list, err := h.svc.MyInterviews(c.Request.Context(), userID, status)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}
