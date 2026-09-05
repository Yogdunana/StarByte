package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	qrcode "github.com/skip2/go-qrcode"
)

func (s *meetingService) CreateMeeting(ctx context.Context, operator uuid.UUID, req *dto.CreateMeetingRequest) (*dto.MeetingResponse, error) {
	if !req.EndTime.After(req.StartTime) {
		return nil, response.NewError(response.CodeBadRequest, "结束时间必须晚于开始时间")
	}
	now := time.Now()
	m := &model.Meeting{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		Status:      model.MeetingPending,
		MeetingType: req.MeetingType,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Location:    req.Location,
		OnlineLink:  req.OnlineLink,
		OrganizerID: operator,
		QRToken:     newQRToken(),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.meetings.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("create meeting: %w", err)
	}
	ids := uniqueUUIDs(append(parseUUIDList(req.UserIDs), operator))
	if err := s.addAttendeeIDs(ctx, m.ID, ids); err != nil {
		return nil, err
	}
	s.notifyMeeting(ctx, ids, tplMeetingNotice, m)
	return s.GetMeeting(ctx, m.ID, nil)
}

func (s *meetingService) ListMeetings(ctx context.Context, viewer uuid.UUID, req *dto.ListMeetingRequest, scope *rbacModel.DataScopeCondition) ([]*dto.MeetingResponse, int64, error) {
	rows, total, err := s.meetings.List(ctx, req, rewriteMeetingScope(scope, viewer))
	if err != nil {
		return nil, 0, fmt.Errorf("list meetings: %w", err)
	}
	out := make([]*dto.MeetingResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapMeeting(&rows[i]))
	}
	return out, total, nil
}

func (s *meetingService) GetMeeting(ctx context.Context, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.MeetingResponse, error) {
	row, err := s.meetings.GetByIDWithNames(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get meeting: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeMeetingNotFound, "会议不存在")
	}
	_ = scope
	return mapMeeting(row), nil
}

func (s *meetingService) UpdateMeeting(ctx context.Context, id uuid.UUID, req *dto.UpdateMeetingRequest) (*dto.MeetingResponse, error) {
	m, err := s.mustMeeting(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canUpdateMeeting(m.Status) {
		return nil, response.NewError(response.CodeMeetingInvalidState, "仅待开始会议可修改")
	}
	applyMeetingPatch(m, req)
	if !m.EndTime.After(m.StartTime) {
		return nil, response.NewError(response.CodeBadRequest, "结束时间必须晚于开始时间")
	}
	m.UpdatedAt = time.Now()
	if err := s.meetings.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("update meeting: %w", err)
	}
	return s.GetMeeting(ctx, id, nil)
}

func (s *meetingService) DeleteMeeting(ctx context.Context, id uuid.UUID) error {
	m, err := s.mustMeeting(ctx, id)
	if err != nil {
		return err
	}
	if !canDeleteMeeting(m.Status) {
		return response.NewError(response.CodeMeetingInvalidState, "进行中或已结束的会议不可删除")
	}
	if err := s.meetings.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete meeting: %w", err)
	}
	return nil
}

func (s *meetingService) StartMeeting(ctx context.Context, id uuid.UUID) (*dto.MeetingResponse, error) {
	m, err := s.mustMeeting(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canStartMeeting(m.Status) {
		return nil, response.NewError(response.CodeMeetingInvalidState, "仅待开始会议可开始")
	}
	m.Status = model.MeetingOngoing
	m.UpdatedAt = time.Now()
	if err := s.meetings.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("start meeting: %w", err)
	}
	ids, _ := s.attendeeIDs(ctx, id)
	s.notifyMeeting(ctx, ids, tplMeetingStarted, m)
	return s.GetMeeting(ctx, id, nil)
}

func (s *meetingService) EndMeeting(ctx context.Context, id uuid.UUID) (*dto.MeetingResponse, error) {
	m, err := s.mustMeeting(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canEndMeeting(m.Status) {
		return nil, response.NewError(response.CodeMeetingInvalidState, "仅进行中会议可结束")
	}
	m.Status = model.MeetingEnded
	m.UpdatedAt = time.Now()
	if err := s.meetings.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("end meeting: %w", err)
	}
	ids, _ := s.attendeeIDs(ctx, id)
	s.notifyMeeting(ctx, ids, tplMeetingEnded, m)
	return s.GetMeeting(ctx, id, nil)
}

func (s *meetingService) CancelMeeting(ctx context.Context, id uuid.UUID, reason string) (*dto.MeetingResponse, error) {
	m, err := s.mustMeeting(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canCancelMeeting(m.Status) {
		return nil, response.NewError(response.CodeMeetingInvalidState, "当前状态不可取消")
	}
	m.Status = model.MeetingCancelled
	m.CancelReason = reason
	m.UpdatedAt = time.Now()
	if err := s.meetings.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("cancel meeting: %w", err)
	}
	return s.GetMeeting(ctx, id, nil)
}

func (s *meetingService) UpdateMinutes(ctx context.Context, id uuid.UUID, minutes string) (*dto.MeetingResponse, error) {
	m, err := s.mustMeeting(ctx, id)
	if err != nil {
		return nil, err
	}
	m.Minutes = minutes
	m.UpdatedAt = time.Now()
	if err := s.meetings.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("update minutes: %w", err)
	}
	return s.GetMeeting(ctx, id, nil)
}

func (s *meetingService) MeetingQRCode(ctx context.Context, id uuid.UUID) (*dto.QRCodeResponse, []byte, error) {
	m, err := s.mustMeeting(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if m.QRToken == "" {
		m.QRToken = newQRToken()
		m.UpdatedAt = time.Now()
		if err := s.meetings.Update(ctx, m); err != nil {
			return nil, nil, fmt.Errorf("save qr token: %w", err)
		}
	}
	path := "/meeting/checkin?meeting_id=" + id.String() + "&token=" + m.QRToken
	png, err := qrcode.Encode(path, qrcode.Medium, 256)
	if err != nil {
		return nil, nil, fmt.Errorf("encode qr: %w", err)
	}
	return &dto.QRCodeResponse{
		MeetingID: id.String(), Token: m.QRToken, CheckinPath: path,
		PNGBase64: base64.StdEncoding.EncodeToString(png),
	}, png, nil
}

func (s *meetingService) mustMeeting(ctx context.Context, id uuid.UUID) (*model.Meeting, error) {
	m, err := s.meetings.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get meeting: %w", err)
	}
	if m == nil {
		return nil, response.NewError(response.CodeMeetingNotFound, "会议不存在")
	}
	return m, nil
}

func applyMeetingPatch(m *model.Meeting, req *dto.UpdateMeetingRequest) {
	if req.Title != nil {
		m.Title = *req.Title
	}
	if req.Description != nil {
		m.Description = *req.Description
	}
	if req.StartTime != nil {
		m.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		m.EndTime = *req.EndTime
	}
	if req.Location != nil {
		m.Location = *req.Location
	}
	if req.OnlineLink != nil {
		m.OnlineLink = *req.OnlineLink
	}
	if req.MeetingType != nil {
		m.MeetingType = *req.MeetingType
	}
}

func rewriteMeetingScope(scope *rbacModel.DataScopeCondition, userID uuid.UUID) *rbacModel.DataScopeCondition {
	if scope == nil || scope.IsEmpty() {
		return scope
	}
	if scope.Query == "1 = 0" {
		return &rbacModel.DataScopeCondition{
			Query: "m.organizer_id = ? OR m.id IN (SELECT meeting_id FROM meeting_attendees WHERE user_id = ?)",
			Args:  []interface{}{userID, userID},
		}
	}
	q := strings.ReplaceAll(scope.Query, "department_id", "u.department_id")
	return &rbacModel.DataScopeCondition{Query: q, Args: scope.Args}
}

func newQRToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString([]byte(uuid.New().String()))
	}
	return hex.EncodeToString(b)
}

func parseUUIDList(raw []string) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(raw))
	for _, s := range raw {
		id, err := uuid.Parse(s)
		if err == nil {
			out = append(out, id)
		}
	}
	return out
}

func uniqueUUIDs(ids []uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]struct{}{}
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
