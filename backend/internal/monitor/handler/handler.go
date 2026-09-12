package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/monitor/service"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct{ svc service.Service }

func New(svc service.Service) *Handler { return &Handler{svc: svc} }

func withPermission(group *gin.RouterGroup, permCode string, cache rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cache))
	return g
}

// RegisterRoutes 注册 /api/v1/monitor。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, cache rbacService.PermissionCacheService) {
	g := r.Group("/monitor")
	read := withPermission(g, "monitor:read", cache)
	read.GET("/server", h.Server)
	read.GET("/app", h.App)
	read.GET("/database", h.Database)
	read.GET("/redis", h.Redis)
	read.GET("/api-stats", h.APIStats)
}

// Server 服务器状态
// @Summary 服务器状态
// @Description CPU / 内存 / 磁盘 / 负载（不返回主机凭据）
// @Tags 监控
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /monitor/server [get]
// @Security BearerAuth
func (h *Handler) Server(c *gin.Context) {
	out, err := h.svc.Server(c.Request.Context())
	write(c, out, err)
}

// App 应用健康
// @Summary 应用健康
// @Description 运行时长、Goroutine、MemStats / GC
// @Tags 监控
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /monitor/app [get]
// @Security BearerAuth
func (h *Handler) App(c *gin.Context) {
	out, err := h.svc.App(c.Request.Context())
	write(c, out, err)
}

// Database 数据库连接池
// @Summary 数据库连接池
// @Description sql.DB Stats（活跃 / 空闲 / 等待）
// @Tags 监控
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /monitor/database [get]
// @Security BearerAuth
func (h *Handler) Database(c *gin.Context) {
	out, err := h.svc.Database(c.Request.Context())
	write(c, out, err)
}

// Redis Redis 状态
// @Summary Redis 状态
// @Description INFO clients / memory / stats / keyspace
// @Tags 监控
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /monitor/redis [get]
// @Security BearerAuth
func (h *Handler) Redis(c *gin.Context) {
	out, err := h.svc.Redis(c.Request.Context())
	write(c, out, err)
}

// APIStats API 调用计数
// @Summary API 调用计数
// @Description 复用 Prometheus 计数器；P50/P95 持久化见 percentiles_note
// @Tags 监控
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /monitor/api-stats [get]
// @Security BearerAuth
func (h *Handler) APIStats(c *gin.Context) {
	out, err := h.svc.APIStats(c.Request.Context())
	write(c, out, err)
}

func write(c *gin.Context, data any, err error) {
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, data)
}
