package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/internship/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// CompleteInternship 完成实习
// @Summary 完成实习
// @Tags 实习
// @Accept json
// @Produce json
// @Param id path string true "实习ID"
// @Param request body dto.CompleteRequest true "报告与成果"
// @Success 200 {object} response.Response
// @Router /internships/{id}/complete [post]
func (h *InternshipHandler) CompleteInternship(c *gin.Context) {
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
	var req dto.CompleteRequest
	_ = c.ShouldBindJSON(&req)
	out, err := h.svc.Complete(c.Request.Context(), userID, id, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// SubmitReport 提交实习报告
// @Summary 提交实习报告
// @Tags 实习
// @Accept json
// @Produce json
// @Param id path string true "实习ID"
// @Param request body dto.ReportRequest true "报告"
// @Success 200 {object} response.Response
// @Router /internships/{id}/report [post]
func (h *InternshipHandler) SubmitReport(c *gin.Context) {
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
	var req dto.ReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.SubmitReport(c.Request.Context(), userID, id, req.Report, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// MentorComment 指导老师评价
// @Summary 指导老师评价
// @Tags 实习
// @Accept json
// @Produce json
// @Param id path string true "实习ID"
// @Param request body dto.MentorCommentRequest true "评价"
// @Success 200 {object} response.Response
// @Router /internships/{id}/mentor-comment [post]
func (h *InternshipHandler) MentorComment(c *gin.Context) {
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
	var req dto.MentorCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.MentorComment(c.Request.Context(), userID, id, req.MentorComment, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// GetConfig 获取实习配置
// @Summary 获取实习配置
// @Tags 实习
// @Produce json
// @Success 200 {object} response.Response
// @Router /system/internship-config [get]
func (h *InternshipHandler) GetConfig(c *gin.Context) {
	out, err := h.svc.GetConfig(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// UpdateConfig 更新实习配置
// @Summary 更新实习配置
// @Tags 实习
// @Accept json
// @Produce json
// @Param request body dto.InternshipConfigRequest true "配置"
// @Success 200 {object} response.Response
// @Router /system/internship-config [put]
func (h *InternshipHandler) UpdateConfig(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.InternshipConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.UpdateConfig(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
