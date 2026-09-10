package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func TestMeetingClosureClosesPollsAndPreservesHistory(t *testing.T) {
	for _, cancel := range []bool{false, true} {
		t.Run(map[bool]string{true: "cancel", false: "end"}[cancel], func(t *testing.T) {
			ctx := context.Background()
			s, mm, _, _ := newTestSvc()
			m, owner := seedMeeting(t, s, mm, model.MeetingOngoing)
			v, err := s.CreateVote(ctx, m.ID, &dto.CreateVoteRequest{Title: "议题", VoteType: 1, Options: []dto.VoteOptionInput{{Key: "a", Label: "甲"}, {Key: "b", Label: "乙"}}})
			require.NoError(t, err)
			id := uuid.MustParse(v.ID)
			require.NoError(t, s.CastVote(ctx, id, owner, "a"))
			if cancel {
				_, err = s.CancelMeeting(ctx, m.ID, "调整安排")
			} else {
				_, err = s.EndMeeting(ctx, m.ID)
			}
			require.NoError(t, err)
			result, err := s.VoteResult(ctx, id)
			require.NoError(t, err)
			require.Equal(t, model.VoteClosed, result.Status)
			require.Equal(t, 1, result.TotalVoters)
			requireAppError(t, s.CastVote(ctx, id, uuid.New(), "b"), response.CodeVoteNotOpen)
			requireAppError(t, s.DeleteMeeting(ctx, m.ID), response.CodeMeetingInvalidState)
		})
	}
}
func TestRemoveAttendeePreservesParticipation(t *testing.T) {
	ctx := context.Background()
	s, mm, _, _ := newTestSvc()
	m, owner := seedMeeting(t, s, mm, model.MeetingOngoing)
	requireAppError(t, s.RemoveAttendee(ctx, m.ID, owner), response.CodeMeetingInvalidState)
	person := uuid.New()
	_, err := s.AddAttendees(ctx, m.ID, []uuid.UUID{person})
	require.NoError(t, err)
	_, err = s.Checkin(ctx, m.ID, person, "")
	require.NoError(t, err)
	requireAppError(t, s.RemoveAttendee(ctx, m.ID, person), response.CodeMeetingInvalidState)
}
func TestVoteValidation(t *testing.T) {
	s, mm, _, _ := newTestSvc()
	m, _ := seedMeeting(t, s, mm, model.MeetingOngoing)
	for _, req := range []dto.CreateVoteRequest{
		{Title: " ", Options: []dto.VoteOptionInput{{Key: "a", Label: "甲"}, {Key: "b", Label: "乙"}}},
		{Title: "议题", Duration: -1}, {Title: "议题", Options: []dto.VoteOptionInput{{Key: "a", Label: "甲"}}},
	} {
		_, err := s.CreateVote(context.Background(), m.ID, &req)
		requireAppError(t, err, response.CodeBadRequest)
	}
}
