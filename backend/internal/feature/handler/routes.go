package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/feature/dto"
	"github.com/Yogdunana/StarByte/backend/internal/feature/model"
	"github.com/Yogdunana/StarByte/backend/internal/feature/service"
	formdto "github.com/Yogdunana/StarByte/backend/internal/form/dto"
	formsvc "github.com/Yogdunana/StarByte/backend/internal/form/service"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func withPermission(group *gin.RouterGroup, permCode string, cache rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cache))
	return g
}

// RegisterRoutes 注册管理 API 与当前用户评估。静态 /audit 须在 /:id 之前。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, cache rbacService.PermissionCacheService) {
	r.GET("/features/me", h.EvaluateMe)

	g := r.Group("/system/features")
	read := withPermission(g, "feature:read", cache)
	read.GET("", h.List)
	read.GET("/audit", h.Audit)
	read.GET("/:id", h.Get)
	read.GET("/:id/evaluate", h.Evaluate)

	withPermission(g, "feature:create", cache).POST("", h.Create)
	withPermission(g, "feature:update", cache).PUT("/:id", h.Update)
	withPermission(g, "feature:manage", cache).POST("/:id/toggle", h.Toggle)
}

// RegisterGates 挂上招新前要灰度的新表面。不拦截 /auth 与 CAS。
func RegisterGates(r *gin.RouterGroup, feat service.Service, forms formsvc.FormService) {
	cms := r.Group("/cms")
	cms.Use(RequireFlag(feat, model.KeyCMSPublic, nil))
	cms.GET("/pages", func(c *gin.Context) {
		if forms == nil {
			response.OK(c, []dto.CMSPage{})
			return
		}
		pub := int16(1)
		list, _, _, _, err := forms.List(c.Request.Context(), formdto.ListQuery{Page: 1, PageSize: 50, Status: &pub}, false)
		if err != nil {
			response.Error(c, err)
			return
		}
		pages := make([]dto.CMSPage, 0, len(list))
		for _, item := range list {
			pages = append(pages, dto.CMSPage{ID: item.ID, Name: item.Name, Description: item.Description, UpdatedAt: item.UpdatedAt})
		}
		response.OK(c, pages)
	})
}
