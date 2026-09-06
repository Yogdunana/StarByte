package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/dict/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListItems 字典项列表
// @Summary 字典项列表
// @Description 按类型编码列出字典项；默认仅启用项，all=1 需 dict:read
// @Tags 数据字典
// @Produce json
// @Param type path string true "类型编码"
// @Param all query string false "传 1 返回全部含停用项"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /system/dicts/{type} [get]
// @Security BearerAuth
func (h *DictHandler) ListItems(c *gin.Context) {
	enabledOnly := c.Query("all") != "1"
	if !enabledOnly {
		ok, err := h.canReadAll(c)
		if err != nil {
			response.Error(c, err)
			return
		}
		if !ok {
			response.Error(c, response.NewForbiddenError("权限不足: dict:read"))
			return
		}
	}
	list, err := h.svc.ListItems(c.Request.Context(), c.Param("type"), enabledOnly)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

// CreateItem 创建字典项
// @Summary 创建字典项
// @Description 在指定字典类型下新增选项
// @Tags 数据字典
// @Accept json
// @Produce json
// @Param request body dto.CreateItemRequest true "字典项"
// @Success 200 {object} response.Response{data=dto.ItemResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/dicts [post]
// @Security BearerAuth
func (h *DictHandler) CreateItem(c *gin.Context) {
	var req dto.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.CreateItem(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// UpdateItem 更新字典项
// @Summary 更新字典项
// @Description 更新字典项标签、排序或状态
// @Tags 数据字典
// @Accept json
// @Produce json
// @Param id path string true "字典项 ID"
// @Param request body dto.UpdateItemRequest true "更新内容"
// @Success 200 {object} response.Response{data=dto.ItemResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/dicts/{id} [put]
// @Security BearerAuth
func (h *DictHandler) UpdateItem(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.UpdateItem(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// DeleteItem 删除字典项
// @Summary 删除字典项
// @Description 删除一条数据字典项
// @Tags 数据字典
// @Produce json
// @Param id path string true "字典项 ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /system/dicts/{id} [delete]
// @Security BearerAuth
func (h *DictHandler) DeleteItem(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteItem(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithoutData(c)
}
