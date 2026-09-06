package service

import (
	"context"
	"sync"

	"github.com/Yogdunana/StarByte/backend/internal/configstore/model"
	"github.com/google/uuid"
)

type memRepo struct {
	mu   sync.Mutex
	rows map[uuid.UUID]*model.Config
}

func newMemRepo() *memRepo {
	return &memRepo{rows: map[uuid.UUID]*model.Config{}}
}

func (m *memRepo) Create(_ context.Context, row *model.Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *row
	m.rows[row.ID] = &cp
	return nil
}

func (m *memRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Config, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.rows[id]
	if row == nil {
		return nil, nil
	}
	cp := *row
	return &cp, nil
}

func (m *memRepo) GetByKey(_ context.Context, key string) (*model.Config, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, row := range m.rows {
		if row.ConfigKey == key {
			cp := *row
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *memRepo) List(_ context.Context, category, keyword string) ([]model.Config, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.Config, 0, len(m.rows))
	for _, row := range m.rows {
		if category != "" && row.Category != category {
			continue
		}
		if keyword != "" && row.ConfigKey != keyword && row.Description != keyword {
			continue
		}
		out = append(out, *row)
	}
	return out, nil
}

func (m *memRepo) Update(_ context.Context, row *model.Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *row
	m.rows[row.ID] = &cp
	return nil
}

func (m *memRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rows, id)
	return nil
}
