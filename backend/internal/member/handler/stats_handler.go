package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// ApplicationStats 申请统计
// @Summary 申请统计
// @Tags 会员
// @Produce json
// @Param start_date query string false "开始日期"
// @Param end_date query string false "结束日期"
// @Param group_by query string false "date/department/type"
// @Success 200 {object} response.Response
// @Router /member/stats/applications [get]
func (h *MemberHandler) ApplicationStats(c *gin.Context) {
	var q dto.StatsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.ApplicationStats(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// MemberStats 会员分布
// @Summary 会员分布
// @Tags 会员
// @Produce json
// @Param group_by query string false "department/grade/type/status"
// @Success 200 {object} response.Response
// @Router /member/stats/members [get]
func (h *MemberHandler) MemberStats(c *gin.Context) {
	var q dto.StatsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.MemberStats(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}
