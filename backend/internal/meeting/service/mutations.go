package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
)

func (s *meetingService) CreateMeeting(ctx context.Context, operator uuid.UUID, req *dto.CreateMeetingRequest) (*dto.MeetingResponse, error) {
	return meetingMutation(ctx, s, uuid.Nil, func(bound *meetingService) (*dto.MeetingResponse, error) {
		return bound.createMeeting(ctx, operator, req)
	})
}
func (s *meetingService) UpdateMeeting(ctx context.Context, id uuid.UUID, req *dto.UpdateMeetingRequest) (*dto.MeetingResponse, error) {
	return meetingMutation(ctx, s, id, func(bound *meetingService) (*dto.MeetingResponse, error) { return bound.updateMeeting(ctx, id, req) })
}
func (s *meetingService) DeleteMeeting(ctx context.Context, id uuid.UUID) error {
	return s.meetingTransaction(ctx, id, func(bound *meetingService) error { return bound.deleteMeeting(ctx, id) })
}
func (s *meetingService) StartMeeting(ctx context.Context, id uuid.UUID) (*dto.MeetingResponse, error) {
	return meetingMutation(ctx, s, id, func(bound *meetingService) (*dto.MeetingResponse, error) { return bound.startMeeting(ctx, id) })
}
func (s *meetingService) EndMeeting(ctx context.Context, id uuid.UUID) (*dto.MeetingResponse, error) {
	return meetingMutation(ctx, s, id, func(bound *meetingService) (*dto.MeetingResponse, error) { return bound.endMeeting(ctx, id) })
}
func (s *meetingService) CancelMeeting(ctx context.Context, id uuid.UUID, reason string) (*dto.MeetingResponse, error) {
	return meetingMutation(ctx, s, id, func(bound *meetingService) (*dto.MeetingResponse, error) { return bound.cancelMeeting(ctx, id, reason) })
}
func (s *meetingService) UpdateMinutes(ctx context.Context, id uuid.UUID, minutes string) (*dto.MeetingResponse, error) {
	return meetingMutation(ctx, s, id, func(bound *meetingService) (*dto.MeetingResponse, error) {
		return bound.updateMinutes(ctx, id, minutes)
	})
}
func (s *meetingService) AddAttendees(ctx context.Context, meetingID uuid.UUID, userIDs []uuid.UUID) ([]dto.AttendeeResponse, error) {
	return meetingMutation(ctx, s, meetingID, func(bound *meetingService) ([]dto.AttendeeResponse, error) {
		return bound.addAttendees(ctx, meetingID, userIDs)
	})
}
func (s *meetingService) RemoveAttendee(ctx context.Context, meetingID, userID uuid.UUID) error {
	return s.meetingTransaction(ctx, meetingID, func(bound *meetingService) error { return bound.removeAttendee(ctx, meetingID, userID) })
}
func (s *meetingService) Checkin(ctx context.Context, meetingID, userID uuid.UUID, token string) (*dto.AttendeeResponse, error) {
	return meetingMutation(ctx, s, meetingID, func(bound *meetingService) (*dto.AttendeeResponse, error) {
		return bound.checkin(ctx, meetingID, userID, token)
	})
}
func (s *meetingService) AddAgenda(ctx context.Context, meetingID uuid.UUID, req *dto.CreateAgendaRequest) (*dto.AgendaResponse, error) {
	return meetingMutation(ctx, s, meetingID, func(bound *meetingService) (*dto.AgendaResponse, error) { return bound.addAgenda(ctx, meetingID, req) })
}
func (s *meetingService) UpdateAgenda(ctx context.Context, meetingID, agendaID uuid.UUID, req *dto.UpdateAgendaRequest) (*dto.AgendaResponse, error) {
	return meetingMutation(ctx, s, meetingID, func(bound *meetingService) (*dto.AgendaResponse, error) {
		return bound.updateAgenda(ctx, meetingID, agendaID, req)
	})
}
func (s *meetingService) DeleteAgenda(ctx context.Context, meetingID, agendaID uuid.UUID) error {
	return s.meetingTransaction(ctx, meetingID, func(bound *meetingService) error { return bound.deleteAgenda(ctx, meetingID, agendaID) })
}
func (s *meetingService) SortAgendas(ctx context.Context, meetingID uuid.UUID, ids []uuid.UUID) ([]dto.AgendaResponse, error) {
	return meetingMutation(ctx, s, meetingID, func(bound *meetingService) ([]dto.AgendaResponse, error) {
		return bound.sortAgendas(ctx, meetingID, ids)
	})
}
func (s *meetingService) CreateVote(ctx context.Context, meetingID uuid.UUID, req *dto.CreateVoteRequest) (*dto.VoteResponse, error) {
	return meetingMutation(ctx, s, meetingID, func(bound *meetingService) (*dto.VoteResponse, error) { return bound.createVote(ctx, meetingID, req) })
}

func (s *meetingService) MeetingQRCode(ctx context.Context, id uuid.UUID) (*dto.QRCodeResponse, []byte, error) {
	var result *dto.QRCodeResponse
	var png []byte
	err := s.meetingTransaction(ctx, id, func(bound *meetingService) error {
		var err error
		result, png, err = bound.meetingQRCode(ctx, id)
		return err
	})
	return result, png, err
}
