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

func TestVoteAutoCloseOnExpire(t *testing.T) {
	svc, mm, _, _ := newTestSvc()
	m, org := seedMeeting(t, svc, mm, model.MeetingOngoing)
	vote, err := svc.CreateVote(context.Background(), m.ID, &dto.CreateVoteRequest{
		Title: "限时", VoteType: 1, Duration: 1,
		Options: []dto.VoteOptionInput{{Key: "a", Label: "A"}, {Key: "b", Label: "B"}},
	})
	require.NoError(t, err)
	vid := uuid.MustParse(vote.ID)
	past := time.Now().Add(-time.Second)
	row, _ := svc.votes.GetVote(context.Background(), vid)
	row.EndTime = &past
	_ = svc.votes.UpdateVote(context.Background(), row)
	err = svc.CastVote(context.Background(), vid, org, "a")
	requireAppError(t, err, response.CodeVoteNotOpen)
	got, err := svc.GetVote(context.Background(), vid, org)
	require.NoError(t, err)
	require.Equal(t, model.VoteClosed, got.Status)
}
