package service

import (
	"context"
	"encoding/json"
	"fmt"
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
	StartHotReload(ctx context.Context) error
}

type flagService struct {
	rows   repo.Repository
	store  SnapshotStore
	bus    Broadcaster
	roles  RoleLookup
	users  UserLookup
	perms  PermLookup
	mu     sync.RWMutex
	memory map[string]model.Flag
}

func New(rows repo.Repository, store SnapshotStore, bus Broadcaster, roles RoleLookup, users UserLookup, perms PermLookup) Service {
	if store == nil {
		store = NewMemorySnapshot()
	}
	if bus == nil {
		bus = NewMemoryBus()
	}
	return &flagService{rows: rows, store: store, bus: bus, roles: roles, users: users, perms: perms, memory: map[string]model.Flag{}}
}

func (s *flagService) StartHotReload(ctx context.Context) error {
	if err := s.reload(ctx); err != nil {
		return err
	}
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
		out = append(out, dto.ToFlag(&rows[i]))
	}
	return out, total, nil
}

func (s *flagService) Get(ctx context.Context, id uuid.UUID) (*dto.FlagResponse, error) {
	row, err := s.require(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := dto.ToFlag(row)
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
	now := time.Now()
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
	resp := dto.ToFlag(row)
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
	row.UpdatedAt = time.Now()
	if err := s.rows.Update(ctx, row); err != nil {
		return nil, fmt.Errorf("update flag: %w", err)
	}
	s.audit(ctx, row, actor, model.ActionUpdate, &before, row, "")
	s.invalidate(ctx)
	resp := dto.ToFlag(row)
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
	row.UpdatedAt = time.Now()
	if err := s.rows.Update(ctx, row); err != nil {
		return nil, fmt.Errorf("toggle flag: %w", err)
	}
	s.audit(ctx, row, actor, model.ActionToggle, &before, row, reason)
	s.invalidate(ctx)
	resp := dto.ToFlag(row)
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
	return &dto.EvaluateResponse{Key: row.FlagKey, Enabled: got.Enabled, Reason: got.Reason, Type: row.FlagType}, nil
}

func (s *flagService) EvaluateMe(ctx context.Context, userID uuid.UUID, keys []string) (map[string]dto.EvaluateResponse, error) {
	sub, err := s.Resolve(ctx, userID)
	if err != nil {
		return nil, err
	}
	staff := s.announcementStaff(ctx, userID)
	if len(keys) == 0 {
		keys = []string{model.KeyCMSPublic, model.KeyAnnouncementFeed, model.KeyMembershipPortal}
	}
	out := make(map[string]dto.EvaluateResponse, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if key == model.KeyAnnouncementFeed && staff {
			out[key] = dto.EvaluateResponse{Key: key, Enabled: true, Reason: "staff_bypass", Type: model.TypeBoolean}
			continue
		}
		flag := s.lookup(ctx, key)
		got := feature.Evaluate(flag, sub)
		typ := ""
		if flag != nil {
			typ = flag.FlagType
		}
		out[key] = dto.EvaluateResponse{Key: key, Enabled: got.Enabled, Reason: got.Reason, Type: typ}
	}
	return out, nil
}

func (s *flagService) Enabled(ctx context.Context, key string, sub feature.Subject) bool {
	return feature.Evaluate(s.lookup(ctx, key), sub).Enabled
}

func (s *flagService) Resolve(ctx context.Context, userID uuid.UUID) (feature.Subject, error) {
	sub := feature.Subject{UserID: userID}
	if userID == uuid.Nil {
		return sub, nil
	}
	if s.roles != nil {
		codes, err := s.roles.GetUserRoleCodes(ctx, userID)
		if err != nil {
			return sub, fmt.Errorf("resolve roles: %w", err)
		}
		sub.RoleCodes = codes
	}
	if s.users != nil {
		dept, err := s.users.DepartmentOf(ctx, userID)
		if err != nil {
			return sub, fmt.Errorf("resolve dept: %w", err)
		}
		if dept != nil && *dept != uuid.Nil {
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
	s.mu.RUnlock()
	if err := s.reload(ctx); err != nil {
		logCacheErr("lookup-reload", err)
		row, err := s.rows.GetByKey(ctx, key)
		if err != nil || row == nil {
			return nil
		}
		return row
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if f, ok := s.memory[key]; ok {
		cp := f
		return &cp
	}
	return nil
}

func (s *flagService) reload(ctx context.Context) error {
	flags, err := s.store.Get(ctx)
	if err != nil {
		logCacheErr("snapshot-get", err)
		flags = nil
	}
	if flags == nil {
		flags, err = s.rows.ListAll(ctx)
		if err != nil {
			return err
		}
		logCacheErr("snapshot-set", s.store.Set(ctx, flags))
	}
	next := make(map[string]model.Flag, len(flags))
	for i := range flags {
		next[flags[i].FlagKey] = flags[i]
	}
	s.mu.Lock()
	s.memory = next
	s.mu.Unlock()
	return nil
}

func (s *flagService) invalidate(ctx context.Context) {
	logCacheErr("snapshot-del", s.store.Delete(ctx))
	s.mu.Lock()
	s.memory = map[string]model.Flag{}
	s.mu.Unlock()
	logCacheErr("publish", s.bus.Publish(ctx, time.Now().UTC().Format(time.RFC3339Nano)))
	_ = s.reload(ctx)
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
	}
	return nil
}
