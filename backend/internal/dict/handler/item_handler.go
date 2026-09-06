package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/dict/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

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
