package service

import (
	"context"
	"sync"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/scheduler/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type memRepo struct {
	mu    sync.Mutex
	tasks map[uuid.UUID]*model.Task
	runs  map[uuid.UUID]*model.Run
	logs  map[uuid.UUID][]model.RunLog
}

func newMemRepo() *memRepo {
	return &memRepo{
		tasks: map[uuid.UUID]*model.Task{},
		runs:  map[uuid.UUID]*model.Run{},
		logs:  map[uuid.UUID][]model.RunLog{},
	}
}

func (m *memRepo) CreateTask(_ context.Context, rec *model.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *rec
	m.tasks[rec.ID] = &cp
	return nil
}

func (m *memRepo) UpdateTask(_ context.Context, rec *model.Task) error {
	return m.CreateTask(context.Background(), rec)
}

func (m *memRepo) GetTask(_ context.Context, id uuid.UUID) (*model.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec := m.tasks[id]
	if rec == nil {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *rec
	return &cp, nil
}

func (m *memRepo) GetTaskByCode(_ context.Context, code string) (*model.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, rec := range m.tasks {
		if rec.Code == code {
			cp := *rec
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *memRepo) ListTasks(_ context.Context, keyword string, status *int16, offset, limit int) ([]model.Task, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var all []model.Task
	for _, rec := range m.tasks {
		if rec.Status == model.StatusDeleted {
			continue
		}
		if status != nil && rec.Status != *status {
			continue
		}
		if keyword != "" && rec.Name != keyword && rec.Code != keyword {
			continue
		}
		all = append(all, *rec)
	}
	total := int64(len(all))
	if offset > len(all) {
		return nil, total, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], total, nil
}

func (m *memRepo) ListDue(_ context.Context, now time.Time, limit int) ([]model.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var list []model.Task
	for _, rec := range m.tasks {
		if rec.Status != model.StatusActive || rec.NextRunAt == nil || rec.NextRunAt.After(now) {
			continue
		}
		list = append(list, *rec)
		if len(list) >= limit {
			break
		}
	}
	return list, nil
}

func (m *memRepo) ListByIDs(_ context.Context, ids []uuid.UUID) ([]model.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var list []model.Task
	for _, id := range ids {
		if rec := m.tasks[id]; rec != nil {
			list = append(list, *rec)
		}
	}
	return list, nil
}

func (m *memRepo) CreateRun(_ context.Context, rec *model.Run) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *rec
	m.runs[rec.ID] = &cp
	return nil
}

func (m *memRepo) UpdateRun(_ context.Context, rec *model.Run) error {
	return m.CreateRun(context.Background(), rec)
}

func (m *memRepo) GetRun(_ context.Context, id uuid.UUID) (*model.Run, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec := m.runs[id]
	if rec == nil {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *rec
	return &cp, nil
}

func (m *memRepo) LatestSuccess(_ context.Context, taskID uuid.UUID) (*model.Run, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var best *model.Run
	for _, rec := range m.runs {
		if rec.TaskID == taskID && rec.Status == model.RunSuccess {
			cp := *rec
			best = &cp
		}
	}
	if best == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return best, nil
}

func (m *memRepo) ListRuns(_ context.Context, taskID uuid.UUID, offset, limit int) ([]model.Run, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var all []model.Run
	for _, rec := range m.runs {
		if rec.TaskID == taskID {
			all = append(all, *rec)
		}
	}
	total := int64(len(all))
	if offset > len(all) {
		return nil, total, nil
	}
	end := offset + limit
	if end > len(all) || limit <= 0 {
		end = len(all)
	}
	return all[offset:end], total, nil
}

func (m *memRepo) AddLog(_ context.Context, rec *model.RunLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs[rec.RunID] = append(m.logs[rec.RunID], *rec)
	return nil
}

func (m *memRepo) ListLogs(_ context.Context, runID uuid.UUID) ([]model.RunLog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]model.RunLog{}, m.logs[runID]...), nil
}
