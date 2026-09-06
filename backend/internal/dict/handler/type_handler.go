package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/dict/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func (h *DictHandler) ListTypes(c *gin.Context) {
	list, err := h.svc.ListTypes(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

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
