package service

import (
	"context"
	"sync"

	"github.com/Yogdunana/StarByte/backend/internal/dict/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type memRepo struct {
	mu    sync.Mutex
	types map[uuid.UUID]*model.DictType
	items map[uuid.UUID]*model.DictItem
}

func newMemRepo() *memRepo {
	return &memRepo{
		types: map[uuid.UUID]*model.DictType{},
		items: map[uuid.UUID]*model.DictItem{},
	}
}

func (m *memRepo) ListTypes(_ context.Context) ([]model.DictType, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.DictType, 0, len(m.types))
	for _, rec := range m.types {
		out = append(out, *rec)
	}
	return out, nil
}

func (m *memRepo) GetTypeByID(_ context.Context, id uuid.UUID) (*model.DictType, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec := m.types[id]
	if rec == nil {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *rec
	return &cp, nil
}

func (m *memRepo) GetTypeByCode(_ context.Context, code string) (*model.DictType, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, rec := range m.types {
		if rec.Code == code {
			cp := *rec
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *memRepo) CreateType(_ context.Context, rec *model.DictType) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *rec
	m.types[rec.ID] = &cp
	return nil
}

func (m *memRepo) UpdateType(_ context.Context, rec *model.DictType) error {
	return m.CreateType(context.Background(), rec)
}

func (m *memRepo) DeleteType(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.types, id)
	for itemID, item := range m.items {
		if item.TypeID == id {
			delete(m.items, itemID)
		}
	}
	return nil
}

func (m *memRepo) ListItemsByTypeID(_ context.Context, typeID uuid.UUID, enabledOnly bool) ([]model.DictItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.DictItem, 0)
	for _, rec := range m.items {
		if rec.TypeID != typeID {
			continue
		}
		if enabledOnly && rec.Status != model.StatusEnabled {
			continue
		}
		out = append(out, *rec)
	}
	return out, nil
}

func (m *memRepo) GetItemByID(_ context.Context, id uuid.UUID) (*model.DictItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec := m.items[id]
	if rec == nil {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *rec
	return &cp, nil
}

func (m *memRepo) CreateItem(_ context.Context, rec *model.DictItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *rec
	m.items[rec.ID] = &cp
	return nil
}

func (m *memRepo) UpdateItem(_ context.Context, rec *model.DictItem) error {
	return m.CreateItem(context.Background(), rec)
}

func (m *memRepo) DeleteItem(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, id)
	return nil
}
