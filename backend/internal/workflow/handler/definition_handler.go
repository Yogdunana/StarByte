package handler

import (
	"strconv"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/dto"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/service"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// DefinitionHandler 流程定义处理器
type DefinitionHandler struct {
	defService service.DefinitionService
}

// NewDefinitionHandler 创建流程定义处理器
func NewDefinitionHandler(defService service.DefinitionService) *DefinitionHandler {
	return &DefinitionHandler{defService: defService}
}

// List 流程定义列表
// @Summary 获取流程定义列表
// @Description 分页查询流程定义列表
// @Tags 流程定义
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param keyword query string false "关键词"
// @Param category query string false "分类"
// @Param status query int false "状态"
// @Success 200 {object} response.Response{data=response.PageResponse{list=[]dto.DefinitionResponse}}
// @Router /workflow/definitions [get]
func (h *DefinitionHandler) List(c *gin.Context) {
	page, pageSize := parsePagination(c)

	keyword := c.Query("keyword")
	category := c.Query("category")

	var status *int
	if statusStr := c.Query("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			status = &s
		}
	}

	list, total, err := h.defService.List(c.Request.Context(), page, pageSize, keyword, category, status)
	if err != nil {
		response.Error(c, err)
		return
	}

	result := make([]dto.DefinitionResponse, 0, len(list))
	for _, def := range list {
		result = append(result, dto.ToDefinitionResponse(&def))
	}

	response.Page(c, result, total, page, pageSize)
}

// GetByID 获取流程定义详情
// @Summary 获取流程定义详情
// @Description 根据ID获取流程定义详细信息
// @Tags 流程定义
// @Produce json
// @Security Bearer
// @Param id path string true "流程定义ID"
// @Success 200 {object} response.Response{data=dto.DefinitionResponse}
// @Router /workflow/definitions/{id} [get]
func (h *DefinitionHandler) GetByID(c *gin.Context) {
	id, err := parseUUIDParam(c, "id", "无效的流程定义ID")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	def, err := h.defService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, dto.ToDefinitionResponse(def))
}

// Create 创建流程定义
// @Summary 创建流程定义
// @Description 创建新的流程定义（草稿状态）
// @Tags 流程定义
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreateDefinitionRequest true "流程定义信息"
// @Success 200 {object} response.Response{data=dto.DefinitionResponse}
// @Router /workflow/definitions [post]
func (h *DefinitionHandler) Create(c *gin.Context) {
	var req dto.CreateDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, parseBindingError(err))
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	def, err := h.defService.Create(c.Request.Context(), &req, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, dto.ToDefinitionResponse(def))
}

// Update 更新流程定义
// @Summary 更新流程定义
// @Description 更新流程定义基本信息
// @Tags 流程定义
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "流程定义ID"
// @Param request body dto.UpdateDefinitionRequest true "流程定义信息"
// @Success 200 {object} response.Response{data=dto.DefinitionResponse}
// @Router /workflow/definitions/{id} [put]
func (h *DefinitionHandler) Update(c *gin.Context) {
	id, err := parseUUIDParam(c, "id", "无效的流程定义ID")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var req dto.UpdateDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, parseBindingError(err))
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	def, err := h.defService.Update(c.Request.Context(), id, &req, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, dto.ToDefinitionResponse(def))
}

// Delete 删除流程定义
// @Summary 删除流程定义
// @Description 删除指定流程定义
// @Tags 流程定义
// @Produce json
// @Security Bearer
// @Param id path string true "流程定义ID"
// @Success 200 {object} response.Response
// @Router /workflow/definitions/{id} [delete]
func (h *DefinitionHandler) Delete(c *gin.Context) {
	id, err := parseUUIDParam(c, "id", "无效的流程定义ID")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.defService.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// Publish 发布流程定义
// @Summary 发布流程定义
// @Description 发布流程定义，创建新版本
// @Tags 流程定义
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "流程定义ID"
// @Param request body dto.PublishDefinitionRequest true "流程图数据"
// @Success 200 {object} response.Response{data=dto.VersionResponse}
// @Router /workflow/definitions/{id}/publish [post]
func (h *DefinitionHandler) Publish(c *gin.Context) {
	id, err := parseUUIDParam(c, "id", "无效的流程定义ID")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var req dto.PublishDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, parseBindingError(err))
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	version, err := h.defService.Publish(c.Request.Context(), id, &req, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, dto.ToVersionResponse(version))
}

// ListVersions 获取版本列表
// @Summary 获取流程定义版本列表
// @Description 获取指定流程定义的所有版本
// @Tags 流程定义
// @Produce json
// @Security Bearer
// @Param id path string true "流程定义ID"
// @Success 200 {object} response.Response{data=[]dto.VersionResponse}
// @Router /workflow/definitions/{id}/versions [get]
func (h *DefinitionHandler) ListVersions(c *gin.Context) {
	id, err := parseUUIDParam(c, "id", "无效的流程定义ID")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	versions, err := h.defService.ListVersions(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	result := make([]dto.VersionResponse, 0, len(versions))
	for _, v := range versions {
		result = append(result, dto.ToVersionResponse(&v))
	}

	response.OK(c, result)
}

// GetVersionByID 获取指定版本详情
// @Summary 获取指定版本详情
// @Description 根据版本ID获取流程定义版本详情，并校验版本归属
// @Tags 流程定义
// @Produce json
// @Security Bearer
// @Param id path string true "流程定义ID"
// @Param versionId path string true "版本ID"
// @Success 200 {object} response.Response{data=dto.VersionResponse}
// @Router /workflow/definitions/{id}/versions/{versionId} [get]
func (h *DefinitionHandler) GetVersionByID(c *gin.Context) {
	defID, err := parseUUIDParam(c, "id", "无效的流程定义ID")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	versionID, err := parseUUIDParam(c, "versionId", "无效的版本ID")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	version, err := h.defService.GetVersionByID(c.Request.Context(), versionID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// 校验版本归属关系
	if version.DefinitionID != defID {
		response.BadRequest(c, "版本不属于该流程定义")
		return
	}

	response.OK(c, dto.ToVersionResponse(version))
}
