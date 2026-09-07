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
	if scope.IsSelf {
		return &rbacModel.DataScopeCondition{Query: "r.created_by = ?", Args: []interface{}{userID}, IsSelf: true}
	}
	if scope.Query == "1 = 0" {
		return scope
	}
	q := strings.ReplaceAll(scope.Query, "department_id", "r.department_id")
	return &rbacModel.DataScopeCondition{Query: q, Args: scope.Args, IsSelf: scope.IsSelf}
}

func canAccess(scope *rbacModel.DataScopeCondition, createdBy *uuid.UUID, dept *uuid.UUID, viewer uuid.UUID) bool {
	rewritten := rewriteScope(scope, viewer)
	if rewritten == nil || rewritten.IsEmpty() {
		return true
	}
	if rewritten.Query == "1 = 0" {
		return false
	}
	if rewritten.Query == "r.created_by = ?" || rewritten.IsSelf {
		return createdBy != nil && *createdBy == viewer
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
