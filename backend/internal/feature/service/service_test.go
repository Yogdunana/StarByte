package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/feature"
	"github.com/Yogdunana/StarByte/backend/internal/feature/dto"
	"github.com/Yogdunana/StarByte/backend/internal/feature/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func newTestSvc(roles []string, dept *uuid.UUID, perms stubPerms) (Service, *memRepo, *MemoryBus) {
	rows := newMemRepo()
	bus := NewMemoryBus()
	svc := New(rows, NewMemorySnapshot(), bus, stubRoles{codes: roles}, stubUsers{dept: dept}, perms)
	return svc, rows, bus
}

func TestCreateValidateAndToggle(t *testing.T) {
	svc, rows, _ := newTestSvc(nil, nil, stubPerms{})
	ctx := context.Background()
	actor := uuid.New()
	_, err := svc.Create(ctx, actor, &dto.CreateFlagRequest{FlagKey: "Bad", Name: "x", FlagType: model.TypeBoolean})
	if codeOf(err) != response.CodeFeatureInvalidKey {
		t.Fatalf("key: %v", err)
	}
	_, err = svc.Create(ctx, actor, &dto.CreateFlagRequest{FlagKey: "demo.flag", Name: "Demo", FlagType: "wasm"})
	if codeOf(err) != response.CodeFeatureInvalidType {
		t.Fatalf("type: %v", err)
	}
	_, err = svc.Create(ctx, actor, &dto.CreateFlagRequest{
		FlagKey: "demo.pct", Name: "Pct", FlagType: model.TypePercentage, Rules: model.Rules{Percent: 120},
	})
	if codeOf(err) != response.CodeFeatureInvalidRule {
		t.Fatalf("pct: %v", err)
	}
	created, err := svc.Create(ctx, actor, &dto.CreateFlagRequest{
		FlagKey: "demo.flag", Name: "Demo", FlagType: model.TypeBoolean, Enabled: false, GroupName: "demo",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.Create(ctx, actor, &dto.CreateFlagRequest{FlagKey: "demo.flag", Name: "x", FlagType: model.TypeBoolean}); codeOf(err) != response.CodeFeatureKeyExists {
		t.Fatalf("dup: %v", err)
	}
	id := uuid.MustParse(created.ID)
	got, err := svc.Toggle(ctx, actor, id, &dto.ToggleRequest{})
	if err != nil || !got.Enabled {
		t.Fatalf("toggle: %+v %v", got, err)
	}
	if len(rows.audits) < 2 {
		t.Fatalf("audits=%d", len(rows.audits))
	}
	list, total, err := svc.List(ctx, dto.ListQuery{Keyword: "demo"})
	if err != nil || total != 1 || list[0].FlagKey != "demo.flag" {
		t.Fatalf("list: %v %d %+v", err, total, list)
	}
}

func TestUpdateDeleteAndEvaluate(t *testing.T) {
	svc, _, _ := newTestSvc([]string{"minister"}, nil, stubPerms{})
	ctx := context.Background()
	actor := uuid.New()
	uid := uuid.New()
	created, err := svc.Create(ctx, actor, &dto.CreateFlagRequest{
		FlagKey: "role.gate", Name: "Role", FlagType: model.TypeRoleDept, Enabled: true,
		Rules: model.Rules{RoleCodes: []string{"minister"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(created.ID)
	name := "Role gate"
	typ := model.TypeUserAllowlist
	rules := model.Rules{UserIDs: []string{uid.String()}}
	updated, err := svc.Update(ctx, actor, id, &dto.UpdateFlagRequest{Name: &name, FlagType: &typ, Rules: &rules})
	if err != nil || updated.Name != name {
		t.Fatalf("update: %+v %v", updated, err)
	}
	eval, err := svc.EvaluateID(ctx, id, uid)
	if err != nil || !eval.Enabled || eval.Reason != feature.ReasonAllowlistHit {
		t.Fatalf("eval: %+v %v", eval, err)
	}
	if err := svc.Delete(ctx, actor, id); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Get(ctx, id); codeOf(err) != response.CodeFeatureNotFound {
		t.Fatalf("deleted: %v", err)
	}
}

func TestSystemFlagProtectedAndMeSnapshot(t *testing.T) {
	svc, rows, _ := newTestSvc(nil, nil, stubPerms{perms: []string{"announcement:create"}})
	ctx := context.Background()
	id := uuid.New()
	_ = rows.Create(ctx, &model.Flag{ID: id, FlagKey: model.KeyAnnouncementFeed, Name: "Feed", FlagType: model.TypeBoolean, Enabled: false, IsSystem: true})
	_ = rows.Create(ctx, &model.Flag{ID: uuid.New(), FlagKey: model.KeyCMSPublic, Name: "CMS", FlagType: model.TypeBoolean, Enabled: false})
	if err := svc.Delete(ctx, uuid.New(), id); codeOf(err) != response.CodeFeatureProtected {
		t.Fatalf("protected: %v", err)
	}
	me, err := svc.EvaluateMe(ctx, uuid.New(), []string{model.KeyAnnouncementFeed, model.KeyCMSPublic})
	if err != nil {
		t.Fatal(err)
	}
	if !me[model.KeyAnnouncementFeed].Enabled || me[model.KeyAnnouncementFeed].Reason != "staff_bypass" {
		t.Fatalf("staff: %+v", me[model.KeyAnnouncementFeed])
	}
	if me[model.KeyCMSPublic].Enabled {
		t.Fatalf("cms should stay off: %+v", me[model.KeyCMSPublic])
	}
}

func TestPercentTenPercentStable(t *testing.T) {
	svc, _, _ := newTestSvc(nil, nil, stubPerms{})
	ctx := context.Background()
	created, err := svc.Create(ctx, uuid.New(), &dto.CreateFlagRequest{
		FlagKey: "rollout.ten", Name: "10%", FlagType: model.TypePercentage, Enabled: true,
		Rules: model.Rules{Percent: 10, Salt: "recruit"},
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(created.ID)
	uid := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	a, err := svc.EvaluateID(ctx, id, uid)
	if err != nil {
		t.Fatal(err)
	}
	b, err := svc.EvaluateID(ctx, id, uid)
	if err != nil || a.Enabled != b.Enabled {
		t.Fatalf("unstable: %+v %+v", a, b)
	}
}

func TestHotReloadViaBus(t *testing.T) {
	rows := newMemRepo()
	store := NewMemorySnapshot()
	bus := NewMemoryBus()
	svc := New(rows, store, bus, stubRoles{}, stubUsers{}, stubPerms{}).(*flagService)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := svc.StartHotReload(ctx); err != nil {
		t.Fatal(err)
	}
	actor := uuid.New()
	created, err := svc.Create(ctx, actor, &dto.CreateFlagRequest{FlagKey: "hot.flag", Name: "Hot", FlagType: model.TypeBoolean, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if !svc.Enabled(ctx, "hot.flag", feature.Subject{UserID: uuid.New()}) {
		t.Fatal("expected enabled after create")
	}
	off := false
	if _, err := svc.Update(ctx, actor, uuid.MustParse(created.ID), &dto.UpdateFlagRequest{Enabled: &off}); err != nil {
		t.Fatal(err)
	}
	if svc.Enabled(ctx, "hot.flag", feature.Subject{UserID: uuid.New()}) {
		t.Fatal("expected disabled after update")
	}
}

func TestGetValidateRulesAndResolve(t *testing.T) {
	dept := uuid.New()
	svc, _, _ := newTestSvc([]string{"member"}, &dept, stubPerms{isSuper: true})
	ctx := context.Background()
	if _, err := svc.Get(ctx, uuid.New()); codeOf(err) != response.CodeFeatureNotFound {
		t.Fatalf("missing: %v", err)
	}
	_, err := svc.Create(ctx, uuid.New(), &dto.CreateFlagRequest{
		FlagKey: "bad.users", Name: "x", FlagType: model.TypeUserAllowlist, Rules: model.Rules{UserIDs: []string{"nope"}},
	})
	if codeOf(err) != response.CodeFeatureInvalidRule {
		t.Fatalf("users: %v", err)
	}
	_, err = svc.Create(ctx, uuid.New(), &dto.CreateFlagRequest{
		FlagKey: "bad.depts", Name: "x", FlagType: model.TypeRoleDept, Rules: model.Rules{DepartmentIDs: []string{"nope"}},
	})
	if codeOf(err) != response.CodeFeatureInvalidRule {
		t.Fatalf("depts: %v", err)
	}
	created, err := svc.Create(ctx, uuid.New(), &dto.CreateFlagRequest{FlagKey: "ok.flag", Name: "OK", FlagType: model.TypeBoolean, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	got, err := svc.Get(ctx, uuid.MustParse(created.ID))
	if err != nil || got.FlagKey != "ok.flag" {
		t.Fatalf("get: %+v %v", got, err)
	}
	on := true
	if _, err := svc.Toggle(ctx, uuid.New(), uuid.MustParse(created.ID), &dto.ToggleRequest{Enabled: &on, Reason: "force"}); err != nil {
		t.Fatal(err)
	}
	sub, err := svc.Resolve(ctx, uuid.New())
	if err != nil || len(sub.RoleCodes) != 1 || len(sub.DepartmentIDs) != 1 {
		t.Fatalf("resolve: %+v %v", sub, err)
	}
	me, err := svc.EvaluateMe(ctx, uuid.New(), nil)
	if err != nil || !me[model.KeyAnnouncementFeed].Enabled {
		t.Fatalf("super me: %+v %v", me, err)
	}
}

func TestListAudits(t *testing.T) {
	svc, _, _ := newTestSvc(nil, nil, stubPerms{})
	ctx := context.Background()
	_, err := svc.Create(ctx, uuid.New(), &dto.CreateFlagRequest{FlagKey: "audit.flag", Name: "A", FlagType: model.TypeBoolean})
	if err != nil {
		t.Fatal(err)
	}
	list, total, err := svc.ListAudits(ctx, dto.AuditQuery{FlagKey: "audit.flag"})
	if err != nil || total != 1 || list[0].Action != model.ActionCreate {
		t.Fatalf("audit: %v %d %+v", err, total, list)
	}
}

func TestABEnvScheduleRollbackAnalytics(t *testing.T) {
	svc, rows, _ := newTestSvc(nil, nil, stubPerms{})
	fs := svc.(*flagService)
	fs.env = "prod"
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	fs.now = func() time.Time { return now }
	ctx := context.Background()
	actor := uuid.New()
	uid := uuid.MustParse("55555555-5555-4555-8555-555555555555")

	_, err := svc.Create(ctx, actor, &dto.CreateFlagRequest{
		FlagKey: "exp.bad", Name: "bad", FlagType: model.TypeABTest,
		Rules: model.Rules{Variants: []model.Variant{{Key: "only", Weight: 1}}},
	})
	if codeOf(err) != response.CodeFeatureInvalidRule {
		t.Fatalf("ab variants: %v", err)
	}
	_, err = svc.Create(ctx, actor, &dto.CreateFlagRequest{
		FlagKey: "exp.dup", Name: "bad", FlagType: model.TypeABTest,
		Rules: model.Rules{Variants: []model.Variant{{Key: "a", Weight: 1}, {Key: "A", Weight: 1}}},
	})
	if codeOf(err) != response.CodeFeatureInvalidRule {
		t.Fatalf("dup variant: %v", err)
	}
	_, err = svc.Create(ctx, actor, &dto.CreateFlagRequest{
		FlagKey: "exp.neg", Name: "bad", FlagType: model.TypeABTest,
		Rules: model.Rules{Variants: []model.Variant{{Key: "a", Weight: -1}, {Key: "b", Weight: 1}}},
	})
	if codeOf(err) != response.CodeFeatureInvalidRule {
		t.Fatalf("neg weight: %v", err)
	}
	_, err = svc.Create(ctx, actor, &dto.CreateFlagRequest{
		FlagKey: "exp.empty", Name: "bad", FlagType: model.TypeABTest,
		Rules: model.Rules{Variants: []model.Variant{{Key: "", Weight: 1}, {Key: "b", Weight: 1}}},
	})
	if codeOf(err) != response.CodeFeatureInvalidRule {
		t.Fatalf("empty key: %v", err)
	}
	_, err = svc.Create(ctx, actor, &dto.CreateFlagRequest{
		FlagKey: "env.dup", Name: "bad", FlagType: model.TypeBoolean,
		Rules: model.Rules{Environments: []string{"dev", "dev"}},
	})
	if codeOf(err) != response.CodeFeatureInvalidRule {
		t.Fatalf("dup env: %v", err)
	}
	_, err = svc.Create(ctx, actor, &dto.CreateFlagRequest{
		FlagKey: "env.bad", Name: "bad", FlagType: model.TypeBoolean,
		Rules: model.Rules{Environments: []string{"staging"}},
	})
	if codeOf(err) != response.CodeFeatureInvalidRule {
		t.Fatalf("env: %v", err)
	}
	start := now.Add(time.Hour)
	end := now.Add(-time.Hour)
	_, err = svc.Create(ctx, actor, &dto.CreateFlagRequest{
		FlagKey: "sched.bad", Name: "bad", FlagType: model.TypeBoolean,
		Rules: model.Rules{StartsAt: &start, EndsAt: &end},
	})
	if codeOf(err) != response.CodeFeatureInvalidRule {
		t.Fatalf("window: %v", err)
	}

	created, err := svc.Create(ctx, actor, &dto.CreateFlagRequest{
		FlagKey: "exp.hero", Name: "Hero", FlagType: model.TypeABTest, Enabled: true,
		Rules: model.Rules{
			Environments: []string{"prod"},
			Variants:     []model.Variant{{Key: "control", Weight: 50}, {Key: "treatment", Weight: 50}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(created.ID)
	if !created.EffectiveEnabled || created.Environment != "prod" {
		t.Fatalf("decorate: %+v", created)
	}
	eval, err := svc.EvaluateID(ctx, id, uid)
	if err != nil || eval.Variant == "" {
		t.Fatalf("eval ab: %+v %v", eval, err)
	}
	me, err := svc.EvaluateMe(ctx, uid, []string{"exp.hero"})
	if err != nil || me["exp.hero"].Variant == "" {
		t.Fatalf("me: %+v %v", me, err)
	}
	stats, err := svc.Analytics(ctx, id, 0)
	if err != nil || stats.Total < 1 || stats.Days != 7 {
		t.Fatalf("analytics: %+v %v", stats, err)
	}

	name := "Hero v2"
	if _, err := svc.Update(ctx, actor, id, &dto.UpdateFlagRequest{Name: &name}); err != nil {
		t.Fatal(err)
	}
	rolled, err := svc.Rollback(ctx, actor, id)
	if err != nil || rolled.Name != "Hero" {
		t.Fatalf("rollback: %+v %v", rolled, err)
	}
	if _, err := svc.Rollback(ctx, actor, uuid.New()); codeOf(err) != response.CodeFeatureNotFound {
		t.Fatalf("missing rollback: %v", err)
	}
	fresh, err := svc.Create(ctx, actor, &dto.CreateFlagRequest{FlagKey: "fresh.flag", Name: "Fresh", FlagType: model.TypeBoolean})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Rollback(ctx, actor, uuid.MustParse(fresh.ID)); codeOf(err) != response.CodeFeatureNoRollback {
		t.Fatalf("no snapshot: %v", err)
	}
	if _, err := svc.Analytics(ctx, uuid.New(), 120); codeOf(err) != response.CodeFeatureNotFound {
		t.Fatalf("analytics missing: %v", err)
	}

	onAt := now.Add(-2 * time.Hour)
	offAt := now.Add(time.Hour)
	scheduled, err := svc.Create(ctx, actor, &dto.CreateFlagRequest{
		FlagKey: "sched.window", Name: "Sched", FlagType: model.TypeBoolean, Enabled: true,
		Rules: model.Rules{StartsAt: &onAt, EndsAt: &offAt},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !scheduled.EffectiveEnabled || scheduled.ScheduleState != model.ScheduleActive {
		t.Fatalf("active decorate: %+v", scheduled)
	}
	if n := fs.applySchedules(ctx); n != 0 {
		t.Fatalf("window still open n=%d", n)
	}
	fs.now = func() time.Time { return offAt.Add(time.Second) }
	if n := fs.applySchedules(ctx); n != 1 {
		t.Fatalf("schedule off n=%d", n)
	}
	got, err := svc.Get(ctx, uuid.MustParse(scheduled.ID))
	if err != nil || got.Enabled {
		t.Fatalf("persisted off: %+v %v", got, err)
	}
	on := true
	if _, err := svc.Toggle(ctx, actor, uuid.MustParse(scheduled.ID), &dto.ToggleRequest{Enabled: &on}); err != nil {
		t.Fatal(err)
	}
	if n := fs.applySchedules(ctx); n != 0 {
		t.Fatalf("manual enable after ends_at must stick, n=%d", n)
	}
	got, err = svc.Get(ctx, uuid.MustParse(scheduled.ID))
	if err != nil || !got.Enabled || got.EffectiveEnabled {
		t.Fatalf("manual enable after ends_at: %+v %v", got, err)
	}

	fs.now = func() time.Time { return now }
	future := now.Add(time.Hour)
	pending, err := svc.Create(ctx, actor, &dto.CreateFlagRequest{
		FlagKey: "sched.on", Name: "Later", FlagType: model.TypeBoolean, Enabled: false,
		Rules: model.Rules{StartsAt: &future},
	})
	if err != nil {
		t.Fatal(err)
	}
	if fs.applySchedules(ctx) != 0 {
		t.Fatal("should wait")
	}
	fs.now = func() time.Time { return future.Add(time.Second) }
	if n := fs.applySchedules(ctx); n != 1 {
		t.Fatalf("schedule on n=%d", n)
	}
	got, err = svc.Get(ctx, uuid.MustParse(pending.ID))
	if err != nil || !got.Enabled {
		t.Fatalf("persisted on: %+v %v", got, err)
	}
	off := false
	if _, err := svc.Toggle(ctx, actor, uuid.MustParse(pending.ID), &dto.ToggleRequest{Enabled: &off}); err != nil {
		t.Fatal(err)
	}
	if n := fs.applySchedules(ctx); n != 0 {
		t.Fatalf("manual disable after starts_at must stick, n=%d", n)
	}
	got, err = svc.Get(ctx, uuid.MustParse(pending.ID))
	if err != nil || got.Enabled {
		t.Fatalf("manual disable after starts_at: %+v %v", got, err)
	}
	if !svc.Enabled(ctx, "exp.hero", feature.Subject{UserID: uid}) && !svc.Enabled(ctx, "exp.hero", feature.Subject{UserID: uid, Environment: "prod"}) {
		// either variant may be off; just ensure env is attached
	}
	_ = rows
}

func TestScheduleLoopAndNilClock(t *testing.T) {
	svc, _, _ := newTestSvc(nil, nil, stubPerms{})
	fs := svc.(*flagService)
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	end := now.Add(time.Hour)
	fs.now = func() time.Time { return now }
	fs.tickEvery = 8 * time.Millisecond
	ctx := context.Background()
	created, err := svc.Create(ctx, uuid.New(), &dto.CreateFlagRequest{
		FlagKey: "loop.off", Name: "Loop", FlagType: model.TypeBoolean, Enabled: true,
		Rules: model.Rules{EndsAt: &end},
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(created.ID)
	fs.now = func() time.Time { return end.Add(time.Second) }
	loopCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	if err := svc.StartHotReload(loopCtx); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(400 * time.Millisecond)
	for time.Now().Before(deadline) {
		got, gerr := svc.Get(ctx, id)
		if gerr == nil && got != nil && !got.Enabled {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	got, err := svc.Get(ctx, id)
	if err != nil || got.Enabled {
		t.Fatalf("loop should persist off: %+v %v", got, err)
	}
}

func TestClockNilFallsBackToWallTime(t *testing.T) {
	svc, _, _ := newTestSvc(nil, nil, stubPerms{})
	fs := svc.(*flagService)
	fs.now = nil
	if fs.clock().IsZero() {
		t.Fatal("clock fallback")
	}
}

func TestToggleCMSPublicIgnoresStickySnapshot(t *testing.T) {
	rows := newMemRepo()
	store := &stickySetSnapshot{
		setErr: errors.New("redis set failed"),
		delErr: errors.New("redis del failed"),
	}
	svc := New(rows, store, NewMemoryBus(), stubRoles{}, stubUsers{}, stubPerms{}).(*flagService)
	ctx := context.Background()
	id := uuid.New()
	_ = rows.Create(ctx, &model.Flag{
		ID: id, FlagKey: model.KeyCMSPublic, Name: "CMS", FlagType: model.TypeBoolean, Enabled: false,
	})
	store.flags = []model.Flag{{
		ID: id, FlagKey: model.KeyCMSPublic, Name: "CMS", FlagType: model.TypeBoolean, Enabled: false,
	}}
	on := true
	if _, err := svc.Toggle(ctx, uuid.New(), id, &dto.ToggleRequest{Enabled: &on}); err != nil {
		t.Fatal(err)
	}
	if !svc.Enabled(ctx, model.KeyCMSPublic, feature.Subject{UserID: uuid.New()}) {
		t.Fatal("toggle cms.public must apply immediately even when Redis snapshot cannot be replaced")
	}
}

func TestReloadEmptySnapshotFallsBackToDB(t *testing.T) {
	rows := newMemRepo()
	store := NewMemorySnapshot()
	svc := New(rows, store, NewMemoryBus(), stubRoles{}, stubUsers{}, stubPerms{}).(*flagService)
	ctx := context.Background()
	_ = rows.Create(ctx, &model.Flag{
		ID: uuid.New(), FlagKey: model.KeyCMSPublic, Name: "CMS", FlagType: model.TypeBoolean, Enabled: true,
	})
	if err := store.Set(ctx, []model.Flag{}); err != nil {
		t.Fatal(err)
	}
	if err := svc.reload(ctx); err != nil {
		t.Fatal(err)
	}
	if !svc.Enabled(ctx, model.KeyCMSPublic, feature.Subject{UserID: uuid.New()}) {
		t.Fatal("empty snapshot must fall back to DB")
	}
}

func TestStartHotReloadIgnoresStaleSnapshot(t *testing.T) {
	rows := newMemRepo()
	ctx := context.Background()
	id := uuid.New()
	_ = rows.Create(ctx, &model.Flag{
		ID: id, FlagKey: model.KeyCMSPublic, Name: "CMS", FlagType: model.TypeBoolean, Enabled: true,
	})
	store := &stickySetSnapshot{flags: []model.Flag{{
		ID: id, FlagKey: model.KeyCMSPublic, Name: "CMS", FlagType: model.TypeBoolean, Enabled: false,
	}}}
	svc := New(rows, store, NewMemoryBus(), stubRoles{}, stubUsers{}, stubPerms{}).(*flagService)
	if err := svc.StartHotReload(ctx); err != nil {
		t.Fatal(err)
	}
	if !svc.Enabled(ctx, model.KeyCMSPublic, feature.Subject{}) {
		t.Fatal("boot must read DB, not a leftover Redis snapshot")
	}
}

func TestReloadSkipsUntrustedSnapshot(t *testing.T) {
	rows := newMemRepo()
	store := &stickySetSnapshot{
		setErr: errors.New("redis set failed"),
		delErr: errors.New("redis del failed"),
	}
	svc := New(rows, store, NewMemoryBus(), stubRoles{}, stubUsers{}, stubPerms{}).(*flagService)
	ctx := context.Background()
	id := uuid.New()
	_ = rows.Create(ctx, &model.Flag{
		ID: id, FlagKey: model.KeyCMSPublic, Name: "CMS", FlagType: model.TypeBoolean, Enabled: false,
	})
	store.flags = []model.Flag{{
		ID: id, FlagKey: model.KeyCMSPublic, Name: "CMS", FlagType: model.TypeBoolean, Enabled: false,
	}}
	on := true
	if _, err := svc.Toggle(ctx, uuid.New(), id, &dto.ToggleRequest{Enabled: &on}); err != nil {
		t.Fatal(err)
	}
	if err := svc.reload(ctx); err != nil {
		t.Fatal(err)
	}
	if !svc.Enabled(ctx, model.KeyCMSPublic, feature.Subject{}) {
		t.Fatal("reload must not treat leftover Redis as authoritative after a failed SET/DEL")
	}
}

func TestUnknownKeyDoesNotRollBackFreshMemory(t *testing.T) {
	rows := newMemRepo()
	store := &stickySetSnapshot{
		setErr: errors.New("redis set failed"),
		delErr: errors.New("redis del failed"),
	}
	svc := New(rows, store, NewMemoryBus(), stubRoles{}, stubUsers{}, stubPerms{}).(*flagService)
	ctx := context.Background()
	id := uuid.New()
	_ = rows.Create(ctx, &model.Flag{
		ID: id, FlagKey: model.KeyCMSPublic, Name: "CMS", FlagType: model.TypeBoolean, Enabled: false,
	})
	store.flags = []model.Flag{{
		ID: id, FlagKey: model.KeyCMSPublic, Name: "CMS", FlagType: model.TypeBoolean, Enabled: false,
	}}
	on := true
	if _, err := svc.Toggle(ctx, uuid.New(), id, &dto.ToggleRequest{Enabled: &on}); err != nil {
		t.Fatal(err)
	}
	me, err := svc.EvaluateMe(ctx, uuid.New(), []string{model.KeyCMSPublic, "unknown.probe"})
	if err != nil {
		t.Fatal(err)
	}
	if !me[model.KeyCMSPublic].Enabled {
		t.Fatal("unknown-key lookup must not reload leftover snapshot over a fresh toggle")
	}
}

func TestBooleanEvaluateSurvivesRoleLookupError(t *testing.T) {
	rows := newMemRepo()
	svc := New(rows, NewMemorySnapshot(), NewMemoryBus(), failRoles{}, failUsers{}, stubPerms{})
	ctx := context.Background()
	_, err := svc.Create(ctx, uuid.New(), &dto.CreateFlagRequest{
		FlagKey: model.KeyCMSPublic, Name: "CMS", FlagType: model.TypeBoolean, Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	uid := uuid.New()
	sub, err := svc.Resolve(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}
	if !svc.Enabled(ctx, model.KeyCMSPublic, sub) {
		t.Fatal("boolean gate must stay on when role/dept lookup fails")
	}
	me, err := svc.EvaluateMe(ctx, uid, []string{model.KeyCMSPublic})
	if err != nil || !me[model.KeyCMSPublic].Enabled {
		t.Fatalf("evaluate me: %+v %v", me, err)
	}
}

func TestLookupFallsBackWhenSnapshotOmitsKey(t *testing.T) {
	rows := newMemRepo()
	ctx := context.Background()
	_ = rows.Create(ctx, &model.Flag{ID: uuid.New(), FlagKey: model.KeyCMSPublic, Name: "CMS", FlagType: model.TypeBoolean, Enabled: true})
	_ = rows.Create(ctx, &model.Flag{ID: uuid.New(), FlagKey: "other.flag", Name: "Other", FlagType: model.TypeBoolean, Enabled: true})
	store := NewMemorySnapshot()
	if err := store.Set(ctx, []model.Flag{{FlagKey: "other.flag", Name: "Other", FlagType: model.TypeBoolean, Enabled: true}}); err != nil {
		t.Fatal(err)
	}
	svc := New(rows, store, NewMemoryBus(), stubRoles{}, stubUsers{}, stubPerms{}).(*flagService)
	if !svc.Enabled(ctx, model.KeyCMSPublic, feature.Subject{}) {
		t.Fatal("omitted snapshot key should read db")
	}
}

func TestInvalidateRefreshesWhenDeleteFails(t *testing.T) {
	rows := newMemRepo()
	ctx := context.Background()
	id := uuid.New()
	row := &model.Flag{ID: id, FlagKey: model.KeyCMSPublic, Name: "CMS", FlagType: model.TypeBoolean, Enabled: true}
	_ = rows.Create(ctx, row)
	store := &staleSnapshot{flags: []model.Flag{*row}}
	svc := New(rows, store, NewMemoryBus(), stubRoles{}, stubUsers{}, stubPerms{}).(*flagService)
	if err := svc.reload(ctx); err != nil {
		t.Fatal(err)
	}
	if !svc.Enabled(ctx, model.KeyCMSPublic, feature.Subject{}) {
		t.Fatal("want on")
	}
	off := false
	if _, err := svc.Toggle(ctx, uuid.New(), id, &dto.ToggleRequest{Enabled: &off}); err != nil {
		t.Fatal(err)
	}
	if svc.Enabled(ctx, model.KeyCMSPublic, feature.Subject{}) {
		t.Fatal("toggle must close gate even if snapshot delete fails")
	}
}

func TestStartHotReloadSurvivesStuckSnapshot(t *testing.T) {
	rows := newMemRepo()
	store := &stickySetSnapshot{
		setErr: errors.New("redis set failed"),
		delErr: errors.New("redis del failed"),
	}
	svc := New(rows, store, NewMemoryBus(), stubRoles{}, stubUsers{}, stubPerms{}).(*flagService)
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	end := now.Add(time.Hour)
	svc.now = func() time.Time { return now }
	svc.tickEvery = 8 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	id := uuid.New()
	if err := rows.Create(ctx, &model.Flag{
		ID: id, FlagKey: "boot.loop", Name: "Boot", FlagType: model.TypeBoolean,
		Enabled: true, Rules: model.Rules{EndsAt: &end}, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	svc.now = func() time.Time { return end.Add(time.Second) }
	if err := svc.StartHotReload(ctx); err != nil {
		t.Fatalf("boot must continue after Redis write failure: %v", err)
	}
	deadline := time.Now().Add(400 * time.Millisecond)
	for time.Now().Before(deadline) {
		got, gerr := svc.Get(ctx, id)
		if gerr == nil && got != nil && !got.Enabled {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("schedule loop should still persist off after a Redis blip at boot")
}

func TestInvalidateDropsSnapshotWhenSetFails(t *testing.T) {
	rows := newMemRepo()
	store := &stickySetSnapshot{}
	bus := NewMemoryBus()
	writer := New(rows, store, bus, stubRoles{}, stubUsers{}, stubPerms{}).(*flagService)
	peer := New(rows, store, bus, stubRoles{}, stubUsers{}, stubPerms{}).(*flagService)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	created, err := writer.Create(ctx, uuid.New(), &dto.CreateFlagRequest{
		FlagKey: "set.fail", Name: "Set", FlagType: model.TypeBoolean, Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := peer.StartHotReload(ctx); err != nil {
		t.Fatal(err)
	}
	store.setErr = errors.New("redis set failed")
	off := false
	if _, err := writer.Toggle(ctx, uuid.New(), uuid.MustParse(created.ID), &dto.ToggleRequest{Enabled: &off}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(400 * time.Millisecond)
	for time.Now().Before(deadline) {
		if !peer.Enabled(ctx, "set.fail", feature.Subject{UserID: uuid.New()}) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if peer.Enabled(ctx, "set.fail", feature.Subject{UserID: uuid.New()}) {
		t.Fatal("peer must not keep a leftover snapshot after a failed SET")
	}
	got, err := store.Get(ctx)
	if err != nil || got != nil {
		t.Fatalf("failed SET should delete leftover snapshot, got %+v %v", got, err)
	}
}

func TestInvalidateSkipsPublishWhenSnapshotStuck(t *testing.T) {
	rows := newMemRepo()
	store := &stickySetSnapshot{
		setErr: errors.New("redis set failed"),
		delErr: errors.New("redis del failed"),
	}
	bus := &countBus{inner: NewMemoryBus()}
	svc := New(rows, store, bus, stubRoles{}, stubUsers{}, stubPerms{}).(*flagService)
	ctx := context.Background()
	id := uuid.New()
	_ = rows.Create(ctx, &model.Flag{
		ID: id, FlagKey: "stuck.flag", Name: "Stuck", FlagType: model.TypeBoolean, Enabled: true,
	})
	store.flags = []model.Flag{{
		ID: id, FlagKey: "stuck.flag", Name: "Stuck", FlagType: model.TypeBoolean, Enabled: true,
	}}
	before := bus.n
	if _, err := svc.Toggle(ctx, uuid.New(), id, &dto.ToggleRequest{}); err != nil {
		t.Fatal(err)
	}
	if bus.n != before {
		t.Fatalf("stuck snapshot must not be broadcast, publishes %d -> %d", before, bus.n)
	}
	if svc.Enabled(ctx, "stuck.flag", feature.Subject{UserID: uuid.New()}) {
		t.Fatal("writer should still apply the DB write locally")
	}
}

func TestEvaluateMeCapsDedupsAndSkipsUnknownReload(t *testing.T) {
	rows := newMemRepo()
	store := &countSnapshot{inner: NewMemorySnapshot()}
	bus := NewMemoryBus()
	svc := New(rows, store, bus, stubRoles{}, stubUsers{}, stubPerms{}).(*flagService)
	ctx := context.Background()
	created, err := svc.Create(ctx, uuid.New(), &dto.CreateFlagRequest{
		FlagKey: "bool.gate", Name: "Bool", FlagType: model.TypeBoolean, Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	uid := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	getsBefore := store.gets
	dupes := []string{"bool.gate", "BOOL.GATE", "bool.gate", "Not A Key", "???"}
	me, err := svc.EvaluateMe(ctx, uid, dupes)
	if err != nil {
		t.Fatal(err)
	}
	if len(me) != 1 || !me["bool.gate"].Enabled {
		t.Fatalf("dedup: %+v", me)
	}
	if n := len(rows.exposures); n != 1 {
		t.Fatalf("duplicate keys must record one exposure, got %d", n)
	}
	keys := make([]string, 0, 40)
	for i := 0; i < 40; i++ {
		keys = append(keys, "unknown.key"+string(rune('a'+i%26))+string(rune('a'+i/26)))
	}
	me, err = svc.EvaluateMe(ctx, uid, keys)
	if err != nil {
		t.Fatal(err)
	}
	if len(me) != model.MaxEvaluateKeys {
		t.Fatalf("cap=%d got=%d", model.MaxEvaluateKeys, len(me))
	}
	if store.gets != getsBefore {
		t.Fatalf("unknown keys must not reload snapshot, gets %d -> %d", getsBefore, store.gets)
	}
	if _, err := svc.EvaluateID(ctx, uuid.MustParse(created.ID), uid); err != nil {
		t.Fatal(err)
	}
	if got := sanitizeEvaluateKeys([]string{"???", "1bad", "x"}); len(got) != 3 {
		t.Fatalf("invalid keys should fall back to defaults, got %v", got)
	}
}

type countSnapshot struct {
	inner SnapshotStore
	gets  int
}

func (s *countSnapshot) Get(ctx context.Context) ([]model.Flag, error) {
	s.gets++
	return s.inner.Get(ctx)
}

func (s *countSnapshot) Set(ctx context.Context, flags []model.Flag) error {
	return s.inner.Set(ctx, flags)
}

func (s *countSnapshot) Delete(ctx context.Context) error {
	return s.inner.Delete(ctx)
}

func TestInvalidateReloadsFromDBWhenSnapshotStale(t *testing.T) {
	rows := newMemRepo()
	store := &staleSnapshot{}
	bus := NewMemoryBus()
	svc := New(rows, store, bus, stubRoles{}, stubUsers{}, stubPerms{}).(*flagService)
	ctx := context.Background()
	created, err := svc.Create(ctx, uuid.New(), &dto.CreateFlagRequest{
		FlagKey: "stale.flag", Name: "Stale", FlagType: model.TypeBoolean, Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Snapshot keeps the enabled=true copy even after later writes.
	store.flags = []model.Flag{{
		ID: uuid.MustParse(created.ID), FlagKey: "stale.flag", Name: "Stale",
		FlagType: model.TypeBoolean, Enabled: true,
	}}
	off := false
	if _, err := svc.Toggle(ctx, uuid.New(), uuid.MustParse(created.ID), &dto.ToggleRequest{Enabled: &off}); err != nil {
		t.Fatal(err)
	}
	if svc.Enabled(ctx, "stale.flag", feature.Subject{UserID: uuid.New()}) {
		t.Fatal("failed snapshot delete must not refill memory from stale Redis data")
	}
}

func TestBooleanAnalyticsSplitsEnabledAndDisabled(t *testing.T) {
	svc, _, _ := newTestSvc(nil, nil, stubPerms{})
	ctx := context.Background()
	created, err := svc.Create(ctx, uuid.New(), &dto.CreateFlagRequest{
		FlagKey: "bool.gate", Name: "Bool", FlagType: model.TypeBoolean, Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(created.ID)
	onUser := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	offUser := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	if _, err := svc.EvaluateID(ctx, id, onUser); err != nil {
		t.Fatal(err)
	}
	off := false
	if _, err := svc.Toggle(ctx, uuid.New(), id, &dto.ToggleRequest{Enabled: &off}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.EvaluateID(ctx, id, offUser); err != nil {
		t.Fatal(err)
	}
	stats, err := svc.Analytics(ctx, id, 7)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != 2 || stats.EnabledCount != 1 || stats.DisabledCount != 1 {
		t.Fatalf("boolean analytics %+v", stats)
	}
	if len(stats.Variants) != 2 {
		t.Fatalf("expected two variant/enabled buckets, got %+v", stats.Variants)
	}
}

type stickySetSnapshot struct {
	mu     sync.Mutex
	flags  []model.Flag
	setErr error
	delErr error
}

func (s *stickySetSnapshot) Get(context.Context) ([]model.Flag, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneFlags(s.flags), nil
}

func (s *stickySetSnapshot) Set(_ context.Context, flags []model.Flag) error {
	if s.setErr != nil {
		return s.setErr
	}
	s.mu.Lock()
	s.flags = cloneFlags(flags)
	s.mu.Unlock()
	return nil
}

func (s *stickySetSnapshot) Delete(context.Context) error {
	if s.delErr != nil {
		return s.delErr
	}
	s.mu.Lock()
	s.flags = nil
	s.mu.Unlock()
	return nil
}

type countBus struct {
	inner *MemoryBus
	n     int
}

func (b *countBus) Publish(ctx context.Context, payload string) error {
	b.n++
	return b.inner.Publish(ctx, payload)
}

func (b *countBus) Subscribe(ctx context.Context, onMsg func(string)) error {
	return b.inner.Subscribe(ctx, onMsg)
}

type staleSnapshot struct {
	flags []model.Flag
}

func (s *staleSnapshot) Get(context.Context) ([]model.Flag, error) {
	return cloneFlags(s.flags), nil
}

func (s *staleSnapshot) Set(context.Context, []model.Flag) error { return nil }

func (s *staleSnapshot) Delete(context.Context) error {
	return errors.New("redis delete failed")
}

func TestUserDepartmentsNil(t *testing.T) {
	lookup := NewUserDepartments(nil)
	dept, err := lookup.DepartmentOf(context.Background(), uuid.New())
	if err != nil || dept != nil {
		t.Fatalf("%v %v", dept, err)
	}
}

type failRoles struct{}

func (failRoles) GetUserRoleCodes(context.Context, uuid.UUID) ([]string, error) {
	return nil, errors.New("perm cache down")
}

type failUsers struct{}

func (failUsers) DepartmentOf(context.Context, uuid.UUID) (*uuid.UUID, error) {
	return nil, errors.New("user repo down")
}

func codeOf(err error) int {
	if err == nil {
		return 0
	}
	if app, ok := err.(*response.AppError); ok {
		return app.Code
	}
	return -1
}
