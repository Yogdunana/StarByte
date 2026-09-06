package service

import (
	"context"
	"sync"

	"github.com/Yogdunana/StarByte/backend/internal/form/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type memRepo struct {
	mu    sync.Mutex
	forms map[uuid.UUID]*model.Form
	subs  map[uuid.UUID]*model.Submission
	names map[uuid.UUID]string
}

func newMemRepo() *memRepo {
	return &memRepo{
		forms: map[uuid.UUID]*model.Form{},
		subs:  map[uuid.UUID]*model.Submission{},
		names: map[uuid.UUID]string{},
	}
}

func (m *memRepo) Create(_ context.Context, rec *model.Form) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *rec
	m.forms[rec.ID] = &cp
	return nil
}

func (m *memRepo) Update(_ context.Context, rec *model.Form) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.forms[rec.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	cp := *rec
	m.forms[rec.ID] = &cp
	return nil
}

func (m *memRepo) Get(_ context.Context, id uuid.UUID) (*model.Form, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.forms[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *rec
	return &cp, nil
}

func (m *memRepo) GetByName(_ context.Context, name string) (*model.Form, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, rec := range m.forms {
		if rec.Name == name {
			cp := *rec
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *memRepo) List(_ context.Context, keyword string, status *int16, offset, limit int) ([]model.Form, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var all []model.Form
	for _, rec := range m.forms {
		if status != nil && rec.Status != *status {
			continue
		}
		if keyword != "" && rec.Name != keyword {
			continue
		}
		all = append(all, *rec)
	}
	total := int64(len(all))
	if offset > len(all) {
		return []model.Form{}, total, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], total, nil
}

func (m *memRepo) CountSubmissions(_ context.Context, formID uuid.UUID) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for _, s := range m.subs {
		if s.FormID == formID {
			n++
		}
	}
	return n, nil
}

func (m *memRepo) CreateSubmission(_ context.Context, rec *model.Submission) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *rec
	m.subs[rec.ID] = &cp
	return nil
}

func (m *memRepo) ListSubmissions(_ context.Context, formID uuid.UUID, offset, limit int) ([]model.Submission, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var all []model.Submission
	for _, s := range m.subs {
		if s.FormID == formID {
			all = append(all, *s)
		}
	}
	total := int64(len(all))
	if offset > len(all) {
		return []model.Submission{}, total, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], total, nil
}

func (m *memRepo) UserDisplayNames(_ context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[uuid.UUID]string, len(ids))
	for _, id := range ids {
		if n, ok := m.names[id]; ok {
			out[id] = n
		}
	}
	return out, nil
}
