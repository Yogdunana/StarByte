package handler

import (
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/internal/report/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	svc service.Service
}

func New(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

func RegisterRoutes(
	r *gin.RouterGroup,
	h *Handler,
	cache rbacService.PermissionCacheService,
	db *gorm.DB,
	departments rbacRepo.DepartmentRepo,
) {
	g := r.Group("/reports")
	g.Use(middleware.RequirePermission("report:read"))
	g.Use(middleware.RequireDataScope("report:read"))
	g.Use(middleware.PermissionRequired(cache))
	g.Use(middleware.DataScopeMiddleware(db, departments, cache))
	g.GET("", h.List)
}
