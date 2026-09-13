package service

import (
	"strings"

	"github.com/google/uuid"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
)

// rewriteScope 把中间件条件套到带别名的联表查询上。
// self（1 = 0）按 Issue 改为仅本人：alias.user_id = 当前用户。
func rewriteScope(scope *rbacModel.DataScopeCondition, alias string, userID uuid.UUID) *rbacModel.DataScopeCondition {
	if scope == nil || userID == uuid.Nil {
		return &rbacModel.DataScopeCondition{Query: "1 = 0"}
	}
	if scope.IsEmpty() {
		return scope
	}
	if scope.IsSelf {
		return &rbacModel.DataScopeCondition{
			Query: alias + ".user_id = ?",
			Args:  []interface{}{userID},
		}
	}
	q := strings.ReplaceAll(scope.Query, "department_id", alias+".department_id")
	return &rbacModel.DataScopeCondition{Query: q, Args: scope.Args}
}

func canAccessRecord(scope *rbacModel.DataScopeCondition, ownerID uuid.UUID, deptID *uuid.UUID, viewer uuid.UUID) bool {
	if scope == nil || viewer == uuid.Nil {
		return false
	}
	rewritten := rewriteScope(scope, "x", viewer)
	if rewritten == nil || rewritten.IsEmpty() {
		return true
	}
	if rewritten.Query == "x.user_id = ?" {
		return ownerID == viewer
	}
	if rewritten.Query != "x.department_id = ?" && rewritten.Query != "x.department_id IN ?" {
		return false
	}
	if deptID == nil {
		return false
	}
	for _, arg := range rewritten.Args {
		switch v := arg.(type) {
		case uuid.UUID:
			if v == *deptID {
				return true
			}
		case []uuid.UUID:
			for _, id := range v {
				if id == *deptID {
					return true
				}
			}
		}
	}
	return strings.Contains(rewritten.Query, "IN") && deptInArgs(rewritten.Args, *deptID)
}

func deptInArgs(args []interface{}, deptID uuid.UUID) bool {
	for _, arg := range args {
		ids, ok := arg.([]uuid.UUID)
		if !ok {
			continue
		}
		for _, id := range ids {
			if id == deptID {
				return true
			}
		}
	}
	return false
}

func applicationListScope(scope *rbacModel.DataScopeCondition, viewer uuid.UUID) *rbacModel.DataScopeCondition {
	out := rewriteScope(scope, "a", viewer)
	if scope == nil || viewer == uuid.Nil || out.IsEmpty() {
		return out
	}
	query := strings.ReplaceAll(out.Query, "a.department_id", "CASE WHEN a.charter_policy THEN a.review_department_id ELSE a.department_id END")
	// Committee readers must currently hold a committee office and have an assigned committee task.
	query = "(" + query + ") OR (a.charter_policy AND EXISTS (SELECT 1 FROM flow_tasks ft JOIN user_roles ur ON ur.user_id=ft.assignee_id JOIN roles r ON r.id=ur.role_id WHERE ft.instance_id=a.flow_instance_id AND ft.node_id='committee' AND ft.assignee_id=? AND r.status=0 AND r.code IN ('president','vice_president','center_director','minister') AND (ur.expired_at IS NULL OR ur.expired_at>NOW())))"
	args := append([]interface{}{}, out.Args...)
	args = append(args, viewer)
	return &rbacModel.DataScopeCondition{Query: query, Args: args}
}
