package service

import (
	"context"
	"strings"
	"sync"

	"github.com/Yogdunana/StarByte/backend/internal/feature/dto"
	"github.com/Yogdunana/StarByte/backend/internal/feature/model"
	"github.com/google/uuid"
)

type memRepo struct {
	mu     sync.Mutex
	flags  map[uuid.UUID]*model.Flag
	byKey  map[string]uuid.UUID
	audits []model.Audit
}

func newMemRepo() *memRepo {
	return &memRepo{flags: map[uuid.UUID]*model.Flag{}, byKey: map[string]uuid.UUID{}}
}

func (m *memRepo) List(_ context.Context, q dto.ListQuery) ([]model.Flag, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.Flag
	for _, f := range m.flags {
		if kw := strings.TrimSpace(q.Keyword); kw != "" && !strings.Contains(f.FlagKey, kw) && !strings.Contains(f.Name, kw) {
			continue
		}
		if g := strings.TrimSpace(q.GroupName); g != "" && f.GroupName != g {
			continue
		}
		if q.Enabled != nil && f.Enabled != *q.Enabled {
			continue
		}
		out = append(out, *f)
	}
	return out, int64(len(out)), nil
}

func (m *memRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Flag, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	f, ok := m.flags[id]
	if !ok {
		return nil, nil
	}
	cp := *f
	return &cp, nil
}

func (m *memRepo) GetByKey(_ context.Context, key string) (*model.Flag, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.byKey[key]
	if !ok {
		return nil, nil
	}
	cp := *m.flags[id]
	return &cp, nil
}

func (m *memRepo) ListAll(_ context.Context) ([]model.Flag, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.Flag, 0, len(m.flags))
	for _, f := range m.flags {
		out = append(out, *f)
	}
	return out, nil
}

func (m *memRepo) Create(_ context.Context, flag *model.Flag) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *flag
	m.flags[flag.ID] = &cp
	m.byKey[flag.FlagKey] = flag.ID
	return nil
}

func (m *memRepo) Update(_ context.Context, flag *model.Flag) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *flag
	m.flags[flag.ID] = &cp
	m.byKey[flag.FlagKey] = flag.ID
	return nil
}

func (m *memRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if f, ok := m.flags[id]; ok {
		delete(m.byKey, f.FlagKey)
		delete(m.flags, id)
	}
	return nil
}

func (m *memRepo) CreateAudit(_ context.Context, row *model.Audit) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.audits = append(m.audits, *row)
	return nil
}

func (m *memRepo) ListAudits(_ context.Context, q dto.AuditQuery) ([]model.Audit, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.Audit
	for _, a := range m.audits {
		if k := strings.TrimSpace(q.FlagKey); k != "" && a.FlagKey != k {
			continue
		}
		out = append(out, a)
	}
	return out, int64(len(out)), nil
}

type stubRoles struct{ codes []string }

func (s stubRoles) GetUserRoleCodes(context.Context, uuid.UUID) ([]string, error) {
	return s.codes, nil
}

type stubUsers struct{ dept *uuid.UUID }

func (s stubUsers) DepartmentOf(context.Context, uuid.UUID) (*uuid.UUID, error) {
	return s.dept, nil
}

type stubPerms struct {
	perms   []string
	isSuper bool
}

func (s stubPerms) GetUserPermissionsAndSuperAdmin(context.Context, uuid.UUID) ([]string, bool, error) {
	return s.perms, s.isSuper, nil
}
