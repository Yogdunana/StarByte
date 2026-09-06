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

// Stats 缓存统计
// @Summary 缓存统计
// @Description 查看 Redis 连接池与按 pattern 扫描的键
// @Tags 缓存
// @Produce json
// @Param pattern query string false "键匹配模式"
// @Success 200 {object} response.Response{data=dto.Stats}
// @Failure 401 {object} response.Response
// @Router /system/cache/stats [get]
// @Security BearerAuth
func (h *CacheHandler) Stats(c *gin.Context) {
	out, err := h.svc.Stats(c.Request.Context(), c.Query("pattern"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// DeleteKey 删除单个缓存键
// @Summary 删除单个缓存键
// @Description 按完整键名删除 Redis 缓存
// @Tags 缓存
// @Produce json
// @Param key path string true "缓存键"
// @Success 200 {object} response.Response{data=dto.DeleteResult}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/cache/{key} [delete]
// @Security BearerAuth
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

// DeletePattern 按模式清除缓存
// @Summary 按模式清除缓存
// @Description 按 glob 模式批量删除缓存键
// @Tags 缓存
// @Produce json
// @Param pattern path string true "匹配模式"
// @Success 200 {object} response.Response{data=dto.DeleteResult}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/cache/pattern/{pattern} [delete]
// @Security BearerAuth
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

// Warmup 缓存预热
// @Summary 缓存预热
// @Description 将指定键写入 L1/L2 缓存
// @Tags 缓存
// @Accept json
// @Produce json
// @Param request body dto.WarmupRequest true "预热条目"
// @Success 200 {object} response.Response{data=dto.WarmupResult}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/cache/warmup [post]
// @Security BearerAuth
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
