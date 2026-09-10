package service

import (
	"strings"

	"github.com/google/uuid"

	rbac "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
)

func rewriteTaskScope(scope *rbac.DataScopeCondition, userID uuid.UUID) *rbac.DataScopeCondition {
	return rewriteTaskScopeAlias(scope, userID, "t")
}
func rewriteTaskScopeAlias(scope *rbac.DataScopeCondition, userID uuid.UUID, alias string) *rbac.DataScopeCondition {
	if userID == uuid.Nil {
		return &rbac.DataScopeCondition{Query: "1 = 0"}
	}
	if scope != nil && !scope.IsSelf && strings.TrimSpace(scope.Query) == "" {
		return &rbac.DataScopeCondition{}
	}
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}
	q := prefix + "creator_id = ? OR " + prefix + "assignee_id = ?"
	args := []interface{}{userID, userID}
	if scope != nil && !scope.IsSelf {
		switch strings.TrimSpace(scope.Query) {
		case "department_id = ?", "department_id IN ?":
			q += " OR " + prefix + scope.Query
			args = append(args, scope.Args...)
		}
	}
	return &rbac.DataScopeCondition{Query: "(" + q + ")", Args: args}
}
func canViewTask(t *model.Task, viewer uuid.UUID, scope *rbac.DataScopeCondition) bool {
	if viewer == uuid.Nil {
		return false
	}
	if t.CreatorID == viewer || (t.AssigneeID != nil && *t.AssigneeID == viewer) {
		return true
	}
	if scope == nil || scope.IsSelf {
		return false
	}
	switch strings.TrimSpace(scope.Query) {
	case "":
		return true
	case "department_id = ?", "department_id IN ?":
		if t.DepartmentID == nil {
			return false
		}
		for _, arg := range scope.Args {
			switch v := arg.(type) {
			case uuid.UUID:
				if v == *t.DepartmentID {
					return true
				}
			case []uuid.UUID:
				for _, id := range v {
					if id == *t.DepartmentID {
						return true
					}
				}
			}
		}
	}
	return false
}
