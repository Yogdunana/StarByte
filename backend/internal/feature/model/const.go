package model

const (
	TypeBoolean       = "boolean"
	TypeUserAllowlist = "user_allowlist"
	TypeRoleDept      = "role_dept"
	TypePercentage    = "percentage"

	ActionCreate = "create"
	ActionUpdate = "update"
	ActionToggle = "toggle"
	ActionDelete = "delete"

	KeyCMSPublic        = "cms.public"
	KeyAnnouncementFeed = "announcement.feed"
	KeyMembershipPortal = "membership.portal"
)

func ValidType(t string) bool {
	switch t {
	case TypeBoolean, TypeUserAllowlist, TypeRoleDept, TypePercentage:
		return true
	default:
		return false
	}
}
