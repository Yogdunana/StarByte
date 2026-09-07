package service

import (
	"context"
	"sync"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/contract/dto"
	"github.com/Yogdunana/StarByte/backend/internal/contract/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
)

type memRepo struct {
	mu   sync.Mutex
	rows map[uuid.UUID]*model.Contract
	tpls map[uuid.UUID]*model.Template
}

func newMem() *memRepo {
	return &memRepo{rows: map[uuid.UUID]*model.Contract{}, tpls: map[uuid.UUID]*model.Template{}}
}

func (m *memRepo) Create(_ context.Context, row *model.Contract) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *row
	m.rows[row.ID] = &cp
	return nil
}

func (m *memRepo) Update(_ context.Context, row *model.Contract) error {
	return m.Create(context.Background(), row)
}

func (m *memRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rows, id)
	return nil
}

func (m *memRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Contract, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.rows[id] == nil {
		return nil, nil
	}
	cp := *m.rows[id]
	return &cp, nil
}

func (m *memRepo) named(row *model.Contract) *model.ContractNamed {
	out := &model.ContractNamed{Contract: *row, OwnerName: "owner"}
	if row.TemplateID != nil {
		if t := m.tpls[*row.TemplateID]; t != nil {
			out.TemplateName = t.Name
		}
	}
	return out
}

func (m *memRepo) GetNamed(ctx context.Context, id uuid.UUID) (*model.ContractNamed, error) {
	row, err := m.GetByID(ctx, id)
	if err != nil || row == nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.named(row), nil
}

func (m *memRepo) List(_ context.Context, _ *dto.ListContractRequest, _ *rbacModel.DataScopeCondition) ([]model.ContractNamed, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.ContractNamed, 0, len(m.rows))
	for _, row := range m.rows {
		out = append(out, *m.named(row))
	}
	return out, int64(len(out)), nil
}

func (m *memRepo) ListExpiring(_ context.Context, until time.Time, _ *rbacModel.DataScopeCondition) ([]model.ContractNamed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	var out []model.ContractNamed
	for _, row := range m.rows {
		if row.Status == model.StatusActive && row.ExpiredAt != nil && !row.ExpiredAt.After(until) && !row.ExpiredAt.Before(now) {
			out = append(out, *m.named(row))
		}
	}
	return out, nil
}

func (m *memRepo) MarkExpired(_ context.Context, now time.Time) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for _, row := range m.rows {
		if row.Status == model.StatusActive && row.ExpiredAt != nil && row.ExpiredAt.Before(now) {
			row.Status = model.StatusExpired
			n++
		}
	}
	return n, nil
}

func (m *memRepo) ListTemplates(_ context.Context) ([]model.Template, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.Template, 0, len(m.tpls))
	for _, t := range m.tpls {
		out = append(out, *t)
	}
	return out, nil
}

func (m *memRepo) GetTemplate(_ context.Context, id uuid.UUID) (*model.Template, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.tpls[id] == nil {
		return nil, nil
	}
	cp := *m.tpls[id]
	return &cp, nil
}

func (m *memRepo) GetUser(_ context.Context, id uuid.UUID) (*model.NamedUser, error) {
	return &model.NamedUser{ID: id, RealName: "owner"}, nil
}
