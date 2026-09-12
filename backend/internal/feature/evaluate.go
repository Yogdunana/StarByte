package feature

import (
	"hash/fnv"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/feature/model"
	"github.com/google/uuid"
)

// Subject is the evaluation context. Anonymous callers use a zero UserID.
type Subject struct {
	UserID        uuid.UUID
	RoleCodes     []string
	DepartmentIDs []uuid.UUID
	Environment   string
	Now           time.Time
}

// Result explains a single evaluation.
type Result struct {
	Enabled bool
	Reason  string
	Variant string
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
	ReasonInvalidRule   = "invalid_rule"
	ReasonEnvMiss       = "env_miss"
	ReasonScheduleWait  = "schedule_pending"
	ReasonScheduleEnd   = "schedule_expired"
	ReasonVariantHit    = "variant_hit"
)

// Evaluate applies targeting, schedule window, environment, and type rules.
// A disabled flag is always off. Schedule further restricts an enabled flag
// without waiting for the persist ticker.
func Evaluate(flag *model.Flag, sub Subject) Result {
	if flag == nil {
		return Result{Enabled: false, Reason: ReasonNotFound}
	}
	if !flag.Enabled {
		return Result{Enabled: false, Reason: ReasonDisabled}
	}
	if !envAllows(flag, sub.Environment) {
		return Result{Enabled: false, Reason: ReasonEnvMiss}
	}
	now := evalNow(sub)
	if flag.Rules.StartsAt != nil && now.Before(flag.Rules.StartsAt.UTC()) {
		return Result{Enabled: false, Reason: ReasonScheduleWait}
	}
	if flag.Rules.EndsAt != nil && !now.Before(flag.Rules.EndsAt.UTC()) {
		return Result{Enabled: false, Reason: ReasonScheduleEnd}
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
	case model.TypeABTest:
		return evalAB(flag, sub)
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

func evalAB(flag *model.Flag, sub Subject) Result {
	variants := flag.Rules.Variants
	if len(variants) == 0 {
		return Result{Enabled: false, Reason: ReasonInvalidRule}
	}
	if sub.UserID == uuid.Nil {
		return Result{Enabled: false, Reason: ReasonAnonymous}
	}
	picked, ok := pickVariant(flag.FlagKey, flag.Rules.Salt, sub.UserID.String(), variants)
	if !ok {
		return Result{Enabled: false, Reason: ReasonInvalidRule}
	}
	return Result{Enabled: variantGateOn(picked), Reason: ReasonVariantHit, Variant: strings.TrimSpace(picked.Key)}
}

func variantGateOn(v model.Variant) bool {
	if v.Enabled != nil {
		return *v.Enabled
	}
	key := strings.ToLower(strings.TrimSpace(v.Key))
	return key != "" && key != "control" && key != "off"
}

func pickVariant(flagKey, salt, userID string, variants []model.Variant) (model.Variant, bool) {
	total := 0
	for _, v := range variants {
		if v.Weight > 0 {
			total += v.Weight
		}
	}
	if total <= 0 {
		return model.Variant{}, false
	}
	pick := int(hash64(flagKey, salt, userID) % uint64(total))
	acc := 0
	for _, v := range variants {
		w := v.Weight
		if w < 0 {
			w = 0
		}
		acc += w
		if pick < acc {
			return v, true
		}
	}
	return variants[len(variants)-1], true
}

func envAllows(flag *model.Flag, env string) bool {
	if len(flag.Rules.Environments) == 0 {
		return true
	}
	want := strings.ToLower(strings.TrimSpace(env))
	if want == "" {
		return false
	}
	for _, item := range flag.Rules.Environments {
		if strings.ToLower(strings.TrimSpace(item)) == want {
			return true
		}
	}
	return false
}

func evalNow(sub Subject) time.Time {
	if !sub.Now.IsZero() {
		return sub.Now.UTC()
	}
	return time.Now().UTC()
}

// ScheduleState is the window status for admin UI (independent of user targeting).
func ScheduleState(flag *model.Flag, now time.Time) string {
	if flag == nil {
		return model.ScheduleNone
	}
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	hasStart := flag.Rules.StartsAt != nil
	hasEnd := flag.Rules.EndsAt != nil
	if !hasStart && !hasEnd {
		return model.ScheduleNone
	}
	if hasStart && now.Before(flag.Rules.StartsAt.UTC()) {
		return model.SchedulePending
	}
	if hasEnd && !now.Before(flag.Rules.EndsAt.UTC()) {
		return model.ScheduleExpired
	}
	return model.ScheduleActive
}

// MasterOn is enabled && env && schedule window (no user targeting).
func MasterOn(flag *model.Flag, env string, now time.Time) bool {
	if flag == nil || !flag.Enabled {
		return false
	}
	sub := Subject{Environment: env, Now: now}
	if !envAllows(flag, env) {
		return false
	}
	if flag.Rules.StartsAt != nil && evalNow(sub).Before(flag.Rules.StartsAt.UTC()) {
		return false
	}
	if flag.Rules.EndsAt != nil && !evalNow(sub).Before(flag.Rules.EndsAt.UTC()) {
		return false
	}
	return true
}

// ShouldScheduleOn is true when starts_at has arrived, the flag is still off,
// and the last write happened before starts_at (so a later manual disable is kept).
func ShouldScheduleOn(flag *model.Flag, now time.Time) bool {
	if flag == nil || flag.Enabled || flag.Rules.StartsAt == nil {
		return false
	}
	now = now.UTC()
	start := flag.Rules.StartsAt.UTC()
	if now.Before(start) {
		return false
	}
	if flag.Rules.EndsAt != nil && !now.Before(flag.Rules.EndsAt.UTC()) {
		return false
	}
	if writtenAtOrAfter(flag.UpdatedAt, start) {
		return false
	}
	return true
}

// ShouldScheduleOff is true when ends_at has passed, the flag is still on,
// and the last write happened before ends_at (so a later manual enable is kept).
func ShouldScheduleOff(flag *model.Flag, now time.Time) bool {
	if flag == nil || !flag.Enabled || flag.Rules.EndsAt == nil {
		return false
	}
	end := flag.Rules.EndsAt.UTC()
	if now.UTC().Before(end) {
		return false
	}
	if writtenAtOrAfter(flag.UpdatedAt, end) {
		return false
	}
	return true
}

func writtenAtOrAfter(updated, boundary time.Time) bool {
	if updated.IsZero() {
		return false
	}
	return !updated.UTC().Before(boundary)
}

// Bucket returns a stable 0-99 bucket for percentage rollout.
func Bucket(flagKey, salt, userID string) uint64 {
	return bucket(flagKey, salt, userID)
}

func bucket(flagKey, salt, userID string) uint64 {
	return hash64(flagKey, salt, userID) % 100
}

func hash64(flagKey, salt, userID string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(flagKey))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(salt))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(userID))
	return h.Sum64()
}
