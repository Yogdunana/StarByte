package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func requireAppError(t *testing.T, err error, code int) {
	t.Helper()
	require.Error(t, err)
	var app *response.AppError
	require.ErrorAs(t, err, &app)
	require.Equal(t, code, app.Code)
}

func seedMeeting(t *testing.T, svc *meetingService, mm *memMeetings, status int16) (*model.Meeting, uuid.UUID) {
	t.Helper()
	org := uuid.New()
	start := time.Now().Add(time.Hour)
	out, err := svc.CreateMeeting(context.Background(), org, &dto.CreateMeetingRequest{
		Title: "例会", StartTime: start, EndTime: start.Add(time.Hour), MeetingType: 1, Location: "A101",
	})
	require.NoError(t, err)
	id := uuid.MustParse(out.ID)
	m, _ := mm.GetByID(context.Background(), id)
	m.Status = status
	_ = mm.Update(context.Background(), m)
	return m, org
}

func TestCreateAndStartMeeting(t *testing.T) {
	svc, mm, _, _ := newTestSvc()
	start := time.Now().Add(time.Hour)
	out, err := svc.CreateMeeting(context.Background(), uuid.New(), &dto.CreateMeetingRequest{
		Title: "秋招例会", StartTime: start, EndTime: start.Add(2 * time.Hour), MeetingType: 1,
	})
	require.NoError(t, err)
	require.Equal(t, model.MeetingPending, out.Status)
	started, err := svc.StartMeeting(context.Background(), uuid.MustParse(out.ID))
	require.NoError(t, err)
	require.Equal(t, model.MeetingOngoing, started.Status)
	_, err = svc.StartMeeting(context.Background(), uuid.MustParse(out.ID))
	requireAppError(t, err, response.CodeMeetingInvalidState)
	_ = mm
}

func TestCheckinRules(t *testing.T) {
	svc, mm, _, _ := newTestSvc()
	m, org := seedMeeting(t, svc, mm, model.MeetingOngoing)
	_, err := svc.Checkin(context.Background(), m.ID, uuid.New(), "")
	requireAppError(t, err, response.CodeMeetingNotAttendee)
	first, err := svc.Checkin(context.Background(), m.ID, org, "")
	require.NoError(t, err)
	require.True(t, first.Attended)
	_, err = svc.Checkin(context.Background(), m.ID, org, "")
	requireAppError(t, err, response.CodeMeetingDupCheckin)
}

func TestCastVoteRules(t *testing.T) {
	svc, mm, _, vv := newTestSvc()
	m, org := seedMeeting(t, svc, mm, model.MeetingOngoing)
	guest := uuid.New()
	vote, err := svc.CreateVote(context.Background(), m.ID, &dto.CreateVoteRequest{
		Title: "方向", VoteType: model.VoteEqual,
		Options: []dto.VoteOptionInput{{Key: "web", Label: "Web"}, {Key: "ai", Label: "AI"}},
	})
	require.NoError(t, err)
	vid := uuid.MustParse(vote.ID)
	err = svc.CastVote(context.Background(), vid, guest, "web")
	requireAppError(t, err, response.CodeVoteNoAccess)
	err = svc.CastVote(context.Background(), vid, org, "none")
	requireAppError(t, err, response.CodeVoteOptionGone)
	require.NoError(t, svc.CastVote(context.Background(), vid, org, "web"))
	err = svc.CastVote(context.Background(), vid, org, "ai")
	requireAppError(t, err, response.CodeVoteDuplicate)
	_, err = svc.MyVote(context.Background(), vid, org)
	require.NoError(t, err)
	res, err := svc.VoteResult(context.Background(), vid)
	require.NoError(t, err)
	require.Equal(t, 1, res.TotalVoters)
	require.Equal(t, 1.0, res.TotalWeight)
	_ = vv
}

func TestAnonymousMyVoteHidden(t *testing.T) {
	svc, mm, _, _ := newTestSvc()
	m, org := seedMeeting(t, svc, mm, model.MeetingOngoing)
	vote, err := svc.CreateVote(context.Background(), m.ID, &dto.CreateVoteRequest{
		Title: "匿名", VoteType: 1, IsAnonymous: true,
		Options: []dto.VoteOptionInput{{Key: "a", Label: "A"}, {Key: "b", Label: "B"}},
	})
	require.NoError(t, err)
	vid := uuid.MustParse(vote.ID)
	require.NoError(t, svc.CastVote(context.Background(), vid, org, "a"))
	_, err = svc.MyVote(context.Background(), vid, org)
	requireAppError(t, err, response.CodeVoteAnonymousHidden)
}

func TestWeightedVoteUsesPosition(t *testing.T) {
	svc, mm, _, _ := newTestSvc()
	m, org := seedMeeting(t, svc, mm, model.MeetingOngoing)
	mm.users[org] = &model.NamedUser{ID: org, RealName: "社长", PositionCode: "president"}
	vote, err := svc.CreateVote(context.Background(), m.ID, &dto.CreateVoteRequest{
		Title: "加权", VoteType: model.VoteWeighted,
		Options: []dto.VoteOptionInput{{Key: "yes", Label: "同意"}, {Key: "no", Label: "反对"}},
	})
	require.NoError(t, err)
	vid := uuid.MustParse(vote.ID)
	require.NoError(t, svc.CastVote(context.Background(), vid, org, "yes"))
	res, err := svc.VoteResult(context.Background(), vid)
	require.NoError(t, err)
	require.Equal(t, 5.0, res.TotalWeight)
}
