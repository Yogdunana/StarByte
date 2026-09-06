package handler

import (
	"net/http"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListProfiles 档案列表
// @Summary 档案列表
// @Description 档案列表
// @Tags 会员
// @Produce json
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /member/profiles [get]
// @Security BearerAuth
func (h *MemberHandler) ListProfiles(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ListProfileRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	list, total, err := h.svc.ListProfiles(c.Request.Context(), userID, &req, dataScope(c))
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

// GetProfile 档案详情
// @Summary 档案详情
// @Description 档案详情
// @Tags 会员
// @Produce json
// @Param id path string true "档案ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /member/profiles/{id} [get]
// @Security BearerAuth
func (h *MemberHandler) GetProfile(c *gin.Context) {
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
	result, err := h.svc.GetProfile(c.Request.Context(), userID, id, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// UpdateProfile 更新档案
// @Summary 更新档案
// @Description 更新档案
// @Tags 会员
// @Accept json
// @Produce json
// @Param id path string true "档案ID"
// @Param request body dto.UpdateProfileRequest true "档案"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /member/profiles/{id} [put]
// @Security BearerAuth
func (h *MemberHandler) UpdateProfile(c *gin.Context) {
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
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.UpdateProfile(c.Request.Context(), userID, id, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// ProfileHistory 档案变更历史
// @Summary 档案变更历史
// @Description 档案变更历史
// @Tags 会员
// @Produce json
// @Param id path string true "档案ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /member/profiles/{id}/history [get]
// @Security BearerAuth
func (h *MemberHandler) ProfileHistory(c *gin.Context) {
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
	list, err := h.svc.ProfileHistory(c.Request.Context(), userID, id, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

// UpdateProfileStatus 变更档案状态
// @Summary 变更档案状态
// @Description 变更档案状态
// @Tags 会员
// @Accept json
// @Produce json
// @Param id path string true "档案ID"
// @Param request body dto.UpdateProfileStatusRequest true "状态"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /member/profiles/{id}/status [put]
// @Security BearerAuth
func (h *MemberHandler) UpdateProfileStatus(c *gin.Context) {
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
	var req dto.UpdateProfileStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.UpdateProfileStatus(c.Request.Context(), userID, id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// ExportProfiles 导出档案 PDF
// @Summary 导出档案 PDF
// @Description 导出档案 PDF
// @Tags 会员
// @Produce application/pdf
// @Success 200 {file} file
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /member/profiles/export [get]
// @Security BearerAuth
func (h *MemberHandler) ExportProfiles(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ListProfileRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	pdf, err := h.svc.ExportProfiles(c.Request.Context(), userID, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	c.Header("Content-Disposition", `attachment; filename="member-profiles.pdf"`)
	c.Data(http.StatusOK, "application/pdf", pdf)
}
