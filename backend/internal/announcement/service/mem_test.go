package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/announcement/dto"
	"github.com/Yogdunana/StarByte/backend/internal/announcement/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type memRepo struct {
	mu        sync.Mutex
	items     map[uuid.UUID]*model.Announcement
	reads     map[string]model.AnnouncementRead
	users     map[uuid.UUID]model.NamedUser
	activeIDs []uuid.UUID
}

func newMemRepo() *memRepo {
	return &memRepo{
		items: map[uuid.UUID]*model.Announcement{},
		reads: map[string]model.AnnouncementRead{},
		users: map[uuid.UUID]model.NamedUser{},
	}
}

func (m *memRepo) addUser(id uuid.UUID, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[id] = model.NamedUser{ID: id, RealName: name, Username: name}
	m.activeIDs = append(m.activeIDs, id)
}

func (m *memRepo) Create(_ context.Context, a *model.Announcement) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *a
	m.items[a.ID] = &cp
	return nil
}

func (m *memRepo) Update(_ context.Context, a *model.Announcement) error {
	return m.Create(context.Background(), a)
}

func (m *memRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if row, ok := m.items[id]; ok {
		row.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	}
	return nil
}

func (m *memRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Announcement, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.items[id]
	if row == nil || row.DeletedAt.Valid {
		return nil, nil
	}
	cp := *row
	return &cp, nil
}

func (m *memRepo) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Announcement, error) {
	return m.GetByID(ctx, id)
}

func (m *memRepo) named(a *model.Announcement, viewer uuid.UUID) model.AnnouncementNamed {
	name := "作者"
	if u, ok := m.users[a.AuthorID]; ok && u.RealName != "" {
		name = u.RealName
	}
	_, read := m.reads[a.ID.String()+"/"+viewer.String()]
	return model.AnnouncementNamed{Announcement: *a, AuthorName: name, IsRead: read}
}

func (m *memRepo) GetByIDNamed(_ context.Context, id, viewer uuid.UUID) (*model.AnnouncementNamed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.items[id]
	if row == nil || row.DeletedAt.Valid {
		return nil, nil
	}
	named := m.named(row, viewer)
	return &named, nil
}

func (m *memRepo) inAudience(a *model.Announcement, viewer uuid.UUID) bool {
	switch model.NormalizeAudience(a.AudienceType) {
	case model.AudienceAll:
		return true
	case model.AudienceUsers:
		for _, id := range a.AudienceIDs {
			if id == viewer.String() {
				return true
			}
		}
		return false
	case model.AudienceDepartment:
		u, ok := m.users[viewer]
		if !ok || u.DepartmentID == nil {
			return false
		}
		for _, id := range a.AudienceIDs {
			if id == u.DepartmentID.String() {
				return true
			}
		}
		return false
	case model.AudienceRole:
		u := m.users[viewer]
		roleSet := map[string]struct{}{}
		for _, rid := range u.RoleIDs {
			roleSet[rid.String()] = struct{}{}
		}
		for _, id := range a.AudienceIDs {
			if _, ok := roleSet[id]; ok {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (m *memRepo) List(_ context.Context, viewer uuid.UUID, staff, manage bool, req *dto.ListAnnouncementRequest) ([]model.AnnouncementNamed, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if req == nil {
		req = &dto.ListAnnouncementRequest{}
	}
	var out []model.AnnouncementNamed
	for _, a := range m.items {
		if a.DeletedAt.Valid {
			continue
		}
		if req.Category != "" && a.Category != req.Category {
			continue
		}
		if kw := strings.TrimSpace(req.Keyword); kw != "" &&
			!strings.Contains(a.Title, kw) && !strings.Contains(a.Content, kw) {
			continue
		}
		if req.PinnedOnly && !a.Pinned {
			continue
		}
		if req.UnreadOnly {
			if a.Status != model.StatusPublished {
				continue
			}
			if _, ok := m.reads[a.ID.String()+"/"+viewer.String()]; ok {
				continue
			}
		}
		if req.Status != nil {
			if a.Status != *req.Status {
				continue
			}
			if *req.Status == model.StatusDraft && !manage && a.AuthorID != viewer {
				continue
			}
			if !manage && *req.Status != model.StatusDraft && !m.inAudience(a, viewer) {
				continue
			}
		} else if manage {
			// all statuses
		} else if staff {
			if a.Status != model.StatusPublished && a.Status != model.StatusArchived && a.AuthorID != viewer {
				continue
			}
			if (a.Status == model.StatusPublished || a.Status == model.StatusArchived) && !m.inAudience(a, viewer) && a.AuthorID != viewer {
				continue
			}
		} else if a.Status != model.StatusPublished && a.Status != model.StatusArchived {
			continue
		} else if !m.inAudience(a, viewer) {
			continue
		}
		out = append(out, m.named(a, viewer))
	}
	return out, int64(len(out)), nil
}

func (m *memRepo) ListDueDrafts(_ context.Context, now time.Time, _ int) ([]model.Announcement, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.Announcement
	for _, a := range m.items {
		if a.DeletedAt.Valid || a.Status != model.StatusDraft || a.ScheduledAt == nil {
			continue
		}
		if !a.ScheduledAt.After(now) {
			cp := *a
			out = append(out, cp)
		}
	}
	return out, nil
}

func (m *memRepo) ListExpiredPublished(_ context.Context, now time.Time, _ int) ([]model.Announcement, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.Announcement
	for _, a := range m.items {
		if a.DeletedAt.Valid || a.Status != model.StatusPublished || a.ExpiresAt == nil {
			continue
		}
		if !a.ExpiresAt.After(now) {
			cp := *a
			out = append(out, cp)
		}
	}
	return out, nil
}

func (m *memRepo) MarkRead(_ context.Context, announcementID, userID uuid.UUID, at time.Time, duration int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := announcementID.String() + "/" + userID.String()
	if prev, ok := m.reads[key]; ok {
		if duration > prev.DurationSeconds {
			prev.DurationSeconds = duration
			m.reads[key] = prev
		}
		return nil
	}
	if duration < 0 {
		duration = 0
	}
	m.reads[key] = model.AnnouncementRead{
		ID:              uuid.New(),
		AnnouncementID:  announcementID,
		UserID:          userID,
		ReadAt:          at,
		DurationSeconds: duration,
	}
	return nil
}

func (m *memRepo) UnreadCount(_ context.Context, userID uuid.UUID) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for _, a := range m.items {
		if a.DeletedAt.Valid || a.Status != model.StatusPublished {
			continue
		}
		if !m.inAudience(a, userID) {
			continue
		}
		if _, ok := m.reads[a.ID.String()+"/"+userID.String()]; !ok {
			n++
		}
	}
	return n, nil
}

func (m *memRepo) ListReaders(_ context.Context, announcementID uuid.UUID) ([]model.ReaderNamed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.ReaderNamed
	prefix := announcementID.String() + "/"
	for key, rd := range m.reads {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		name := "读者"
		if u, ok := m.users[rd.UserID]; ok && u.RealName != "" {
			name = u.RealName
		}
		out = append(out, model.ReaderNamed{
			UserID:          rd.UserID,
			RealName:        name,
			Username:        name,
			ReadAt:          rd.ReadAt,
			DurationSeconds: rd.DurationSeconds,
		})
	}
	return out, nil
}

func (m *memRepo) CountActiveUsers(_ context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return int64(len(m.activeIDs)), nil
}

func (m *memRepo) ListActiveUserIDs(_ context.Context) ([]uuid.UUID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := append([]uuid.UUID{}, m.activeIDs...)
	return out, nil
}

func (m *memRepo) ListNamedUsers(_ context.Context, ids []uuid.UUID) ([]model.NamedUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.NamedUser
	for _, id := range ids {
		if u, ok := m.users[id]; ok {
			out = append(out, u)
			continue
		}
		out = append(out, model.NamedUser{ID: id, RealName: id.String()[:8], Username: "user"})
	}
	return out, nil
}

func (m *memRepo) ListActiveUserIDsAmong(_ context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	allow := map[uuid.UUID]struct{}{}
	for _, id := range m.activeIDs {
		allow[id] = struct{}{}
	}
	var out []uuid.UUID
	for _, id := range ids {
		if _, ok := allow[id]; ok {
			out = append(out, id)
		}
	}
	return out, nil
}

func (m *memRepo) ListActiveUserIDsByDepartments(_ context.Context, deptIDs []uuid.UUID) ([]uuid.UUID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	want := map[uuid.UUID]struct{}{}
	for _, id := range deptIDs {
		want[id] = struct{}{}
	}
	var out []uuid.UUID
	for _, id := range m.activeIDs {
		u := m.users[id]
		if u.DepartmentID != nil {
			if _, ok := want[*u.DepartmentID]; ok {
				out = append(out, id)
			}
		}
	}
	return out, nil
}

func (m *memRepo) ListActiveUserIDsByRoles(_ context.Context, roleIDs []uuid.UUID) ([]uuid.UUID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	want := map[uuid.UUID]struct{}{}
	for _, id := range roleIDs {
		want[id] = struct{}{}
	}
	var out []uuid.UUID
	for _, id := range m.activeIDs {
		u := m.users[id]
		for _, rid := range u.RoleIDs {
			if _, ok := want[rid]; ok {
				out = append(out, id)
				break
			}
		}
	}
	return out, nil
}

func (m *memRepo) UserInAudience(_ context.Context, userID uuid.UUID, audienceType string, audienceIDs []string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.inAudience(&model.Announcement{AudienceType: audienceType, AudienceIDs: audienceIDs}, userID), nil
}

func (m *memRepo) GetUser(_ context.Context, id uuid.UUID) (*model.NamedUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if u, ok := m.users[id]; ok {
		cp := u
		return &cp, nil
	}
	return &model.NamedUser{ID: id, RealName: "测试用户", Username: "tester"}, nil
}

type recNotify struct {
	mu    sync.Mutex
	calls []notifyCall
}

type notifyCall struct {
	users    []uuid.UUID
	template string
	vars     map[string]interface{}
}

func (n *recNotify) Send(_ context.Context, userIDs []uuid.UUID, template string, vars map[string]interface{}) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	cp := append([]uuid.UUID{}, userIDs...)
	n.calls = append(n.calls, notifyCall{users: cp, template: template, vars: vars})
	return nil
}

func newTestSvc() (*announcementService, *memRepo, *recNotify) {
	repo := newMemRepo()
	n := &recNotify{}
	svc := &announcementService{rows: repo, notify: n, now: func() time.Time {
		return time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC)
	}}
	return svc, repo, n
}
