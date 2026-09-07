package service

import (
	"strings"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
)

func rewriteScope(scope *rbacModel.DataScopeCondition, userID uuid.UUID) *rbacModel.DataScopeCondition {
	if scope == nil || scope.IsEmpty() {
		return scope
	}
	if scope.Query == "1 = 0" {
		if !scope.IsSelf {
			return scope
		}
		return &rbacModel.DataScopeCondition{Query: "c.user_id = ?", Args: []interface{}{userID}, IsSelf: true}
	}
	q := strings.ReplaceAll(scope.Query, "department_id", "u.department_id")
	return &rbacModel.DataScopeCondition{Query: q, Args: scope.Args, IsSelf: scope.IsSelf}
}

func canAccess(scope *rbacModel.DataScopeCondition, owner uuid.UUID, dept *uuid.UUID, viewer uuid.UUID) bool {
	rewritten := rewriteScope(scope, viewer)
	if rewritten == nil || rewritten.IsEmpty() {
		return true
	}
	if rewritten.Query == "1 = 0" {
		return false
	}
	if rewritten.Query == "c.user_id = ?" || rewritten.IsSelf {
		return owner == viewer
	}
	if dept == nil {
		return false
	}
	for _, arg := range rewritten.Args {
		switch v := arg.(type) {
		case uuid.UUID:
			if v == *dept {
				return true
			}
		case []uuid.UUID:
			for _, id := range v {
				if id == *dept {
					return true
				}
			}
		}
	}
	return false
}
