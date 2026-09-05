package service

import (
	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/internal/interview/model"
	"github.com/google/uuid"
)

func summarizeEvaluations(
	interviewID uuid.UUID,
	applicant dto.Person,
	rows []model.EvaluationNamed,
	dims []model.Dimension,
) *dto.EvaluationSummary {
	weights := map[string]float64{}
	for _, d := range dims {
		weights[d.Name] = d.Weight
	}
	byEval := map[uuid.UUID]*dto.EvaluatorScores{}
	order := []uuid.UUID{}
	dimScores := map[string][]float64{}
	for _, row := range rows {
		item, ok := byEval[row.InterviewerID]
		if !ok {
			item = &dto.EvaluatorScores{
				Evaluator: dto.Person{ID: row.InterviewerID.String(), Name: row.EvaluatorName},
				Scores:    []dto.DimensionScore{},
			}
			byEval[row.InterviewerID] = item
			order = append(order, row.InterviewerID)
		}
		item.Scores = append(item.Scores, dto.DimensionScore{
			Dimension: row.Dimension,
			Score:     row.Score,
			Comment:   row.Comment,
		})
		dimScores[row.Dimension] = append(dimScores[row.Dimension], row.Score)
	}
	evals := make([]dto.EvaluatorScores, 0, len(order))
	var evalSum float64
	for _, id := range order {
		item := byEval[id]
		item.TotalScore = evaluatorTotal(item.Scores, weights)
		evalSum += item.TotalScore
		evals = append(evals, *item)
	}
	var avg float64
	if len(evals) > 0 {
		avg = round1(evalSum / float64(len(evals)))
	}
	return &dto.EvaluationSummary{
		InterviewID:   interviewID.String(),
		Applicant:     applicant,
		Evaluations:   evals,
		AverageScore:  avg,
		WeightedScore: weightedAverage(dimScores, weights),
	}
}

func evaluatorTotal(scores []dto.DimensionScore, weights map[string]float64) float64 {
	var sum, wsum float64
	var usedWeight bool
	for _, s := range scores {
		if w, ok := weights[s.Dimension]; ok && w > 0 {
			sum += s.Score * w
			wsum += w
			usedWeight = true
			continue
		}
		sum += s.Score
		wsum++
	}
	if wsum == 0 {
		return 0
	}
	if usedWeight {
		return round1(sum / wsum)
	}
	return round1(sum / float64(len(scores)))
}

func weightedAverage(dimScores map[string][]float64, weights map[string]float64) float64 {
	var sum, wsum float64
	for name, scores := range dimScores {
		if len(scores) == 0 {
			continue
		}
		var avg float64
		for _, s := range scores {
			avg += s
		}
		avg /= float64(len(scores))
		w := weights[name]
		if w <= 0 {
			w = 1
		}
		sum += avg * w
		wsum += w
	}
	if wsum == 0 {
		return 0
	}
	return round1(sum / wsum)
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
