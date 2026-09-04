package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// Submit 提交入会申请
// @Summary 提交入会申请
// @Tags 会员
// @Accept json
// @Produce json
// @Param request body dto.SubmitApplicationRequest true "申请"
// @Success 200 {object} response.Response
// @Router /member/applications [post]
func (h *MemberHandler) Submit(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.SubmitApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.Submit(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// MyApplications 我的申请
// @Summary 我的申请
// @Tags 会员
// @Produce json
// @Success 200 {object} response.Response
// @Router /member/applications/my [get]
func (h *MemberHandler) MyApplications(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	list, err := h.svc.MyApplications(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

// Resubmit 补充材料后重新提交
// @Summary 补充材料后重新提交
// @Tags 会员
// @Accept json
// @Produce json
// @Param id path string true "申请ID"
// @Param request body dto.ResubmitApplicationRequest true "补充内容"
// @Success 200 {object} response.Response
// @Router /member/applications/{id}/resubmit [post]
func (h *MemberHandler) Resubmit(c *gin.Context) {
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
	var req dto.ResubmitApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.Resubmit(c.Request.Context(), userID, id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// ListApplications 申请列表
// @Summary 申请列表
// @Tags 会员
// @Produce json
// @Success 200 {object} response.Response
// @Router /member/applications [get]
func (h *MemberHandler) ListApplications(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ListApplicationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	list, total, err := h.svc.ListApplications(c.Request.Context(), userID, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	response.Page(c, list, total, req.Page, req.PageSize)
}

// GetApplication 申请详情
// @Summary 申请详情
// @Tags 会员
// @Produce json
// @Param id path string true "申请ID"
// @Success 200 {object} response.Response
// @Router /member/applications/{id} [get]
func (h *MemberHandler) GetApplication(c *gin.Context) {
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
	result, err := h.svc.GetApplication(c.Request.Context(), userID, id, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// ApplicationHistory 申请历史
// @Summary 申请历史
// @Tags 会员
// @Produce json
// @Param id path string true "申请ID"
// @Success 200 {object} response.Response
// @Router /member/applications/{id}/history [get]
func (h *MemberHandler) ApplicationHistory(c *gin.Context) {
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
	list, err := h.svc.ApplicationHistory(c.Request.Context(), userID, id, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

// ListDepartments 意向部门
// @Summary 入会意向部门
// @Tags 会员
// @Produce json
// @Success 200 {object} response.Response
// @Router /member/departments [get]
func (h *MemberHandler) ListDepartments(c *gin.Context) {
	list, err := h.svc.ListDepartments(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}
