package middleware

import (
	"strings"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ResolveDataScope builds the caller's data-range condition for a resource.
// Used when the resource is chosen at request time (e.g. unified search).
func ResolveDataScope(
	c *gin.Context,
	db *gorm.DB,
	deptRepo rbacRepo.DepartmentRepo,
	cacheService rbacService.PermissionCacheService,
	resource string,
) (*rbacModel.DataScopeCondition, error) {
	denied := &rbacModel.DataScopeCondition{Query: "1 = 0"}
	if c == nil || strings.TrimSpace(resource) == "" {
		return denied, nil
	}
	if isSuper, exists := c.Get("is_super_admin"); exists {
		if b, ok := isSuper.(bool); ok && b {
			return &rbacModel.DataScopeCondition{}, nil
		}
	}
	userID, err := uuid.Parse(auth.GetUserID(c))
	if err != nil {
		return denied, nil
	}
	return buildDataScopeCondition(c.Request.Context(), db, deptRepo, cacheService, userID, resource, c)
}
