package service

import (
	"context"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/internal/interview/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func defaultDims() []model.Dimension {
	return []model.Dimension{
		{Name: "技术能力", Weight: 0.4, MaxScore: 100},
		{Name: "沟通能力", Weight: 0.3, MaxScore: 100},
	}
}

func TestSubmitEvaluations_Duplicate(t *testing.T) {
	records := &mockInterviewRepo{}
	evals := &mockEvalRepo{}
	svc := NewInterviewService(&mockSessionRepo{}, records, evals, nil, nil)
	id := uuid.New()
	evalID := uuid.New()
	records.On("GetByID", mock.Anything, id).Return(&model.Interview{ID: id, Status: model.InterviewOngoing}, nil)
	records.On("ListInterviewers", mock.Anything, []uuid.UUID{id}).Return([]model.InterviewerNamed{{InterviewID: id, InterviewerID: evalID}}, nil)
	evals.On("HasEvaluatorScores", mock.Anything, id, evalID).Return(true, nil)
	_, err := svc.SubmitEvaluations(context.Background(), evalID, id, &dto.SubmitEvaluationsRequest{
		Evaluations: []dto.EvaluationItem{{Dimension: "技术能力", Score: 80}},
	})
	requireAppError(t, err, response.CodeInterviewDupEval)
}

func TestSubmitEvaluations_ScoreRange(t *testing.T) {
	records := &mockInterviewRepo{}
	evals := &mockEvalRepo{}
	svc := NewInterviewService(&mockSessionRepo{}, records, evals, nil, nil)
	id := uuid.New()
	evalID := uuid.New()
	records.On("GetByID", mock.Anything, id).Return(&model.Interview{ID: id, Status: model.InterviewOngoing}, nil)
	records.On("ListInterviewers", mock.Anything, []uuid.UUID{id}).Return([]model.InterviewerNamed{{InterviewID: id, InterviewerID: evalID}}, nil)
	evals.On("HasEvaluatorScores", mock.Anything, id, evalID).Return(false, nil)
	evals.On("ListDimensions", mock.Anything).Return(defaultDims(), nil)
	_, err := svc.SubmitEvaluations(context.Background(), evalID, id, &dto.SubmitEvaluationsRequest{
		Evaluations: []dto.EvaluationItem{{Dimension: "技术能力", Score: 120}},
	})
	requireAppError(t, err, response.CodeInterviewScoreRange)
}

func TestSubmitEvaluations_DimGone(t *testing.T) {
	records := &mockInterviewRepo{}
	evals := &mockEvalRepo{}
	svc := NewInterviewService(&mockSessionRepo{}, records, evals, nil, nil)
	id := uuid.New()
	evalID := uuid.New()
	records.On("GetByID", mock.Anything, id).Return(&model.Interview{ID: id, Status: model.InterviewOngoing}, nil)
	records.On("ListInterviewers", mock.Anything, []uuid.UUID{id}).Return([]model.InterviewerNamed{{InterviewID: id, InterviewerID: evalID}}, nil)
	evals.On("HasEvaluatorScores", mock.Anything, id, evalID).Return(false, nil)
	evals.On("ListDimensions", mock.Anything).Return(defaultDims(), nil)
	_, err := svc.SubmitEvaluations(context.Background(), evalID, id, &dto.SubmitEvaluationsRequest{
		Evaluations: []dto.EvaluationItem{{Dimension: "不存在", Score: 80}},
	})
	requireAppError(t, err, response.CodeInterviewDimGone)
}

func TestSubmitEvaluations_NoEvaluator(t *testing.T) {
	records := &mockInterviewRepo{}
	svc := NewInterviewService(&mockSessionRepo{}, records, &mockEvalRepo{}, nil, nil)
	id := uuid.New()
	records.On("GetByID", mock.Anything, id).Return(&model.Interview{ID: id, Status: model.InterviewOngoing}, nil)
	records.On("ListInterviewers", mock.Anything, []uuid.UUID{id}).Return([]model.InterviewerNamed{}, nil)
	_, err := svc.SubmitEvaluations(context.Background(), uuid.New(), id, &dto.SubmitEvaluationsRequest{
		Evaluations: []dto.EvaluationItem{{Dimension: "技术能力", Score: 80}},
	})
	requireAppError(t, err, response.CodeInterviewNoEvaluator)
}

func TestSubmitEvaluations_OK(t *testing.T) {
	records := &mockInterviewRepo{}
	evals := &mockEvalRepo{}
	svc := NewInterviewService(&mockSessionRepo{}, records, evals, nil, nil)
	id := uuid.New()
	evalID := uuid.New()
	iv := &model.Interview{ID: id, Status: model.InterviewOngoing, ApplicantID: uuid.New()}
	records.On("GetByID", mock.Anything, id).Return(iv, nil)
	records.On("ListInterviewers", mock.Anything, []uuid.UUID{id}).Return([]model.InterviewerNamed{{InterviewID: id, InterviewerID: evalID, Name: "官"}}, nil)
	evals.On("HasEvaluatorScores", mock.Anything, id, evalID).Return(false, nil)
	evals.On("ListDimensions", mock.Anything).Return(defaultDims(), nil)
	evals.On("CreateBatch", mock.Anything, mock.Anything).Return(nil)
	evals.On("ListByInterview", mock.Anything, id).Return([]model.EvaluationNamed{
		{Evaluation: model.Evaluation{InterviewerID: evalID, Dimension: "技术能力", Score: 80}, EvaluatorName: "官"},
	}, nil)
	records.On("GetByIDWithNames", mock.Anything, id).Return(namedInterview(iv, "张三"), nil)
	records.On("Update", mock.Anything, mock.AnythingOfType("*model.Interview")).Return(nil)
	out, err := svc.SubmitEvaluations(context.Background(), evalID, id, &dto.SubmitEvaluationsRequest{
		Evaluations: []dto.EvaluationItem{{Dimension: "技术能力", Score: 80, Comment: "扎实"}},
	})
	require.NoError(t, err)
	require.Equal(t, 80.0, out.AverageScore)
}
