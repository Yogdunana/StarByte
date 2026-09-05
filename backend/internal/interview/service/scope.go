package service

import (
	"strings"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
)

func rewriteSessionScope(scope *rbacModel.DataScopeCondition, userID uuid.UUID) *rbacModel.DataScopeCondition {
	if scope == nil || scope.IsEmpty() {
		return scope
	}
	if scope.Query == "1 = 0" {
		return &rbacModel.DataScopeCondition{
			Query: "s.created_by = ?",
			Args:  []interface{}{userID},
		}
	}
	q := strings.ReplaceAll(scope.Query, "department_id", "s.department_id")
	return &rbacModel.DataScopeCondition{Query: q, Args: scope.Args}
}

func rewriteInterviewScope(scope *rbacModel.DataScopeCondition, userID uuid.UUID) *rbacModel.DataScopeCondition {
	if scope == nil || scope.IsEmpty() {
		return scope
	}
	if scope.Query == "1 = 0" {
		return &rbacModel.DataScopeCondition{
			Query: "i.applicant_id = ? OR i.id IN (SELECT interview_id FROM interview_interviewers WHERE interviewer_id = ?)",
			Args:  []interface{}{userID, userID},
		}
	}
	q := strings.ReplaceAll(scope.Query, "department_id", "s.department_id")
	return &rbacModel.DataScopeCondition{Query: q, Args: scope.Args}
}

func canAccessInterview(scope *rbacModel.DataScopeCondition, applicantID uuid.UUID, deptID *uuid.UUID, viewer uuid.UUID) bool {
	rewritten := rewriteInterviewScope(scope, viewer)
	if rewritten == nil || rewritten.IsEmpty() {
		return true
	}
	if strings.Contains(rewritten.Query, "applicant_id") {
		return applicantID == viewer
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
	return strings.Contains(rewritten.Query, "IN")
}
