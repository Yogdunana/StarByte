package handler

import (
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct{ svc service.Service }

func New(svc service.Service) *Handler { return &Handler{svc: svc} }

func withScope(
	group *gin.RouterGroup,
	code string,
	cache rbacService.PermissionCacheService,
	db *gorm.DB,
	deptRepo rbacRepo.DepartmentRepo,
) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(code))
	g.Use(middleware.RequireDataScope("schedule"))
	g.Use(middleware.PermissionRequired(cache))
	g.Use(middleware.DataScopeMiddleware(db, deptRepo, cache))
	return g
}

// RegisterRoutes 注册 /api/v1/schedules。静态路径须在 /events/:id 之前。
func RegisterRoutes(
	r *gin.RouterGroup,
	h *Handler,
	cache rbacService.PermissionCacheService,
	db *gorm.DB,
	deptRepo rbacRepo.DepartmentRepo,
) {
	g := r.Group("/schedules")

	read := withScope(g, "schedule:read", cache, db, deptRepo)
	read.GET("/calendars", h.ListCalendars)
	read.GET("/calendars/:id", h.GetCalendar)
	read.GET("/calendars/:id/members", h.ListMembers)
	read.GET("/events/range", h.RangeEvents)
	read.GET("/events", h.ListEvents)
	read.GET("/events/:id", h.GetEvent)

	create := withScope(g, "schedule:create", cache, db, deptRepo)
	create.POST("/calendars", h.CreateCalendar)
	create.POST("/events", h.CreateEvent)

	update := withScope(g, "schedule:update", cache, db, deptRepo)
	update.PUT("/calendars/:id", h.UpdateCalendar)
	update.POST("/calendars/:id/members", h.AddMember)
	update.DELETE("/calendars/:id/members/:uid", h.RemoveMember)
	update.PUT("/events/:id", h.UpdateEvent)
	update.POST("/events/:id/remind", h.SetReminders)
	update.POST("/events/:id/rsvp", h.RSVP)

	del := withScope(g, "schedule:delete", cache, db, deptRepo)
	del.DELETE("/calendars/:id", h.DeleteCalendar)
	del.DELETE("/events/:id", h.DeleteEvent)
}
