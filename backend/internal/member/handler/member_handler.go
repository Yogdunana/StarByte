package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/member/service"
)

// MemberHandler 入会申请与人员档案。
type MemberHandler struct {
	svc service.MemberService
}

// NewMemberHandler 创建处理器。
func NewMemberHandler(svc service.MemberService) *MemberHandler {
	return &MemberHandler{svc: svc}
}
