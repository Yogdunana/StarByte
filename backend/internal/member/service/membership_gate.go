package service

import (
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func hasMembershipRole(roles []string) bool {
	for _, role := range roles {
		switch role {
		case "member", "officer", "probationary", "president", "vice_president", "minister", "vice_minister", "center_director", "vice_center_director", "honorary":
			return true
		}
	}
	return false
}

func rejectDuplicateMembership(applicantType int, profile *model.MemberProfile, roles []string) error {
	if profile != nil {
		if profile.Status == model.ProfileDisabled {
			return admissionDenied("档案已停用，请联系管理人员处理")
		}
		if profile.Status == model.ProfileProbation {
			return response.NewError(response.CodeMemberAppDuplicate, "已有预备期申请，请等待处理")
		}
		if profile.Status == model.ProfileActive && profile.MemberType >= int16(applicantType) {
			return response.NewError(response.CodeMemberAppDuplicate, "已具有该成员身份，无需重复申请")
		}
	}
	if applicantType == int(model.ApplicantMember) && hasMembershipRole(roles) {
		return response.NewError(response.CodeMemberAppDuplicate, "已具有该成员身份，无需重复申请")
	}
	return nil
}
