package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/internal/interview/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// Viewer carries server-resolved scopes. ReviewScope is a separate grant for
// confidential interview material; a general read permission is insufficient.
type Viewer struct {
	ID          uuid.UUID
	Scope       *rbacModel.DataScopeCondition
	ReviewScope *rbacModel.DataScopeCondition
}

func assignedTo(people []dto.Person, viewer uuid.UUID) bool {
	for _, p := range people {
		if p.ID == viewer.String() {
			return true
		}
	}
	return false
}

func (v Viewer) canRead(row *model.InterviewWithNames, people []dto.Person) bool {
	if v.Scope != nil && v.Scope.IsSelf && assignedTo(people, v.ID) {
		return true
	}
	return canAccessInterview(v.Scope, row.ApplicantID, row.DepartmentID, v.ID)
}

func (v Viewer) canReview(row *model.InterviewWithNames, people []dto.Person) bool {
	// A candidate must never read their own internal assessment, even when they
	// also hold an administrative role or were accidentally assigned as evaluator.
	if v.ID == uuid.Nil || v.ID == row.ApplicantID {
		return false
	}
	if assignedTo(people, v.ID) {
		return true
	}
	return canAccessInterview(v.ReviewScope, uuid.Nil, row.DepartmentID, v.ID)
}

func (s *interviewService) readInterview(ctx context.Context, viewer Viewer, id uuid.UUID) (*model.InterviewWithNames, []dto.Person, error) {
	row, err := s.records.GetByIDWithNames(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("get interview: %w", err)
	}
	if row == nil {
		return nil, nil, response.NewError(response.CodeInterviewRecordGone, "面试记录不存在")
	}
	named, err := s.records.ListInterviewers(ctx, []uuid.UUID{id})
	if err != nil {
		return nil, nil, fmt.Errorf("list interviewers: %w", err)
	}
	people := groupEvaluators(named)[id]
	if !viewer.canRead(row, people) {
		return nil, nil, response.NewError(response.CodeForbidden, "无权访问该面试")
	}
	return row, people, nil
}

// CanManageDepartment checks a server-resolved management scope. Self access
// never grants authority to create sessions or decide a candidate's result.
func (v Viewer) CanManageDepartment(departmentID *uuid.UUID) bool {
	return v.Scope != nil && !v.Scope.IsSelf && canAccessInterview(v.Scope, uuid.Nil, departmentID, v.ID)
}
