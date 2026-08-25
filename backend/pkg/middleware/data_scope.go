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

// DataScopeCondition 表示数据权限过滤的 SQL 条件
// 当 Query 为空时表示不附加任何过滤（用户可访问全部数据，如超级管理员或 all 范围）
// 当 Query 为 "1 = 0" 时表示不允许访问任何记录
// 所有条件均通过 Args 参数化传递，避免 SQL 注入风险
type DataScopeCondition struct {
	Query string
	Args  []interface{}
}

// dataScopeContextKey 数据权限条件在 Gin context 中的存储键名
const dataScopeContextKey = "data_scope_condition"

// DataScopeMiddleware 创建数据权限中间件
// 该中间件在请求处理前计算用户的数据权限范围，并将过滤条件存入 context
// 下游 handler 和 service 可通过 GetDataScopeFromContext 获取过滤条件
// 这样可以确保数据权限统一生效，避免因遗漏调用而产生越权漏洞
//
// db 和 deptRepo 通过闭包注入，便于单元测试时替换为 mock 实现
func DataScopeMiddleware(db *gorm.DB, deptRepo rbacRepo.DepartmentRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 context 获取用户 ID（由 JWTAuth 中间件设置）
		userIDStr, exists := c.Get("user_id")
		if !exists {
			// 没有用户身份，跳过（由鉴权中间件处理未登录情况）
			c.Next()
			return
		}

		userID, err := uuid.Parse(fmt.Sprintf("%v", userIDStr))
		if err != nil {
			logger.Warn("data scope: invalid user id in context", zap.Error(err))
			c.Next()
			return
		}

		// 从路由参数中获取资源名（通过 RequireDataScope 设置）
		resource, _ := c.Get("data_scope_resource")
		resourceStr, _ := resource.(string)
		if resourceStr == "" {
			// 未声明数据权限资源，默认不限制（由权限码控制访问即可）
			c.Next()
			return
		}

		// 构建数据权限条件
		condition, err := buildDataScopeCondition(c.Request.Context(), db, deptRepo, userID, resourceStr, c)
		if err != nil {
			logger.Error("build data scope condition failed",
				zap.Stringer("user_id", userID),
				zap.String("resource", resourceStr),
				zap.Error(err))
			// 构建失败时默认只允许查看自己的数据（fail closed 策略）
			c.Set(dataScopeContextKey, &DataScopeCondition{
				Query: "created_by = ?",
				Args:  []interface{}{userID},
			})
			c.Next()
			return
		}

		c.Set(dataScopeContextKey, condition)
		c.Next()
	}
}

// RequireDataScope 声明路由需要的数据权限资源
// 应在路由注册时使用，类似于 RequirePermission
func RequireDataScope(resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("data_scope_resource", resource)
		c.Next()
	}
}

// GetDataScopeFromContext 从 context 中获取数据权限条件
// 供下游 handler 和 service 层使用
// 如果 context 中没有数据权限条件（如未启用数据权限中间件），返回 nil
func GetDataScopeFromContext(c *gin.Context) *DataScopeCondition {
	val, exists := c.Get(dataScopeContextKey)
	if !exists {
		return nil
	}
	cond, ok := val.(*DataScopeCondition)
	if !ok {
		return nil
	}
	return cond
}

// buildDataScopeCondition 构建用户在指定资源上的数据权限过滤条件
// 超级管理员或数据范围为 all 时返回空条件（不限制）
// 其他情况根据数据范围返回对应的过滤条件
func buildDataScopeCondition(ctx context.Context, db *gorm.DB, deptRepo rbacRepo.DepartmentRepo, userID uuid.UUID, resource string, c *gin.Context) (*DataScopeCondition, error) {
	if db == nil {
		return nil, fmt.Errorf("data scope: database handle is nil")
	}

	// 超级管理员不进行数据过滤
	if isSuper, exists := c.Get("is_super_admin"); exists {
		if b, ok := isSuper.(bool); ok && b {
			return &DataScopeCondition{}, nil
		}
	}

	// 获取用户在指定资源上的数据权限范围集合
	scopes, err := fetchUserDataScopes(ctx, db, userID, resource)
	if err != nil {
		return nil, fmt.Errorf("fetch user data scopes: %w", err)
	}

	// 用户在该资源上没有任何数据权限，退化为仅本人数据（最安全的非空默认）
	if len(scopes) == 0 {
		return &DataScopeCondition{
			Query: "created_by = ?",
			Args:  []interface{}{userID},
		}, nil
	}

	// 取最宽松的数据权限范围
	scopeType := mostPermissiveScope(scopes)

	switch scopeType {
	case rbacModel.DataScopeAll:
		// 全部数据：不附加过滤条件
		return &DataScopeCondition{}, nil

	case rbacModel.DataScopeSelf:
		return &DataScopeCondition{
			Query: "created_by = ?",
			Args:  []interface{}{userID},
		}, nil

	case rbacModel.DataScopeDepartment:
		deptID, err := fetchUserDepartmentID(ctx, db, userID)
		if err != nil {
			return nil, fmt.Errorf("fetch user department: %w", err)
		}
		if deptID == nil {
			// 缺少部门信息时退化为仅本人数据
			return &DataScopeCondition{
				Query: "created_by = ?",
				Args:  []interface{}{userID},
			}, nil
		}
		return &DataScopeCondition{
			Query: "department_id = ?",
			Args:  []interface{}{*deptID},
		}, nil

	case rbacModel.DataScopeDepartmentAndSub:
		deptID, err := fetchUserDepartmentID(ctx, db, userID)
		if err != nil {
			return nil, fmt.Errorf("fetch user department: %w", err)
		}
		if deptID == nil {
			return &DataScopeCondition{
				Query: "created_by = ?",
				Args:  []interface{}{userID},
			}, nil
		}
		deptIDs, err := deptRepo.GetDepartmentAndSubIDs(ctx, *deptID)
		if err != nil {
			return nil, fmt.Errorf("get department and sub ids: %w", err)
		}
		if len(deptIDs) == 0 {
			return &DataScopeCondition{Query: "1 = 0"}, nil
		}
		return &DataScopeCondition{
			Query: "department_id IN ?",
			Args:  []interface{}{deptIDs},
		}, nil

	case rbacModel.DataScopeCustom:
		customDeptIDs, err := fetchCustomDepartmentIDs(ctx, db, userID, resource)
		if err != nil {
			return nil, fmt.Errorf("fetch custom department ids: %w", err)
		}
		if len(customDeptIDs) == 0 {
			return &DataScopeCondition{Query: "1 = 0"}, nil
		}
		return &DataScopeCondition{
			Query: "department_id IN ?",
			Args:  []interface{}{customDeptIDs},
		}, nil

	default:
		return nil, rbac.NewInvalidDataScopeError(scopeType)
	}
}

// fetchUserDataScopes 返回用户在指定资源上各角色授予的不同 data_scope 值（去重）
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

// fetchUserDepartmentID 返回用户所属部门 ID，未分配部门时返回 nil
func fetchUserDepartmentID(ctx context.Context, db *gorm.DB, userID uuid.UUID) (*uuid.UUID, error) {
	var row struct {
		DepartmentID *uuid.UUID
	}
	if err := db.WithContext(ctx).Raw(`SELECT department_id FROM users WHERE id = ?`, userID).Scan(&row).Error; err != nil {
		return nil, err
	}
	return row.DepartmentID, nil
}

// fetchCustomDepartmentIDs 返回用户通过 custom 数据权限授予的自定义部门 ID 列表
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

// scopeRank 为数据权限范围分配宽松度等级，数值越大越宽松
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

// mostPermissiveScope 返回给定范围列表中最宽松的那个
// 若列表为空或所有值均不识别，返回 "self"（最严格的非空默认）
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
