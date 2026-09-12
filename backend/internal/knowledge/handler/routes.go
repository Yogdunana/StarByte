package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/knowledge/service"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	authmiddleware "github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Handler struct{ svc service.Service }

func New(svc service.Service) *Handler { return &Handler{svc: svc} }

func withPermission(group *gin.RouterGroup, permCode string, cache rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cache))
	return g
}

// RegisterPublicRoutes 公开阅读：visibility=public 无需 JWT；其余按登录/权限码判定。
func RegisterPublicRoutes(r *gin.RouterGroup, h *Handler, jwtCfg *config.JWTConfig, rdb *redis.Client, cache rbacService.PermissionCacheService) {
	g := r.Group("/knowledge/public")
	g.Use(authmiddleware.OptionalJWT(jwtCfg, rdb))
	g.Use(middleware.AttachPermissions(cache))
	g.GET("/pages/:slug", h.PublicPage)
	g.GET("/docs", h.PublicDocs)
	g.GET("/docs/:slug", h.PublicDoc)
	g.GET("/tree", h.PublicTree)
	g.GET("/search", h.PublicSearch)
}

// RegisterRoutes 注册 /api/v1/knowledge。静态路径须在 /:id 之前。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, cache rbacService.PermissionCacheService) {
	g := r.Group("/knowledge")

	read := withPermission(g, "doc:read", cache)
	read.GET("/docs", h.List)
	read.GET("/docs/:id", h.Get)
	read.GET("/docs/:id/history", h.History)
	read.GET("/docs/:id/versions/:version", h.GetVersion)
	read.GET("/search", h.Search)
	read.GET("/tree", h.Tree)
	read.GET("/categories", h.ListCategories)

	create := withPermission(g, "doc:create", cache)
	create.POST("/docs", h.Create)
	create.POST("/categories", h.CreateCategory)

	update := withPermission(g, "doc:update", cache)
	update.PUT("/docs/:id", h.Update)
	update.POST("/docs/:id/rollback", h.Rollback)
	update.POST("/docs/:id/attachments", h.Attach)
	update.DELETE("/docs/:id/attachments/:fileId", h.Detach)
	update.PUT("/categories/:id", h.UpdateCategory)

	del := withPermission(g, "doc:delete", cache)
	del.DELETE("/docs/:id", h.Delete)
	del.DELETE("/categories/:id", h.DeleteCategory)

	publish := withPermission(g, "doc:publish", cache)
	publish.POST("/docs/:id/publish", h.Publish)
}
