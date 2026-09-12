package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/feature/dto"
	"github.com/Yogdunana/StarByte/backend/internal/feature/service"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct{ svc service.Service }

func New(svc service.Service) *Handler { return &Handler{svc: svc} }

// List 特性开关列表
// @Summary 特性开关列表
// @Tags 特性开关
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页条数"
// @Param keyword query string false "关键词"
// @Param group query string false "分组"
// @Param enabled query bool false "是否启用"
// @Success 200 {object} response.Response
// @Router /system/features [get]
// @Security BearerAuth
func (h *Handler) List(c *gin.Context) {
	var q dto.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	q.Page, q.PageSize = defaultPage(q.Page, q.PageSize)
	list, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

// Create 创建特性开关
// @Summary 创建特性开关
// @Tags 特性开关
// @Accept json
// @Produce json
// @Param request body dto.CreateFlagRequest true "开关"
// @Success 200 {object} response.Response
// @Router /system/features [post]
// @Security BearerAuth
func (h *Handler) Create(c *gin.Context) {
	uid, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	out, err := h.svc.Create(c.Request.Context(), uid, &req)
	write(c, out, err)
}

// Get 开关详情
// @Summary 开关详情
// @Tags 特性开关
// @Produce json
// @Param id path string true "开关 ID"
// @Success 200 {object} response.Response
// @Router /system/features/{id} [get]
// @Security BearerAuth
func (h *Handler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.Get(c.Request.Context(), id)
	write(c, out, err)
}

// Update 更新开关
// @Summary 更新开关
// @Tags 特性开关
// @Accept json
// @Produce json
// @Param id path string true "开关 ID"
// @Param request body dto.UpdateFlagRequest true "更新"
// @Success 200 {object} response.Response
// @Router /system/features/{id} [put]
// @Security BearerAuth
func (h *Handler) Update(c *gin.Context) {
	uid, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	out, err := h.svc.Update(c.Request.Context(), uid, id, &req)
	write(c, out, err)
}

// Toggle 切换开关
// @Summary 切换开关
// @Tags 特性开关
// @Accept json
// @Produce json
// @Param id path string true "开关 ID"
// @Param request body dto.ToggleRequest false "目标状态"
// @Success 200 {object} response.Response
// @Router /system/features/{id}/toggle [post]
// @Security BearerAuth
func (h *Handler) Toggle(c *gin.Context) {
	uid, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ToggleRequest
	_ = c.ShouldBindJSON(&req)
	out, err := h.svc.Toggle(c.Request.Context(), uid, id, &req)
	write(c, out, err)
}

// Evaluate 评估指定用户
// @Summary 评估指定用户
// @Tags 特性开关
// @Produce json
// @Param id path string true "开关 ID"
// @Param user_id query string false "用户 ID，默认当前用户"
// @Success 200 {object} response.Response
// @Router /system/features/{id}/evaluate [get]
// @Security BearerAuth
func (h *Handler) Evaluate(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	uid, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var q dto.EvaluateQuery
	_ = c.ShouldBindQuery(&q)
	if q.UserID != "" {
		parsed, perr := uuid.Parse(q.UserID)
		if perr != nil {
			response.BadRequest(c, "无效的用户ID")
			return
		}
		uid = parsed
	}
	out, err := h.svc.EvaluateID(c.Request.Context(), id, uid)
	write(c, out, err)
}

// Audit 开关审计
// @Summary 开关审计
// @Tags 特性开关
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页条数"
// @Param flag_key query string false "开关键"
// @Success 200 {object} response.Response
// @Router /system/features/audit [get]
// @Security BearerAuth
func (h *Handler) Audit(c *gin.Context) {
	var q dto.AuditQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	q.Page, q.PageSize = defaultPage(q.Page, q.PageSize)
	list, total, err := h.svc.ListAudits(c.Request.Context(), q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

// EvaluateMe 当前调用方开关快照（公开；有 JWT 则按用户评估，否则匿名 subject）
// @Summary 当前用户开关快照
// @Tags 特性开关
// @Produce json
// @Param keys query string false "逗号分隔的 key"
// @Success 200 {object} response.Response
// @Router /features/me [get]
func (h *Handler) EvaluateMe(c *gin.Context) {
	var q dto.EvaluateMeQuery
	_ = c.ShouldBindQuery(&q)
	out, err := h.svc.EvaluateMe(c.Request.Context(), optionalUserID(c), splitKeys(q.Keys))
	write(c, out, err)
}
