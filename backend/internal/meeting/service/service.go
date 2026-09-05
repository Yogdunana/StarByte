package service

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/internal/meeting/repo"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
)

type Notifier interface {
	Send(ctx context.Context, userIDs []uuid.UUID, template string, vars map[string]interface{}) error
}

type MeetingService interface {
	CreateMeeting(ctx context.Context, operator uuid.UUID, req *dto.CreateMeetingRequest) (*dto.MeetingResponse, error)
	ListMeetings(ctx context.Context, viewer uuid.UUID, req *dto.ListMeetingRequest, scope *rbacModel.DataScopeCondition) ([]*dto.MeetingResponse, int64, error)
	GetMeeting(ctx context.Context, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.MeetingResponse, error)
	UpdateMeeting(ctx context.Context, id uuid.UUID, req *dto.UpdateMeetingRequest) (*dto.MeetingResponse, error)
	DeleteMeeting(ctx context.Context, id uuid.UUID) error
	StartMeeting(ctx context.Context, id uuid.UUID) (*dto.MeetingResponse, error)
	EndMeeting(ctx context.Context, id uuid.UUID) (*dto.MeetingResponse, error)
	CancelMeeting(ctx context.Context, id uuid.UUID, reason string) (*dto.MeetingResponse, error)
	UpdateMinutes(ctx context.Context, id uuid.UUID, minutes string) (*dto.MeetingResponse, error)
	MeetingQRCode(ctx context.Context, id uuid.UUID) (*dto.QRCodeResponse, []byte, error)

	AddAgenda(ctx context.Context, meetingID uuid.UUID, req *dto.CreateAgendaRequest) (*dto.AgendaResponse, error)
	UpdateAgenda(ctx context.Context, meetingID, agendaID uuid.UUID, req *dto.UpdateAgendaRequest) (*dto.AgendaResponse, error)
	DeleteAgenda(ctx context.Context, meetingID, agendaID uuid.UUID) error
	SortAgendas(ctx context.Context, meetingID uuid.UUID, ids []uuid.UUID) ([]dto.AgendaResponse, error)
	ListAgendas(ctx context.Context, meetingID uuid.UUID) ([]dto.AgendaResponse, error)

	ListAttendees(ctx context.Context, meetingID uuid.UUID) ([]dto.AttendeeResponse, error)
	AddAttendees(ctx context.Context, meetingID uuid.UUID, userIDs []uuid.UUID) ([]dto.AttendeeResponse, error)
	RemoveAttendee(ctx context.Context, meetingID, userID uuid.UUID) error
	Checkin(ctx context.Context, meetingID, userID uuid.UUID, token string) (*dto.AttendeeResponse, error)

	CreateVote(ctx context.Context, meetingID uuid.UUID, req *dto.CreateVoteRequest) (*dto.VoteResponse, error)
	ListVotes(ctx context.Context, meetingID, viewer uuid.UUID) ([]*dto.VoteResponse, error)
	GetVote(ctx context.Context, id, viewer uuid.UUID) (*dto.VoteResponse, error)
	CastVote(ctx context.Context, voteID, userID uuid.UUID, optionKey string) error
	VoteResult(ctx context.Context, id uuid.UUID) (*dto.VoteResultResponse, error)
	CloseVote(ctx context.Context, id uuid.UUID) (*dto.VoteResponse, error)
	MyVote(ctx context.Context, voteID, userID uuid.UUID) (*dto.MyVoteResponse, error)

	GetWeightConfig(ctx context.Context) (*dto.VoteWeightConfigResponse, error)
	UpdateWeightConfig(ctx context.Context, operator uuid.UUID, req *dto.VoteWeightConfigRequest) (*dto.VoteWeightConfigResponse, error)
}

type meetingService struct {
	meetings  repo.MeetingRepo
	agendas   repo.AgendaRepo
	attendees repo.AttendeeRepo
	votes     repo.VoteRepo
	notify    Notifier
}

func NewMeetingService(
	meetings repo.MeetingRepo,
	agendas repo.AgendaRepo,
	attendees repo.AttendeeRepo,
	votes repo.VoteRepo,
	notify Notifier,
) MeetingService {
	return &meetingService{
		meetings: meetings, agendas: agendas, attendees: attendees, votes: votes, notify: notify,
	}
}
