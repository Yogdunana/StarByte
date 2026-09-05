package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// CreateSession 创建面试场次
// @Summary 创建面试场次
// @Tags 面试
// @Accept json
// @Produce json
// @Param request body dto.CreateSessionRequest true "场次"
// @Success 200 {object} response.Response
// @Router /interviews/sessions [post]
func (h *InterviewHandler) CreateSession(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.CreateSession(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// ListSessions 场次列表
// @Summary 场次列表
// @Tags 面试
// @Produce json
// @Success 200 {object} response.Response
// @Router /interviews/sessions [get]
func (h *InterviewHandler) ListSessions(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ListSessionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	list, total, err := h.svc.ListSessions(c.Request.Context(), userID, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	page, size := defaultPage(req.Page, req.PageSize)
	response.Page(c, list, total, page, size)
}

// GetSession 场次详情
// @Summary 场次详情
// @Tags 面试
// @Produce json
// @Param id path string true "场次ID"
// @Success 200 {object} response.Response
// @Router /interviews/sessions/{id} [get]
func (h *InterviewHandler) GetSession(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.GetSession(c.Request.Context(), id, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// UpdateSession 更新场次
// @Summary 更新场次
// @Tags 面试
// @Accept json
// @Produce json
// @Param id path string true "场次ID"
// @Success 200 {object} response.Response
// @Router /interviews/sessions/{id} [put]
func (h *InterviewHandler) UpdateSession(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.UpdateSession(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// DeleteSession 删除场次
// @Summary 删除场次
// @Tags 面试
// @Param id path string true "场次ID"
// @Success 200 {object} response.Response
// @Router /interviews/sessions/{id} [delete]
func (h *InterviewHandler) DeleteSession(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteSession(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// StartSession 开始场次
// @Summary 开始场次
// @Tags 面试
// @Param id path string true "场次ID"
// @Success 200 {object} response.Response
// @Router /interviews/sessions/{id}/start [post]
func (h *InterviewHandler) StartSession(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.StartSession(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// EndSession 结束场次
// @Summary 结束场次
// @Tags 面试
// @Param id path string true "场次ID"
// @Success 200 {object} response.Response
// @Router /interviews/sessions/{id}/end [post]
func (h *InterviewHandler) EndSession(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.EndSession(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// SessionQRCode 签到二维码
// @Summary 获取签到二维码
// @Tags 面试
// @Param id path string true "场次ID"
// @Success 200 {object} response.Response
// @Router /interviews/sessions/{id}/qrcode [get]
func (h *InterviewHandler) SessionQRCode(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	meta, png, err := h.svc.SessionQRCode(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	if c.Query("format") == "png" {
		c.Header("Content-Type", "image/png")
		c.Writer.WriteHeader(200)
		_, _ = c.Writer.Write(png)
		return
	}
	response.OK(c, meta)
}
