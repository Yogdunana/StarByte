package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/dict/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListTypes 字典类型列表
// @Summary 字典类型列表
// @Description 列出全部数据字典类型
// @Tags 数据字典
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/dicts/types [get]
// @Security BearerAuth
func (h *DictHandler) ListTypes(c *gin.Context) {
	list, err := h.svc.ListTypes(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

// CreateType 创建字典类型
// @Summary 创建字典类型
// @Description 新建数据字典类型
// @Tags 数据字典
// @Accept json
// @Produce json
// @Param request body dto.CreateTypeRequest true "字典类型"
// @Success 200 {object} response.Response{data=dto.TypeResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/dicts/types [post]
// @Security BearerAuth
func (h *DictHandler) CreateType(c *gin.Context) {
	var req dto.CreateTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.CreateType(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// UpdateType 更新字典类型
// @Summary 更新字典类型
// @Description 更新数据字典类型名称、排序或状态
// @Tags 数据字典
// @Accept json
// @Produce json
// @Param id path string true "类型 ID"
// @Param request body dto.UpdateTypeRequest true "更新内容"
// @Success 200 {object} response.Response{data=dto.TypeResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/dicts/types/{id} [put]
// @Security BearerAuth
func (h *DictHandler) UpdateType(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.UpdateType(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// DeleteType 删除字典类型
// @Summary 删除字典类型
// @Description 删除非系统字典类型及其字典项
// @Tags 数据字典
// @Produce json
// @Param id path string true "类型 ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /system/dicts/types/{id} [delete]
// @Security BearerAuth
func (h *DictHandler) DeleteType(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteType(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithoutData(c)
}
