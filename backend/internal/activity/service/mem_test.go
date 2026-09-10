package service

import (
	"context"
	"sync"

	"github.com/Yogdunana/StarByte/backend/internal/activity/dto"
	"github.com/Yogdunana/StarByte/backend/internal/activity/model"
	"github.com/google/uuid"
)

// ===== 内存活动 repo =====

type memActivities struct {
	mu    sync.Mutex
	items map[uuid.UUID]*model.Activity
}

func newMemActivities() *memActivities {
	return &memActivities{items: map[uuid.UUID]*model.Activity{}}
}

func (m *memActivities) Create(_ context.Context, a *model.Activity) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *a
	m.items[a.ID] = &cp
	return nil
}

func (m *memActivities) Update(_ context.Context, a *model.Activity) error {
	return m.Create(context.Background(), a)
}

func (m *memActivities) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, id)
	return nil
}

func (m *memActivities) GetByID(_ context.Context, id uuid.UUID) (*model.Activity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.items[id]
	if row == nil {
		return nil, nil
	}
	cp := *row
	return &cp, nil
}

func (m *memActivities) GetByIDWithNames(_ context.Context, id uuid.UUID) (*model.ActivityWithNames, error) {
	row, err := m.GetByID(context.Background(), id)
	if err != nil || row == nil {
		return nil, err
	}
	return &model.ActivityWithNames{
		Activity:      *row,
		OrganizerName: "组织者",
	}, nil
}

func (m *memActivities) List(_ context.Context, _ *dto.ListActivityRequest) ([]model.ActivityWithNames, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.ActivityWithNames
	for _, a := range m.items {
		if a.DeletedAt.Valid {
			continue
		}
		out = append(out, model.ActivityWithNames{Activity: *a, OrganizerName: "组织者"})
	}
	return out, int64(len(out)), nil
}

func (m *memActivities) GetUser(_ context.Context, id uuid.UUID) (*model.NamedUser, error) {
	return &model.NamedUser{ID: id, RealName: "测试用户", Username: "tester"}, nil
}

// ===== 内存报名 repo =====

type memRegs struct {
	mu    sync.Mutex
	items map[uuid.UUID]*model.ActivityRegistration
}

func newMemRegs() *memRegs {
	return &memRegs{items: map[uuid.UUID]*model.ActivityRegistration{}}
}

func regKey(activityID, userID uuid.UUID) string {
	return activityID.String() + ":" + userID.String()
}

func (m *memRegs) Create(_ context.Context, r *model.ActivityRegistration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *r
	m.items[r.ID] = &cp
	return nil
}

func (m *memRegs) Update(_ context.Context, r *model.ActivityRegistration) error {
	return m.Create(context.Background(), r)
}

func (m *memRegs) GetByID(_ context.Context, id uuid.UUID) (*model.ActivityRegistration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range m.items {
		if r.ID == id {
			cp := *r
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *memRegs) GetByActivityAndUser(_ context.Context, activityID, userID uuid.UUID) (*model.ActivityRegistration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range m.items {
		if r.ActivityID == activityID && r.UserID == userID {
			cp := *r
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *memRegs) ListByActivity(_ context.Context, activityID uuid.UUID) ([]model.RegistrationNamed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.RegistrationNamed
	for _, r := range m.items {
		if r.ActivityID == activityID {
			out = append(out, model.RegistrationNamed{
				ActivityRegistration: *r,
				RealName:             "测试用户",
				Username:             "tester",
			})
		}
	}
	return out, nil
}

func (m *memRegs) CountByActivityAndStatus(_ context.Context, activityID uuid.UUID, status int16) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for _, r := range m.items {
		if r.ActivityID == activityID && r.Status == status {
			n++
		}
	}
	return n, nil
}

func (m *memRegs) CountCheckedIn(_ context.Context, activityID uuid.UUID) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for _, r := range m.items {
		if r.ActivityID == activityID && r.CheckinStatus == model.CheckinDone {
			n++
		}
	}
	return n, nil
}

func (m *memRegs) ListWaitlist(_ context.Context, activityID uuid.UUID) ([]model.ActivityRegistration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.ActivityRegistration
	for _, r := range m.items {
		if r.ActivityID == activityID && r.Status == model.RegWaitlist {
			out = append(out, *r)
		}
	}
	return out, nil
}

// ===== 内存满意度 repo =====

type memSurveys struct {
	mu    sync.Mutex
	items map[uuid.UUID]*model.ActivitySurvey
}

func newMemSurveys() *memSurveys {
	return &memSurveys{items: map[uuid.UUID]*model.ActivitySurvey{}}
}

func surveyKey(activityID, userID uuid.UUID) string {
	return activityID.String() + ":" + userID.String()
}

func (m *memSurveys) Create(_ context.Context, s *model.ActivitySurvey) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *s
	m.items[s.ID] = &cp
	return nil
}

func (m *memSurveys) GetByActivityAndUser(_ context.Context, activityID, userID uuid.UUID) (*model.ActivitySurvey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.items {
		if s.ActivityID == activityID && s.UserID == userID {
			cp := *s
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *memSurveys) StatsByActivity(_ context.Context, activityID uuid.UUID) (int64, float64, map[int16]int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	dist := map[int16]int64{}
	var total float64
	var count int64
	for _, s := range m.items {
		if s.ActivityID == activityID {
			dist[s.Rating]++
			total += float64(s.Rating)
			count++
		}
	}
	avg := 0.0
	if count > 0 {
		avg = total / float64(count)
	}
	return count, avg, dist, nil
}

// ===== 内存通知器 =====

type memNotifier struct {
	mu      sync.Mutex
	sent    []string
	targets [][]uuid.UUID
}

func (n *memNotifier) Send(_ context.Context, userIDs []uuid.UUID, template string, _ map[string]interface{}) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.sent = append(n.sent, template)
	n.targets = append(n.targets, append([]uuid.UUID{}, userIDs...))
	return nil
}

// ===== 测试辅助 =====

func newTestSvc() (*activityService, *memActivities, *memRegs, *memSurveys, *memNotifier) {
	aa := newMemActivities()
	rr := newMemRegs()
	ss := newMemSurveys()
	nn := &memNotifier{}
	svc := NewActivityService(aa, rr, ss, nn).(*activityService)
	return svc, aa, rr, ss, nn
}

// 防止未使用告警
var _ = regKey
var _ = surveyKey
