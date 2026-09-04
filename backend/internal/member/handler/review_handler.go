package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// Approve 审核通过
// @Summary 审核通过
// @Tags 会员
// @Accept json
// @Produce json
// @Param id path string true "申请ID"
// @Param request body dto.ReviewCommentRequest true "意见"
// @Success 200 {object} response.Response
// @Router /member/applications/{id}/approve [post]
func (h *MemberHandler) Approve(c *gin.Context) {
	h.review(c, "approve")
}

// Reject 审核拒绝
// @Summary 审核拒绝
// @Tags 会员
// @Accept json
// @Produce json
// @Param id path string true "申请ID"
// @Param request body dto.ReviewCommentRequest true "意见"
// @Success 200 {object} response.Response
// @Router /member/applications/{id}/reject [post]
func (h *MemberHandler) Reject(c *gin.Context) {
	h.review(c, "reject")
}

// Supplement 要求补充材料
// @Summary 要求补充材料
// @Tags 会员
// @Accept json
// @Produce json
// @Param id path string true "申请ID"
// @Param request body dto.SupplementRequest true "补充要求"
// @Success 200 {object} response.Response
// @Router /member/applications/{id}/supplement [post]
func (h *MemberHandler) Supplement(c *gin.Context) {
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
	var req dto.SupplementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.Supplement(c.Request.Context(), userID, id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *MemberHandler) review(c *gin.Context, action string) {
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
	var req dto.ReviewCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	var (
		result any
		svcErr error
	)
	if action == "approve" {
		result, svcErr = h.svc.Approve(c.Request.Context(), userID, id, req.Comment)
	} else {
		result, svcErr = h.svc.Reject(c.Request.Context(), userID, id, req.Comment)
	}
	if svcErr != nil {
		response.Error(c, svcErr)
		return
	}
	response.OK(c, result)
}
