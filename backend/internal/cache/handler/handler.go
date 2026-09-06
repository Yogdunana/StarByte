package handler

import (
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/cache/dto"
	"github.com/Yogdunana/StarByte/backend/internal/cache/service"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type CacheHandler struct {
	svc service.CacheService
}

func NewCacheHandler(svc service.CacheService) *CacheHandler {
	return &CacheHandler{svc: svc}
}

func (h *CacheHandler) Stats(c *gin.Context) {
	out, err := h.svc.Stats(c.Request.Context(), c.Query("pattern"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *CacheHandler) DeleteKey(c *gin.Context) {
	key := strings.TrimSpace(c.Param("key"))
	if key == "" {
		response.BadRequest(c, "缓存键不能为空")
		return
	}
	if err := h.svc.DeleteKey(c.Request.Context(), key); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, &dto.DeleteResult{Deleted: 1, Keys: []string{key}})
}

func (h *CacheHandler) DeletePattern(c *gin.Context) {
	pattern := strings.TrimSpace(c.Param("pattern"))
	if pattern == "" {
		response.BadRequest(c, "清除模式不能为空")
		return
	}
	out, err := h.svc.DeletePattern(c.Request.Context(), pattern)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *CacheHandler) Warmup(c *gin.Context) {
	var req dto.WarmupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Warmup(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
