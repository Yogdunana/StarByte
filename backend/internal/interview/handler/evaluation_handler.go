package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// SubmitEvaluations 提交评分
// @Summary 提交评分
// @Tags 面试
// @Accept json
// @Produce json
// @Param id path string true "面试ID"
// @Param request body dto.SubmitEvaluationsRequest true "评分"
// @Success 200 {object} response.Response
// @Router /interviews/{id}/evaluations [post]
func (h *InterviewHandler) SubmitEvaluations(c *gin.Context) {
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
	var req dto.SubmitEvaluationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.SubmitEvaluations(c.Request.Context(), userID, id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// GetEvaluations 查看评分汇总
// @Summary 查看评分
// @Tags 面试
// @Param id path string true "面试ID"
// @Success 200 {object} response.Response
// @Router /interviews/{id}/evaluations [get]
func (h *InterviewHandler) GetEvaluations(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.GetEvaluations(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// UpdateEvaluation 更新评分
// @Summary 更新评分
// @Tags 面试
// @Param id path string true "面试ID"
// @Param eid path string true "评分ID"
// @Success 200 {object} response.Response
// @Router /interviews/{id}/evaluations/{eid} [put]
func (h *InterviewHandler) UpdateEvaluation(c *gin.Context) {
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
	eid, err := parseNamedID(c, "eid")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.UpdateEvaluation(c.Request.Context(), userID, id, eid, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// SubmitResult 提交面试结果
// @Summary 提交面试结果
// @Tags 面试
// @Param id path string true "面试ID"
// @Success 200 {object} response.Response
// @Router /interviews/{id}/result [post]
func (h *InterviewHandler) SubmitResult(c *gin.Context) {
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
	var req dto.SubmitResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.SubmitResult(c.Request.Context(), userID, id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// ListDimensions 维度列表
// @Summary 评分维度列表
// @Tags 面试
// @Success 200 {object} response.Response
// @Router /interviews/dimensions [get]
func (h *InterviewHandler) ListDimensions(c *gin.Context) {
	out, err := h.svc.ListDimensions(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// CreateDimension 创建维度
// @Summary 创建评分维度
// @Tags 面试
// @Success 200 {object} response.Response
// @Router /interviews/dimensions [post]
func (h *InterviewHandler) CreateDimension(c *gin.Context) {
	var req dto.CreateDimensionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.CreateDimension(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// UpdateDimension 更新维度
// @Summary 更新评分维度
// @Tags 面试
// @Success 200 {object} response.Response
// @Router /interviews/dimensions/{id} [put]
func (h *InterviewHandler) UpdateDimension(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateDimensionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.UpdateDimension(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// DeleteDimension 删除维度
// @Summary 删除评分维度
// @Tags 面试
// @Success 200 {object} response.Response
// @Router /interviews/dimensions/{id} [delete]
func (h *InterviewHandler) DeleteDimension(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteDimension(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// Stats 面试统计
// @Summary 面试统计
// @Tags 面试
// @Success 200 {object} response.Response
// @Router /interviews/stats [get]
func (h *InterviewHandler) Stats(c *gin.Context) {
	var q dto.StatsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Stats(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
