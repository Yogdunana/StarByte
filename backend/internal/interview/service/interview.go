package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/internal/interview/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *interviewService) CreateInterview(ctx context.Context, operator uuid.UUID, req *dto.CreateInterviewRequest) (*dto.InterviewResponse, error) {
	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		return nil, response.NewError(response.CodeBadRequest, "无效的场次ID")
	}
	sess, err := s.mustSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if !canAddCandidate(sess.Status) {
		return nil, response.NewError(response.CodeInterviewInvalidState, "当前场次不可添加面试者")
	}
	n, err := s.sessions.CountCandidates(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("count candidates: %w", err)
	}
	if int(n) >= sess.MaxCandidates {
		return nil, response.NewError(response.CodeInterviewSessionFull, "超过最大候选人数")
	}
	applicantID, appID, err := s.resolveApplicant(ctx, req)
	if err != nil {
		return nil, err
	}
	exist, err := s.records.FindBySessionApplicant(ctx, sessionID, applicantID)
	if err != nil {
		return nil, fmt.Errorf("find interview: %w", err)
	}
	if exist != nil {
		return nil, response.NewError(response.CodeConflict, "该面试者已在本场次中")
	}
	now := time.Now()
	iv := &model.Interview{
		ID:            uuid.New(),
		SessionID:     &sessionID,
		ApplicationID: appID,
		ApplicantID:   applicantID,
		Round:         sess.Round,
		Type:          1,
		Status:        model.InterviewPending,
		ScheduledAt:   req.ScheduledTime,
		Location:      firstNonEmpty(req.Location, sess.Location),
		Duration:      req.Duration,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.records.Create(ctx, iv); err != nil {
		return nil, fmt.Errorf("create interview: %w", err)
	}
	s.notifyInvite(ctx, applicantID, sess, iv)
	_ = operator
	return s.GetInterview(ctx, iv.ID, nil)
}

func (s *interviewService) resolveApplicant(ctx context.Context, req *dto.CreateInterviewRequest) (uuid.UUID, *uuid.UUID, error) {
	if req.ApplicationID != "" {
		appID, err := uuid.Parse(req.ApplicationID)
		if err != nil {
			return uuid.Nil, nil, response.NewError(response.CodeBadRequest, "无效的申请ID")
		}
		app, err := s.records.GetApplication(ctx, appID)
		if err != nil {
			return uuid.Nil, nil, fmt.Errorf("get application: %w", err)
		}
		if app == nil {
			return uuid.Nil, nil, response.NewError(response.CodeMemberAppNotFound, "申请不存在")
		}
		return app.UserID, &appID, nil
	}
	if req.ApplicantID == "" {
		return uuid.Nil, nil, response.NewError(response.CodeBadRequest, "请指定面试者或入会申请")
	}
	uid, err := uuid.Parse(req.ApplicantID)
	if err != nil {
		return uuid.Nil, nil, response.NewError(response.CodeBadRequest, "无效的面试者ID")
	}
	u, err := s.records.GetUser(ctx, uid)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("get user: %w", err)
	}
	if u == nil {
		return uuid.Nil, nil, response.NewError(response.CodeUserNotFound, "用户不存在")
	}
	return uid, nil, nil
}

func (s *interviewService) ListInterviews(ctx context.Context, viewer uuid.UUID, req *dto.ListInterviewRequest, scope *rbacModel.DataScopeCondition) ([]*dto.InterviewResponse, int64, error) {
	rows, total, err := s.records.List(ctx, req, rewriteInterviewScope(scope, viewer))
	if err != nil {
		return nil, 0, fmt.Errorf("list interviews: %w", err)
	}
	out, err := s.attachEvaluators(ctx, rows)
	return out, total, err
}

func (s *interviewService) GetInterview(ctx context.Context, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.InterviewResponse, error) {
	row, err := s.records.GetByIDWithNames(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get interview: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeInterviewRecordGone, "面试记录不存在")
	}
	_ = scope
	evals, err := s.records.ListInterviewers(ctx, []uuid.UUID{row.ID})
	if err != nil {
		return nil, fmt.Errorf("list interviewers: %w", err)
	}
	return mapInterview(row, groupEvaluators(evals)[row.ID]), nil
}

func (s *interviewService) MyInterviews(ctx context.Context, userID uuid.UUID, status *int16) ([]*dto.InterviewResponse, error) {
	rows, err := s.records.ListMine(ctx, userID, status)
	if err != nil {
		return nil, fmt.Errorf("list my interviews: %w", err)
	}
	return s.attachEvaluators(ctx, rows)
}

func (s *interviewService) attachEvaluators(ctx context.Context, rows []model.InterviewWithNames) ([]*dto.InterviewResponse, error) {
	ids := make([]uuid.UUID, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	named, err := s.records.ListInterviewers(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("list interviewers: %w", err)
	}
	grouped := groupEvaluators(named)
	out := make([]*dto.InterviewResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapInterview(&rows[i], grouped[rows[i].ID]))
	}
	return out, nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
