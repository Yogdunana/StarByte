package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/discipline/dto"
	"github.com/Yogdunana/StarByte/backend/internal/discipline/service"
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

// ListRecords 处分列表
// @Summary 处分记录列表
// @Tags 处分
// @Produce json
// @Router /discipline/records [get]
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

// CreateRecord 登记处分
// @Summary 登记处分
// @Tags 处分
// @Accept json
// @Produce json
// @Param request body dto.CreateRecordRequest true "处分"
// @Router /discipline/records [post]
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

// GetRecord 处分详情
// @Summary 处分详情
// @Tags 处分
// @Produce json
// @Param id path string true "记录ID"
// @Router /discipline/records/{id} [get]
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

// UpdateRecord 更新处分
// @Summary 更新处分
// @Tags 处分
// @Accept json
// @Produce json
// @Param id path string true "记录ID"
// @Router /discipline/records/{id} [put]
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

// Approve 审批处分
// @Summary 审批处分
// @Tags 处分
// @Accept json
// @Produce json
// @Param id path string true "记录ID"
// @Router /discipline/records/{id}/approve [post]
// @Security BearerAuth
func (h *Handler) Approve(c *gin.Context) {
	h.act(c, func(uid, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.RecordResponse, error) {
		var req dto.ApproveRequest
		_ = c.ShouldBindJSON(&req)
		return h.svc.Approve(c.Request.Context(), uid, id, req.Comment, scope)
	})
}

// Revoke 撤销处分
// @Summary 撤销处分
// @Tags 处分
// @Accept json
// @Produce json
// @Param id path string true "记录ID"
// @Router /discipline/records/{id}/revoke [post]
// @Security BearerAuth
func (h *Handler) Revoke(c *gin.Context) {
	h.act(c, func(uid, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.RecordResponse, error) {
		var req dto.RevokeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return nil, response.NewError(response.CodeBadRequest, "请填写撤销原因")
		}
		return h.svc.Revoke(c.Request.Context(), uid, id, req.Reason, scope)
	})
}

// Appeal 申诉
// @Summary 提交申诉
// @Tags 处分
// @Accept json
// @Produce json
// @Param id path string true "记录ID"
// @Router /discipline/records/{id}/appeal [post]
// @Security BearerAuth
func (h *Handler) Appeal(c *gin.Context) {
	h.act(c, func(uid, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.RecordResponse, error) {
		var req dto.AppealRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return nil, response.NewError(response.CodeBadRequest, "请填写申诉理由")
		}
		return h.svc.Appeal(c.Request.Context(), uid, id, req.Reason, scope)
	})
}

func (h *Handler) act(c *gin.Context, fn func(uuid.UUID, uuid.UUID, *rbacModel.DataScopeCondition) (*dto.RecordResponse, error)) {
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
	out, err := fn(userID, id, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
