package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/contract/dto"
	"github.com/Yogdunana/StarByte/backend/internal/contract/service"
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

// List 合同列表
// @Summary 合同列表
// @Tags 合同
// @Produce json
// @Router /contracts [get]
// @Security BearerAuth
func (h *Handler) List(c *gin.Context) {
	var req dto.ListContractRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	list, total, page, size, err := h.svc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, page, size)
}

// Create 新增合同
// @Summary 新增合同
// @Tags 合同
// @Accept json
// @Produce json
// @Param request body dto.CreateContractRequest true "合同"
// @Router /contracts [post]
// @Security BearerAuth
func (h *Handler) Create(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Create(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// Get 合同详情
// @Summary 合同详情
// @Tags 合同
// @Produce json
// @Param id path string true "合同ID"
// @Router /contracts/{id} [get]
// @Security BearerAuth
func (h *Handler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// Update 更新合同
// @Summary 更新合同
// @Tags 合同
// @Accept json
// @Produce json
// @Param id path string true "合同ID"
// @Router /contracts/{id} [put]
// @Security BearerAuth
func (h *Handler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// Delete 删除合同
// @Summary 删除合同
// @Tags 合同
// @Produce json
// @Param id path string true "合同ID"
// @Router /contracts/{id} [delete]
// @Security BearerAuth
func (h *Handler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// Templates 合同模板
// @Summary 合同模板列表
// @Tags 合同
// @Produce json
// @Router /contracts/templates [get]
// @Security BearerAuth
func (h *Handler) Templates(c *gin.Context) {
	out, err := h.svc.Templates(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// Expiring 即将到期合同
// @Summary 即将到期合同
// @Tags 合同
// @Produce json
// @Router /contracts/expiring [get]
// @Security BearerAuth
func (h *Handler) Expiring(c *gin.Context) {
	var q dto.ExpiringQuery
	_ = c.ShouldBindQuery(&q)
	out, err := h.svc.Expiring(c.Request.Context(), q.Days)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
