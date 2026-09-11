package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/internal/interview/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func TestInterviewPrivacy(t *testing.T) {
	applicant, examiner, manager, outsider, dept := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	all := &rbacModel.DataScopeCondition{}
	self := &rbacModel.DataScopeCondition{Query: "1 = 0", IsSelf: true}
	department := &rbacModel.DataScopeCondition{Query: "department_id IN ?", Args: []interface{}{[]uuid.UUID{dept}}}
	otherDept := &rbacModel.DataScopeCondition{Query: "department_id IN ?", Args: []interface{}{[]uuid.UUID{uuid.New()}}}
	tests := []struct {
		name              string
		viewer            Viewer
		readable, private bool
	}{
		{"candidate", Viewer{ID: applicant, Scope: self}, true, false},
		{"candidate with admin grants", Viewer{ID: applicant, Scope: all, ReviewScope: all}, true, false},
		{"assigned examiner", Viewer{ID: examiner, Scope: self}, true, true},
		{"scoped reviewer", Viewer{ID: manager, Scope: department, ReviewScope: department}, true, true},
		{"read only colleague", Viewer{ID: outsider, Scope: department}, true, false},
		{"unrelated department", Viewer{ID: outsider, Scope: otherDept, ReviewScope: all}, false, false},
		{"explicit deny", Viewer{ID: examiner, Scope: &rbacModel.DataScopeCondition{Query: "1 = 0"}}, false, false},
		{"missing scope", Viewer{ID: manager, ReviewScope: all}, false, false},
		{"missing identity", Viewer{Scope: all, ReviewScope: all}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records, evals := &mockInterviewRepo{}, &mockEvalRepo{}
			svc := NewInterviewService(&mockSessionRepo{}, records, evals, nil, nil)
			id := uuid.New()
			score := 83.0
			row := &model.InterviewWithNames{Interview: model.Interview{ID: id, ApplicantID: applicant, ResultComment: "CONFIDENTIAL", Score: &score}, DepartmentID: &dept}
			records.On("GetByIDWithNames", mock.Anything, id).Return(row, nil)
			records.On("ListInterviewers", mock.Anything, []uuid.UUID{id}).Return([]model.InterviewerNamed{{InterviewID: id, InterviewerID: examiner}}, nil)
			out, err := svc.GetInterview(context.Background(), tt.viewer, id)
			if !tt.readable {
				requireAppError(t, err, response.CodeForbidden)
			} else {
				require.NoError(t, err)
				body, err := json.Marshal(out)
				require.NoError(t, err)
				if tt.private {
					require.Contains(t, string(body), "CONFIDENTIAL")
				} else {
					require.NotContains(t, string(body), "result_comment")
					require.NotContains(t, string(body), "score")
				}
			}
			if tt.private {
				evals.On("ListByInterview", mock.Anything, id).Return([]model.EvaluationNamed{}, nil)
				evals.On("ListDimensions", mock.Anything).Return([]model.Dimension{}, nil)
			}
			_, err = svc.GetEvaluations(context.Background(), tt.viewer, id)
			if tt.private {
				require.NoError(t, err)
			} else {
				requireAppError(t, err, response.CodeForbidden)
			}
		})
	}
}

func TestMyInterviewsNeverContainsInternalAssessment(t *testing.T) {
	records := &mockInterviewRepo{}
	svc := NewInterviewService(&mockSessionRepo{}, records, &mockEvalRepo{}, nil, nil)
	user, id := uuid.New(), uuid.New()
	score := 75.0
	records.On("ListMine", mock.Anything, user, mock.Anything).Return([]model.InterviewWithNames{{Interview: model.Interview{ID: id, ApplicantID: user, ResultComment: "CONFIDENTIAL", Score: &score}}}, nil)
	records.On("ListInterviewers", mock.Anything, []uuid.UUID{id}).Return([]model.InterviewerNamed{}, nil)
	out, err := svc.MyInterviews(context.Background(), user, nil)
	require.NoError(t, err)
	body, err := json.Marshal(out)
	require.NoError(t, err)
	require.NotContains(t, string(body), "result_comment")
	require.NotContains(t, string(body), "score")
}

func TestResultNotificationDoesNotReceiveInternalComment(t *testing.T) {
	records, notify := &mockInterviewRepo{}, &mockNotify{}
	svc := NewInterviewService(&mockSessionRepo{}, records, &mockEvalRepo{}, notify, nil).(*interviewService)
	user := uuid.New()
	records.On("GetUser", mock.Anything, user).Return(&model.NamedUser{ID: user}, nil)
	notify.On("Send", mock.Anything, []uuid.UUID{user}, tplResult, mock.MatchedBy(func(vars map[string]interface{}) bool {
		body, err := json.Marshal(vars)
		require.NoError(t, err)
		require.NotContains(t, string(body), "CONFIDENTIAL")
		return vars["comment"] == ""
	})).Return(nil).Once()
	svc.notifyResult(context.Background(), &model.Interview{ApplicantID: user, ResultCode: model.ResultPass, ResultComment: "CONFIDENTIAL"})
	notify.AssertExpectations(t)
}

func TestCandidateCannotScoreThemselves(t *testing.T) {
	records := &mockInterviewRepo{}
	user, id := uuid.New(), uuid.New()
	svc := NewInterviewService(&mockSessionRepo{}, records, &mockEvalRepo{}, nil, nil)
	records.On("GetByID", mock.Anything, id).Return(&model.Interview{ID: id, ApplicantID: user, Status: model.InterviewOngoing}, nil)
	_, err := svc.SubmitEvaluations(context.Background(), user, id, &dto.SubmitEvaluationsRequest{})
	requireAppError(t, err, response.CodeForbidden)
}

func TestScopeDenyIsNotRewrittenAsSelf(t *testing.T) {
	uid := uuid.New()
	deny := &rbacModel.DataScopeCondition{Query: "1 = 0"}
	require.Equal(t, "1 = 0", rewriteInterviewScope(deny, uid).Query)
	require.Equal(t, "1 = 0", rewriteSessionScope(deny, uid).Query)
}

func TestCandidateCannotAssignOrDecideOwnInterview(t *testing.T) {
	records := &mockInterviewRepo{}
	user, id := uuid.New(), uuid.New()
	svc := NewInterviewService(&mockSessionRepo{}, records, &mockEvalRepo{}, nil, nil)
	records.On("GetByID", mock.Anything, id).Return(&model.Interview{ID: id, ApplicantID: user, Status: model.InterviewOngoing}, nil)
	_, err := svc.AssignEvaluators(context.Background(), id, &dto.AssignEvaluatorsRequest{EvaluatorIDs: []string{user.String()}})
	requireAppError(t, err, response.CodeForbidden)
	_, err = svc.SubmitResult(context.Background(), user, id, &dto.SubmitResultRequest{Result: model.ResultPass})
	requireAppError(t, err, response.CodeForbidden)
}
