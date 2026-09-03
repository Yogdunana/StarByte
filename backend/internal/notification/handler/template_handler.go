package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	"github.com/Yogdunana/StarByte/backend/internal/notification/service"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TemplateHandler 通知模板管理处理器
type TemplateHandler struct {
	templateService service.TemplateService
}

// NewTemplateHandler 创建模板处理器
func NewTemplateHandler(templateService service.TemplateService) *TemplateHandler {
	return &TemplateHandler{templateService: templateService}
}

// Create POST /api/v1/system/notification-templates
// @Summary 创建通知模板
// @Description 创建通知模板（管理员）
// @Tags 系统管理-通知模板
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreateTemplateRequest true "模板信息"
// @Success 200 {object} response.Response
// @Router /system/notification-templates [post]
func (h *TemplateHandler) Create(c *gin.Context) {
	var req dto.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.templateService.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// Get GET /api/v1/system/notification-templates/:id
// @Summary 模板详情
// @Description 获取通知模板详情
// @Tags 系统管理-通知模板
// @Produce json
// @Security Bearer
// @Param id path string true "模板 ID"
// @Success 200 {object} response.Response
// @Router /system/notification-templates/{id} [get]
func (h *TemplateHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	result, err := h.templateService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// List GET /api/v1/system/notification-templates
// @Summary 模板列表
// @Description 获取通知模板列表（分页）
// @Tags 系统管理-通知模板
// @Produce json
// @Security Bearer
// @Param page query int false "页码"
// @Param page_size query int false "每页条数"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Response
// @Router /system/notification-templates [get]
func (h *TemplateHandler) List(c *gin.Context) {
	var req dto.ListTemplatesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	list, total, err := h.templateService.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Page(c, list, total, req.Page, req.PageSize)
}

// Update PUT /api/v1/system/notification-templates/:id
// @Summary 更新模板
// @Description 更新通知模板
// @Tags 系统管理-通知模板
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "模板 ID"
// @Param request body dto.UpdateTemplateRequest true "更新内容"
// @Success 200 {object} response.Response
// @Router /system/notification-templates/{id} [put]
func (h *TemplateHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	var req dto.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.templateService.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// Delete DELETE /api/v1/system/notification-templates/:id
// @Summary 删除模板
// @Description 删除通知模板
// @Tags 系统管理-通知模板
// @Produce json
// @Security Bearer
// @Param id path string true "模板 ID"
// @Success 200 {object} response.Response
// @Router /system/notification-templates/{id} [delete]
func (h *TemplateHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	if err := h.templateService.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// Test POST /api/v1/system/notification-templates/:id/test
// @Summary 测试模板
// @Description 使用提供的变量测试渲染通知模板
// @Tags 系统管理-通知模板
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "模板 ID"
// @Param request body dto.TestTemplateRequest true "测试变量"
// @Success 200 {object} response.Response
// @Router /system/notification-templates/{id}/test [post]
func (h *TemplateHandler) Test(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	var req dto.TestTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.templateService.Test(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}
