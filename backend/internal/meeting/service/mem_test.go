package service

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
)

type memMeetings struct {
	mu    sync.Mutex
	items map[uuid.UUID]*model.Meeting
	users map[uuid.UUID]*model.NamedUser
}

func newMemMeetings() *memMeetings {
	return &memMeetings{items: map[uuid.UUID]*model.Meeting{}, users: map[uuid.UUID]*model.NamedUser{}}
}

func (m *memMeetings) Create(_ context.Context, row *model.Meeting) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *row
	m.items[row.ID] = &cp
	return nil
}
func (m *memMeetings) Update(_ context.Context, row *model.Meeting) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *row
	m.items[row.ID] = &cp
	return nil
}
func (m *memMeetings) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, id)
	return nil
}
func (m *memMeetings) GetByID(_ context.Context, id uuid.UUID) (*model.Meeting, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.items[id]
	if row == nil {
		return nil, nil
	}
	cp := *row
	return &cp, nil
}
func (m *memMeetings) GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.MeetingWithNames, error) {
	row, err := m.GetByID(ctx, id)
	if err != nil || row == nil {
		return nil, err
	}
	return &model.MeetingWithNames{Meeting: *row, OrganizerName: "组织者"}, nil
}
func (m *memMeetings) List(_ context.Context, _ *dto.ListMeetingRequest, _ *rbacModel.DataScopeCondition) ([]model.MeetingWithNames, int64, error) {
	return nil, 0, nil
}
func (m *memMeetings) GetUser(_ context.Context, id uuid.UUID) (*model.NamedUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.users[id], nil
}

type memAgendas struct {
	mu    sync.Mutex
	items map[uuid.UUID]*model.Agenda
}

func newMemAgendas() *memAgendas { return &memAgendas{items: map[uuid.UUID]*model.Agenda{}} }

func (m *memAgendas) Create(_ context.Context, a *model.Agenda) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *a
	m.items[a.ID] = &cp
	return nil
}
func (m *memAgendas) Update(_ context.Context, a *model.Agenda) error {
	return m.Create(context.Background(), a)
}
func (m *memAgendas) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, id)
	return nil
}
func (m *memAgendas) GetByID(_ context.Context, id uuid.UUID) (*model.Agenda, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.items[id] == nil {
		return nil, nil
	}
	cp := *m.items[id]
	return &cp, nil
}
func (m *memAgendas) ListByMeeting(_ context.Context, meetingID uuid.UUID) ([]model.Agenda, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.Agenda
	for _, a := range m.items {
		if a.MeetingID == meetingID {
			out = append(out, *a)
		}
	}
	return out, nil
}
func (m *memAgendas) SaveSort(_ context.Context, items []model.Agenda) error {
	for i := range items {
		_ = m.Create(context.Background(), &items[i])
	}
	return nil
}

type memAttendees struct {
	mu    sync.Mutex
	items map[string]*model.Attendee
}

func newMemAttendees() *memAttendees { return &memAttendees{items: map[string]*model.Attendee{}} }

func attKey(mid, uid uuid.UUID) string { return mid.String() + ":" + uid.String() }

func (m *memAttendees) Add(_ context.Context, items []model.Attendee) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range items {
		cp := items[i]
		m.items[attKey(cp.MeetingID, cp.UserID)] = &cp
	}
	return nil
}
func (m *memAttendees) Remove(_ context.Context, meetingID, userID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, attKey(meetingID, userID))
	return nil
}
func (m *memAttendees) Get(_ context.Context, meetingID, userID uuid.UUID) (*model.Attendee, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.items[attKey(meetingID, userID)]
	if row == nil {
		return nil, nil
	}
	cp := *row
	return &cp, nil
}
func (m *memAttendees) Update(_ context.Context, a *model.Attendee) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *a
	m.items[attKey(a.MeetingID, a.UserID)] = &cp
	return nil
}
func (m *memAttendees) List(_ context.Context, meetingID uuid.UUID) ([]model.AttendeeNamed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.AttendeeNamed
	for _, a := range m.items {
		if a.MeetingID == meetingID {
			out = append(out, model.AttendeeNamed{Attendee: *a, RealName: "用户"})
		}
	}
	return out, nil
}
func (m *memAttendees) IsAttendee(_ context.Context, meetingID, userID uuid.UUID) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.items[attKey(meetingID, userID)]
	return ok, nil
}
