package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/internal/interview/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *interviewService) SubmitEvaluations(ctx context.Context, evaluator, id uuid.UUID, req *dto.SubmitEvaluationsRequest) (*dto.EvaluationSummary, error) {
	iv, err := s.mustInterview(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canScore(iv.Status) {
		return nil, response.NewError(response.CodeInterviewInvalidState, "当前状态不允许评分")
	}
	if err := s.ensureAssigned(ctx, id, evaluator); err != nil {
		return nil, err
	}
	dup, err := s.evals.HasEvaluatorScores(ctx, id, evaluator)
	if err != nil {
		return nil, fmt.Errorf("check dup eval: %w", err)
	}
	if dup {
		return nil, response.NewError(response.CodeInterviewDupEval, "重复评分")
	}
	dims, err := s.evals.ListDimensions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list dimensions: %w", err)
	}
	rows, err := buildEvalRows(id, evaluator, req.Evaluations, dims)
	if err != nil {
		return nil, err
	}
	if err := s.evals.CreateBatch(ctx, rows); err != nil {
		return nil, fmt.Errorf("create evaluations: %w", err)
	}
	if err := s.persistWeightedScore(ctx, iv); err != nil {
		return nil, err
	}
	return s.GetEvaluations(ctx, id)
}

func (s *interviewService) GetEvaluations(ctx context.Context, id uuid.UUID) (*dto.EvaluationSummary, error) {
	row, err := s.records.GetByIDWithNames(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get interview: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeInterviewRecordGone, "面试记录不存在")
	}
	evals, err := s.evals.ListByInterview(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list evaluations: %w", err)
	}
	dims, err := s.evals.ListDimensions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list dimensions: %w", err)
	}
	return summarizeEvaluations(id, dto.Person{ID: row.ApplicantID.String(), Name: row.ApplicantName}, evals, dims), nil
}

func (s *interviewService) UpdateEvaluation(ctx context.Context, evaluator, interviewID, eid uuid.UUID, req *dto.UpdateEvaluationRequest) (*dto.EvaluationResponse, error) {
	ev, err := s.evals.GetByID(ctx, eid)
	if err != nil {
		return nil, fmt.Errorf("get evaluation: %w", err)
	}
	if ev == nil || ev.InterviewID != interviewID {
		return nil, response.NewError(response.CodeNotFound, "评分不存在")
	}
	if ev.InterviewerID != evaluator {
		return nil, response.NewError(response.CodeForbidden, "只能修改自己的评分")
	}
	iv, err := s.mustInterview(ctx, interviewID)
	if err != nil {
		return nil, err
	}
	if !canScore(iv.Status) {
		return nil, response.NewError(response.CodeInterviewInvalidState, "当前状态不允许改分")
	}
	if req.Score != nil {
		dim, err := s.evals.GetDimensionByName(ctx, ev.Dimension)
		if err != nil {
			return nil, fmt.Errorf("get dimension: %w", err)
		}
		if dim != nil && (*req.Score < 0 || *req.Score > dim.MaxScore) {
			return nil, response.NewError(response.CodeInterviewScoreRange, "评分超出范围")
		}
		ev.Score = *req.Score
	}
	if req.Comment != nil {
		ev.Comment = *req.Comment
	}
	ev.UpdatedAt = time.Now()
	if err := s.evals.Update(ctx, ev); err != nil {
		return nil, fmt.Errorf("update evaluation: %w", err)
	}
	_ = s.persistWeightedScore(ctx, iv)
	return &dto.EvaluationResponse{
		ID: ev.ID.String(), InterviewID: ev.InterviewID.String(),
		Evaluator: dto.Person{ID: ev.InterviewerID.String()},
		Dimension: ev.Dimension, Score: ev.Score, Comment: ev.Comment,
		CreatedAt: ev.CreatedAt, UpdatedAt: ev.UpdatedAt,
	}, nil
}

func (s *interviewService) SubmitResult(ctx context.Context, operator, id uuid.UUID, req *dto.SubmitResultRequest) (*dto.InterviewResponse, error) {
	iv, err := s.mustInterview(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canSubmitResult(iv.Status) {
		return nil, response.NewError(response.CodeInterviewInvalidState, "当前状态不允许提交结果")
	}
	if err := s.refreshScore(ctx, iv); err != nil {
		return nil, err
	}
	now := time.Now()
	iv.ResultCode = req.Result
	iv.ResultComment = req.Comment
	iv.Result = resultLabel(req.Result)
	if iv.Status == model.InterviewOngoing {
		iv.Status = model.InterviewDone
		iv.ActualEndTime = &now
	}
	iv.UpdatedAt = now
	if err := s.records.Update(ctx, iv); err != nil {
		return nil, fmt.Errorf("submit result: %w", err)
	}
	s.notifyResult(ctx, iv)
	if iv.ApplicationID != nil && s.syncer != nil && req.Result != model.ResultPending {
		if err := s.syncer.SyncFromInterview(ctx, operator, *iv.ApplicationID, req.Result, req.Comment); err != nil {
			return nil, err
		}
	}
	return s.GetInterview(ctx, id, nil)
}

func (s *interviewService) ensureAssigned(ctx context.Context, interviewID, evaluator uuid.UUID) error {
	rows, err := s.records.ListInterviewers(ctx, []uuid.UUID{interviewID})
	if err != nil {
		return fmt.Errorf("list interviewers: %w", err)
	}
	if len(rows) == 0 {
		return response.NewError(response.CodeInterviewNoEvaluator, "面试官未分配")
	}
	for _, r := range rows {
		if r.InterviewerID == evaluator {
			return nil
		}
	}
	return response.NewError(response.CodeForbidden, "你不是本场面试官")
}

func buildEvalRows(interviewID, evaluator uuid.UUID, items []dto.EvaluationItem, dims []model.Dimension) ([]model.Evaluation, error) {
	byName := map[string]model.Dimension{}
	for _, d := range dims {
		byName[d.Name] = d
	}
	now := time.Now()
	seen := map[string]struct{}{}
	rows := make([]model.Evaluation, 0, len(items))
	for _, item := range items {
		if _, dup := seen[item.Dimension]; dup {
			return nil, response.NewError(response.CodeInterviewDupEval, "重复评分")
		}
		seen[item.Dimension] = struct{}{}
		dim, ok := byName[item.Dimension]
		if !ok {
			return nil, response.NewError(response.CodeInterviewDimGone, "维度不存在")
		}
		if item.Score < 0 || item.Score > dim.MaxScore {
			return nil, response.NewError(response.CodeInterviewScoreRange, "评分超出范围")
		}
		rows = append(rows, model.Evaluation{
			ID: uuid.New(), InterviewID: interviewID, InterviewerID: evaluator,
			Dimension: item.Dimension, Score: item.Score, Comment: item.Comment,
			Recommendation: 3, CreatedAt: now, UpdatedAt: now,
		})
	}
	return rows, nil
}

func (s *interviewService) persistWeightedScore(ctx context.Context, iv *model.Interview) error {
	if err := s.refreshScore(ctx, iv); err != nil {
		return err
	}
	iv.UpdatedAt = time.Now()
	return s.records.Update(ctx, iv)
}
