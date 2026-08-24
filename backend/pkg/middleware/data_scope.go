package middleware

import (
	"context"
	"fmt"

	rbac "github.com/Yogdunana/StarByte/backend/internal/rbac"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DataScopeCondition represents a SQL fragment and its arguments used to apply
// data-scope filtering to a repository query. When Query is empty the caller
// should apply no filtering, i.e. the user is permitted to access all records
// (super administrator or "all" data scope).
//
// For conditions that deny access to any record (for example a "custom" scope
// without any granted department) Query is set to "1 = 0".
type DataScopeCondition struct {
	Query string
	Args  []interface{}
}

// GetDataScopeCondition builds the data-scope filter for the given user and
// resource. Super administrators (flagged via the "is_super_admin" context
// value set by PermissionRequired) and users whose effective scope is "all"
// receive an empty condition, meaning no filtering is applied.
//
// The data scope is resolved directly from role_permissions (joined with
// user_roles and permissions) via the provided *gorm.DB. For the
// "department_and_sub" scope the department subtree is resolved through
// deptRepo.GetDepartmentAndSubIDs.
//
// Note: a *gorm.DB is accepted in addition to deptRepo because resolving the
// user's effective data scope requires cross-table joins that the
// DepartmentRepo interface does not expose. This keeps the data access explicit
// and testable rather than relying on a global database handle.
func GetDataScopeCondition(c *gin.Context, db *gorm.DB, deptRepo rbacRepo.DepartmentRepo, userID uuid.UUID, resource string) (*DataScopeCondition, error) {
	ctx := c.Request.Context()

	if db == nil {
		return nil, fmt.Errorf("data scope: database handle is nil")
	}

	// 超级管理员不进行数据过滤
	if isSuper, exists := c.Get("is_super_admin"); exists {
		if b, ok := isSuper.(bool); ok && b {
			return &DataScopeCondition{}, nil
		}
	}

	// 解析用户在指定资源上的数据权限范围
	scopes, err := fetchUserDataScopes(ctx, db, userID, resource)
	if err != nil {
		logger.Error("fetch user data scopes failed",
			zap.Stringer("user_id", userID),
			zap.String("resource", resource),
			zap.Error(err),
		)
		return nil, fmt.Errorf("fetch user data scopes: %w", err)
	}

	// 用户在该资源上没有任何数据权限，退化为仅本人数据（最安全的非空默认）
	if len(scopes) == 0 {
		return &DataScopeCondition{Query: "created_by = ?", Args: []interface{}{userID}}, nil
	}

	scopeType := mostPermissiveScope(scopes)

	switch scopeType {
	case rbacModel.DataScopeAll:
		// 全部数据：不附加过滤条件
		return &DataScopeCondition{}, nil

	case rbacModel.DataScopeSelf:
		return &DataScopeCondition{Query: "created_by = ?", Args: []interface{}{userID}}, nil

	case rbacModel.DataScopeDepartment:
		deptID, err := fetchUserDepartmentID(ctx, db, userID)
		if err != nil {
			logger.Error("fetch user department failed", zap.Stringer("user_id", userID), zap.Error(err))
			return nil, fmt.Errorf("fetch user department: %w", err)
		}
		if deptID == nil {
			// 缺少部门信息时退化为仅本人数据
			return &DataScopeCondition{Query: "created_by = ?", Args: []interface{}{userID}}, nil
		}
		return &DataScopeCondition{Query: "department_id = ?", Args: []interface{}{*deptID}}, nil

	case rbacModel.DataScopeDepartmentAndSub:
		deptID, err := fetchUserDepartmentID(ctx, db, userID)
		if err != nil {
			logger.Error("fetch user department failed", zap.Stringer("user_id", userID), zap.Error(err))
			return nil, fmt.Errorf("fetch user department: %w", err)
		}
		if deptID == nil {
			return &DataScopeCondition{Query: "created_by = ?", Args: []interface{}{userID}}, nil
		}
		deptIDs, err := deptRepo.GetDepartmentAndSubIDs(ctx, *deptID)
		if err != nil {
			logger.Error("get department and sub ids failed", zap.Stringer("department_id", *deptID), zap.Error(err))
			return nil, fmt.Errorf("get department and sub ids: %w", err)
		}
		if len(deptIDs) == 0 {
			return &DataScopeCondition{Query: "1 = 0"}, nil
		}
		return &DataScopeCondition{Query: "department_id IN ?", Args: []interface{}{deptIDs}}, nil

	case rbacModel.DataScopeCustom:
		customDeptIDs, err := fetchCustomDepartmentIDs(ctx, db, userID, resource)
		if err != nil {
			logger.Error("fetch custom department ids failed", zap.Stringer("user_id", userID), zap.Error(err))
			return nil, fmt.Errorf("fetch custom department ids: %w", err)
		}
		if len(customDeptIDs) == 0 {
			return &DataScopeCondition{Query: "1 = 0"}, nil
		}
		return &DataScopeCondition{Query: "department_id IN ?", Args: []interface{}{customDeptIDs}}, nil

	default:
		return nil, rbac.NewInvalidDataScopeError(scopeType)
	}
}

// fetchUserDataScopes returns the distinct data_scope values granted to the
// user for the given resource across all of the user's active roles.
func fetchUserDataScopes(ctx context.Context, db *gorm.DB, userID uuid.UUID, resource string) ([]string, error) {
	type scopeRow struct {
		DataScope string
	}
	var rows []scopeRow
	err := db.WithContext(ctx).Raw(`
		SELECT DISTINCT rp.data_scope
		FROM role_permissions rp
		JOIN user_roles ur ON ur.role_id = rp.role_id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE ur.user_id = ?
		  AND p.resource = ?
		  AND (ur.expired_at IS NULL OR ur.expired_at > NOW())
	`, userID, resource).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	scopes := make([]string, 0, len(rows))
	for _, r := range rows {
		scopes = append(scopes, r.DataScope)
	}
	return scopes, nil
}

// fetchUserDepartmentID returns the department id of the given user, or nil if
// the user is not assigned to any department.
func fetchUserDepartmentID(ctx context.Context, db *gorm.DB, userID uuid.UUID) (*uuid.UUID, error) {
	var row struct {
		DepartmentID *uuid.UUID
	}
	if err := db.WithContext(ctx).Raw(`SELECT department_id FROM users WHERE id = ?`, userID).Scan(&row).Error; err != nil {
		return nil, err
	}
	return row.DepartmentID, nil
}

// fetchCustomDepartmentIDs returns the distinct department ids granted to the
// user via "custom" data scopes for the given resource.
func fetchCustomDepartmentIDs(ctx context.Context, db *gorm.DB, userID uuid.UUID, resource string) ([]uuid.UUID, error) {
	type deptRow struct {
		DepartmentID uuid.UUID
	}
	var rows []deptRow
	err := db.WithContext(ctx).Raw(`
		SELECT DISTINCT rds.department_id
		FROM role_data_scopes rds
		JOIN role_permissions rp ON rp.id = rds.role_permission_id
		JOIN user_roles ur ON ur.role_id = rp.role_id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE ur.user_id = ?
		  AND p.resource = ?
		  AND rp.data_scope = 'custom'
		  AND (ur.expired_at IS NULL OR ur.expired_at > NOW())
	`, userID, resource).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.DepartmentID)
	}
	return ids, nil
}

// scopeRank assigns a permissiveness rank to a data scope value; higher means
// more permissive.
func scopeRank(scope string) int {
	switch scope {
	case rbacModel.DataScopeAll:
		return 5
	case rbacModel.DataScopeDepartmentAndSub:
		return 4
	case rbacModel.DataScopeCustom:
		return 3
	case rbacModel.DataScopeDepartment:
		return 2
	case rbacModel.DataScopeSelf:
		return 1
	default:
		return 0
	}
}

// mostPermissiveScope returns the most permissive scope from the given list,
// defaulting to "self" when none of the provided scopes are recognized.
func mostPermissiveScope(scopes []string) string {
	best := rbacModel.DataScopeSelf
	bestRank := 0
	for _, s := range scopes {
		if r := scopeRank(s); r > bestRank {
			bestRank = r
			best = s
		}
	}
	return best
}
