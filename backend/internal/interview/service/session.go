package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/internal/interview/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	qrcode "github.com/skip2/go-qrcode"
)

func (s *interviewService) CreateSession(ctx context.Context, operator uuid.UUID, req *dto.CreateSessionRequest) (*dto.SessionResponse, error) {
	if !req.EndTime.After(req.StartTime) {
		return nil, response.NewError(response.CodeBadRequest, "结束时间必须晚于开始时间")
	}
	now := time.Now()
	sess := &model.Session{
		ID:            uuid.New(),
		Title:         req.Title,
		Round:         int16(req.Round),
		DepartmentID:  parseOptionalUUID(req.DepartmentID),
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		Location:      req.Location,
		OnlineLink:    req.OnlineLink,
		Status:        model.SessionPending,
		MaxCandidates: req.MaxCandidates,
		Description:   req.Description,
		CreatedBy:     &operator,
		QRToken:       newQRToken(),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.sessions.Create(ctx, sess); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return s.GetSession(ctx, sess.ID, nil)
}

func (s *interviewService) ListSessions(ctx context.Context, viewer uuid.UUID, req *dto.ListSessionRequest, scope *rbacModel.DataScopeCondition) ([]*dto.SessionResponse, int64, error) {
	rows, total, err := s.sessions.List(ctx, req, rewriteSessionScope(scope, viewer))
	if err != nil {
		return nil, 0, fmt.Errorf("list sessions: %w", err)
	}
	out := make([]*dto.SessionResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapSession(&rows[i]))
	}
	return out, total, nil
}

func (s *interviewService) GetSession(ctx context.Context, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.SessionResponse, error) {
	row, err := s.sessions.GetByIDWithNames(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeInterviewNotFound, "面试场次不存在")
	}
	_ = scope
	return mapSession(row), nil
}

func (s *interviewService) UpdateSession(ctx context.Context, id uuid.UUID, req *dto.UpdateSessionRequest) (*dto.SessionResponse, error) {
	sess, err := s.mustSession(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canUpdateSession(sess.Status) {
		return nil, response.NewError(response.CodeInterviewInvalidState, "仅待开始场次可修改")
	}
	applySessionPatch(sess, req)
	if !sess.EndTime.After(sess.StartTime) {
		return nil, response.NewError(response.CodeBadRequest, "结束时间必须晚于开始时间")
	}
	sess.UpdatedAt = time.Now()
	if err := s.sessions.Update(ctx, sess); err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}
	return s.GetSession(ctx, id, nil)
}

func (s *interviewService) DeleteSession(ctx context.Context, id uuid.UUID) error {
	sess, err := s.mustSession(ctx, id)
	if err != nil {
		return err
	}
	if !canDeleteSession(sess.Status) {
		return response.NewError(response.CodeInterviewInvalidState, "进行中或已结束的场次不可删除")
	}
	n, err := s.sessions.CountCandidates(ctx, id)
	if err != nil {
		return fmt.Errorf("count candidates: %w", err)
	}
	if n > 0 {
		sess.Status = model.SessionCancelled
		sess.UpdatedAt = time.Now()
		return s.sessions.Update(ctx, sess)
	}
	return s.sessions.Delete(ctx, id)
}

func (s *interviewService) StartSession(ctx context.Context, id uuid.UUID) (*dto.SessionResponse, error) {
	sess, err := s.mustSession(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canStartSession(sess.Status) {
		return nil, response.NewError(response.CodeInterviewInvalidState, "场次状态不允许开始")
	}
	sess.Status = model.SessionOngoing
	if sess.QRToken == "" {
		sess.QRToken = newQRToken()
	}
	sess.UpdatedAt = time.Now()
	if err := s.sessions.Update(ctx, sess); err != nil {
		return nil, fmt.Errorf("start session: %w", err)
	}
	return s.GetSession(ctx, id, nil)
}

func (s *interviewService) EndSession(ctx context.Context, id uuid.UUID) (*dto.SessionResponse, error) {
	sess, err := s.mustSession(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canEndSession(sess.Status) {
		return nil, response.NewError(response.CodeInterviewInvalidState, "场次状态不允许结束")
	}
	if err := s.records.MarkAbsentBySession(ctx, id); err != nil {
		return nil, fmt.Errorf("mark absent: %w", err)
	}
	sess.Status = model.SessionEnded
	sess.UpdatedAt = time.Now()
	if err := s.sessions.Update(ctx, sess); err != nil {
		return nil, fmt.Errorf("end session: %w", err)
	}
	return s.GetSession(ctx, id, nil)
}

func (s *interviewService) SessionQRCode(ctx context.Context, id uuid.UUID) (*dto.QRCodeResponse, []byte, error) {
	sess, err := s.mustSession(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if sess.QRToken == "" {
		sess.QRToken = newQRToken()
		sess.UpdatedAt = time.Now()
		if err := s.sessions.Update(ctx, sess); err != nil {
			return nil, nil, fmt.Errorf("save qr token: %w", err)
		}
	}
	path := "/interview/checkin?session_id=" + id.String() + "&token=" + sess.QRToken
	png, err := qrcode.Encode(path, qrcode.Medium, 256)
	if err != nil {
		return nil, nil, fmt.Errorf("encode qr: %w", err)
	}
	return &dto.QRCodeResponse{
		SessionID:   id.String(),
		Token:       sess.QRToken,
		CheckinPath: path,
		PNGBase64:   base64.StdEncoding.EncodeToString(png),
	}, png, nil
}

func (s *interviewService) mustSession(ctx context.Context, id uuid.UUID) (*model.Session, error) {
	sess, err := s.sessions.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	if sess == nil {
		return nil, response.NewError(response.CodeInterviewNotFound, "面试场次不存在")
	}
	return sess, nil
}

func applySessionPatch(sess *model.Session, req *dto.UpdateSessionRequest) {
	if req.Title != nil {
		sess.Title = *req.Title
	}
	if req.Round != nil {
		sess.Round = int16(*req.Round)
	}
	if req.DepartmentID != nil {
		sess.DepartmentID = parseOptionalUUID(*req.DepartmentID)
	}
	if req.StartTime != nil {
		sess.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		sess.EndTime = *req.EndTime
	}
	if req.Location != nil {
		sess.Location = *req.Location
	}
	if req.OnlineLink != nil {
		sess.OnlineLink = *req.OnlineLink
	}
	if req.MaxCandidates != nil {
		sess.MaxCandidates = *req.MaxCandidates
	}
	if req.Description != nil {
		sess.Description = *req.Description
	}
}

func newQRToken() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return uuid.NewString()
	}
	return hex.EncodeToString(buf)
}
