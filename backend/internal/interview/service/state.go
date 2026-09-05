package service

import "github.com/Yogdunana/StarByte/backend/internal/interview/model"

func canStartSession(status int16) bool {
	return status == model.SessionPending
}

func canEndSession(status int16) bool {
	return status == model.SessionOngoing
}

func canUpdateSession(status int16) bool {
	return status == model.SessionPending
}

func canDeleteSession(status int16) bool {
	return status == model.SessionPending || status == model.SessionCancelled
}

func canAddCandidate(status int16) bool {
	return status == model.SessionPending || status == model.SessionOngoing
}

func canCheckin(status int16) bool {
	return status == model.InterviewPending
}

func canStartInterview(status int16) bool {
	return status == model.InterviewPending || status == model.InterviewCheckedIn
}

func canEndInterview(status int16) bool {
	return status == model.InterviewOngoing
}

func canScore(status int16) bool {
	return status == model.InterviewOngoing || status == model.InterviewDone
}

func canSubmitResult(status int16) bool {
	return status == model.InterviewDone || status == model.InterviewOngoing
}

func resultLabel(code int16) string {
	switch code {
	case model.ResultPass:
		return "通过"
	case model.ResultFail:
		return "不通过"
	case model.ResultPending:
		return "待定"
	default:
		return ""
	}
}
