package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/search/dto"
	"github.com/Yogdunana/StarByte/backend/internal/search/service"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	svc service.SearchService
}

func NewSearchHandler(svc service.SearchService) *SearchHandler {
	return &SearchHandler{svc: svc}
}

func (h *SearchHandler) Resources(c *gin.Context) {
	response.OK(c, h.svc.Resources())
}

func (h *SearchHandler) Query(c *gin.Context) {
	var req dto.QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Query(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
