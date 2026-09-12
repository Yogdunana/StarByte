package feature

import (
	"hash/fnv"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/feature/model"
	"github.com/google/uuid"
)

// Subject is the evaluation context. Anonymous callers use a zero UserID.
type Subject struct {
	UserID        uuid.UUID
	RoleCodes     []string
	DepartmentIDs []uuid.UUID
}

// Result explains a single evaluation.
type Result struct {
	Enabled bool
	Reason  string
}

const (
	ReasonDisabled      = "disabled"
	ReasonNotFound      = "not_found"
	ReasonBoolean       = "boolean"
	ReasonAllowlistHit  = "allowlist_hit"
	ReasonAllowlistMiss = "allowlist_miss"
	ReasonRoleHit       = "role_hit"
	ReasonDeptHit       = "dept_hit"
	ReasonRoleDeptMiss  = "role_dept_miss"
	ReasonPercentHit    = "percent_hit"
	ReasonPercentMiss   = "percent_miss"
	ReasonAnonymous     = "anonymous"
	ReasonInvalidType   = "invalid_type"
)

// Evaluate applies phase-1 rules. A disabled flag is always off.
func Evaluate(flag *model.Flag, sub Subject) Result {
	if flag == nil {
		return Result{Enabled: false, Reason: ReasonNotFound}
	}
	if !flag.Enabled {
		return Result{Enabled: false, Reason: ReasonDisabled}
	}
	switch flag.FlagType {
	case model.TypeBoolean:
		return Result{Enabled: true, Reason: ReasonBoolean}
	case model.TypeUserAllowlist:
		return evalAllowlist(flag, sub)
	case model.TypeRoleDept:
		return evalRoleDept(flag, sub)
	case model.TypePercentage:
		return evalPercent(flag, sub)
	default:
		return Result{Enabled: false, Reason: ReasonInvalidType}
	}
}

func evalAllowlist(flag *model.Flag, sub Subject) Result {
	if sub.UserID == uuid.Nil {
		return Result{Enabled: false, Reason: ReasonAnonymous}
	}
	want := strings.ToLower(sub.UserID.String())
	for _, id := range flag.Rules.UserIDs {
		if strings.ToLower(strings.TrimSpace(id)) == want {
			return Result{Enabled: true, Reason: ReasonAllowlistHit}
		}
	}
	return Result{Enabled: false, Reason: ReasonAllowlistMiss}
}

func evalRoleDept(flag *model.Flag, sub Subject) Result {
	roles := make(map[string]struct{}, len(sub.RoleCodes))
	for _, r := range sub.RoleCodes {
		roles[strings.ToLower(strings.TrimSpace(r))] = struct{}{}
	}
	for _, code := range flag.Rules.RoleCodes {
		if _, ok := roles[strings.ToLower(strings.TrimSpace(code))]; ok {
			return Result{Enabled: true, Reason: ReasonRoleHit}
		}
	}
	depts := make(map[string]struct{}, len(sub.DepartmentIDs))
	for _, id := range sub.DepartmentIDs {
		if id != uuid.Nil {
			depts[strings.ToLower(id.String())] = struct{}{}
		}
	}
	for _, id := range flag.Rules.DepartmentIDs {
		if _, ok := depts[strings.ToLower(strings.TrimSpace(id))]; ok {
			return Result{Enabled: true, Reason: ReasonDeptHit}
		}
	}
	return Result{Enabled: false, Reason: ReasonRoleDeptMiss}
}

func evalPercent(flag *model.Flag, sub Subject) Result {
	pct := flag.Rules.Percent
	if pct <= 0 {
		return Result{Enabled: false, Reason: ReasonPercentMiss}
	}
	if pct >= 100 {
		return Result{Enabled: true, Reason: ReasonPercentHit}
	}
	if sub.UserID == uuid.Nil {
		return Result{Enabled: false, Reason: ReasonAnonymous}
	}
	if bucket(flag.FlagKey, flag.Rules.Salt, sub.UserID.String()) < uint64(pct) {
		return Result{Enabled: true, Reason: ReasonPercentHit}
	}
	return Result{Enabled: false, Reason: ReasonPercentMiss}
}

// Bucket returns a stable 0-99 bucket for percentage rollout.
func Bucket(flagKey, salt, userID string) uint64 {
	return bucket(flagKey, salt, userID)
}

func bucket(flagKey, salt, userID string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(flagKey))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(salt))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(userID))
	return h.Sum64() % 100
}
