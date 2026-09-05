package service

import "github.com/Yogdunana/StarByte/backend/internal/meeting/model"

func canUpdateMeeting(status int16) bool {
	return status == model.MeetingPending
}

func canDeleteMeeting(status int16) bool {
	return status == model.MeetingPending || status == model.MeetingCancelled
}

func canStartMeeting(status int16) bool {
	return status == model.MeetingPending
}

func canEndMeeting(status int16) bool {
	return status == model.MeetingOngoing
}

func canCancelMeeting(status int16) bool {
	return status == model.MeetingPending || status == model.MeetingOngoing
}

func canCheckinMeeting(status int16) bool {
	return status == model.MeetingPending || status == model.MeetingOngoing
}

func canCreateVote(status int16) bool {
	return status == model.MeetingPending || status == model.MeetingOngoing
}

func canCastVote(status int16) bool {
	return status == model.VoteOpen
}
