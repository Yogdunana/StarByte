package service

import (
	"strings"

	"github.com/google/uuid"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
)

func rewriteSessionScope(scope *rbacModel.DataScopeCondition, userID uuid.UUID) *rbacModel.DataScopeCondition {
	if scope == nil {
		return &rbacModel.DataScopeCondition{Query: "1 = 0"}
	}
	if scope.IsEmpty() {
		return scope
	}
	if scope.IsSelf {
		return &rbacModel.DataScopeCondition{
			Query: "s.created_by = ? OR s.id IN (SELECT i.session_id FROM interviews i JOIN interview_interviewers ii ON i.id = ii.interview_id WHERE ii.interviewer_id = ? AND i.session_id IS NOT NULL)",
			Args:  []interface{}{userID, userID},
		}
	}
	q := strings.ReplaceAll(scope.Query, "department_id", "s.department_id")
	return &rbacModel.DataScopeCondition{Query: q, Args: scope.Args}
}

func rewriteInterviewScope(scope *rbacModel.DataScopeCondition, userID uuid.UUID) *rbacModel.DataScopeCondition {
	if scope == nil {
		return &rbacModel.DataScopeCondition{Query: "1 = 0"}
	}
	if scope.IsEmpty() {
		return scope
	}
	if scope.IsSelf {
		return &rbacModel.DataScopeCondition{
			Query: "i.applicant_id = ? OR i.id IN (SELECT interview_id FROM interview_interviewers WHERE interviewer_id = ?)",
			Args:  []interface{}{userID, userID},
		}
	}
	q := strings.ReplaceAll(scope.Query, "department_id", "s.department_id")
	return &rbacModel.DataScopeCondition{Query: q, Args: scope.Args}
}

func canAccessInterview(scope *rbacModel.DataScopeCondition, applicantID uuid.UUID, deptID *uuid.UUID, viewer uuid.UUID) bool {
	if scope == nil || viewer == uuid.Nil {
		return false
	}
	if scope.IsSelf {
		return applicantID == viewer
	}
	if scope.IsEmpty() {
		return true
	}
	if scope.Query != "department_id = ?" && scope.Query != "department_id IN ?" {
		return false
	}
	if deptID == nil {
		return false
	}
	for _, arg := range scope.Args {
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
	return false
}
