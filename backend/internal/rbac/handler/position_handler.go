package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/rbac/dto"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PositionHandler 职位处理器
type PositionHandler struct {
	positionService service.PositionService
}

// NewPositionHandler 创建职位处理器
func NewPositionHandler(positionService service.PositionService) *PositionHandler {
	return &PositionHandler{positionService: positionService}
}

// List 职位列表
// @Summary 获取职位列表
// @Description 分页查询职位列表
// @Tags 职位管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Response{data=response.PageResponse{list=[]dto.PositionResponse}}
// @Router /system/positions [get]
func (h *PositionHandler) List(c *gin.Context) {
	var req dto.ListPositionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	list, total, err := h.positionService.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Page(c, list, total, req.Page, req.PageSize)
}

// GetByID 获取职位详情
// @Summary 获取职位详情
// @Description 根据ID获取职位详细信息
// @Tags 职位管理
// @Produce json
// @Security Bearer
// @Param id path string true "职位ID"
// @Success 200 {object} response.Response{data=dto.PositionResponse}
// @Router /system/positions/{id} [get]
func (h *PositionHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的职位ID")
		return
	}

	result, err := h.positionService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// Create 创建职位
// @Summary 创建职位
// @Description 创建新职位
// @Tags 职位管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreatePositionRequest true "职位信息"
// @Success 200 {object} response.Response{data=dto.PositionResponse}
// @Router /system/positions [post]
func (h *PositionHandler) Create(c *gin.Context) {
	var req dto.CreatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	result, err := h.positionService.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// Update 更新职位
// @Summary 更新职位
// @Description 更新职位信息
// @Tags 职位管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "职位ID"
// @Param request body dto.UpdatePositionRequest true "职位信息"
// @Success 200 {object} response.Response{data=dto.PositionResponse}
// @Router /system/positions/{id} [put]
func (h *PositionHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的职位ID")
		return
	}

	var req dto.UpdatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	result, err := h.positionService.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// Delete 删除职位
// @Summary 删除职位
// @Description 删除指定职位
// @Tags 职位管理
// @Produce json
// @Security Bearer
// @Param id path string true "职位ID"
// @Success 200 {object} response.Response
// @Router /system/positions/{id} [delete]
func (h *PositionHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的职位ID")
		return
	}

	if err := h.positionService.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}
