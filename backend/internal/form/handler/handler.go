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

// List 表单列表
// @Summary 表单列表
// @Description 列出表单定义；无 form:read 时仅返回已发布表单
// @Tags 表单
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页条数"
// @Param keyword query string false "关键词"
// @Param status query int false "状态 0草稿 1已发布 2已停用"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /forms [get]
// @Security BearerAuth
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

// Get 表单定义
// @Summary 表单定义
// @Description 按 ID 获取表单字段定义
// @Tags 表单
// @Produce json
// @Param id path string true "表单 ID"
// @Success 200 {object} response.Response{data=dto.FormDetail}
// @Failure 404 {object} response.Response
// @Router /forms/{id} [get]
// @Security BearerAuth
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

// Create 创建表单
// @Summary 创建表单
// @Description 创建动态表单定义，需要 form:write
// @Tags 表单
// @Accept json
// @Produce json
// @Param body body dto.CreateFormRequest true "表单"
// @Success 200 {object} response.Response{data=dto.FormDetail}
// @Failure 400 {object} response.Response
// @Router /forms [post]
// @Security BearerAuth
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

// Update 更新表单
// @Summary 更新表单
// @Description 更新表单定义或状态，需要 form:write
// @Tags 表单
// @Accept json
// @Produce json
// @Param id path string true "表单 ID"
// @Param body body dto.UpdateFormRequest true "更新内容"
// @Success 200 {object} response.Response{data=dto.FormDetail}
// @Failure 400 {object} response.Response
// @Router /forms/{id} [put]
// @Security BearerAuth
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

// Submit 提交表单
// @Summary 提交表单
// @Description 提交已发布表单的填写数据，需要 form:submit
// @Tags 表单
// @Accept json
// @Produce json
// @Param id path string true "表单 ID"
// @Param body body object true "字段值"
// @Success 200 {object} response.Response{data=dto.SubmitResponse}
// @Failure 400 {object} response.Response
// @Router /forms/{id}/submit [post]
// @Security BearerAuth
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

// ListSubmissions 提交记录
// @Summary 提交记录
// @Description 列出某表单的提交记录，需要 form:read
// @Tags 表单
// @Produce json
// @Param id path string true "表单 ID"
// @Param page query int false "页码"
// @Param page_size query int false "每页条数"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /forms/{id}/submissions [get]
// @Security BearerAuth
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
