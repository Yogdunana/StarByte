package feature

import (
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/feature/model"
	"github.com/google/uuid"
)

func TestEvaluate_NilAndDisabled(t *testing.T) {
	if got := Evaluate(nil, Subject{}); got.Reason != ReasonNotFound || got.Enabled {
		t.Fatalf("nil: %+v", got)
	}
	flag := &model.Flag{FlagType: model.TypeBoolean, Enabled: false}
	if got := Evaluate(flag, Subject{}); got.Reason != ReasonDisabled || got.Enabled {
		t.Fatalf("disabled: %+v", got)
	}
}

func TestEvaluate_Boolean(t *testing.T) {
	flag := &model.Flag{FlagType: model.TypeBoolean, Enabled: true}
	if got := Evaluate(flag, Subject{}); !got.Enabled || got.Reason != ReasonBoolean {
		t.Fatalf("boolean: %+v", got)
	}
}

func TestEvaluate_Allowlist(t *testing.T) {
	uid := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	flag := &model.Flag{
		FlagType: model.TypeUserAllowlist,
		Enabled:  true,
		Rules:    model.Rules{UserIDs: []string{uid.String()}},
	}
	if got := Evaluate(flag, Subject{}); got.Enabled || got.Reason != ReasonAnonymous {
		t.Fatalf("anon: %+v", got)
	}
	if got := Evaluate(flag, Subject{UserID: uid}); !got.Enabled || got.Reason != ReasonAllowlistHit {
		t.Fatalf("hit: %+v", got)
	}
	if got := Evaluate(flag, Subject{UserID: uuid.New()}); got.Enabled || got.Reason != ReasonAllowlistMiss {
		t.Fatalf("miss: %+v", got)
	}
}

func TestEvaluate_RoleDept(t *testing.T) {
	dept := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	flag := &model.Flag{
		FlagType: model.TypeRoleDept,
		Enabled:  true,
		Rules:    model.Rules{RoleCodes: []string{"minister"}, DepartmentIDs: []string{dept.String()}},
	}
	if got := Evaluate(flag, Subject{RoleCodes: []string{"MINISTER"}}); !got.Enabled || got.Reason != ReasonRoleHit {
		t.Fatalf("role: %+v", got)
	}
	if got := Evaluate(flag, Subject{DepartmentIDs: []uuid.UUID{dept}}); !got.Enabled || got.Reason != ReasonDeptHit {
		t.Fatalf("dept: %+v", got)
	}
	if got := Evaluate(flag, Subject{RoleCodes: []string{"member"}}); got.Enabled || got.Reason != ReasonRoleDeptMiss {
		t.Fatalf("miss: %+v", got)
	}
}

func TestEvaluate_Percentage(t *testing.T) {
	uid := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	zero := &model.Flag{FlagKey: "a", FlagType: model.TypePercentage, Enabled: true, Rules: model.Rules{Percent: 0}}
	if got := Evaluate(zero, Subject{UserID: uid}); got.Enabled {
		t.Fatalf("0%% should miss: %+v", got)
	}
	all := &model.Flag{FlagKey: "a", FlagType: model.TypePercentage, Enabled: true, Rules: model.Rules{Percent: 100}}
	if got := Evaluate(all, Subject{UserID: uid}); !got.Enabled {
		t.Fatalf("100%% should hit: %+v", got)
	}
	mid := &model.Flag{FlagKey: "cms.public", FlagType: model.TypePercentage, Enabled: true, Rules: model.Rules{Percent: 10, Salt: "recruit"}}
	first := Evaluate(mid, Subject{UserID: uid})
	second := Evaluate(mid, Subject{UserID: uid})
	if first != second {
		t.Fatalf("hash must be stable: %+v vs %+v", first, second)
	}
	anon := Evaluate(mid, Subject{})
	if anon.Enabled || anon.Reason != ReasonAnonymous {
		t.Fatalf("anon percent: %+v", anon)
	}
}

func TestEvaluate_InvalidType(t *testing.T) {
	flag := &model.Flag{FlagType: "wasm", Enabled: true}
	if got := Evaluate(flag, Subject{}); got.Enabled || got.Reason != ReasonInvalidType {
		t.Fatalf("invalid: %+v", got)
	}
}

func TestEvaluate_Environment(t *testing.T) {
	flag := &model.Flag{
		FlagType: model.TypeBoolean, Enabled: true,
		Rules: model.Rules{Environments: []string{"prod"}},
	}
	if got := Evaluate(flag, Subject{Environment: "dev"}); got.Enabled || got.Reason != ReasonEnvMiss {
		t.Fatalf("dev: %+v", got)
	}
	if got := Evaluate(flag, Subject{}); got.Enabled || got.Reason != ReasonEnvMiss {
		t.Fatalf("empty env: %+v", got)
	}
	if got := Evaluate(flag, Subject{Environment: "prod"}); !got.Enabled {
		t.Fatalf("prod: %+v", got)
	}
	open := &model.Flag{FlagType: model.TypeBoolean, Enabled: true}
	if got := Evaluate(open, Subject{Environment: "dev"}); !got.Enabled {
		t.Fatalf("no env list should match: %+v", got)
	}
}

func TestEvaluate_ScheduleWindow(t *testing.T) {
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	start := now.Add(time.Hour)
	end := now.Add(-time.Hour)
	pending := &model.Flag{FlagType: model.TypeBoolean, Enabled: true, Rules: model.Rules{StartsAt: &start}}
	if got := Evaluate(pending, Subject{Now: now}); got.Enabled || got.Reason != ReasonScheduleWait {
		t.Fatalf("pending: %+v", got)
	}
	expired := &model.Flag{FlagType: model.TypeBoolean, Enabled: true, Rules: model.Rules{EndsAt: &end}}
	if got := Evaluate(expired, Subject{Now: now}); got.Enabled || got.Reason != ReasonScheduleEnd {
		t.Fatalf("expired: %+v", got)
	}
	activeEnd := now.Add(time.Hour)
	active := &model.Flag{FlagType: model.TypeBoolean, Enabled: true, Rules: model.Rules{StartsAt: &end, EndsAt: &activeEnd}}
	if got := Evaluate(active, Subject{Now: now}); !got.Enabled {
		t.Fatalf("active: %+v", got)
	}
}

func TestEvaluate_ABVariants(t *testing.T) {
	uid := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	on := true
	off := false
	flag := &model.Flag{
		FlagKey:  "exp.hero",
		FlagType: model.TypeABTest,
		Enabled:  true,
		Rules: model.Rules{
			Salt: "v1",
			Variants: []model.Variant{
				{Key: "control", Weight: 50, Enabled: &off},
				{Key: "treatment", Weight: 50, Enabled: &on},
			},
		},
	}
	anon := Evaluate(flag, Subject{})
	if anon.Variant != "control" || anon.Enabled || anon.Reason != ReasonAnonymous {
		t.Fatalf("anon ab: %+v", anon)
	}
	a := Evaluate(flag, Subject{UserID: uid})
	b := Evaluate(flag, Subject{UserID: uid})
	if a != b || a.Variant == "" || a.Reason != ReasonVariantHit {
		t.Fatalf("unstable ab: %+v vs %+v", a, b)
	}
	empty := Evaluate(&model.Flag{FlagType: model.TypeABTest, Enabled: true}, Subject{UserID: uid})
	if empty.Enabled || empty.Reason != ReasonInvalidRule {
		t.Fatalf("empty variants: %+v", empty)
	}
	zeroW := Evaluate(&model.Flag{
		FlagType: model.TypeABTest, Enabled: true,
		Rules: model.Rules{Variants: []model.Variant{{Key: "a", Weight: 0}}},
	}, Subject{UserID: uid})
	if zeroW.Enabled || zeroW.Reason != ReasonInvalidRule {
		t.Fatalf("zero weight: %+v", zeroW)
	}
	implicit := Evaluate(&model.Flag{
		FlagKey: "exp.imp", FlagType: model.TypeABTest, Enabled: true,
		Rules: model.Rules{Variants: []model.Variant{{Key: "control", Weight: 1}, {Key: "on", Weight: 1}}},
	}, Subject{UserID: uid})
	if implicit.Variant != "control" && implicit.Variant != "on" {
		t.Fatalf("implicit: %+v", implicit)
	}
	if implicit.Variant == "control" && implicit.Enabled {
		t.Fatalf("control should be off: %+v", implicit)
	}
}

func TestScheduleHelpers(t *testing.T) {
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)
	if ScheduleState(nil, now) != model.ScheduleNone {
		t.Fatal("nil")
	}
	plain := &model.Flag{}
	if ScheduleState(plain, now) != model.ScheduleNone {
		t.Fatal("none")
	}
	if ScheduleState(&model.Flag{Rules: model.Rules{StartsAt: &future}}, now) != model.SchedulePending {
		t.Fatal("pending")
	}
	if ScheduleState(&model.Flag{Rules: model.Rules{EndsAt: &past}}, now) != model.ScheduleExpired {
		t.Fatal("expired")
	}
	if ScheduleState(&model.Flag{Rules: model.Rules{StartsAt: &past, EndsAt: &future}}, time.Time{}) != model.ScheduleActive {
		t.Fatal("active with zero now")
	}
	off := &model.Flag{Enabled: false, Rules: model.Rules{StartsAt: &past, EndsAt: &future}}
	if !ShouldScheduleOn(off, now) || ShouldScheduleOff(off, now) {
		t.Fatal("should turn on")
	}
	on := &model.Flag{Enabled: true, Rules: model.Rules{EndsAt: &past}}
	if !ShouldScheduleOff(on, now) || ShouldScheduleOn(on, now) {
		t.Fatal("should turn off")
	}
	if MasterOn(nil, "prod", now) || MasterOn(&model.Flag{Enabled: true, Rules: model.Rules{Environments: []string{"prod"}}}, "dev", now) {
		t.Fatal("master off")
	}
	if !MasterOn(&model.Flag{Enabled: true}, "dev", now) {
		t.Fatal("master on")
	}
}

func TestBucket_RangeAndStability(t *testing.T) {
	a := Bucket("cms.public", "s", "u1")
	b := Bucket("cms.public", "s", "u1")
	if a != b || a > 99 {
		t.Fatalf("bucket=%d", a)
	}
	if Bucket("cms.public", "s", "u1") == Bucket("cms.public", "other", "u1") && Bucket("cms.public", "s", "u1") == Bucket("other", "s", "u1") {
		t.Fatal("salt/key should usually change the bucket")
	}
}
