package model

const (
	TypeBoolean       = "boolean"
	TypeUserAllowlist = "user_allowlist"
	TypeRoleDept      = "role_dept"
	TypePercentage    = "percentage"
	TypeABTest        = "ab_test"

	ActionCreate      = "create"
	ActionUpdate      = "update"
	ActionToggle      = "toggle"
	ActionDelete      = "delete"
	ActionRollback    = "rollback"
	ActionScheduleOn  = "schedule_on"
	ActionScheduleOff = "schedule_off"

	EnvDev  = "dev"
	EnvTest = "test"
	EnvProd = "prod"

	ScheduleNone    = ""
	SchedulePending = "pending"
	ScheduleActive  = "active"
	ScheduleExpired = "expired"

	KeyCMSPublic        = "cms.public"
	KeyAnnouncementFeed = "announcement.feed"
	KeyMembershipPortal = "membership.portal"
)

func ValidType(t string) bool {
	switch t {
	case TypeBoolean, TypeUserAllowlist, TypeRoleDept, TypePercentage, TypeABTest:
		return true
	default:
		return false
	}
}

func ValidEnv(env string) bool {
	switch env {
	case EnvDev, EnvTest, EnvProd:
		return true
	default:
		return false
	}
}
