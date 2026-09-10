package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/member/service"
)

// MemberHandler 入会申请与人员档案。
type MemberHandler struct {
	svc       service.MemberService
	admission service.AdmissionService
}

// NewMemberHandler 创建处理器。
func NewMemberHandler(svc service.MemberService, admissions ...service.AdmissionService) *MemberHandler {
	var admission service.AdmissionService
	if len(admissions) > 0 {
		admission = admissions[0]
	}
	return &MemberHandler{svc: svc, admission: admission}
}
