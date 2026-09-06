package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/dict/service"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type DictHandler struct {
	svc   service.DictService
	cache rbacService.PermissionCacheService
}

func NewDictHandler(svc service.DictService, cache rbacService.PermissionCacheService) *DictHandler {
	return &DictHandler{svc: svc, cache: cache}
}

func withPermission(group *gin.RouterGroup, permCode string, cacheService rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cacheService))
	return g
}

// RegisterRoutes 注册 /api/v1/system/dicts。静态 /types 必须在 /:type 之前。
func RegisterRoutes(r *gin.RouterGroup, h *DictHandler, cacheService rbacService.PermissionCacheService) {
	g := r.Group("/system/dicts")

	read := withPermission(g, "dict:read", cacheService)
	read.GET("/types", h.ListTypes)

	create := withPermission(g, "dict:create", cacheService)
	create.POST("/types", h.CreateType)
	create.POST("", h.CreateItem)

	update := withPermission(g, "dict:update", cacheService)
	update.PUT("/types/:id", h.UpdateType)
	update.PUT("/:id", h.UpdateItem)

	del := withPermission(g, "dict:delete", cacheService)
	del.DELETE("/types/:id", h.DeleteType)
	del.DELETE("/:id", h.DeleteItem)

	// 登录即可读启用项；?all=1 在 handler 内校验 dict:read
	g.GET("/:type", h.ListItems)
}
