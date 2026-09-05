package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *meetingService) ListAttendees(ctx context.Context, meetingID uuid.UUID) ([]dto.AttendeeResponse, error) {
	if _, err := s.mustMeeting(ctx, meetingID); err != nil {
		return nil, err
	}
	rows, err := s.attendees.List(ctx, meetingID)
	if err != nil {
		return nil, fmt.Errorf("list attendees: %w", err)
	}
	out := make([]dto.AttendeeResponse, 0, len(rows))
	for _, a := range rows {
		out = append(out, mapAttendee(a))
	}
	return out, nil
}

func (s *meetingService) AddAttendees(ctx context.Context, meetingID uuid.UUID, userIDs []uuid.UUID) ([]dto.AttendeeResponse, error) {
	m, err := s.mustMeeting(ctx, meetingID)
	if err != nil {
		return nil, err
	}
	if m.Status == model.MeetingEnded || m.Status == model.MeetingCancelled {
		return nil, response.NewError(response.CodeMeetingInvalidState, "已结束或已取消的会议不可加人")
	}
	ids := uniqueUUIDs(userIDs)
	if err := s.addAttendeeIDs(ctx, meetingID, ids); err != nil {
		return nil, err
	}
	s.notifyMeeting(ctx, ids, tplMeetingNotice, m)
	return s.ListAttendees(ctx, meetingID)
}

func (s *meetingService) RemoveAttendee(ctx context.Context, meetingID, userID uuid.UUID) error {
	if _, err := s.mustMeeting(ctx, meetingID); err != nil {
		return err
	}
	if err := s.attendees.Remove(ctx, meetingID, userID); err != nil {
		return fmt.Errorf("remove attendee: %w", err)
	}
	return nil
}

func (s *meetingService) Checkin(ctx context.Context, meetingID, userID uuid.UUID, token string) (*dto.AttendeeResponse, error) {
	m, err := s.mustMeeting(ctx, meetingID)
	if err != nil {
		return nil, err
	}
	if !canCheckinMeeting(m.Status) {
		return nil, response.NewError(response.CodeMeetingInvalidState, "当前会议不可签到")
	}
	if token != "" && m.QRToken != "" && token != m.QRToken {
		return nil, response.NewError(response.CodeBadRequest, "签到二维码无效")
	}
	att, err := s.attendees.Get(ctx, meetingID, userID)
	if err != nil {
		return nil, fmt.Errorf("get attendee: %w", err)
	}
	if att == nil {
		return nil, response.NewError(response.CodeMeetingNotAttendee, "非参会人，无权签到")
	}
	if att.Attended {
		return nil, response.NewError(response.CodeMeetingDupCheckin, "重复签到")
	}
	now := time.Now()
	att.Attended = true
	att.CheckedInAt = &now
	if err := s.attendees.Update(ctx, att); err != nil {
		return nil, fmt.Errorf("checkin: %w", err)
	}
	list, err := s.attendees.List(ctx, meetingID)
	if err != nil {
		return nil, err
	}
	for _, row := range list {
		if row.UserID == userID {
			out := mapAttendee(row)
			return &out, nil
		}
	}
	out := mapAttendee(model.AttendeeNamed{Attendee: *att})
	return &out, nil
}

func (s *meetingService) addAttendeeIDs(ctx context.Context, meetingID uuid.UUID, userIDs []uuid.UUID) error {
	now := time.Now()
	items := make([]model.Attendee, 0, len(userIDs))
	for _, uid := range userIDs {
		items = append(items, model.Attendee{
			ID: uuid.New(), MeetingID: meetingID, UserID: uid, CreatedAt: now,
		})
	}
	if err := s.attendees.Add(ctx, items); err != nil {
		return fmt.Errorf("add attendees: %w", err)
	}
	return nil
}

func (s *meetingService) attendeeIDs(ctx context.Context, meetingID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := s.attendees.List(ctx, meetingID)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, a := range rows {
		ids = append(ids, a.UserID)
	}
	return ids, nil
}
