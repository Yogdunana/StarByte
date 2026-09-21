package service

import (
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
)

// activityScopeClause 把 RBAC 数据范围条件改写为可作用于 activities（别名 a）的 WHERE 片段。
// activities 表没有 department_id，归属部门由 organizer_id 关联的 users.department_id 决定，
// 因此部门范围必须改写成子查询 —— 这样在带 users JOIN 的 namedQuery 和不带 JOIN 的 Count
// 查询里都能正确求值。返回 nil 表示不限制。
func activityScopeClause(scope *rbacModel.DataScopeCondition, viewer uuid.UUID) *rbacModel.DataScopeCondition {
	if scope == nil || scope.IsEmpty() {
		return nil
	}
	if scope.IsSelf {
		return &rbacModel.DataScopeCondition{Query: "a.organizer_id = ?", Args: []interface{}{viewer}, IsSelf: true}
	}
	if scope.Query == "1 = 0" {
		return scope
	}
	// 部门范围：activities 没有 department_id，需通过 organizer 关联 users 解析其部门，
	// 用子查询保证在 Count（无 users join）与 namedQuery（users 别名 u）两种写法下均成立。
	return &rbacModel.DataScopeCondition{
		Query: "EXISTS (SELECT 1 FROM users su WHERE su.id = a.organizer_id AND su.department_id IN ?)",
		Args:  scope.Args,
	}
}

// canAccessActivity 判定 viewer 能否访问 organizerID 创建的活动。
// organizerDept 为 nil 表示无法解析组织者部门；此时除 all/self 外一律拒绝（fail-closed）。
func canAccessActivity(scope *rbacModel.DataScopeCondition, organizerID uuid.UUID, organizerDept *uuid.UUID, viewer uuid.UUID) bool {
	if scope == nil || scope.IsEmpty() {
		return true
	}
	if scope.IsSelf {
		return organizerID == viewer
	}
	if scope.Query == "1 = 0" {
		return false
	}
	// 部门范围：组织者为 nil 部门时无法命中，直接拒绝（fail-closed）。
	if organizerDept == nil {
		return false
	}
	for _, arg := range scope.Args {
		switch v := arg.(type) {
		case uuid.UUID:
			if v == *organizerDept {
				return true
			}
		case []uuid.UUID:
			for _, id := range v {
				if id == *organizerDept {
					return true
				}
			}
		}
	}
	return false
}
