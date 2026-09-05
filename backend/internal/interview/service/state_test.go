package service

import (
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/interview/model"
	"github.com/stretchr/testify/require"
)

func TestSessionStateMachine(t *testing.T) {
	require.True(t, canStartSession(model.SessionPending))
	require.False(t, canStartSession(model.SessionOngoing))
	require.True(t, canEndSession(model.SessionOngoing))
	require.False(t, canEndSession(model.SessionPending))
	require.True(t, canUpdateSession(model.SessionPending))
	require.False(t, canDeleteSession(model.SessionOngoing))
	require.True(t, canAddCandidate(model.SessionOngoing))
}

func TestInterviewStateMachine(t *testing.T) {
	require.True(t, canCheckin(model.InterviewPending))
	require.False(t, canCheckin(model.InterviewDone))
	require.True(t, canStartInterview(model.InterviewCheckedIn))
	require.True(t, canEndInterview(model.InterviewOngoing))
	require.True(t, canScore(model.InterviewOngoing))
	require.True(t, canSubmitResult(model.InterviewDone))
	require.Equal(t, "通过", resultLabel(model.ResultPass))
	require.Equal(t, "不通过", resultLabel(model.ResultFail))
}
