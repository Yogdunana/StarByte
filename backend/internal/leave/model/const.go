package model

const (
	ApprovalStatusPending  = "pending"
	ApprovalStatusApproved = "approved"
	ApprovalStatusRejected = "rejected"

	TypeAnnual       = "annual"
	TypeSick         = "sick"
	TypePersonal     = "personal"
	TypeCompensatory = "compensatory"

	StageDepartment = "minister"
	StageOrg        = "president"

	MaxAttachments = 5
	MaxTypeCodeLen = 20
	MaxTypeNameLen = 50
)
