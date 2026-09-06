package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/form/dto"
	"github.com/Yogdunana/StarByte/backend/internal/form/service"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FormHandler struct {
	svc   service.FormService
	cache rbacService.PermissionCacheService
}

func NewFormHandler(svc service.FormService, cache rbacService.PermissionCacheService) *FormHandler {
	return &FormHandler{svc: svc, cache: cache}
}

func parseID(c *gin.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, response.NewError(response.CodeBadRequest, "无效的ID")
	}
	return id, nil
}

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

func (h *FormHandler) hasPerm(c *gin.Context, code string) (bool, error) {
	userID, err := currentUser(c)
	if err != nil {
		return false, err
	}
	perms, isSuper, err := h.cache.GetUserPermissionsAndSuperAdmin(c.Request.Context(), userID)
	if err != nil {
		return false, err
	}
	if isSuper {
		return true, nil
	}
	for _, p := range perms {
		if p == code {
			return true, nil
		}
	}
	return false, nil
}

func (h *FormHandler) List(c *gin.Context) {
	var q dto.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	canRead, err := h.hasPerm(c, "form:read")
	if err != nil {
		response.Error(c, err)
		return
	}
	list, total, page, size, err := h.svc.List(c.Request.Context(), q, canRead)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, page, size)
}

func (h *FormHandler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	canRead, err := h.hasPerm(c, "form:read")
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.Get(c.Request.Context(), id, canRead)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *FormHandler) Create(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *FormHandler) Update(c *gin.Context) {
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
	var req dto.UpdateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Update(c.Request.Context(), userID, id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *FormHandler) Submit(c *gin.Context) {
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
	var values map[string]interface{}
	if err := c.ShouldBindJSON(&values); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Submit(c.Request.Context(), userID, id, values)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *FormHandler) ListSubmissions(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var q dto.SubmissionQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	list, total, page, size, err := h.svc.ListSubmissions(c.Request.Context(), id, q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, page, size)
}
