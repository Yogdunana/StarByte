package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	qrcode "github.com/skip2/go-qrcode"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	"github.com/Yogdunana/StarByte/backend/internal/meeting/repo"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *meetingService) createMeeting(ctx context.Context, operator uuid.UUID, req *dto.CreateMeetingRequest) (*dto.MeetingResponse, error) {
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
	if err := validateMeeting(m); err != nil {
		return nil, err
	}
	for _, raw := range req.UserIDs {
		if _, err := uuid.Parse(raw); err != nil {
			return nil, response.NewError(response.CodeBadRequest, "参会人 ID 无效")
		}
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
	filter := rewriteMeetingScope(scope, viewer)
	if s.access != nil {
		filter = repo.MeetingScope(model.ViewerFromContext(ctx))
	}
	rows, total, err := s.meetings.List(ctx, req, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("list meetings: %w", err)
	}
	out := make([]*dto.MeetingResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapMeeting(&rows[i]))
	}
	if err := s.meetingCapabilities(ctx, out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (s *meetingService) GetMeeting(ctx context.Context, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.MeetingResponse, error) {
	if err := s.requireMeetingAccess(ctx, id); err != nil {
		return nil, err
	}
	row, err := s.meetings.GetByIDWithNames(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get meeting: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeMeetingNotFound, "会议不存在")
	}
	_ = scope
	out := mapMeeting(row)
	if err := s.meetingCapabilities(ctx, []*dto.MeetingResponse{out}); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *meetingService) updateMeeting(ctx context.Context, id uuid.UUID, req *dto.UpdateMeetingRequest) (*dto.MeetingResponse, error) {
	m, err := s.mustMeeting(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canUpdateMeeting(m.Status) {
		return nil, response.NewError(response.CodeMeetingInvalidState, "仅待开始会议可修改")
	}
	applyMeetingPatch(m, req)
	if err := validateMeeting(m); err != nil {
		return nil, err
	}
	if !m.EndTime.After(m.StartTime) {
		return nil, response.NewError(response.CodeBadRequest, "结束时间必须晚于开始时间")
	}
	m.UpdatedAt = time.Now()
	if err := s.meetings.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("update meeting: %w", err)
	}
	return s.GetMeeting(ctx, id, nil)
}

func (s *meetingService) deleteMeeting(ctx context.Context, id uuid.UUID) error {
	m, err := s.mustMeeting(ctx, id)
	if err != nil {
		return err
	}
	if !canDeleteMeeting(m.Status) {
		return response.NewError(response.CodeMeetingInvalidState, "进行中或已结束的会议不可删除")
	}
	votes, err := s.votes.ListByMeeting(ctx, id)
	if err != nil {
		return fmt.Errorf("check meeting vote history: %w", err)
	}
	if len(votes) > 0 {
		return response.NewError(response.CodeMeetingInvalidState, "会议含投票记录，请保留历史，不可删除")
	}
	if err := s.meetings.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete meeting: %w", err)
	}
	return nil
}

func (s *meetingService) startMeeting(ctx context.Context, id uuid.UUID) (*dto.MeetingResponse, error) {
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
	ids, err := s.attendeeIDs(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list notification recipients: %w", err)
	}
	s.notifyMeeting(ctx, ids, tplMeetingStarted, m)
	return s.GetMeeting(ctx, id, nil)
}

func (s *meetingService) endMeeting(ctx context.Context, id uuid.UUID) (*dto.MeetingResponse, error) {
	m, err := s.mustMeeting(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canEndMeeting(m.Status) {
		return nil, response.NewError(response.CodeMeetingInvalidState, "仅进行中会议可结束")
	}
	if err := s.closeMeetingVotes(ctx, id); err != nil {
		return nil, err
	}
	m.Status = model.MeetingEnded
	m.UpdatedAt = time.Now()
	if err := s.meetings.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("end meeting: %w", err)
	}
	ids, err := s.attendeeIDs(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list notification recipients: %w", err)
	}
	s.notifyMeeting(ctx, ids, tplMeetingEnded, m)
	return s.GetMeeting(ctx, id, nil)
}

func (s *meetingService) cancelMeeting(ctx context.Context, id uuid.UUID, reason string) (*dto.MeetingResponse, error) {
	m, err := s.mustMeeting(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canCancelMeeting(m.Status) {
		return nil, response.NewError(response.CodeMeetingInvalidState, "当前状态不可取消")
	}
	if err := s.closeMeetingVotes(ctx, id); err != nil {
		return nil, err
	}
	m.Status = model.MeetingCancelled
	m.CancelReason = reason
	m.UpdatedAt = time.Now()
	if err := s.meetings.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("cancel meeting: %w", err)
	}
	return s.GetMeeting(ctx, id, nil)
}

func (s *meetingService) updateMinutes(ctx context.Context, id uuid.UUID, minutes string) (*dto.MeetingResponse, error) {
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

func (s *meetingService) meetingQRCode(ctx context.Context, id uuid.UUID) (*dto.QRCodeResponse, []byte, error) {
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
	if err := s.requireMeetingAccess(ctx, id); err != nil {
		return nil, err
	}
	m, err := s.meetings.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get meeting: %w", err)
	}
	if m == nil {
		return nil, response.NewError(response.CodeMeetingNotFound, "会议不存在")
	}
	return m, nil
}
