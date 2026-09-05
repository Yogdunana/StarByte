package service

import (
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	"github.com/stretchr/testify/require"
)

func TestMeetingStateGuards(t *testing.T) {
	require.True(t, canStartMeeting(model.MeetingPending))
	require.False(t, canStartMeeting(model.MeetingOngoing))
	require.True(t, canEndMeeting(model.MeetingOngoing))
	require.False(t, canEndMeeting(model.MeetingPending))
	require.True(t, canCancelMeeting(model.MeetingPending))
	require.True(t, canCancelMeeting(model.MeetingOngoing))
	require.False(t, canCancelMeeting(model.MeetingEnded))
	require.True(t, canDeleteMeeting(model.MeetingPending))
	require.False(t, canDeleteMeeting(model.MeetingOngoing))
	require.True(t, canCheckinMeeting(model.MeetingOngoing))
	require.False(t, canCreateVote(model.MeetingEnded))
	require.True(t, canCastVote(model.VoteOpen))
	require.False(t, canCastVote(model.VoteClosed))
}
