package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/member/service"
)

// MemberHandler 入会申请与人员档案。
type MemberHandler struct {
	svc       service.MemberService
	admission service.AdmissionService
	importer  service.MemberImportService
}

// NewMemberHandler 创建处理器。
func NewMemberHandler(svc service.MemberService, admissions ...service.AdmissionService) *MemberHandler {
	var admission service.AdmissionService
	if len(admissions) > 0 {
		admission = admissions[0]
	}
	return &MemberHandler{svc: svc, admission: admission}
}

// WithImporter 挂上批量录入服务。单独给一个方法而不是再加可变参数：
// 既有调用方（含单测）不用改签名。
func (h *MemberHandler) WithImporter(svc service.MemberImportService) *MemberHandler {
	h.importer = svc
	return h
}
