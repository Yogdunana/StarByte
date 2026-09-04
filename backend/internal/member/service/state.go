package service

import "github.com/Yogdunana/StarByte/backend/internal/member/model"

const (
	actionApprove    = "approve"
	actionReject     = "reject"
	actionSupplement = "supplement"
	actionResubmit   = "resubmit"
)

// nextStatus 按 Issue #6 状态机计算下一状态。
func nextStatus(applicantType, current int16, action string) (int16, bool) {
	if action == actionResubmit {
		if current == model.AppSupplement {
			return model.AppPending, true
		}
		return 0, false
	}

	if applicantType == model.ApplicantMember {
		return memberNext(current, action)
	}
	return officerNext(current, action)
}

func memberNext(current int16, action string) (int16, bool) {
	if current != model.AppPending {
		return 0, false
	}
	switch action {
	case actionApprove:
		return model.AppApproved, true
	case actionReject:
		return model.AppRejected, true
	case actionSupplement:
		return model.AppSupplement, true
	default:
		return 0, false
	}
}

func officerNext(current int16, action string) (int16, bool) {
	switch current {
	case model.AppPending:
		switch action {
		case actionApprove:
			return model.AppReviewing, true
		case actionReject:
			return model.AppRejected, true
		case actionSupplement:
			return model.AppSupplement, true
		}
	case model.AppReviewing:
		switch action {
		case actionApprove:
			return model.AppInterviewing, true
		case actionReject:
			return model.AppRejected, true
		}
	case model.AppInterviewing:
		switch action {
		case actionApprove:
			return model.AppApproved, true
		case actionReject:
			return model.AppRejected, true
		}
	}
	return 0, false
}

func stageLabel(status int16) string {
	switch status {
	case model.AppPending:
		return "待审核"
	case model.AppReviewing:
		return "审核中"
	case model.AppInterviewing:
		return "面试中"
	case model.AppApproved:
		return "通过"
	case model.AppRejected:
		return "拒绝"
	case model.AppSupplement:
		return "补充材料"
	default:
		return ""
	}
}
