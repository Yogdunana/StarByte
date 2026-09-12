package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/knowledge/dto"
	"github.com/Yogdunana/StarByte/backend/internal/knowledge/model"
	"github.com/Yogdunana/StarByte/backend/internal/knowledge/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

type Viewer struct {
	UserID     uuid.UUID
	Perms      []string
	CanCreate  bool
	CanUpdate  bool
	CanDelete  bool
	CanPublish bool
}

func (v Viewer) Authenticated() bool {
	return v.UserID != uuid.Nil
}

func (v Viewer) Has(code string) bool {
	for _, p := range v.Perms {
		if p == "*" || p == code {
			return true
		}
	}
	return false
}

func (v Viewer) Staff() bool {
	return v.CanCreate || v.CanUpdate || v.CanDelete || v.CanPublish
}

type Service interface {
	CreateCategory(ctx context.Context, viewer Viewer, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error)
	UpdateCategory(ctx context.Context, viewer Viewer, id uuid.UUID, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error)
	DeleteCategory(ctx context.Context, viewer Viewer, id uuid.UUID) error
	ListCategories(ctx context.Context) ([]*dto.CategoryResponse, error)

	Create(ctx context.Context, viewer Viewer, req *dto.CreateDocRequest) (*dto.DocResponse, error)
	Update(ctx context.Context, viewer Viewer, id uuid.UUID, req *dto.UpdateDocRequest) (*dto.DocResponse, error)
	Delete(ctx context.Context, viewer Viewer, id uuid.UUID) error
	Get(ctx context.Context, viewer Viewer, id uuid.UUID) (*dto.DocResponse, error)
	GetBySlug(ctx context.Context, viewer Viewer, kind, slug string) (*dto.DocResponse, error)
	List(ctx context.Context, viewer Viewer, req *dto.ListDocRequest) ([]*dto.DocResponse, int64, error)
	Publish(ctx context.Context, viewer Viewer, id uuid.UUID) (*dto.DocResponse, error)
	History(ctx context.Context, viewer Viewer, id uuid.UUID) ([]*dto.VersionResponse, error)
	GetVersion(ctx context.Context, viewer Viewer, id uuid.UUID, version int) (*dto.VersionResponse, error)
	Rollback(ctx context.Context, viewer Viewer, id uuid.UUID, version int) (*dto.DocResponse, error)
	Search(ctx context.Context, viewer Viewer, req *dto.SearchRequest) ([]*dto.DocResponse, int64, error)
	Tree(ctx context.Context, viewer Viewer) ([]*dto.TreeNode, error)

	Attach(ctx context.Context, viewer Viewer, id, fileID uuid.UUID) (*dto.DocResponse, error)
	Detach(ctx context.Context, viewer Viewer, id, fileID uuid.UUID) error
}

type knowledgeService struct {
	rows repo.Repository
	now  func() time.Time
}

func New(rows repo.Repository) Service {
	return &knowledgeService{rows: rows, now: time.Now}
}

func (s *knowledgeService) clock() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

func (s *knowledgeService) load(ctx context.Context, id uuid.UUID) (*model.Doc, error) {
	d, err := s.rows.GetDoc(ctx, id)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, notFound()
	}
	return d, nil
}

func canRead(v Viewer, d *model.Doc) error {
	if v.Staff() || (v.Authenticated() && d.AuthorID == v.UserID) {
		return nil
	}
	if d.Status != model.StatusPublished {
		if !v.Authenticated() {
			return loginRequired()
		}
		return notFound()
	}
	switch d.Visibility {
	case model.VisibilityPublic:
		return nil
	case model.VisibilityAuthenticated:
		if !v.Authenticated() {
			return loginRequired()
		}
		return nil
	case model.VisibilityPermission:
		if !v.Authenticated() {
			return loginRequired()
		}
		if d.PermissionCode == "" || v.Has(d.PermissionCode) {
			return nil
		}
		return noAccess("无权阅读该文档")
	default:
		return noAccess("无权阅读该文档")
	}
}

func readableForList(v Viewer, d *model.Doc) bool {
	return canRead(v, d) == nil
}

func (s *knowledgeService) snapshot(ctx context.Context, d *model.Doc, editor uuid.UUID) error {
	return s.rows.CreateVersion(ctx, &model.DocVersion{
		ID:        uuid.New(),
		DocID:     d.ID,
		Version:   d.Version,
		Title:     d.Title,
		Summary:   d.Summary,
		Content:   d.Content,
		EditorID:  editor,
		CreatedAt: s.clock(),
	})
}

func (s *knowledgeService) respond(ctx context.Context, id uuid.UUID, includeBody bool) (*dto.DocResponse, error) {
	row, err := s.rows.GetDocNamed(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, notFound()
	}
	atts, err := s.rows.ListAttachments(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDocResponse(row, atts, includeBody), nil
}

func (s *knowledgeService) CreateCategory(ctx context.Context, viewer Viewer, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	if !viewer.CanUpdate && !viewer.CanCreate {
		return nil, noAccess("无权维护分类")
	}
	if req == nil {
		return nil, response.NewError(response.CodeBadRequest, "参数错误")
	}
	slug := strings.TrimSpace(strings.ToLower(req.Slug))
	if !model.ValidSlug(slug) {
		return nil, invalidSlug()
	}
	if existing, err := s.rows.GetCategoryBySlug(ctx, slug); err != nil {
		return nil, err
	} else if existing != nil {
		return nil, slugConflict()
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, response.NewError(response.CodeBadRequest, "分类名称不能为空")
	}
	c := &model.Category{
		ID:        uuid.New(),
		Name:      name,
		Slug:      slug,
		SortOrder: req.SortOrder,
		CreatedAt: s.clock(),
		UpdatedAt: s.clock(),
	}
	if req.ParentID != nil && strings.TrimSpace(*req.ParentID) != "" {
		pid, err := uuid.Parse(strings.TrimSpace(*req.ParentID))
		if err != nil {
			return nil, response.NewError(response.CodeBadRequest, "无效的父分类")
		}
		parent, err := s.rows.GetCategory(ctx, pid)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, categoryNotFound()
		}
		c.ParentID = &pid
	}
	if err := s.rows.CreateCategory(ctx, c); err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}
	return toCategoryResponse(c), nil
}

func (s *knowledgeService) UpdateCategory(ctx context.Context, viewer Viewer, id uuid.UUID, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
	if !viewer.CanUpdate {
		return nil, noAccess("无权维护分类")
	}
	c, err := s.rows.GetCategory(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, categoryNotFound()
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, response.NewError(response.CodeBadRequest, "分类名称不能为空")
		}
		c.Name = name
	}
	if req.Slug != nil {
		slug := strings.TrimSpace(strings.ToLower(*req.Slug))
		if !model.ValidSlug(slug) {
			return nil, invalidSlug()
		}
		if existing, err := s.rows.GetCategoryBySlug(ctx, slug); err != nil {
			return nil, err
		} else if existing != nil && existing.ID != id {
			return nil, slugConflict()
		}
		c.Slug = slug
	}
	if req.SortOrder != nil {
		c.SortOrder = *req.SortOrder
	}
	if req.ClearParent {
		c.ParentID = nil
	} else if req.ParentID != nil && strings.TrimSpace(*req.ParentID) != "" {
		pid, err := uuid.Parse(strings.TrimSpace(*req.ParentID))
		if err != nil || pid == id {
			return nil, response.NewError(response.CodeBadRequest, "无效的父分类")
		}
		parent, err := s.rows.GetCategory(ctx, pid)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, categoryNotFound()
		}
		c.ParentID = &pid
	}
	c.UpdatedAt = s.clock()
	if err := s.rows.UpdateCategory(ctx, c); err != nil {
		return nil, fmt.Errorf("update category: %w", err)
	}
	return toCategoryResponse(c), nil
}

func (s *knowledgeService) DeleteCategory(ctx context.Context, viewer Viewer, id uuid.UUID) error {
	if !viewer.CanDelete {
		return noAccess("无权删除分类")
	}
	c, err := s.rows.GetCategory(ctx, id)
	if err != nil {
		return err
	}
	if c == nil {
		return categoryNotFound()
	}
	n, err := s.rows.CountDocsInCategory(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return invalidState("分类下仍有文档，无法删除")
	}
	return s.rows.DeleteCategory(ctx, id)
}

func (s *knowledgeService) ListCategories(ctx context.Context) ([]*dto.CategoryResponse, error) {
	rows, err := s.rows.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]*dto.CategoryResponse, len(rows))
	for i := range rows {
		byID[rows[i].ID] = toCategoryResponse(&rows[i])
	}
	var roots []*dto.CategoryResponse
	for i := range rows {
		node := byID[rows[i].ID]
		if rows[i].ParentID == nil {
			roots = append(roots, node)
			continue
		}
		if parent, ok := byID[*rows[i].ParentID]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}
	return roots, nil
}

func (s *knowledgeService) Create(ctx context.Context, viewer Viewer, req *dto.CreateDocRequest) (*dto.DocResponse, error) {
	if !viewer.CanCreate {
		return nil, noAccess("无权创建文档")
	}
	if req == nil {
		return nil, response.NewError(response.CodeBadRequest, "参数错误")
	}
	kind := strings.TrimSpace(req.Kind)
	if !model.ValidKind(kind) {
		return nil, response.NewError(response.CodeBadRequest, "文档类型不合法")
	}
	slug := strings.TrimSpace(strings.ToLower(req.Slug))
	if !model.ValidSlug(slug) {
		return nil, invalidSlug()
	}
	taken, err := s.rows.SlugTaken(ctx, kind, slug, nil)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, slugConflict()
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, response.NewError(response.CodeBadRequest, "标题不能为空")
	}
	if len(req.Content) > model.MaxContentLen {
		return nil, response.NewError(response.CodeBadRequest, "正文过长")
	}
	vis := req.Visibility
	if vis == "" {
		vis = model.VisibilityAuthenticated
	}
	if !model.ValidVisibility(vis) {
		return nil, invalidVis()
	}
	if vis == model.VisibilityPermission && strings.TrimSpace(req.PermissionCode) == "" {
		return nil, response.NewError(response.CodeBadRequest, "权限可见性必须填写权限码")
	}
	now := s.clock()
	d := &model.Doc{
		ID:             uuid.New(),
		Kind:           kind,
		Slug:           slug,
		Title:          title,
		Summary:        strings.TrimSpace(req.Summary),
		Content:        req.Content,
		Visibility:     vis,
		PermissionCode: strings.TrimSpace(req.PermissionCode),
		Status:         model.StatusDraft,
		Version:        1,
		AuthorID:       viewer.UserID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if req.CategoryID != nil && strings.TrimSpace(*req.CategoryID) != "" {
		cid, err := uuid.Parse(strings.TrimSpace(*req.CategoryID))
		if err != nil {
			return nil, response.NewError(response.CodeBadRequest, "无效的分类")
		}
		cat, err := s.rows.GetCategory(ctx, cid)
		if err != nil {
			return nil, err
		}
		if cat == nil {
			return nil, categoryNotFound()
		}
		d.CategoryID = &cid
	}
	if err := s.rows.CreateDoc(ctx, d); err != nil {
		return nil, fmt.Errorf("create doc: %w", err)
	}
	if err := s.snapshot(ctx, d, viewer.UserID); err != nil {
		return nil, fmt.Errorf("snapshot doc: %w", err)
	}
	return s.respond(ctx, d.ID, true)
}

func (s *knowledgeService) Update(ctx context.Context, viewer Viewer, id uuid.UUID, req *dto.UpdateDocRequest) (*dto.DocResponse, error) {
	if !viewer.CanUpdate && !viewer.Staff() {
		return nil, noAccess("无权更新文档")
	}
	if req == nil {
		return nil, response.NewError(response.CodeBadRequest, "参数错误")
	}
	d, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	if !viewer.CanUpdate && d.AuthorID != viewer.UserID {
		return nil, noAccess("无权更新该文档")
	}
	changed := false
	if req.Kind != nil {
		if !model.ValidKind(*req.Kind) {
			return nil, response.NewError(response.CodeBadRequest, "文档类型不合法")
		}
		if *req.Kind != d.Kind {
			d.Kind = *req.Kind
			changed = true
		}
	}
	if req.Slug != nil {
		slug := strings.TrimSpace(strings.ToLower(*req.Slug))
		if !model.ValidSlug(slug) {
			return nil, invalidSlug()
		}
		if slug != d.Slug {
			taken, err := s.rows.SlugTaken(ctx, d.Kind, slug, &d.ID)
			if err != nil {
				return nil, err
			}
			if taken {
				return nil, slugConflict()
			}
			d.Slug = slug
			changed = true
		}
	}
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			return nil, response.NewError(response.CodeBadRequest, "标题不能为空")
		}
		if title != d.Title {
			d.Title = title
			changed = true
		}
	}
	if req.Summary != nil && *req.Summary != d.Summary {
		if len(*req.Summary) > model.MaxSummaryLen {
			return nil, response.NewError(response.CodeBadRequest, "摘要过长")
		}
		d.Summary = *req.Summary
		changed = true
	}
	if req.Content != nil && *req.Content != d.Content {
		if len(*req.Content) > model.MaxContentLen {
			return nil, response.NewError(response.CodeBadRequest, "正文过长")
		}
		d.Content = *req.Content
		changed = true
	}
	if req.Visibility != nil {
		if !model.ValidVisibility(*req.Visibility) {
			return nil, invalidVis()
		}
		d.Visibility = *req.Visibility
		changed = true
	}
	if req.PermissionCode != nil {
		d.PermissionCode = strings.TrimSpace(*req.PermissionCode)
		changed = true
	}
	if d.Visibility == model.VisibilityPermission && d.PermissionCode == "" {
		return nil, response.NewError(response.CodeBadRequest, "权限可见性必须填写权限码")
	}
	if req.ClearCategory {
		d.CategoryID = nil
		changed = true
	} else if req.CategoryID != nil && strings.TrimSpace(*req.CategoryID) != "" {
		cid, err := uuid.Parse(strings.TrimSpace(*req.CategoryID))
		if err != nil {
			return nil, response.NewError(response.CodeBadRequest, "无效的分类")
		}
		cat, err := s.rows.GetCategory(ctx, cid)
		if err != nil {
			return nil, err
		}
		if cat == nil {
			return nil, categoryNotFound()
		}
		d.CategoryID = &cid
		changed = true
	}
	if !changed {
		return s.respond(ctx, d.ID, true)
	}
	d.Version++
	d.UpdatedAt = s.clock()
	if err := s.rows.UpdateDoc(ctx, d); err != nil {
		return nil, fmt.Errorf("update doc: %w", err)
	}
	if err := s.snapshot(ctx, d, viewer.UserID); err != nil {
		return nil, fmt.Errorf("snapshot doc: %w", err)
	}
	return s.respond(ctx, d.ID, true)
}

func (s *knowledgeService) Delete(ctx context.Context, viewer Viewer, id uuid.UUID) error {
	d, err := s.load(ctx, id)
	if err != nil {
		return err
	}
	if !viewer.CanDelete && d.AuthorID != viewer.UserID {
		return noAccess("无权删除该文档")
	}
	return s.rows.DeleteDoc(ctx, id)
}

func (s *knowledgeService) Get(ctx context.Context, viewer Viewer, id uuid.UUID) (*dto.DocResponse, error) {
	d, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := canRead(viewer, d); err != nil {
		return nil, err
	}
	return s.respond(ctx, id, true)
}

func (s *knowledgeService) GetBySlug(ctx context.Context, viewer Viewer, kind, slug string) (*dto.DocResponse, error) {
	row, err := s.rows.GetDocNamedBySlug(ctx, kind, slug)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, notFound()
	}
	if err := canRead(viewer, &row.Doc); err != nil {
		return nil, err
	}
	atts, err := s.rows.ListAttachments(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	return toDocResponse(row, atts, true), nil
}

func (s *knowledgeService) List(ctx context.Context, viewer Viewer, req *dto.ListDocRequest) ([]*dto.DocResponse, int64, error) {
	if req == nil {
		req = &dto.ListDocRequest{}
	}
	if req.Kind != "" && !model.ValidKind(req.Kind) {
		return nil, 0, response.NewError(response.CodeBadRequest, "文档类型不合法")
	}
	if req.Visibility != "" && !model.ValidVisibility(req.Visibility) {
		return nil, 0, invalidVis()
	}
	rows, total, err := s.rows.ListDocs(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	list := make([]*dto.DocResponse, 0, len(rows))
	for i := range rows {
		if !readableForList(viewer, &rows[i].Doc) {
			continue
		}
		list = append(list, toDocResponse(&rows[i], nil, req.IncludeBody))
	}
	if !viewer.Staff() {
		total = int64(len(list))
	}
	return list, total, nil
}

func (s *knowledgeService) Publish(ctx context.Context, viewer Viewer, id uuid.UUID) (*dto.DocResponse, error) {
	if !viewer.CanPublish {
		return nil, noAccess("无权发布文档")
	}
	d, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	if d.Status == model.StatusPublished {
		return s.respond(ctx, id, true)
	}
	now := s.clock()
	d.Status = model.StatusPublished
	d.PublishedAt = &now
	d.UpdatedAt = now
	if err := s.rows.UpdateDoc(ctx, d); err != nil {
		return nil, fmt.Errorf("publish doc: %w", err)
	}
	return s.respond(ctx, id, true)
}

func (s *knowledgeService) History(ctx context.Context, viewer Viewer, id uuid.UUID) ([]*dto.VersionResponse, error) {
	d, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := canRead(viewer, d); err != nil {
		return nil, err
	}
	rows, err := s.rows.ListVersions(ctx, id)
	if err != nil {
		return nil, err
	}
	list := make([]*dto.VersionResponse, 0, len(rows))
	for i := range rows {
		list = append(list, toVersionResponse(&rows[i], d.Version, false))
	}
	return list, nil
}

func (s *knowledgeService) GetVersion(ctx context.Context, viewer Viewer, id uuid.UUID, version int) (*dto.VersionResponse, error) {
	d, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := canRead(viewer, d); err != nil {
		return nil, err
	}
	row, err := s.rows.GetVersion(ctx, id, version)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, notFound()
	}
	return toVersionResponse(row, d.Version, true), nil
}

func (s *knowledgeService) Rollback(ctx context.Context, viewer Viewer, id uuid.UUID, version int) (*dto.DocResponse, error) {
	if !viewer.CanUpdate {
		return nil, noAccess("无权回滚文档")
	}
	d, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	old, err := s.rows.GetVersion(ctx, id, version)
	if err != nil {
		return nil, err
	}
	if old == nil {
		return nil, notFound()
	}
	d.Title = old.Title
	d.Summary = old.Summary
	d.Content = old.Content
	d.Version++
	d.UpdatedAt = s.clock()
	if err := s.rows.UpdateDoc(ctx, d); err != nil {
		return nil, fmt.Errorf("rollback doc: %w", err)
	}
	if err := s.snapshot(ctx, d, viewer.UserID); err != nil {
		return nil, fmt.Errorf("snapshot rollback: %w", err)
	}
	return s.respond(ctx, d.ID, true)
}

func (s *knowledgeService) Search(ctx context.Context, viewer Viewer, req *dto.SearchRequest) ([]*dto.DocResponse, int64, error) {
	if req == nil || strings.TrimSpace(req.Query) == "" {
		return nil, 0, response.NewError(response.CodeBadRequest, "请输入搜索词")
	}
	if req.Kind != "" && !model.ValidKind(req.Kind) {
		return nil, 0, response.NewError(response.CodeBadRequest, "文档类型不合法")
	}
	onlyPublic := !viewer.Authenticated()
	var author *uuid.UUID
	if viewer.Authenticated() {
		author = &viewer.UserID
	}
	rows, total, err := s.rows.SearchDocs(ctx, req.Query, req.Kind, req.Page, req.PageSize, onlyPublic, author, viewer.Staff())
	if err != nil {
		return nil, 0, err
	}
	list := make([]*dto.DocResponse, 0, len(rows))
	for i := range rows {
		if !readableForList(viewer, &rows[i].Doc) {
			continue
		}
		list = append(list, toDocResponse(&rows[i], nil, false))
	}
	if !viewer.Staff() {
		total = int64(len(list))
	}
	return list, total, nil
}

func (s *knowledgeService) Tree(ctx context.Context, viewer Viewer) ([]*dto.TreeNode, error) {
	cats, err := s.rows.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	docs, _, err := s.rows.ListDocs(ctx, &dto.ListDocRequest{Page: 1, PageSize: 500})
	if err != nil {
		return nil, err
	}
	catNodes := make(map[uuid.UUID]*dto.TreeNode, len(cats))
	for i := range cats {
		catNodes[cats[i].ID] = &dto.TreeNode{
			ID:    cats[i].ID.String(),
			Type:  "category",
			Title: cats[i].Name,
			Slug:  cats[i].Slug,
		}
	}
	uncategorized := &dto.TreeNode{ID: "uncategorized", Type: "category", Title: "未分类"}
	for i := range docs {
		if !readableForList(viewer, &docs[i].Doc) {
			continue
		}
		node := &dto.TreeNode{
			ID:    docs[i].ID.String(),
			Type:  "doc",
			Title: docs[i].Title,
			Slug:  docs[i].Slug,
			Path:  model.PublicPath(docs[i].Kind, docs[i].Slug),
			Kind:  docs[i].Kind,
		}
		if docs[i].CategoryID != nil {
			if parent, ok := catNodes[*docs[i].CategoryID]; ok {
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		uncategorized.Children = append(uncategorized.Children, node)
	}
	var roots []*dto.TreeNode
	for i := range cats {
		node := catNodes[cats[i].ID]
		if cats[i].ParentID != nil {
			if parent, ok := catNodes[*cats[i].ParentID]; ok {
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}
	if len(uncategorized.Children) > 0 {
		roots = append(roots, uncategorized)
	}
	return pruneEmptyCategories(roots), nil
}

func pruneEmptyCategories(nodes []*dto.TreeNode) []*dto.TreeNode {
	out := make([]*dto.TreeNode, 0, len(nodes))
	for _, n := range nodes {
		if n.Type == "category" {
			n.Children = pruneEmptyCategories(n.Children)
			if len(n.Children) == 0 {
				continue
			}
		}
		out = append(out, n)
	}
	return out
}

func (s *knowledgeService) Attach(ctx context.Context, viewer Viewer, id, fileID uuid.UUID) (*dto.DocResponse, error) {
	if !viewer.CanUpdate {
		return nil, noAccess("无权添加附件")
	}
	if _, err := s.load(ctx, id); err != nil {
		return nil, err
	}
	if err := s.rows.AddAttachment(ctx, &model.Attachment{
		ID:        uuid.New(),
		DocID:     id,
		FileID:    fileID,
		CreatedAt: s.clock(),
	}); err != nil {
		return nil, fmt.Errorf("attach file: %w", err)
	}
	return s.respond(ctx, id, true)
}

func (s *knowledgeService) Detach(ctx context.Context, viewer Viewer, id, fileID uuid.UUID) error {
	if !viewer.CanUpdate {
		return noAccess("无权移除附件")
	}
	if _, err := s.load(ctx, id); err != nil {
		return err
	}
	return s.rows.RemoveAttachment(ctx, id, fileID)
}
