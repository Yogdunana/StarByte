package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Approve 审核通过
// @Summary 审核通过
// @Description 审核通过
// @Tags 会员
// @Accept json
// @Produce json
// @Param id path string true "申请ID"
// @Param request body dto.ReviewCommentRequest true "意见"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /member/applications/{id}/approve [post]
// @Security BearerAuth
func (h *MemberHandler) Approve(c *gin.Context) {
	h.review(c, "approve")
}

// Reject 审核拒绝
// @Summary 审核拒绝
// @Description 审核拒绝
// @Tags 会员
// @Accept json
// @Produce json
// @Param id path string true "申请ID"
// @Param request body dto.ReviewCommentRequest true "意见"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /member/applications/{id}/reject [post]
// @Security BearerAuth
func (h *MemberHandler) Reject(c *gin.Context) {
	h.review(c, "reject")
}

// Supplement 要求补充材料
// @Summary 要求补充材料
// @Description 要求补充材料
// @Tags 会员
// @Accept json
// @Produce json
// @Param id path string true "申请ID"
// @Param request body dto.SupplementRequest true "补充要求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /member/applications/{id}/supplement [post]
// @Security BearerAuth
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

// Transfer 转交当前审批人
// @Summary 转交入会审批
// @Description 将当前入会审批环节转交给其他用户
// @Tags 会员
// @Accept json
// @Produce json
// @Param id path string true "申请ID"
// @Param request body dto.TransferApplicationRequest true "转交"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /member/applications/{id}/transfer [post]
// @Security BearerAuth
func (h *MemberHandler) Transfer(c *gin.Context) {
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
	var req dto.TransferApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	target, err := uuid.Parse(req.TargetUserID)
	if err != nil {
		response.BadRequest(c, "请选择有效的转交对象")
		return
	}
	result, err := h.svc.Transfer(c.Request.Context(), userID, id, target, req.Comment)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// ApplicationProgress 入会流程进度
// @Summary 入会流程进度
// @Description 返回当前申请对应流程实例的可视化步骤
// @Tags 会员
// @Produce json
// @Param id path string true "申请ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /member/applications/{id}/progress [get]
// @Security BearerAuth
func (h *MemberHandler) ApplicationProgress(c *gin.Context) {
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
	result, err := h.svc.ApplicationProgress(c.Request.Context(), userID, id, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// TransferCandidates 转交候选人
// @Summary 入会审批转交候选人
// @Description 搜索可接收当前入会审批的用户
// @Tags 会员
// @Produce json
// @Param id path string true "申请ID"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Response
// @Router /member/applications/{id}/transfer-candidates [get]
// @Security BearerAuth
func (h *MemberHandler) TransferCandidates(c *gin.Context) {
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
	result, err := h.svc.TransferCandidates(c.Request.Context(), userID, id, c.Query("keyword"))
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
