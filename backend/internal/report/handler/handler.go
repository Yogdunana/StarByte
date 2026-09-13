package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/report/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// List 汇报列表
// @Summary 汇报列表
// @Tags 工作汇报
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param report_type query string false "汇报类型" Enums(daily, weekly, monthly)
// @Param review_status query string false "审核状态" Enums(pending, approved, rejected)
// @Param user_id query string false "用户 UUID"
// @Param department_id query string false "部门 UUID"
// @Param period_start query string false "筛选起始日期" Format(date)
// @Param period_end query string false "筛选结束日期" Format(date)
// @Success 200 {object} response.Response{data=response.PageResponse{list=[]dto.ReportResponse}}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /reports [get]
// @Security BearerAuth
func (h *Handler) List(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	var req dto.ListReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	list, total, page, pageSize, err := h.svc.List(
		c.Request.Context(),
		userID,
		&req,
		dataScope(c),
	)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Page(c, list, total, page, pageSize)
}
