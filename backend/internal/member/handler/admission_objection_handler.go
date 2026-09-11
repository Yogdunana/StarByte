package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// AdmissionObjection raises or reviews a probation objection.
// @Summary 候补期异议提出、复核与最终裁决
// @Tags 会员
// @Accept json
// @Produce json
// @Param id path string true "申请ID"
// @Param request body dto.AdmissionObjectionRequest true "异议处理"
// @Success 200 {object} response.Response
// @Router /member/applications/{id}/admission/objection [post]
// @Security BearerAuth
func (h *MemberHandler) AdmissionObjection(c *gin.Context) {
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
	var req dto.AdmissionObjectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if err := h.admission.Objection(c.Request.Context(), viewer, id, &req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
