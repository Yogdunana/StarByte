package service

import (
	"context"
	"sync"
	"time"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/google/uuid"
)

type memTasks struct {
	mu    sync.Mutex
	items map[uuid.UUID]*model.Task
	users map[uuid.UUID]*model.NamedUser
}

func newMemTasks() *memTasks {
	return &memTasks{items: map[uuid.UUID]*model.Task{}, users: map[uuid.UUID]*model.NamedUser{}}
}

func (m *memTasks) Create(_ context.Context, t *model.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *t
	m.items[t.ID] = &cp
	return nil
}
func (m *memTasks) Update(_ context.Context, t *model.Task) error {
	return m.Create(context.Background(), t)
}
func (m *memTasks) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, id)
	return nil
}
func (m *memTasks) GetByID(_ context.Context, id uuid.UUID) (*model.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.items[id] == nil {
		return nil, nil
	}
	cp := *m.items[id]
	return &cp, nil
}
func (m *memTasks) GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.TaskWithNames, error) {
	t, err := m.GetByID(ctx, id)
	if err != nil || t == nil {
		return nil, err
	}
	return &model.TaskWithNames{Task: *t, CreatorName: "创建人", AssigneeName: "负责人"}, nil
}
func (m *memTasks) List(_ context.Context, _ *dto.ListTaskRequest, _ *rbacModel.DataScopeCondition) ([]model.TaskWithNames, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.TaskWithNames, 0, len(m.items))
	for _, t := range m.items {
		out = append(out, model.TaskWithNames{Task: *t})
	}
	return out, int64(len(out)), nil
}
func (m *memTasks) ListMine(_ context.Context, userID uuid.UUID, kind string, _ *dto.MyTaskRequest) ([]model.TaskWithNames, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	out := make([]model.TaskWithNames, 0)
	for _, t := range m.items {
		ok := false
		switch kind {
		case "todo":
			ok = t.AssigneeID != nil && *t.AssigneeID == userID && (t.Status == 0 || t.Status == 1 || t.Status == 4)
		case "done":
			ok = t.AssigneeID != nil && *t.AssigneeID == userID && t.Status == 2
		case "created":
			ok = t.CreatorID == userID
		case "overdue":
			ok = t.AssigneeID != nil && *t.AssigneeID == userID && t.DueDate != nil && t.DueDate.Before(now) && !model.IsClosed(t.Status)
		}
		if ok {
			out = append(out, model.TaskWithNames{Task: *t})
		}
	}
	return out, int64(len(out)), nil
}
func (m *memTasks) ListChildren(_ context.Context, parentID uuid.UUID) ([]model.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.Task
	for _, t := range m.items {
		if t.ParentID != nil && *t.ParentID == parentID {
			out = append(out, *t)
		}
	}
	return out, nil
}
func (m *memTasks) ListDueSoon(_ context.Context, now, until time.Time) ([]model.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.Task
	for _, t := range m.items {
		if t.DueDate != nil && t.DueDate.After(now) && !t.DueDate.After(until) && t.DueRemindedAt == nil && !model.IsClosed(t.Status) {
			out = append(out, *t)
		}
	}
	return out, nil
}
func (m *memTasks) ListOverdue(_ context.Context, now time.Time) ([]model.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.Task
	for _, t := range m.items {
		if t.DueDate != nil && t.DueDate.Before(now) && t.OverdueRemindedAt == nil && !model.IsClosed(t.Status) {
			out = append(out, *t)
		}
	}
	return out, nil
}
func (m *memTasks) Stats(_ context.Context, _ *dto.StatsRequest, now time.Time) (*dto.StatsResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := &dto.StatsResponse{
		ByStatus:   map[string]int64{"pending": 0, "doing": 0, "done": 0, "cancelled": 0, "held": 0},
		ByPriority: map[string]int64{"low": 0, "medium": 0, "high": 0, "urgent": 0},
	}
	keys := map[int16]string{0: "pending", 1: "doing", 2: "done", 3: "cancelled", 4: "held"}
	prios := map[int16]string{0: "low", 1: "medium", 2: "high", 3: "urgent"}
	for _, t := range m.items {
		out.Total++
		out.ByStatus[keys[t.Status]]++
		out.ByPriority[prios[t.Priority]]++
		if t.DueDate != nil && t.DueDate.Before(now) && !model.IsClosed(t.Status) {
			out.Overdue++
		}
	}
	return out, nil
}
func (m *memTasks) GetUser(_ context.Context, id uuid.UUID) (*model.NamedUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.users[id], nil
}
func (m *memTasks) FindUsersByUsername(_ context.Context, names []string) ([]model.NamedUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	want := map[string]struct{}{}
	for _, n := range names {
		want[n] = struct{}{}
	}
	var out []model.NamedUser
	for _, u := range m.users {
		if _, ok := want[u.Username]; ok {
			out = append(out, *u)
		}
	}
	return out, nil
}

type memLogs struct {
	mu    sync.Mutex
	items []model.TaskLog
}

func newMemLogs() *memLogs { return &memLogs{} }
func (m *memLogs) Create(_ context.Context, log *model.TaskLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = append(m.items, *log)
	return nil
}
func (m *memLogs) ListByTask(_ context.Context, taskID uuid.UUID) ([]model.TaskLogNamed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.TaskLogNamed
	for _, l := range m.items {
		if l.TaskID == taskID {
			out = append(out, model.TaskLogNamed{TaskLog: l, OperatorName: "操作者"})
		}
	}
	return out, nil
}

type memComments struct {
	mu    sync.Mutex
	items map[uuid.UUID]*model.TaskComment
}

func newMemComments() *memComments { return &memComments{items: map[uuid.UUID]*model.TaskComment{}} }
func (m *memComments) Create(_ context.Context, c *model.TaskComment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *c
	m.items[c.ID] = &cp
	return nil
}
func (m *memComments) Update(_ context.Context, c *model.TaskComment) error {
	return m.Create(context.Background(), c)
}
func (m *memComments) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, id)
	return nil
}
func (m *memComments) GetByID(_ context.Context, id uuid.UUID) (*model.TaskComment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.items[id] == nil {
		return nil, nil
	}
	cp := *m.items[id]
	return &cp, nil
}
func (m *memComments) ListByTask(_ context.Context, taskID uuid.UUID) ([]model.TaskCommentNamed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.TaskCommentNamed
	for _, c := range m.items {
		if c.TaskID == taskID {
			out = append(out, model.TaskCommentNamed{TaskComment: *c, AuthorName: "作者"})
		}
	}
	return out, nil
}

type memFiles struct {
	mu    sync.Mutex
	items map[uuid.UUID]*model.TaskAttachmentNamed
}

func newMemFiles() *memFiles {
	return &memFiles{items: map[uuid.UUID]*model.TaskAttachmentNamed{}}
}
func (m *memFiles) Create(_ context.Context, a *model.TaskAttachment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[a.ID] = &model.TaskAttachmentNamed{TaskAttachment: *a, FileName: "f.txt"}
	return nil
}
func (m *memFiles) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, id)
	return nil
}
func (m *memFiles) GetByID(_ context.Context, id uuid.UUID) (*model.TaskAttachment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.items[id] == nil {
		return nil, nil
	}
	cp := m.items[id].TaskAttachment
	return &cp, nil
}
func (m *memFiles) GetNamed(_ context.Context, id uuid.UUID) (*model.TaskAttachmentNamed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.items[id], nil
}
func (m *memFiles) ListByTask(_ context.Context, taskID uuid.UUID) ([]model.TaskAttachmentNamed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.TaskAttachmentNamed
	for _, a := range m.items {
		if a.TaskID == taskID {
			out = append(out, *a)
		}
	}
	return out, nil
}

type memNotify struct {
	mu    sync.Mutex
	calls int
}

func (m *memNotify) Send(_ context.Context, _ []uuid.UUID, _ string, _ map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	return nil
}
