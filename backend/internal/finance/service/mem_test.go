package service

import (
	"context"
	"sync"

	"github.com/Yogdunana/StarByte/backend/internal/finance/dto"
	"github.com/Yogdunana/StarByte/backend/internal/finance/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
)

type memRepo struct {
	mu   sync.Mutex
	recs map[uuid.UUID]*model.Record
	cats map[uuid.UUID]*model.Category
}

func newMem() *memRepo {
	return &memRepo{recs: map[uuid.UUID]*model.Record{}, cats: map[uuid.UUID]*model.Category{}}
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

func (m *memRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.recs, id)
	return nil
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

func (m *memRepo) namedOf(row *model.Record) *model.RecordNamed {
	out := &model.RecordNamed{Record: *row}
	if c := m.cats[row.CategoryID]; c != nil {
		out.CategoryName = c.Name
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
	return m.namedOf(row), nil
}

func (m *memRepo) List(_ context.Context, _ *dto.ListRecordRequest, _ *rbacModel.DataScopeCondition) ([]model.RecordNamed, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.RecordNamed, 0, len(m.recs))
	for _, row := range m.recs {
		out = append(out, *m.namedOf(row))
	}
	return out, int64(len(out)), nil
}

func (m *memRepo) ListCategories(_ context.Context) ([]model.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.Category, 0, len(m.cats))
	for _, c := range m.cats {
		out = append(out, *c)
	}
	return out, nil
}

func (m *memRepo) GetCategory(_ context.Context, id uuid.UUID) (*model.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cats[id] == nil {
		return nil, nil
	}
	cp := *m.cats[id]
	return &cp, nil
}

func (m *memRepo) Summary(_ context.Context, _, _, _ string, _ *rbacModel.DataScopeCondition) ([]model.SummaryRow, []model.CategorySumRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	byDir := map[int16]*model.SummaryRow{}
	byCat := map[uuid.UUID]*model.CategorySumRow{}
	for _, row := range m.recs {
		if byDir[row.Direction] == nil {
			byDir[row.Direction] = &model.SummaryRow{Direction: row.Direction}
		}
		byDir[row.Direction].Total += row.Amount
		byDir[row.Direction].Count++
		if byCat[row.CategoryID] == nil {
			name := ""
			if c := m.cats[row.CategoryID]; c != nil {
				name = c.Name
			}
			byCat[row.CategoryID] = &model.CategorySumRow{CategoryID: row.CategoryID, CategoryName: name, Direction: row.Direction}
		}
		byCat[row.CategoryID].Total += row.Amount
		byCat[row.CategoryID].Count++
	}
	totals := make([]model.SummaryRow, 0, len(byDir))
	for _, v := range byDir {
		totals = append(totals, *v)
	}
	cats := make([]model.CategorySumRow, 0, len(byCat))
	for _, v := range byCat {
		cats = append(cats, *v)
	}
	return totals, cats, nil
}
