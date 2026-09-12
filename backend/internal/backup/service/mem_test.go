package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sort"
	"sync"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/backup/dto"
	"github.com/Yogdunana/StarByte/backend/internal/backup/model"
	"github.com/Yogdunana/StarByte/backend/internal/backup/repo"
	"github.com/google/uuid"
)

type memRepo struct {
	mu           sync.Mutex
	records      map[uuid.UUID]*model.Record
	policy       *model.Policy
	tasks        map[string]repo.ScheduledTaskSpec
	syncErr      error
	getErr       error
	getCalls     int
	getFailAfter int
	updateCalls  int
	updateFailN  int
}

func newMemRepo() *memRepo {
	return &memRepo{records: map[uuid.UUID]*model.Record{}, tasks: map[string]repo.ScheduledTaskSpec{}}
}

func (m *memRepo) clone(r *model.Record) *model.Record {
	c := *r
	return &c
}

func (m *memRepo) CreateRecord(_ context.Context, rec *model.Record) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records[rec.ID] = m.clone(rec)
	return nil
}

func (m *memRepo) UpdateRecord(_ context.Context, rec *model.Record) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateCalls++
	if m.updateFailN > 0 && m.updateCalls <= m.updateFailN {
		return errors.New("forced update fail")
	}
	m.records[rec.ID] = m.clone(rec)
	return nil
}

func (m *memRepo) MarkTerminal(_ context.Context, id uuid.UUID, status int16, msg string, now time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	r := m.records[id]
	if r == nil {
		return nil
	}
	if now.IsZero() {
		now = time.Now()
	}
	r.Status = status
	r.ErrorMessage = msg
	r.FinishedAt = &now
	r.UpdatedAt = now
	return nil
}

func (m *memRepo) DeleteRecord(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.records, id)
	return nil
}

func (m *memRepo) GetRecord(_ context.Context, id uuid.UUID) (*model.Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getCalls++
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.getFailAfter > 0 && m.getCalls > m.getFailAfter {
		return nil, errors.New("forced get fail")
	}
	r := m.records[id]
	if r == nil {
		return nil, nil
	}
	return m.clone(r), nil
}

func (m *memRepo) ListRecords(_ context.Context, req *dto.ListRequest) ([]model.Record, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var rows []model.Record
	for _, r := range m.records {
		if req != nil && req.Status != nil && r.Status != *req.Status {
			continue
		}
		rows = append(rows, *m.clone(r))
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].CreatedAt.After(rows[j].CreatedAt) })
	total := int64(len(rows))
	page, size := 1, 20
	if req != nil {
		if req.Page > 0 {
			page = req.Page
		}
		if req.PageSize > 0 {
			size = req.PageSize
		}
	}
	start := (page - 1) * size
	if start > len(rows) {
		return nil, total, nil
	}
	end := start + size
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end], total, nil
}

func (m *memRepo) CountBusy(context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for _, r := range m.records {
		if r.Status == model.StatusPending || r.Status == model.StatusRunning || r.Status == model.StatusRestoring {
			n++
		}
	}
	return n, nil
}

func (m *memRepo) ListStale(_ context.Context, before time.Time) ([]model.Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var rows []model.Record
	for _, r := range m.records {
		if (r.Status == model.StatusPending || r.Status == model.StatusRunning || r.Status == model.StatusRestoring) && r.UpdatedAt.Before(before) {
			rows = append(rows, *m.clone(r))
		}
	}
	return rows, nil
}

func (m *memRepo) ListExpired(_ context.Context, before time.Time) ([]model.Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var rows []model.Record
	for _, r := range m.records {
		when := r.CreatedAt
		if r.FinishedAt != nil {
			when = *r.FinishedAt
		}
		if (r.Status == model.StatusSuccess || r.Status == model.StatusFailed || r.Status == model.StatusRestored || r.Status == model.StatusRestoreFailed) && when.Before(before) {
			rows = append(rows, *m.clone(r))
		}
	}
	return rows, nil
}

func (m *memRepo) StorageStats(context.Context) (int64, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var count, size int64
	for _, r := range m.records {
		if r.Status == model.StatusSuccess || r.Status == model.StatusRestored || r.Status == model.StatusRestoreFailed {
			count++
			size += r.SizeBytes
		}
	}
	return count, size, nil
}

func (m *memRepo) GetPolicy(context.Context) (*model.Policy, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.policy == nil {
		return nil, nil
	}
	c := *m.policy
	return &c, nil
}

func (m *memRepo) UpsertPolicy(_ context.Context, p *model.Policy) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c := *p
	m.policy = &c
	return nil
}

func (m *memRepo) SyncScheduledTask(_ context.Context, spec repo.ScheduledTaskSpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.syncErr != nil {
		return m.syncErr
	}
	m.tasks[spec.Code] = spec
	return nil
}

type memStore struct {
	mu   sync.Mutex
	data map[string][]byte
}

func newMemStore() *memStore { return &memStore{data: map[string][]byte{}} }

func (s *memStore) Put(_ context.Context, key string, r io.Reader, _ int64) (string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = b
	return model.StorageMinIO, nil
}

func (s *memStore) Get(_ context.Context, _, key string) (io.ReadCloser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.data[key]
	if !ok {
		return nil, io.EOF
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

func (s *memStore) Delete(_ context.Context, _, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

type fakeEngine struct {
	dumpErr    error
	restoreErr error
	restored   []byte
	payload    []byte
}

func (e *fakeEngine) Dump(_ context.Context, dest io.Writer) error {
	if e.dumpErr != nil {
		return e.dumpErr
	}
	if len(e.payload) == 0 {
		e.payload = []byte("-- starbyte dump\nSELECT 1;\n")
	}
	_, err := dest.Write(e.payload)
	return err
}

func (e *fakeEngine) Restore(_ context.Context, src io.Reader) error {
	if e.restoreErr != nil {
		return e.restoreErr
	}
	b, err := io.ReadAll(src)
	if err != nil {
		return err
	}
	e.restored = b
	return nil
}

type recAlerter struct{ n int }

func (a *recAlerter) Failed(context.Context, *uuid.UUID, string, string) { a.n++ }
