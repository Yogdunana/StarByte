package feature

import (
	"testing"

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
	flag := &model.Flag{FlagType: "ab_test", Enabled: true}
	if got := Evaluate(flag, Subject{}); got.Enabled || got.Reason != ReasonInvalidType {
		t.Fatalf("invalid: %+v", got)
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
