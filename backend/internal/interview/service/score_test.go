package service

import (
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/internal/interview/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSummarizeEvaluations_AverageAndWeighted(t *testing.T) {
	dims := []model.Dimension{
		{Name: "技术能力", Weight: 0.4},
		{Name: "沟通能力", Weight: 0.3},
		{Name: "逻辑思维", Weight: 0.3},
	}
	a := uuid.New()
	b := uuid.New()
	rows := []model.EvaluationNamed{
		{Evaluation: model.Evaluation{InterviewerID: a, Dimension: "技术能力", Score: 85}, EvaluatorName: "A"},
		{Evaluation: model.Evaluation{InterviewerID: a, Dimension: "沟通能力", Score: 90}, EvaluatorName: "A"},
		{Evaluation: model.Evaluation{InterviewerID: a, Dimension: "逻辑思维", Score: 80}, EvaluatorName: "A"},
		{Evaluation: model.Evaluation{InterviewerID: b, Dimension: "技术能力", Score: 75}, EvaluatorName: "B"},
		{Evaluation: model.Evaluation{InterviewerID: b, Dimension: "沟通能力", Score: 70}, EvaluatorName: "B"},
		{Evaluation: model.Evaluation{InterviewerID: b, Dimension: "逻辑思维", Score: 80}, EvaluatorName: "B"},
	}
	sum := summarizeEvaluations(uuid.New(), dto.Person{Name: "张三"}, rows, dims)
	require.Len(t, sum.Evaluations, 2)
	require.InDelta(t, 85, sum.Evaluations[0].TotalScore, 0.2)
	require.InDelta(t, 75, sum.Evaluations[1].TotalScore, 0.2)
	require.InDelta(t, 80, sum.AverageScore, 0.2)
	require.InDelta(t, 80, sum.WeightedScore, 0.2)
}

func TestRound1(t *testing.T) {
	require.Equal(t, 86.0, round1(85.95))
	require.Equal(t, 87.5, round1(87.54))
}
