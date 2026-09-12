package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/feature/model"
	"github.com/Yogdunana/StarByte/backend/internal/feature/service"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	authmiddleware "github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func withPermission(group *gin.RouterGroup, permCode string, cache rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cache))
	return g
}

// RegisterPublic 匿名也可评估（boolean 开则放行；名单/百分比 fail closed）。不拦 /auth 与 CAS。
func RegisterPublic(r *gin.RouterGroup, h *Handler, jwtCfg *config.JWTConfig, rdb *redis.Client) {
	g := r.Group("")
	g.Use(authmiddleware.OptionalJWT(jwtCfg, rdb))
	g.GET("/features/me", h.EvaluateMe)
}

// RegisterRoutes 注册管理 API。静态 /audit 须在 /:id 之前。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, cache rbacService.PermissionCacheService) {
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

// GatePublicKnowledge 把 cms.public 挂到知识库公开读。OptionalJWT 先于评估，方便名单/百分比。
// 不拦工作台 /knowledge 管理端。
func GatePublicKnowledge(r *gin.RouterGroup, feat service.Service, jwtCfg *config.JWTConfig, rdb *redis.Client) *gin.RouterGroup {
	g := r.Group("")
	g.Use(authmiddleware.OptionalJWT(jwtCfg, rdb))
	g.Use(RequireFlag(feat, model.KeyCMSPublic, nil))
	return g
}
