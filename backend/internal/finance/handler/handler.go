package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/finance/dto"
	"github.com/Yogdunana/StarByte/backend/internal/finance/service"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct{ svc service.Service }

func New(svc service.Service) *Handler { return &Handler{svc: svc} }

func currentUser(c *gin.Context) (uuid.UUID, error) {
	raw := auth.GetUserID(c)
	if raw == "" {
		return uuid.Nil, response.NewUnauthorizedError("用户未认证")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, response.NewError(response.CodeBadRequest, "无效的用户ID")
	}
	return id, nil
}

func parseID(c *gin.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, response.NewError(response.CodeBadRequest, "无效的ID")
	}
	return id, nil
}

func dataScope(c *gin.Context) *rbacModel.DataScopeCondition {
	return middleware.GetDataScopeFromContext(c)
}

// ListRecords 财务记录列表
// @Summary 财务记录列表
// @Tags 财务
// @Produce json
// @Success 200 {object} response.Response
// @Router /finance/records [get]
// @Security BearerAuth
func (h *Handler) ListRecords(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ListRecordRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	list, total, page, size, err := h.svc.List(c.Request.Context(), userID, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, page, size)
}

// CreateRecord 新增财务记录
// @Summary 新增财务记录
// @Tags 财务
// @Accept json
// @Produce json
// @Param request body dto.CreateRecordRequest true "记录"
// @Success 200 {object} response.Response
// @Router /finance/records [post]
// @Security BearerAuth
func (h *Handler) CreateRecord(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Create(c.Request.Context(), userID, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// GetRecord 财务记录详情
// @Summary 财务记录详情
// @Tags 财务
// @Produce json
// @Param id path string true "记录ID"
// @Success 200 {object} response.Response
// @Router /finance/records/{id} [get]
// @Security BearerAuth
func (h *Handler) GetRecord(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.Get(c.Request.Context(), userID, id, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// UpdateRecord 更新财务记录
// @Summary 更新财务记录
// @Tags 财务
// @Accept json
// @Produce json
// @Param id path string true "记录ID"
// @Success 200 {object} response.Response
// @Router /finance/records/{id} [put]
// @Security BearerAuth
func (h *Handler) UpdateRecord(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Update(c.Request.Context(), userID, id, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// DeleteRecord 删除财务记录
// @Summary 删除财务记录
// @Tags 财务
// @Produce json
// @Param id path string true "记录ID"
// @Success 200 {object} response.Response
// @Router /finance/records/{id} [delete]
// @Security BearerAuth
func (h *Handler) DeleteRecord(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), userID, id, dataScope(c)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// Categories 收支分类
// @Summary 收支分类列表
// @Tags 财务
// @Produce json
// @Success 200 {object} response.Response
// @Router /finance/categories [get]
// @Security BearerAuth
func (h *Handler) Categories(c *gin.Context) {
	out, err := h.svc.Categories(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// Summary 财务汇总
// @Summary 财务汇总
// @Tags 财务
// @Produce json
// @Success 200 {object} response.Response
// @Router /finance/summary [get]
// @Security BearerAuth
func (h *Handler) Summary(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var q dto.SummaryQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	out, err := h.svc.Summary(c.Request.Context(), userID, q, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// Export 导出预留
// @Summary 导出 Excel（二期）
// @Tags 财务
// @Produce json
// @Router /finance/export [get]
// @Security BearerAuth
func (h *Handler) Export(c *gin.Context) {
	response.NotImplemented(c, "财务导出将在二期接入导出引擎")
}
