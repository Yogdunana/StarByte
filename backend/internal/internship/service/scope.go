package service

import (
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/internship/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
)

func rewriteScope(scope *rbacModel.DataScopeCondition, userID uuid.UUID) *rbacModel.DataScopeCondition {
	if scope == nil || scope.IsEmpty() {
		return scope
	}
	if scope.Query == "1 = 0" {
		return &rbacModel.DataScopeCondition{
			Query: "i.user_id = ?",
			Args:  []interface{}{userID},
		}
	}
	q := strings.ReplaceAll(scope.Query, "department_id", "i.department_id")
	return &rbacModel.DataScopeCondition{Query: q, Args: scope.Args}
}

func isAllScope(scope *rbacModel.DataScopeCondition) bool {
	return scope == nil || scope.IsEmpty()
}

func canAccessRecord(scope *rbacModel.DataScopeCondition, ownerID uuid.UUID, deptID *uuid.UUID, viewer uuid.UUID) bool {
	rewritten := rewriteScope(scope, viewer)
	if rewritten == nil || rewritten.IsEmpty() {
		return true
	}
	if rewritten.Query == "i.user_id = ?" {
		return ownerID == viewer
	}
	if deptID == nil {
		return false
	}
	return deptInArgs(rewritten.Args, *deptID)
}

func deptInArgs(args []interface{}, deptID uuid.UUID) bool {
	for _, arg := range args {
		switch v := arg.(type) {
		case uuid.UUID:
			if v == deptID {
				return true
			}
		case []uuid.UUID:
			for _, id := range v {
				if id == deptID {
					return true
				}
			}
		}
	}
	return false
}

func canEdit(row *model.Internship, operator uuid.UUID, scope *rbacModel.DataScopeCondition, cfg model.InternshipConfig) bool {
	if isAllScope(scope) {
		return true
	}
	if row.UserID == operator {
		return cfg.AllowStudentEdit
	}
	return cfg.AllowMinisterEdit && canAccessRecord(scope, row.UserID, row.DepartmentID, operator)
}

func canCompleteOrDelete(row *model.Internship, operator uuid.UUID, scope *rbacModel.DataScopeCondition) bool {
	return isAllScope(scope) || row.UserID == operator
}
