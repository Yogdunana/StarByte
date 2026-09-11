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
