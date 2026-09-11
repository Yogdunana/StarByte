package service

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/model"
	"github.com/google/uuid"
)

type memRepo struct {
	mu        sync.Mutex
	cals      map[uuid.UUID]*model.Calendar
	members   map[uuid.UUID]*model.CalendarMember
	events    map[uuid.UUID]*model.Event
	attendees map[uuid.UUID]*model.Attendee
	reminders map[uuid.UUID]*model.Reminder
	users     map[uuid.UUID]*model.NamedUser
	google    map[uuid.UUID]*model.GoogleAccount
}

func newMem() *memRepo {
	return &memRepo{
		cals: map[uuid.UUID]*model.Calendar{}, members: map[uuid.UUID]*model.CalendarMember{},
		events: map[uuid.UUID]*model.Event{}, attendees: map[uuid.UUID]*model.Attendee{},
		reminders: map[uuid.UUID]*model.Reminder{}, users: map[uuid.UUID]*model.NamedUser{},
		google: map[uuid.UUID]*model.GoogleAccount{},
	}
}

func (m *memRepo) addUser(id uuid.UUID, name string, dept *uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[id] = &model.NamedUser{ID: id, RealName: name, Username: name, DepartmentID: dept}
}

func (m *memRepo) CreateCalendar(_ context.Context, row *model.Calendar) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *row
	m.cals[row.ID] = &cp
	return nil
}

func (m *memRepo) UpdateCalendar(_ context.Context, row *model.Calendar) error {
	return m.CreateCalendar(context.Background(), row)
}

func (m *memRepo) DeleteCalendar(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.cals, id)
	for k, mem := range m.members {
		if mem.CalendarID == id {
			delete(m.members, k)
		}
	}
	for k, ev := range m.events {
		if ev.CalendarID == id {
			delete(m.events, k)
		}
	}
	return nil
}

func (m *memRepo) GetCalendar(_ context.Context, id uuid.UUID) (*model.Calendar, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.cals[id]
	if row == nil {
		return nil, nil
	}
	cp := *row
	return &cp, nil
}

func (m *memRepo) namedCal(row *model.Calendar, viewer uuid.UUID) *model.CalendarNamed {
	out := &model.CalendarNamed{Calendar: *row, OwnerName: "owner"}
	if u := m.users[row.OwnerID]; u != nil {
		out.OwnerName = u.RealName
	}
	for _, mem := range m.members {
		if mem.CalendarID == row.ID && mem.UserID == viewer {
			out.MemberRole = mem.Role
		}
	}
	return out
}

func (m *memRepo) GetCalendarNamed(_ context.Context, id, viewer uuid.UUID) (*model.CalendarNamed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.cals[id]
	if row == nil {
		return nil, nil
	}
	return m.namedCal(row, viewer), nil
}

func visibleCalendar(row *model.Calendar, viewer uuid.UUID, memberRole int16, scope *rbacModel.DataScopeCondition) bool {
	return canViewCalendar(scope, row, viewer, memberRole)
}

func (m *memRepo) ListCalendars(_ context.Context, viewer uuid.UUID, req *dto.ListCalendarRequest, scope *rbacModel.DataScopeCondition) ([]model.CalendarNamed, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.CalendarNamed, 0, len(m.cals))
	for _, row := range m.cals {
		named := m.namedCal(row, viewer)
		if !visibleCalendar(row, viewer, named.MemberRole, scope) {
			continue
		}
		if req.CalendarType != nil && row.CalendarType != *req.CalendarType {
			continue
		}
		if kw := strings.TrimSpace(req.Keyword); kw != "" && !strings.Contains(row.Name, kw) {
			continue
		}
		out = append(out, *named)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, int64(len(out)), nil
}

func (m *memRepo) PersonalCalendar(_ context.Context, owner uuid.UUID) (*model.Calendar, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, row := range m.cals {
		if row.OwnerID == owner && row.CalendarType == model.CalendarPersonal && model.NormalizeSource(row.Source) == model.SourcePersonal {
			cp := *row
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *memRepo) AddMember(_ context.Context, row *model.CalendarMember) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, existing := range m.members {
		if existing.CalendarID == row.CalendarID && existing.UserID == row.UserID {
			cp := *row
			cp.ID = id
			m.members[id] = &cp
			return nil
		}
	}
	cp := *row
	m.members[row.ID] = &cp
	return nil
}

func (m *memRepo) RemoveMember(_ context.Context, calendarID, userID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, row := range m.members {
		if row.CalendarID == calendarID && row.UserID == userID {
			delete(m.members, id)
		}
	}
	return nil
}

func (m *memRepo) GetMember(_ context.Context, calendarID, userID uuid.UUID) (*model.CalendarMember, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, row := range m.members {
		if row.CalendarID == calendarID && row.UserID == userID {
			cp := *row
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *memRepo) ListMembers(_ context.Context, calendarID uuid.UUID) ([]model.CalendarMemberNamed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []model.CalendarMemberNamed{}
	for _, row := range m.members {
		if row.CalendarID != calendarID {
			continue
		}
		name := "user"
		if u := m.users[row.UserID]; u != nil {
			name = u.RealName
		}
		out = append(out, model.CalendarMemberNamed{CalendarMember: *row, RealName: name, Username: name})
	}
	return out, nil
}

func (m *memRepo) CreateEvent(_ context.Context, row *model.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *row
	m.events[row.ID] = &cp
	return nil
}

func (m *memRepo) UpdateEvent(_ context.Context, row *model.Event) error {
	return m.CreateEvent(context.Background(), row)
}

func (m *memRepo) DeleteEvent(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.events, id)
	for k, a := range m.attendees {
		if a.EventID == id {
			delete(m.attendees, k)
		}
	}
	for k, r := range m.reminders {
		if r.EventID == id {
			delete(m.reminders, k)
		}
	}
	return nil
}

func (m *memRepo) GetEvent(_ context.Context, id uuid.UUID) (*model.Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.events[id]
	if row == nil {
		return nil, nil
	}
	cp := *row
	return &cp, nil
}

func (m *memRepo) namedEvent(row *model.Event, viewer uuid.UUID) *model.EventNamed {
	out := &model.EventNamed{Event: *row, CalendarName: "cal", CreatorName: "creator"}
	if cal := m.cals[row.CalendarID]; cal != nil {
		out.CalendarName = cal.Name
		out.CalendarColor = cal.Color
		out.CalendarType = cal.CalendarType
		out.OwnerID = cal.OwnerID
		out.DepartmentID = cal.DepartmentID
		for _, mem := range m.members {
			if mem.CalendarID == cal.ID && mem.UserID == viewer {
				out.MemberRole = mem.Role
			}
		}
	}
	if u := m.users[row.CreatedBy]; u != nil {
		out.CreatorName = u.RealName
	}
	for _, a := range m.attendees {
		if a.EventID == row.ID {
			out.AttendeeCount++
		}
	}
	return out
}

func (m *memRepo) GetEventNamed(_ context.Context, id uuid.UUID) (*model.EventNamed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.events[id]
	if row == nil {
		return nil, nil
	}
	return m.namedEvent(row, uuid.Nil), nil
}

func (m *memRepo) ListEvents(_ context.Context, viewer uuid.UUID, req *dto.ListEventRequest, scope *rbacModel.DataScopeCondition) ([]model.EventNamed, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []model.EventNamed{}
	for _, row := range m.events {
		named := m.namedEvent(row, viewer)
		isAttendee := false
		for _, a := range m.attendees {
			if a.EventID == row.ID && a.UserID == viewer {
				isAttendee = true
			}
		}
		if !canViewEvent(scope, named, viewer, named.MemberRole, isAttendee) {
			continue
		}
		if req.CalendarID != "" && row.CalendarID.String() != req.CalendarID {
			continue
		}
		if req.Status != nil && row.Status != *req.Status {
			continue
		}
		if !req.Start.IsZero() && row.EndAt.Before(req.Start) && row.Recurrence == model.RecurrenceNone {
			continue
		}
		if !req.End.IsZero() && row.StartAt.After(req.End) {
			continue
		}
		if kw := strings.TrimSpace(req.Keyword); kw != "" && !strings.Contains(row.Title, kw) {
			continue
		}
		out = append(out, *named)
	}
	return out, int64(len(out)), nil
}

func (m *memRepo) Overlapping(_ context.Context, calendarID, exclude uuid.UUID, start, end time.Time) ([]model.Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.Event
	for _, row := range m.events {
		if row.CalendarID != calendarID || row.Status != model.EventConfirmed || row.ID == exclude {
			continue
		}
		if row.StartAt.Before(end) && row.EndAt.After(start) {
			out = append(out, *row)
		}
	}
	return out, nil
}

func (m *memRepo) ReplaceAttendees(_ context.Context, eventID uuid.UUID, rows []model.Attendee) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, a := range m.attendees {
		if a.EventID == eventID {
			delete(m.attendees, id)
		}
	}
	for i := range rows {
		cp := rows[i]
		m.attendees[cp.ID] = &cp
	}
	return nil
}

func (m *memRepo) ListAttendees(_ context.Context, eventID uuid.UUID) ([]model.AttendeeNamed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []model.AttendeeNamed{}
	for _, a := range m.attendees {
		if a.EventID != eventID {
			continue
		}
		name := "user"
		if u := m.users[a.UserID]; u != nil {
			name = u.RealName
		}
		out = append(out, model.AttendeeNamed{Attendee: *a, RealName: name, Username: name})
	}
	return out, nil
}

func (m *memRepo) GetAttendee(_ context.Context, eventID, userID uuid.UUID) (*model.Attendee, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.attendees {
		if a.EventID == eventID && a.UserID == userID {
			cp := *a
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *memRepo) UpdateAttendee(_ context.Context, row *model.Attendee) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *row
	m.attendees[row.ID] = &cp
	return nil
}

func (m *memRepo) ReplaceReminders(_ context.Context, eventID uuid.UUID, rows []model.Reminder) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, r := range m.reminders {
		if r.EventID == eventID {
			delete(m.reminders, id)
		}
	}
	for i := range rows {
		cp := rows[i]
		m.reminders[cp.ID] = &cp
	}
	return nil
}

func (m *memRepo) ListReminders(_ context.Context, eventID uuid.UUID) ([]model.Reminder, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []model.Reminder{}
	for _, r := range m.reminders {
		if r.EventID == eventID {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (m *memRepo) ListDueReminders(_ context.Context, now time.Time, _ int) ([]model.DueReminder, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []model.DueReminder{}
	for _, r := range m.reminders {
		if r.TriggeredAt != nil {
			continue
		}
		ev := m.events[r.EventID]
		if ev == nil || ev.Status != model.EventConfirmed {
			continue
		}
		if ev.StartAt.Add(-time.Duration(r.MinutesBefore) * time.Minute).After(now) {
			continue
		}
		due := model.DueReminder{Reminder: *r, Title: ev.Title, StartAt: ev.StartAt, CreatedBy: ev.CreatedBy}
		if cal := m.cals[ev.CalendarID]; cal != nil {
			due.OwnerID = cal.OwnerID
		}
		out = append(out, due)
	}
	return out, nil
}

func (m *memRepo) MarkReminderTriggered(_ context.Context, id uuid.UUID, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r := m.reminders[id]; r != nil && r.TriggeredAt == nil {
		r.TriggeredAt = &at
	}
	return nil
}

func (m *memRepo) GetUser(_ context.Context, id uuid.UUID) (*model.NamedUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u := m.users[id]
	if u == nil {
		return nil, nil
	}
	cp := *u
	return &cp, nil
}

func (m *memRepo) CalendarBySource(_ context.Context, owner uuid.UUID, source, sourceKey string) (*model.Calendar, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, row := range m.cals {
		if row.OwnerID != owner || model.NormalizeSource(row.Source) != source {
			continue
		}
		if source == model.SourceImport && sourceKey != "" && row.SourceKey != sourceKey {
			continue
		}
		cp := *row
		return &cp, nil
	}
	return nil, nil
}

func (m *memRepo) ReplaceOriginEvents(_ context.Context, calendarID uuid.UUID, origin string, rows []model.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, ev := range m.events {
		if ev.CalendarID == calendarID && ev.Origin == origin {
			delete(m.events, id)
		}
	}
	for i := range rows {
		cp := rows[i]
		m.events[cp.ID] = &cp
	}
	return nil
}

func (m *memRepo) GetGoogleAccount(_ context.Context, userID uuid.UUID) (*model.GoogleAccount, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.google[userID]
	if row == nil {
		return nil, nil
	}
	cp := *row
	return &cp, nil
}

func (m *memRepo) UpsertGoogleAccount(_ context.Context, row *model.GoogleAccount) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *row
	m.google[row.UserID] = &cp
	return nil
}

func (m *memRepo) DeleteGoogleAccount(_ context.Context, userID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.google, userID)
	return nil
}
