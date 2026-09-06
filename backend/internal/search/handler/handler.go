package handler

import (
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/internal/search/dto"
	"github.com/Yogdunana/StarByte/backend/internal/search/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SearchHandler struct {
	svc   service.SearchService
	db    *gorm.DB
	dept  rbacRepo.DepartmentRepo
	cache rbacService.PermissionCacheService
}

func NewSearchHandler(
	svc service.SearchService,
	db *gorm.DB,
	dept rbacRepo.DepartmentRepo,
	cache rbacService.PermissionCacheService,
) *SearchHandler {
	return &SearchHandler{svc: svc, db: db, dept: dept, cache: cache}
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
	sch, ok := h.svc.Lookup(req.Resource)
	if !ok {
		response.Error(c, response.NewError(response.CodeSearchUnknownResource, "未知检索资源"))
		return
	}
	scope, err := middleware.ResolveDataScope(c, h.db, h.dept, h.cache, sch.RBACResource)
	if err != nil {
		response.Error(c, err)
		return
	}
	viewer, _ := uuid.Parse(auth.GetUserID(c))
	out, err := h.svc.Query(c.Request.Context(), req, scope, viewer)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
