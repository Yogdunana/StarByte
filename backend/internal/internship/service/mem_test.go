package service

import (
	"context"
	"sync"

	"github.com/Yogdunana/StarByte/backend/internal/internship/dto"
	"github.com/Yogdunana/StarByte/backend/internal/internship/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
)

type memRows struct {
	mu     sync.Mutex
	items  map[uuid.UUID]*model.Internship
	users  map[uuid.UUID]*model.NamedUser
	config *model.SystemConfig
}

func newMemRows() *memRows {
	return &memRows{items: map[uuid.UUID]*model.Internship{}, users: map[uuid.UUID]*model.NamedUser{}}
}

func (m *memRows) Create(_ context.Context, row *model.Internship) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *row
	m.items[row.ID] = &cp
	return nil
}

func (m *memRows) Update(_ context.Context, row *model.Internship) error {
	return m.Create(context.Background(), row)
}

func (m *memRows) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, id)
	return nil
}

func (m *memRows) GetByID(_ context.Context, id uuid.UUID) (*model.Internship, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.items[id] == nil {
		return nil, nil
	}
	cp := *m.items[id]
	return &cp, nil
}

func (m *memRows) named(row *model.Internship) model.InternshipWithNames {
	out := model.InternshipWithNames{Internship: *row}
	if u := m.users[row.UserID]; u != nil {
		out.UserName = u.RealName
		out.UserAvatar = u.Avatar
	}
	if row.MentorID != nil {
		if u := m.users[*row.MentorID]; u != nil {
			out.MentorName = u.RealName
		}
	}
	return out
}

func (m *memRows) GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.InternshipWithNames, error) {
	row, err := m.GetByID(ctx, id)
	if err != nil || row == nil {
		return nil, err
	}
	named := m.named(row)
	return &named, nil
}

func (m *memRows) List(_ context.Context, req *dto.ListInternshipRequest, _ *rbacModel.DataScopeCondition) ([]model.InternshipWithNames, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.InternshipWithNames, 0, len(m.items))
	for _, row := range m.items {
		if req.Status != nil && row.Status != *req.Status {
			continue
		}
		if req.UserID != "" && row.UserID.String() != req.UserID {
			continue
		}
		out = append(out, m.named(row))
	}
	return out, int64(len(out)), nil
}

func (m *memRows) ListByUser(_ context.Context, userID uuid.UUID, status *int16) ([]model.InternshipWithNames, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.InternshipWithNames, 0)
	for _, row := range m.items {
		if row.UserID != userID {
			continue
		}
		if status != nil && row.Status != *status {
			continue
		}
		out = append(out, m.named(row))
	}
	return out, nil
}

func (m *memRows) GetUser(_ context.Context, id uuid.UUID) (*model.NamedUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.users[id] == nil {
		return nil, nil
	}
	cp := *m.users[id]
	return &cp, nil
}

func (m *memRows) GetConfig(_ context.Context, _ string) (*model.SystemConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.config == nil {
		return nil, nil
	}
	cp := *m.config
	return &cp, nil
}

func (m *memRows) UpsertConfig(_ context.Context, cfg *model.SystemConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *cfg
	m.config = &cp
	return nil
}

func (m *memRows) ListForStats(_ context.Context, _, _, _ string, _ *rbacModel.DataScopeCondition) ([]model.InternshipWithNames, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.InternshipWithNames, 0, len(m.items))
	for _, row := range m.items {
		out = append(out, m.named(row))
	}
	return out, nil
}

func newTestSvc() (*internshipService, *memRows) {
	mem := newMemRows()
	return NewInternshipService(mem).(*internshipService), mem
}
