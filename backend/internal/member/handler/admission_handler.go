package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// Admission returns the server-authorized signing actions and immutable decisions.
// @Summary 查看正式录用审批
// @Tags 会员
// @Produce json
// @Param id path string true "申请ID"
// @Success 200 {object} response.Response
// @Router /member/applications/{id}/admission [get]
// @Security BearerAuth
func (h *MemberHandler) Admission(c *gin.Context) {
	viewer, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.admission.Snapshot(c.Request.Context(), viewer, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// SignAdmission records an explicit signature; scores do not determine its decision.
// @Summary 正式签字审批
// @Tags 会员
// @Accept json
// @Produce json
// @Param id path string true "申请ID"
// @Param request body dto.SignAdmissionRequest true "签字决定"
// @Success 200 {object} response.Response
// @Router /member/applications/{id}/admission/sign [post]
// @Security BearerAuth
func (h *MemberHandler) SignAdmission(c *gin.Context) {
	viewer, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.SignAdmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.admission.Sign(c.Request.Context(), viewer, id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}
