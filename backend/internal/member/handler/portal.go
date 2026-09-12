package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// Portal 会员门户（灰度，不替代入会申请）
// @Summary 会员门户
// @Tags 会员
// @Produce json
// @Success 200 {object} response.Response
// @Router /member/portal [get]
// @Security BearerAuth
func (h *MemberHandler) Portal(c *gin.Context) {
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
	response.OK(c, dto.PortalResponse{
		Applications: list,
		Hint:         "membership.portal",
	})
}

// RegisterPortal 挂上 /member/portal，由调用方注入特性开关中间件。
func RegisterPortal(r *gin.RouterGroup, h *MemberHandler, gate gin.HandlerFunc) {
	g := r.Group("/member")
	if gate != nil {
		g.Use(gate)
	}
	g.GET("/portal", h.Portal)
}
