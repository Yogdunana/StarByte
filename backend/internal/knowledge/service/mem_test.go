package service

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/knowledge/dto"
	"github.com/Yogdunana/StarByte/backend/internal/knowledge/model"
	"github.com/Yogdunana/StarByte/backend/internal/knowledge/repo"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type memRepo struct {
	mu    sync.Mutex
	cats  map[uuid.UUID]*model.Category
	docs  map[uuid.UUID]*model.Doc
	vers  map[string]*model.DocVersion
	atts  map[string]*model.Attachment
	users map[uuid.UUID]model.NamedUser
}

func newMemRepo() *memRepo {
	return &memRepo{
		cats:  map[uuid.UUID]*model.Category{},
		docs:  map[uuid.UUID]*model.Doc{},
		vers:  map[string]*model.DocVersion{},
		atts:  map[string]*model.Attachment{},
		users: map[uuid.UUID]model.NamedUser{},
	}
}

func (m *memRepo) addUser(id uuid.UUID, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[id] = model.NamedUser{ID: id, RealName: name, Username: name}
}

func (m *memRepo) CreateCategory(_ context.Context, c *model.Category) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *c
	m.cats[c.ID] = &cp
	return nil
}

func (m *memRepo) UpdateCategory(_ context.Context, c *model.Category) error {
	return m.CreateCategory(context.Background(), c)
}

func (m *memRepo) DeleteCategory(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if row, ok := m.cats[id]; ok {
		row.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	}
	return nil
}

func (m *memRepo) GetCategory(_ context.Context, id uuid.UUID) (*model.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.cats[id]
	if row == nil || row.DeletedAt.Valid {
		return nil, nil
	}
	cp := *row
	return &cp, nil
}

func (m *memRepo) GetCategoryBySlug(_ context.Context, slug string) (*model.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.cats {
		if c.Slug == slug && !c.DeletedAt.Valid {
			cp := *c
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *memRepo) ListCategories(_ context.Context) ([]model.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.Category
	for _, c := range m.cats {
		if !c.DeletedAt.Valid {
			out = append(out, *c)
		}
	}
	return out, nil
}

func (m *memRepo) CountDocsInCategory(_ context.Context, id uuid.UUID) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for _, d := range m.docs {
		if d.CategoryID != nil && *d.CategoryID == id && !d.DeletedAt.Valid {
			n++
		}
	}
	return n, nil
}

func (m *memRepo) CreateDoc(_ context.Context, d *model.Doc) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *d
	m.docs[d.ID] = &cp
	return nil
}

func (m *memRepo) UpdateDoc(_ context.Context, d *model.Doc) error {
	return m.CreateDoc(context.Background(), d)
}

func (m *memRepo) DeleteDoc(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if row, ok := m.docs[id]; ok {
		row.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	}
	return nil
}

func (m *memRepo) GetDoc(_ context.Context, id uuid.UUID) (*model.Doc, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.docs[id]
	if row == nil || row.DeletedAt.Valid {
		return nil, nil
	}
	cp := *row
	return &cp, nil
}

func (m *memRepo) GetDocBySlug(_ context.Context, kind, slug string) (*model.Doc, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, d := range m.docs {
		if d.Slug == slug && !d.DeletedAt.Valid && (kind == "" || d.Kind == kind) {
			cp := *d
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *memRepo) named(d *model.Doc) model.DocNamed {
	name := "作者"
	if u, ok := m.users[d.AuthorID]; ok && u.RealName != "" {
		name = u.RealName
	}
	catName := ""
	if d.CategoryID != nil {
		if c, ok := m.cats[*d.CategoryID]; ok {
			catName = c.Name
		}
	}
	return model.DocNamed{Doc: *d, AuthorName: name, CategoryName: catName}
}

func (m *memRepo) GetDocNamed(ctx context.Context, id uuid.UUID) (*model.DocNamed, error) {
	d, err := m.GetDoc(ctx, id)
	if err != nil || d == nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.named(d)
	return &row, nil
}

func (m *memRepo) GetDocNamedBySlug(ctx context.Context, kind, slug string) (*model.DocNamed, error) {
	d, err := m.GetDocBySlug(ctx, kind, slug)
	if err != nil || d == nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.named(d)
	return &row, nil
}

func (m *memRepo) collectDocs(req *dto.ListDocRequest, scope repo.ReadableScope) []model.DocNamed {
	if req == nil {
		req = &dto.ListDocRequest{}
	}
	var all []model.DocNamed
	for _, d := range m.docs {
		if d.DeletedAt.Valid {
			continue
		}
		if !repo.Visible(d, scope) {
			continue
		}
		if req.Kind != "" && d.Kind != req.Kind {
			continue
		}
		if req.Status != nil && d.Status != *req.Status {
			continue
		}
		if req.Visibility != "" && d.Visibility != req.Visibility {
			continue
		}
		if req.CategoryID != "" && (d.CategoryID == nil || d.CategoryID.String() != req.CategoryID) {
			continue
		}
		if kw := strings.TrimSpace(req.Keyword); kw != "" {
			blob := d.Title + d.Summary + d.Content
			if !strings.Contains(strings.ToLower(blob), strings.ToLower(kw)) {
				continue
			}
		}
		all = append(all, m.named(d))
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].UpdatedAt.Equal(all[j].UpdatedAt) {
			return all[i].Title < all[j].Title
		}
		return all[i].UpdatedAt.After(all[j].UpdatedAt)
	})
	return all
}

func pageDocs(all []model.DocNamed, page, pageSize int) ([]model.DocNamed, int64) {
	total := int64(len(all))
	if page <= 0 {
		page = 1
	}
	if pageSize < 0 {
		return all, total
	}
	if pageSize == 0 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start >= len(all) {
		return nil, total
	}
	end := start + pageSize
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], total
}

func (m *memRepo) ListDocs(_ context.Context, req *dto.ListDocRequest, scope repo.ReadableScope) ([]model.DocNamed, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if req == nil {
		req = &dto.ListDocRequest{}
	}
	all := m.collectDocs(req, scope)
	rows, total := pageDocs(all, req.Page, req.PageSize)
	return rows, total, nil
}

func (m *memRepo) SearchDocs(_ context.Context, query, kind string, page, pageSize int, scope repo.ReadableScope) ([]model.DocNamed, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	all := m.collectDocs(&dto.ListDocRequest{Kind: kind, Keyword: query}, scope)
	rows, total := pageDocs(all, page, pageSize)
	return rows, total, nil
}

func (m *memRepo) SlugTaken(_ context.Context, kind, slug string, except *uuid.UUID) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, d := range m.docs {
		if d.DeletedAt.Valid || d.Kind != kind || d.Slug != slug {
			continue
		}
		if except != nil && d.ID == *except {
			continue
		}
		return true, nil
	}
	return false, nil
}

func (m *memRepo) CreateVersion(_ context.Context, v *model.DocVersion) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := v.DocID.String() + "/" + itoa(v.Version)
	cp := *v
	m.vers[key] = &cp
	return nil
}

func (m *memRepo) ListVersions(_ context.Context, docID uuid.UUID) ([]model.VersionNamed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.VersionNamed
	for _, v := range m.vers {
		if v.DocID == docID {
			name := "编辑者"
			if u, ok := m.users[v.EditorID]; ok && u.RealName != "" {
				name = u.RealName
			}
			out = append(out, model.VersionNamed{DocVersion: *v, EditorName: name})
		}
	}
	return out, nil
}

func (m *memRepo) GetVersion(_ context.Context, docID uuid.UUID, version int) (*model.VersionNamed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := m.vers[docID.String()+"/"+itoa(version)]
	if v == nil {
		return nil, nil
	}
	name := "编辑者"
	if u, ok := m.users[v.EditorID]; ok && u.RealName != "" {
		name = u.RealName
	}
	return &model.VersionNamed{DocVersion: *v, EditorName: name}, nil
}

func (m *memRepo) AddAttachment(_ context.Context, a *model.Attachment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *a
	m.atts[a.DocID.String()+"/"+a.FileID.String()] = &cp
	return nil
}

func (m *memRepo) RemoveAttachment(_ context.Context, docID, fileID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.atts, docID.String()+"/"+fileID.String())
	return nil
}

func (m *memRepo) ListAttachments(_ context.Context, docID uuid.UUID) ([]model.AttachmentNamed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.AttachmentNamed
	for _, a := range m.atts {
		if a.DocID == docID {
			out = append(out, model.AttachmentNamed{Attachment: *a, FileName: "file.bin", FileSize: 10})
		}
	}
	return out, nil
}

func (m *memRepo) GetUser(_ context.Context, id uuid.UUID) (*model.NamedUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil, nil
	}
	return &u, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func newTestSvc() (*knowledgeService, *memRepo) {
	mem := newMemRepo()
	svc := New(mem).(*knowledgeService)
	svc.now = func() time.Time { return time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC) }
	return svc, mem
}
