package service

import (
	"context"
	"sync"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/discipline/dto"
	"github.com/Yogdunana/StarByte/backend/internal/discipline/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
)

type memRepo struct {
	mu      sync.Mutex
	recs    map[uuid.UUID]*model.Record
	appeals map[uuid.UUID][]model.Appeal
	users   map[uuid.UUID]*model.NamedUser
}

func newMem() *memRepo {
	return &memRepo{
		recs:    map[uuid.UUID]*model.Record{},
		appeals: map[uuid.UUID][]model.Appeal{},
		users:   map[uuid.UUID]*model.NamedUser{},
	}
}

func (m *memRepo) Create(_ context.Context, row *model.Record) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *row
	m.recs[row.ID] = &cp
	return nil
}

func (m *memRepo) Update(_ context.Context, row *model.Record) error {
	return m.Create(context.Background(), row)
}

func (m *memRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.recs[id] == nil {
		return nil, nil
	}
	cp := *m.recs[id]
	return &cp, nil
}

func (m *memRepo) named(row *model.Record) *model.RecordNamed {
	out := &model.RecordNamed{Record: *row}
	if u := m.users[row.UserID]; u != nil {
		out.UserName = u.RealName
		out.DepartmentID = u.DepartmentID
	}
	return out
}

func (m *memRepo) GetNamed(ctx context.Context, id uuid.UUID) (*model.RecordNamed, error) {
	row, err := m.GetByID(ctx, id)
	if err != nil || row == nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.named(row), nil
}

func (m *memRepo) List(_ context.Context, _ *dto.ListRecordRequest, _ *rbacModel.DataScopeCondition) ([]model.RecordNamed, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.RecordNamed, 0, len(m.recs))
	for _, row := range m.recs {
		out = append(out, *m.named(row))
	}
	return out, int64(len(out)), nil
}

func (m *memRepo) GetUser(_ context.Context, id uuid.UUID) (*model.NamedUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.users[id] == nil {
		return nil, nil
	}
	cp := *m.users[id]
	return &cp, nil
}

func (m *memRepo) CreateAppeal(_ context.Context, row *model.Appeal) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.appeals[row.RecordID] = append(m.appeals[row.RecordID], *row)
	return nil
}

func (m *memRepo) ListAppeals(_ context.Context, recordID uuid.UUID) ([]model.Appeal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]model.Appeal{}, m.appeals[recordID]...), nil
}

func (m *memRepo) HasOpenAppeal(_ context.Context, recordID uuid.UUID) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.appeals[recordID] {
		if a.Status == model.AppealPending {
			return true, nil
		}
	}
	return false, nil
}

func (m *memRepo) ResolveOpenAppeals(_ context.Context, recordID, reviewer uuid.UUID, status int16) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	rows := m.appeals[recordID]
	for i := range rows {
		if rows[i].Status == model.AppealPending {
			rows[i].Status = status
			rows[i].ReviewerID = &reviewer
			rows[i].ReviewedAt = &now
		}
	}
	m.appeals[recordID] = rows
	return nil
}

type captureNotify struct{ n int }

func (c *captureNotify) Send(_ context.Context, _ []uuid.UUID, _ string, _ map[string]interface{}) error {
	c.n++
	return nil
}
