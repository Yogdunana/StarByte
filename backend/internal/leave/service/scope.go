package service

import (
	"strings"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
)

func rewriteApplicantScope(scope *rbacModel.DataScopeCondition, viewerID uuid.UUID) *rbacModel.DataScopeCondition {
	if scope == nil {
		return &rbacModel.DataScopeCondition{Query: "1 = 0"}
	}
	if scope.IsSelf {
		return &rbacModel.DataScopeCondition{
			Query:  "leave_applications.applicant_id = ?",
			Args:   []interface{}{viewerID},
			IsSelf: true,
		}
	}
	if strings.TrimSpace(scope.Query) == "" {
		return &rbacModel.DataScopeCondition{}
	}
	if scope.Query == "1 = 0" {
		return scope
	}
	q := strings.ReplaceAll(scope.Query, "department_id", "applicant.department_id")
	return &rbacModel.DataScopeCondition{Query: q, Args: scope.Args, IsSelf: scope.IsSelf}
}

func isUnrestricted(scope *rbacModel.DataScopeCondition) bool {
	return scope != nil && strings.TrimSpace(scope.Query) == "" && !scope.IsSelf
}

func canAccessApplicant(scope *rbacModel.DataScopeCondition, applicant uuid.UUID, dept *uuid.UUID, viewer uuid.UUID) bool {
	if applicant == viewer {
		return true
	}
	rewritten := rewriteApplicantScope(scope, viewer)
	if rewritten.IsEmpty() {
		return isUnrestricted(scope)
	}
	if rewritten.Query == "1 = 0" {
		return false
	}
	if rewritten.IsSelf || rewritten.Query == "leave_applications.applicant_id = ?" {
		return applicant == viewer
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
