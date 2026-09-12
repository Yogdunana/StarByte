package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/feature"
	"github.com/Yogdunana/StarByte/backend/internal/feature/dto"
	"github.com/Yogdunana/StarByte/backend/internal/feature/model"
	"github.com/Yogdunana/StarByte/backend/internal/feature/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_.]{1,99}$`)

type RoleLookup interface {
	GetUserRoleCodes(ctx context.Context, userID uuid.UUID) ([]string, error)
}

type UserLookup interface {
	DepartmentOf(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error)
}

type PermLookup interface {
	GetUserPermissionsAndSuperAdmin(ctx context.Context, userID uuid.UUID) ([]string, bool, error)
}

type Service interface {
	List(ctx context.Context, q dto.ListQuery) ([]dto.FlagResponse, int64, error)
	Get(ctx context.Context, id uuid.UUID) (*dto.FlagResponse, error)
	Create(ctx context.Context, actor uuid.UUID, req *dto.CreateFlagRequest) (*dto.FlagResponse, error)
	Update(ctx context.Context, actor, id uuid.UUID, req *dto.UpdateFlagRequest) (*dto.FlagResponse, error)
	Toggle(ctx context.Context, actor, id uuid.UUID, req *dto.ToggleRequest) (*dto.FlagResponse, error)
	Delete(ctx context.Context, actor, id uuid.UUID) error
	EvaluateID(ctx context.Context, id, userID uuid.UUID) (*dto.EvaluateResponse, error)
	EvaluateMe(ctx context.Context, userID uuid.UUID, keys []string) (map[string]dto.EvaluateResponse, error)
	Enabled(ctx context.Context, key string, sub feature.Subject) bool
	Resolve(ctx context.Context, userID uuid.UUID) (feature.Subject, error)
	ListAudits(ctx context.Context, q dto.AuditQuery) ([]dto.AuditResponse, int64, error)
	Rollback(ctx context.Context, actor, id uuid.UUID) (*dto.FlagResponse, error)
	Analytics(ctx context.Context, id uuid.UUID, days int) (*dto.AnalyticsResponse, error)
	StartHotReload(ctx context.Context) error
}

type flagService struct {
	rows            repo.Repository
	store           SnapshotStore
	bus             Broadcaster
	roles           RoleLookup
	users           UserLookup
	perms           PermLookup
	env             string
	now             func() time.Time
	tickEvery       time.Duration
	mu              sync.RWMutex
	memory          map[string]model.Flag
	loaded          bool
	snapshotTrusted bool
}

func New(rows repo.Repository, store SnapshotStore, bus Broadcaster, roles RoleLookup, users UserLookup, perms PermLookup) Service {
	if store == nil {
		store = NewMemorySnapshot()
	}
	if bus == nil {
		bus = NewMemoryBus()
	}
	return &flagService{
		rows: rows, store: store, bus: bus, roles: roles, users: users, perms: perms,
		env: normalizeEnv(os.Getenv("APP_ENV")), now: time.Now, tickEvery: 30 * time.Second,
		memory: map[string]model.Flag{},
	}
}

func (s *flagService) StartHotReload(ctx context.Context) error {
	if err := s.reloadFromDB(ctx); err != nil {
		return err
	}
	s.startScheduleLoop(ctx)
	return s.bus.Subscribe(ctx, func(string) {
		if err := s.reload(context.Background()); err != nil {
			logCacheErr("reload", err)
		}
	})
}

func (s *flagService) List(ctx context.Context, q dto.ListQuery) ([]dto.FlagResponse, int64, error) {
	rows, total, err := s.rows.List(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("list flags: %w", err)
	}
	out := make([]dto.FlagResponse, 0, len(rows))
	for i := range rows {
		out = append(out, s.decorate(&rows[i]))
	}
	return out, total, nil
}

func (s *flagService) Get(ctx context.Context, id uuid.UUID) (*dto.FlagResponse, error) {
	row, err := s.require(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := s.decorate(row)
	return &resp, nil
}

func (s *flagService) Create(ctx context.Context, actor uuid.UUID, req *dto.CreateFlagRequest) (*dto.FlagResponse, error) {
	key := strings.TrimSpace(req.FlagKey)
	if !keyPattern.MatchString(key) {
		return nil, response.NewError(response.CodeFeatureInvalidKey, "开关键须为小写字母开头，仅含字母数字._")
	}
	if !model.ValidType(req.FlagType) {
		return nil, response.NewError(response.CodeFeatureInvalidType, "不支持的开关类型")
	}
	if err := validateRules(req.FlagType, req.Rules); err != nil {
		return nil, err
	}
	exist, err := s.rows.GetByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("check key: %w", err)
	}
	if exist != nil {
		return nil, response.NewError(response.CodeFeatureKeyExists, "开关键已存在")
	}
	now := s.clock()
	row := &model.Flag{
		ID: uuid.New(), FlagKey: key, Name: strings.TrimSpace(req.Name), Description: req.Description,
		FlagType: req.FlagType, Enabled: req.Enabled, GroupName: strings.TrimSpace(req.GroupName),
		Priority: req.Priority, Rules: req.Rules, CreatedBy: &actor, UpdatedBy: &actor,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := s.rows.Create(ctx, row); err != nil {
		return nil, fmt.Errorf("create flag: %w", err)
	}
	s.audit(ctx, row, actor, model.ActionCreate, nil, row, "")
	s.invalidate(ctx)
	resp := s.decorate(row)
	return &resp, nil
}

func (s *flagService) Update(ctx context.Context, actor, id uuid.UUID, req *dto.UpdateFlagRequest) (*dto.FlagResponse, error) {
	row, err := s.require(ctx, id)
	if err != nil {
		return nil, err
	}
	before := *row
	if req.Name != nil {
		row.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		row.Description = *req.Description
	}
	if req.FlagType != nil {
		if !model.ValidType(*req.FlagType) {
			return nil, response.NewError(response.CodeFeatureInvalidType, "不支持的开关类型")
		}
		row.FlagType = *req.FlagType
	}
	if req.Enabled != nil {
		row.Enabled = *req.Enabled
	}
	if req.GroupName != nil {
		row.GroupName = strings.TrimSpace(*req.GroupName)
	}
	if req.Priority != nil {
		row.Priority = *req.Priority
	}
	if req.Rules != nil {
		row.Rules = *req.Rules
	}
	if err := validateRules(row.FlagType, row.Rules); err != nil {
		return nil, err
	}
	row.UpdatedBy = &actor
	row.UpdatedAt = s.clock()
	if err := s.rows.Update(ctx, row); err != nil {
		return nil, fmt.Errorf("update flag: %w", err)
	}
	s.audit(ctx, row, actor, model.ActionUpdate, &before, row, "")
	s.invalidate(ctx)
	resp := s.decorate(row)
	return &resp, nil
}

func (s *flagService) Toggle(ctx context.Context, actor, id uuid.UUID, req *dto.ToggleRequest) (*dto.FlagResponse, error) {
	row, err := s.require(ctx, id)
	if err != nil {
		return nil, err
	}
	before := *row
	if req != nil && req.Enabled != nil {
		row.Enabled = *req.Enabled
	} else {
		row.Enabled = !row.Enabled
	}
	reason := ""
	if req != nil {
		reason = strings.TrimSpace(req.Reason)
	}
	row.UpdatedBy = &actor
	row.UpdatedAt = s.clock()
	if err := s.rows.Update(ctx, row); err != nil {
		return nil, fmt.Errorf("toggle flag: %w", err)
	}
	s.audit(ctx, row, actor, model.ActionToggle, &before, row, reason)
	s.invalidate(ctx)
	resp := s.decorate(row)
	return &resp, nil
}

func (s *flagService) Delete(ctx context.Context, actor, id uuid.UUID) error {
	row, err := s.require(ctx, id)
	if err != nil {
		return err
	}
	if row.IsSystem {
		return response.NewError(response.CodeFeatureProtected, "系统预置开关不可删除")
	}
	if err := s.rows.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete flag: %w", err)
	}
	s.audit(ctx, row, actor, model.ActionDelete, row, nil, "")
	s.invalidate(ctx)
	return nil
}

func (s *flagService) EvaluateID(ctx context.Context, id, userID uuid.UUID) (*dto.EvaluateResponse, error) {
	row, err := s.require(ctx, id)
	if err != nil {
		return nil, err
	}
	sub, err := s.Resolve(ctx, userID)
	if err != nil {
		return nil, err
	}
	got := feature.Evaluate(row, sub)
	s.recordExposure(ctx, row, userID, got)
	return toEval(row, got), nil
}

func (s *flagService) EvaluateMe(ctx context.Context, userID uuid.UUID, keys []string) (map[string]dto.EvaluateResponse, error) {
	sub, err := s.Resolve(ctx, userID)
	if err != nil {
		return nil, err
	}
	staff := s.announcementStaff(ctx, userID)
	keys = sanitizeEvaluateKeys(keys)
	out := make(map[string]dto.EvaluateResponse, len(keys))
	for _, key := range keys {
		if key == model.KeyAnnouncementFeed && staff {
			out[key] = dto.EvaluateResponse{Key: key, Enabled: true, Reason: "staff_bypass", Type: model.TypeBoolean}
			continue
		}
		flag := s.lookup(ctx, key)
		got := feature.Evaluate(flag, sub)
		if flag != nil {
			s.recordExposure(ctx, flag, userID, got)
			out[key] = *toEval(flag, got)
			continue
		}
		out[key] = dto.EvaluateResponse{Key: key, Enabled: got.Enabled, Reason: got.Reason}
	}
	return out, nil
}

func sanitizeEvaluateKeys(keys []string) []string {
	if len(keys) == 0 {
		return []string{model.KeyCMSPublic, model.KeyAnnouncementFeed, model.KeyMembershipPortal}
	}
	seen := make(map[string]struct{}, model.MaxEvaluateKeys)
	out := make([]string, 0, min(len(keys), model.MaxEvaluateKeys))
	for _, key := range keys {
		key = strings.ToLower(strings.TrimSpace(key))
		if key == "" || !keyPattern.MatchString(key) {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
		if len(out) >= model.MaxEvaluateKeys {
			break
		}
	}
	if len(out) == 0 {
		return []string{model.KeyCMSPublic, model.KeyAnnouncementFeed, model.KeyMembershipPortal}
	}
	return out
}

func (s *flagService) Enabled(ctx context.Context, key string, sub feature.Subject) bool {
	if sub.Environment == "" {
		sub.Environment = s.env
	}
	if sub.Now.IsZero() {
		sub.Now = s.clock()
	}
	return feature.Evaluate(s.lookup(ctx, key), sub).Enabled
}

func (s *flagService) Resolve(ctx context.Context, userID uuid.UUID) (feature.Subject, error) {
	sub := feature.Subject{UserID: userID, Environment: s.env, Now: s.clock()}
	if userID == uuid.Nil {
		return sub, nil
	}
	if s.roles != nil {
		codes, err := s.roles.GetUserRoleCodes(ctx, userID)
		if err != nil {
			// Boolean / percentage gates do not need roles; fail closed for role_dept.
			logCacheErr("resolve-roles", err)
		} else {
			sub.RoleCodes = codes
		}
	}
	if s.users != nil {
		dept, err := s.users.DepartmentOf(ctx, userID)
		if err != nil {
			logCacheErr("resolve-dept", err)
		} else if dept != nil && *dept != uuid.Nil {
			sub.DepartmentIDs = []uuid.UUID{*dept}
		}
	}
	return sub, nil
}

func (s *flagService) ListAudits(ctx context.Context, q dto.AuditQuery) ([]dto.AuditResponse, int64, error) {
	rows, total, err := s.rows.ListAudits(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("list audits: %w", err)
	}
	out := make([]dto.AuditResponse, 0, len(rows))
	for i := range rows {
		out = append(out, dto.ToAudit(&rows[i]))
	}
	return out, total, nil
}

func (s *flagService) require(ctx context.Context, id uuid.UUID) (*model.Flag, error) {
	row, err := s.rows.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get flag: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeFeatureNotFound, "特性开关不存在")
	}
	return row, nil
}

func (s *flagService) lookup(ctx context.Context, key string) *model.Flag {
	s.mu.RLock()
	if f, ok := s.memory[key]; ok {
		s.mu.RUnlock()
		cp := f
		return &cp
	}
	ready := s.loaded
	s.mu.RUnlock()
	if !ready {
		// Never seed from Redis here — leftover snapshot after a failed
		// SET/DEL must not overwrite a later DB-backed view.
		if err := s.reloadFromDB(ctx); err != nil {
			logCacheErr("lookup-reload", err)
		} else {
			s.mu.RLock()
			if f, ok := s.memory[key]; ok {
				s.mu.RUnlock()
				cp := f
				return &cp
			}
			s.mu.RUnlock()
		}
	}
	row, err := s.rows.GetByKey(ctx, key)
	if err != nil || row == nil {
		return nil
	}
	return row
}

func (s *flagService) reload(ctx context.Context) error {
	s.mu.RLock()
	trusted := s.snapshotTrusted
	s.mu.RUnlock()
	// After SET+DEL both fail, Redis still holds a leftover generation.
	// Do not Get it — that would roll gates back to the stale snapshot.
	if !trusted {
		return s.reloadFromDB(ctx)
	}
	flags, err := s.store.Get(ctx)
	if err != nil {
		logCacheErr("snapshot-get", err)
		return s.reloadFromDB(ctx)
	}
	if len(flags) == 0 {
		return s.reloadFromDB(ctx)
	}
	s.replaceMemory(flags)
	return nil
}

func (s *flagService) reloadFromDB(ctx context.Context) error {
	_, err := s.loadFromDB(ctx)
	return err
}

// loadFromDB fills memory from Postgres. snapshotOK is false when Redis still
// holds a leftover generation (SET and DEL both failed). Boot treats that as
// non-fatal; invalidate must not publish in that case.
func (s *flagService) loadFromDB(ctx context.Context) (snapshotOK bool, err error) {
	flags, err := s.rows.ListAll(ctx)
	if err != nil {
		return false, err
	}
	ok := s.syncSnapshot(ctx, flags)
	s.replaceMemory(flags)
	return ok, nil
}

func (s *flagService) syncSnapshot(ctx context.Context, flags []model.Flag) bool {
	if err := s.store.Set(ctx, flags); err != nil {
		logCacheErr("snapshot-set", err)
		if delErr := s.store.Delete(ctx); delErr != nil {
			logCacheErr("snapshot-del", delErr)
			s.mu.Lock()
			s.snapshotTrusted = false
			s.mu.Unlock()
			return false
		}
	}
	s.mu.Lock()
	s.snapshotTrusted = true
	s.mu.Unlock()
	return true
}

func (s *flagService) replaceMemory(flags []model.Flag) {
	next := make(map[string]model.Flag, len(flags))
	for i := range flags {
		next[flags[i].FlagKey] = flags[i]
	}
	s.mu.Lock()
	s.memory = next
	s.loaded = true
	s.mu.Unlock()
}

func (s *flagService) invalidate(ctx context.Context) {
	// Rebuild from DB first so Set overwrites a stuck snapshot before peers Get.
	ok, err := s.loadFromDB(ctx)
	if err != nil {
		logCacheErr("reload-db", err)
		if delErr := s.store.Delete(ctx); delErr != nil {
			logCacheErr("snapshot-del", delErr)
			return
		}
		s.mu.Lock()
		s.memory = map[string]model.Flag{}
		s.loaded = false
		s.mu.Unlock()
	} else if !ok {
		// Local memory is fresh; leftover snapshot is still readable — do not
		// tell peers to Get it.
		return
	}
	logCacheErr("publish", s.bus.Publish(ctx, time.Now().UTC().Format(time.RFC3339Nano)))
}

func (s *flagService) audit(ctx context.Context, flag *model.Flag, actor uuid.UUID, action string, before, after *model.Flag, reason string) {
	row := &model.Audit{
		ID: uuid.New(), FlagKey: flag.FlagKey, Action: action, ActorID: &actor, Reason: reason, CreatedAt: time.Now(),
	}
	if flag.ID != uuid.Nil {
		id := flag.ID
		row.FlagID = &id
	}
	if before != nil {
		row.BeforeJSON, _ = json.Marshal(before)
	}
	if after != nil {
		row.AfterJSON, _ = json.Marshal(after)
	}
	if err := s.rows.CreateAudit(ctx, row); err != nil {
		logCacheErr("audit", err)
	}
}

func (s *flagService) announcementStaff(ctx context.Context, userID uuid.UUID) bool {
	if s.perms == nil || userID == uuid.Nil {
		return false
	}
	perms, isSuper, err := s.perms.GetUserPermissionsAndSuperAdmin(ctx, userID)
	if err != nil {
		return false
	}
	if isSuper {
		return true
	}
	for _, p := range perms {
		switch p {
		case "announcement:create", "announcement:update", "announcement:delete", "announcement:publish", "announcement:manage":
			return true
		}
	}
	return false
}

func validateRules(flagType string, rules model.Rules) error {
	if err := validateSharedRules(rules); err != nil {
		return err
	}
	switch flagType {
	case model.TypePercentage:
		if rules.Percent < 0 || rules.Percent > 100 {
			return response.NewError(response.CodeFeatureInvalidRule, "百分比须在 0-100")
		}
	case model.TypeUserAllowlist:
		for _, id := range rules.UserIDs {
			if strings.TrimSpace(id) == "" {
				continue
			}
			if _, err := uuid.Parse(strings.TrimSpace(id)); err != nil {
				return response.NewError(response.CodeFeatureInvalidRule, "用户名单含无效 UUID")
			}
		}
	case model.TypeRoleDept:
		for _, id := range rules.DepartmentIDs {
			if strings.TrimSpace(id) == "" {
				continue
			}
			if _, err := uuid.Parse(strings.TrimSpace(id)); err != nil {
				return response.NewError(response.CodeFeatureInvalidRule, "部门名单含无效 UUID")
			}
		}
	case model.TypeABTest:
		if err := validateAB(rules.Variants); err != nil {
			return err
		}
	}
	return nil
}

func validateSharedRules(rules model.Rules) error {
	seen := map[string]struct{}{}
	for _, env := range rules.Environments {
		env = strings.ToLower(strings.TrimSpace(env))
		if env == "" {
			continue
		}
		if !model.ValidEnv(env) {
			return response.NewError(response.CodeFeatureInvalidRule, "环境仅支持 dev/test/prod")
		}
		if _, ok := seen[env]; ok {
			return response.NewError(response.CodeFeatureInvalidRule, "环境列表重复")
		}
		seen[env] = struct{}{}
	}
	if rules.StartsAt != nil && rules.EndsAt != nil && !rules.StartsAt.Before(*rules.EndsAt) {
		return response.NewError(response.CodeFeatureInvalidRule, "定时上线须早于下线")
	}
	return nil
}

func validateAB(variants []model.Variant) error {
	if len(variants) < 2 {
		return response.NewError(response.CodeFeatureInvalidRule, "AB 测试至少需要两个变体")
	}
	keys := map[string]struct{}{}
	total := 0
	for _, v := range variants {
		key := strings.TrimSpace(v.Key)
		if key == "" {
			return response.NewError(response.CodeFeatureInvalidRule, "变体 key 不能为空")
		}
		if _, ok := keys[strings.ToLower(key)]; ok {
			return response.NewError(response.CodeFeatureInvalidRule, "变体 key 重复")
		}
		keys[strings.ToLower(key)] = struct{}{}
		if v.Weight < 0 {
			return response.NewError(response.CodeFeatureInvalidRule, "变体权重不能为负")
		}
		total += v.Weight
	}
	if total <= 0 {
		return response.NewError(response.CodeFeatureInvalidRule, "变体权重之和须大于 0")
	}
	return nil
}

func (s *flagService) Rollback(ctx context.Context, actor, id uuid.UUID) (*dto.FlagResponse, error) {
	row, err := s.require(ctx, id)
	if err != nil {
		return nil, err
	}
	snap, err := s.rows.LatestMutableAudit(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("latest audit: %w", err)
	}
	if snap == nil || len(snap.BeforeJSON) == 0 {
		return nil, response.NewError(response.CodeFeatureNoRollback, "没有可回滚的审计快照")
	}
	var prev model.Flag
	if err := json.Unmarshal(snap.BeforeJSON, &prev); err != nil {
		return nil, response.NewError(response.CodeFeatureNoRollback, "审计快照无法解析")
	}
	if err := validateRules(prev.FlagType, prev.Rules); err != nil {
		return nil, err
	}
	before := *row
	row.Name = prev.Name
	row.Description = prev.Description
	row.FlagType = prev.FlagType
	row.Enabled = prev.Enabled
	row.GroupName = prev.GroupName
	row.Priority = prev.Priority
	row.Rules = prev.Rules
	row.UpdatedBy = &actor
	row.UpdatedAt = s.clock()
	if err := s.rows.Update(ctx, row); err != nil {
		return nil, fmt.Errorf("rollback flag: %w", err)
	}
	s.audit(ctx, row, actor, model.ActionRollback, &before, row, snap.Action)
	s.invalidate(ctx)
	resp := s.decorate(row)
	return &resp, nil
}

func (s *flagService) Analytics(ctx context.Context, id uuid.UUID, days int) (*dto.AnalyticsResponse, error) {
	row, err := s.require(ctx, id)
	if err != nil {
		return nil, err
	}
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}
	since := s.clock().Add(-time.Duration(days) * 24 * time.Hour)
	buckets, err := s.rows.SummarizeExposures(ctx, row.FlagKey, since)
	if err != nil {
		return nil, fmt.Errorf("summarize exposures: %w", err)
	}
	out := &dto.AnalyticsResponse{FlagKey: row.FlagKey, Days: days, Variants: make([]dto.VariantCount, 0, len(buckets))}
	for _, b := range buckets {
		out.Variants = append(out.Variants, dto.VariantCount{Variant: b.Variant, Count: b.Count, Enabled: b.Enabled})
		out.Total += b.Count
		if b.Enabled {
			out.EnabledCount += b.Count
		} else {
			out.DisabledCount += b.Count
		}
	}
	return out, nil
}

func (s *flagService) decorate(flag *model.Flag) dto.FlagResponse {
	out := dto.ToFlag(flag)
	now := s.clock()
	out.EffectiveEnabled = feature.MasterOn(flag, s.env, now)
	out.ScheduleState = feature.ScheduleState(flag, now)
	out.Environment = s.env
	return out
}

func toEval(flag *model.Flag, got feature.Result) *dto.EvaluateResponse {
	typ := ""
	key := ""
	if flag != nil {
		typ = flag.FlagType
		key = flag.FlagKey
	}
	return &dto.EvaluateResponse{Key: key, Enabled: got.Enabled, Reason: got.Reason, Type: typ, Variant: got.Variant}
}

func (s *flagService) recordExposure(ctx context.Context, flag *model.Flag, userID uuid.UUID, got feature.Result) {
	if flag == nil || userID == uuid.Nil {
		return
	}
	id := flag.ID
	uid := userID
	row := &model.Exposure{
		ID: uuid.New(), FlagID: &id, FlagKey: flag.FlagKey, UserID: &uid,
		Variant: got.Variant, Enabled: got.Enabled, Reason: got.Reason,
		Environment: s.env, CreatedAt: s.clock(),
	}
	if err := s.rows.CreateExposure(ctx, row); err != nil {
		logCacheErr("exposure", err)
	}
}

func (s *flagService) clock() time.Time {
	if s.now == nil {
		return time.Now().UTC()
	}
	return s.now().UTC()
}

func normalizeEnv(env string) string {
	return strings.ToLower(strings.TrimSpace(env))
}
