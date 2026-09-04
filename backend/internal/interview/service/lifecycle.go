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

func (s *interviewService) AssignEvaluators(ctx context.Context, id uuid.UUID, req *dto.AssignEvaluatorsRequest) (*dto.InterviewResponse, error) {
	iv, err := s.mustInterview(ctx, id)
	if err != nil {
		return nil, err
	}
	if iv.Status == model.InterviewCancelled || iv.Status == model.InterviewDone {
		return nil, response.NewError(response.CodeInterviewInvalidState, "当前面试不可分配面试官")
	}
	ids, err := parseEvaluatorIDs(req.EvaluatorIDs)
	if err != nil {
		return nil, err
	}
	if err := s.ensureNoConflicts(ctx, iv, ids); err != nil {
		return nil, err
	}
	now := time.Now()
	rows := make([]model.Interviewer, 0, len(ids))
	for _, uid := range ids {
		rows = append(rows, model.Interviewer{
			ID:            uuid.New(),
			InterviewID:   id,
			InterviewerID: uid,
			CreatedAt:     now,
		})
	}
	if err := s.records.ReplaceInterviewers(ctx, id, rows); err != nil {
		return nil, fmt.Errorf("assign evaluators: %w", err)
	}
	s.notifyAssigned(ctx, ids, iv)
	return s.GetInterview(ctx, id, nil)
}

func (s *interviewService) Checkin(ctx context.Context, userID, id uuid.UUID, token string) (*dto.InterviewResponse, error) {
	iv, err := s.mustInterview(ctx, id)
	if err != nil {
		return nil, err
	}
	if iv.ApplicantID != userID {
		return nil, response.NewError(response.CodeForbidden, "仅本人可签到")
	}
	if !canCheckin(iv.Status) {
		return nil, response.NewError(response.CodeInterviewInvalidState, "当前状态不允许签到")
	}
	if token != "" && iv.SessionID != nil {
		sess, err := s.mustSession(ctx, *iv.SessionID)
		if err != nil {
			return nil, err
		}
		if sess.QRToken != token {
			return nil, response.NewError(response.CodeBadRequest, "签到码无效")
		}
	}
	iv.Status = model.InterviewCheckedIn
	iv.UpdatedAt = time.Now()
	if err := s.records.Update(ctx, iv); err != nil {
		return nil, fmt.Errorf("checkin: %w", err)
	}
	return s.GetInterview(ctx, id, nil)
}

func (s *interviewService) StartInterview(ctx context.Context, operator, id uuid.UUID) (*dto.InterviewResponse, error) {
	iv, err := s.mustInterview(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canStartInterview(iv.Status) {
		return nil, response.NewError(response.CodeInterviewInvalidState, "当前状态不允许开始面试")
	}
	evals, err := s.records.ListInterviewers(ctx, []uuid.UUID{id})
	if err != nil {
		return nil, fmt.Errorf("list interviewers: %w", err)
	}
	if len(evals) == 0 {
		return nil, response.NewError(response.CodeInterviewNoEvaluator, "面试官未分配")
	}
	_ = operator
	now := time.Now()
	iv.Status = model.InterviewOngoing
	iv.ActualStartTime = &now
	iv.UpdatedAt = now
	if err := s.records.Update(ctx, iv); err != nil {
		return nil, fmt.Errorf("start interview: %w", err)
	}
	return s.GetInterview(ctx, id, nil)
}

func (s *interviewService) EndInterview(ctx context.Context, operator, id uuid.UUID) (*dto.InterviewResponse, error) {
	iv, err := s.mustInterview(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canEndInterview(iv.Status) {
		return nil, response.NewError(response.CodeInterviewInvalidState, "当前状态不允许结束面试")
	}
	now := time.Now()
	iv.Status = model.InterviewDone
	iv.ActualEndTime = &now
	iv.UpdatedAt = now
	if err := s.refreshScore(ctx, iv); err != nil {
		return nil, err
	}
	if err := s.records.Update(ctx, iv); err != nil {
		return nil, fmt.Errorf("end interview: %w", err)
	}
	_ = operator
	return s.GetInterview(ctx, id, nil)
}

func (s *interviewService) mustInterview(ctx context.Context, id uuid.UUID) (*model.Interview, error) {
	iv, err := s.records.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get interview: %w", err)
	}
	if iv == nil {
		return nil, response.NewError(response.CodeInterviewRecordGone, "面试记录不存在")
	}
	return iv, nil
}

func (s *interviewService) ensureNoConflicts(ctx context.Context, iv *model.Interview, evaluators []uuid.UUID) error {
	if iv.ScheduledAt == nil {
		return nil
	}
	dur := iv.Duration
	if dur <= 0 {
		dur = model.DefaultDuration
	}
	start := *iv.ScheduledAt
	end := start.Add(time.Duration(dur) * time.Minute)
	for _, uid := range evaluators {
		hit, err := s.records.HasInterviewerConflict(ctx, uid, start, end, iv.ID)
		if err != nil {
			return fmt.Errorf("check conflict: %w", err)
		}
		if hit {
			return response.NewError(response.CodeInterviewConflict, "面试时间冲突")
		}
	}
	return nil
}

func parseEvaluatorIDs(raw []string) ([]uuid.UUID, error) {
	seen := map[uuid.UUID]struct{}{}
	out := make([]uuid.UUID, 0, len(raw))
	for _, item := range raw {
		id, err := uuid.Parse(item)
		if err != nil {
			return nil, response.NewError(response.CodeBadRequest, "无效的面试官ID")
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		u := id
		out = append(out, u)
	}
	if len(out) == 0 {
		return nil, response.NewError(response.CodeBadRequest, "请至少选择一位面试官")
	}
	return out, nil
}

func (s *interviewService) refreshScore(ctx context.Context, iv *model.Interview) error {
	summary, err := s.GetEvaluations(ctx, iv.ID)
	if err != nil {
		return err
	}
	if len(summary.Evaluations) == 0 {
		return nil
	}
	score := summary.WeightedScore
	iv.Score = &score
	return nil
}
