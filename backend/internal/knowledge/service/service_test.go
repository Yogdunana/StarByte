package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/knowledge/dto"
	"github.com/Yogdunana/StarByte/backend/internal/knowledge/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func codeOf(err error) int {
	if app, ok := err.(*response.AppError); ok {
		return app.Code
	}
	return 0
}

func editor(id uuid.UUID) Viewer {
	return Viewer{UserID: id, CanCreate: true, CanUpdate: true, CanDelete: true, CanPublish: true, Perms: []string{"*"}}
}

func reader(id uuid.UUID) Viewer {
	return Viewer{UserID: id, Perms: []string{"doc:read"}}
}

func anon() Viewer { return Viewer{} }

func mustCreate(t *testing.T, svc Service, author uuid.UUID, kind, slug, title string) *dto.DocResponse {
	t.Helper()
	resp, err := svc.Create(context.Background(), editor(author), &dto.CreateDocRequest{
		Kind: kind, Slug: slug, Title: title, Content: "正文 **" + title + "**", Visibility: model.VisibilityPublic,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	return resp
}

func TestCreateAndGetPublic(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "管理员")
	created := mustCreate(t, svc, author, model.KindPage, "about-us", "关于我们")
	if created.Status != model.StatusDraft || created.Version != 1 || created.Path != "/about-us" {
		t.Fatalf("got %+v", created)
	}
	id := uuid.MustParse(created.ID)
	if _, err := svc.Publish(context.Background(), editor(author), id); err != nil {
		t.Fatal(err)
	}
	got, err := svc.GetBySlug(context.Background(), anon(), model.KindPage, "about-us")
	if err != nil || got.Title != "关于我们" || got.Content == "" {
		t.Fatalf("public get: %+v err=%v", got, err)
	}
}

func TestCreate_InvalidSlugAndKind(t *testing.T) {
	svc, _ := newTestSvc()
	author := uuid.New()
	_, err := svc.Create(context.Background(), editor(author), &dto.CreateDocRequest{
		Kind: "note", Slug: "ok", Title: "x",
	})
	if err == nil {
		t.Fatal("expected invalid kind")
	}
	_, err = svc.Create(context.Background(), editor(author), &dto.CreateDocRequest{
		Kind: model.KindDoc, Slug: "Bad Slug", Title: "x",
	})
	if codeOf(err) != response.CodeKnowledgeInvalidSlug {
		t.Fatalf("slug code=%d", codeOf(err))
	}
}

func TestCreate_SlugConflict(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	mustCreate(t, svc, author, model.KindDoc, "charter", "章程")
	_, err := svc.Create(context.Background(), editor(author), &dto.CreateDocRequest{
		Kind: model.KindDoc, Slug: "charter", Title: "另一份",
	})
	if codeOf(err) != response.CodeKnowledgeConflict {
		t.Fatalf("code=%d", codeOf(err))
	}
}

func TestMembersOnlyRequiresLogin(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	resp, err := svc.Create(context.Background(), editor(author), &dto.CreateDocRequest{
		Kind: model.KindDoc, Slug: "user-manual", Title: "手册", Visibility: model.VisibilityAuthenticated,
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(resp.ID)
	if _, err := svc.Publish(context.Background(), editor(author), id); err != nil {
		t.Fatal(err)
	}
	_, err = svc.GetBySlug(context.Background(), anon(), model.KindDoc, "user-manual")
	if codeOf(err) != response.CodeKnowledgeLoginRequired {
		t.Fatalf("code=%d err=%v", codeOf(err), err)
	}
	got, err := svc.GetBySlug(context.Background(), reader(uuid.New()), model.KindDoc, "user-manual")
	if err != nil || got.Title != "手册" {
		t.Fatalf("auth read: %+v err=%v", got, err)
	}
}

func TestPermissionVisibility(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	resp, err := svc.Create(context.Background(), editor(author), &dto.CreateDocRequest{
		Kind: model.KindDoc, Slug: "staff-only", Title: "内部", Visibility: model.VisibilityPermission, PermissionCode: "doc:publish",
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(resp.ID)
	if _, err := svc.Publish(context.Background(), editor(author), id); err != nil {
		t.Fatal(err)
	}
	_, err = svc.GetBySlug(context.Background(), reader(uuid.New()), model.KindDoc, "staff-only")
	if codeOf(err) != response.CodeKnowledgeNoAccess {
		t.Fatalf("code=%d", codeOf(err))
	}
	staff := Viewer{UserID: uuid.New(), Perms: []string{"doc:publish"}}
	if _, err := svc.GetBySlug(context.Background(), staff, model.KindDoc, "staff-only"); err != nil {
		t.Fatal(err)
	}
}

func TestVersionAndRollback(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	created := mustCreate(t, svc, author, model.KindDoc, "vdoc", "v1标题")
	id := uuid.MustParse(created.ID)
	title := "v2标题"
	content := "第二版"
	updated, err := svc.Update(context.Background(), editor(author), id, &dto.UpdateDocRequest{Title: &title, Content: &content})
	if err != nil || updated.Version != 2 {
		t.Fatalf("update: %+v err=%v", updated, err)
	}
	hist, err := svc.History(context.Background(), editor(author), id)
	if err != nil || len(hist) < 2 {
		t.Fatalf("history=%d err=%v", len(hist), err)
	}
	ver, err := svc.GetVersion(context.Background(), editor(author), id, 1)
	if err != nil || ver.Title != "v1标题" {
		t.Fatalf("ver1=%+v err=%v", ver, err)
	}
	rolled, err := svc.Rollback(context.Background(), editor(author), id, 1)
	if err != nil || rolled.Title != "v1标题" || rolled.Version != 3 {
		t.Fatalf("rollback %+v err=%v", rolled, err)
	}
}

func TestSearchAndTree(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	cat, err := svc.CreateCategory(context.Background(), editor(author), &dto.CreateCategoryRequest{Name: "公开", Slug: "public"})
	if err != nil {
		t.Fatal(err)
	}
	cid := cat.ID
	created, err := svc.Create(context.Background(), editor(author), &dto.CreateDocRequest{
		Kind: model.KindDoc, Slug: "charter", Title: "计算机协会章程", Content: "总则", CategoryID: &cid, Visibility: model.VisibilityPublic,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Publish(context.Background(), editor(author), uuid.MustParse(created.ID)); err != nil {
		t.Fatal(err)
	}
	list, total, err := svc.Search(context.Background(), anon(), &dto.SearchRequest{Query: "章程"})
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("search total=%d list=%d err=%v", total, len(list), err)
	}
	tree, err := svc.Tree(context.Background(), anon())
	if err != nil || len(tree) == 0 {
		t.Fatalf("tree=%+v err=%v", tree, err)
	}
}

func TestListFiltersDraftFromPublic(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	mustCreate(t, svc, author, model.KindDoc, "draft-doc", "草稿")
	status := model.StatusPublished
	list, _, err := svc.List(context.Background(), anon(), &dto.ListDocRequest{Kind: model.KindDoc, Status: &status})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("anon should not see drafts: %+v", list)
	}
}

func TestDeleteAndNotFound(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	created := mustCreate(t, svc, author, model.KindDoc, "gone", "删")
	id := uuid.MustParse(created.ID)
	if err := svc.Delete(context.Background(), editor(author), id); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Get(context.Background(), editor(author), id); codeOf(err) != response.CodeKnowledgeNotFound {
		t.Fatalf("code=%d", codeOf(err))
	}
}

func TestCategoryCRUD(t *testing.T) {
	svc, _ := newTestSvc()
	author := uuid.New()
	created, err := svc.CreateCategory(context.Background(), editor(author), &dto.CreateCategoryRequest{Name: "手册", Slug: "handbook", SortOrder: 2})
	if err != nil {
		t.Fatal(err)
	}
	name := "成员手册"
	updated, err := svc.UpdateCategory(context.Background(), editor(author), uuid.MustParse(created.ID), &dto.UpdateCategoryRequest{Name: &name})
	if err != nil || updated.Name != name {
		t.Fatalf("%+v err=%v", updated, err)
	}
	if err := svc.DeleteCategory(context.Background(), editor(author), uuid.MustParse(created.ID)); err != nil {
		t.Fatal(err)
	}
}

func TestAttachDetach(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	created := mustCreate(t, svc, author, model.KindDoc, "with-file", "附件")
	id := uuid.MustParse(created.ID)
	fileID := uuid.New()
	got, err := svc.Attach(context.Background(), editor(author), id, fileID)
	if err != nil || len(got.Attachments) != 1 {
		t.Fatalf("%+v err=%v", got, err)
	}
	if err := svc.Detach(context.Background(), editor(author), id, fileID); err != nil {
		t.Fatal(err)
	}
}

func TestCreateDeniedWithoutPerm(t *testing.T) {
	svc, _ := newTestSvc()
	_, err := svc.Create(context.Background(), reader(uuid.New()), &dto.CreateDocRequest{
		Kind: model.KindDoc, Slug: "x", Title: "x",
	})
	if codeOf(err) != response.CodeKnowledgeNoAccess {
		t.Fatalf("code=%d", codeOf(err))
	}
}

func TestErrorCodesInKnowledgeRange(t *testing.T) {
	codes := []int{
		response.CodeKnowledgeNotFound,
		response.CodeKnowledgeInvalidState,
		response.CodeKnowledgeNoAccess,
		response.CodeKnowledgeLoginRequired,
		response.CodeKnowledgeInvalidSlug,
		response.CodeKnowledgeConflict,
		response.CodeKnowledgeInvalidVis,
	}
	for _, c := range codes {
		if c < 33000 || c > 33999 {
			t.Fatalf("code %d outside 33000-33999", c)
		}
	}
}

func TestValidHelpers(t *testing.T) {
	if !model.ValidSlug("about-us") || model.ValidSlug("About") || !model.ValidKind(model.KindPage) {
		t.Fatal("slug/kind helpers")
	}
	if model.PublicPath(model.KindDoc, "x") != "/docs/x" {
		t.Fatal("path")
	}
	v := Viewer{Perms: []string{"doc:read"}}
	if !v.Has("doc:read") || v.Has("doc:publish") || v.Staff() {
		t.Fatal("viewer helpers")
	}
}

func TestPermissionCreateRequiresCode(t *testing.T) {
	svc, _ := newTestSvc()
	_, err := svc.Create(context.Background(), editor(uuid.New()), &dto.CreateDocRequest{
		Kind: model.KindDoc, Slug: "secret", Title: "密", Visibility: model.VisibilityPermission,
	})
	if err == nil {
		t.Fatal("expected permission code required")
	}
}

func TestPublishDenied(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	created := mustCreate(t, svc, author, model.KindDoc, "p", "p")
	_, err := svc.Publish(context.Background(), reader(author), uuid.MustParse(created.ID))
	if codeOf(err) != response.CodeKnowledgeNoAccess {
		t.Fatalf("code=%d", codeOf(err))
	}
}

func TestUpdateDocAndCategoryTree(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	parent, err := svc.CreateCategory(context.Background(), editor(author), &dto.CreateCategoryRequest{Name: "根", Slug: "root"})
	if err != nil {
		t.Fatal(err)
	}
	pid := parent.ID
	child, err := svc.CreateCategory(context.Background(), editor(author), &dto.CreateCategoryRequest{Name: "子", Slug: "child", ParentID: &pid})
	if err != nil {
		t.Fatal(err)
	}
	sort := 9
	slug := "child-2"
	if _, err := svc.UpdateCategory(context.Background(), editor(author), uuid.MustParse(child.ID), &dto.UpdateCategoryRequest{Slug: &slug, SortOrder: &sort, ClearParent: true}); err != nil {
		t.Fatal(err)
	}
	cats, err := svc.ListCategories(context.Background())
	if err != nil || len(cats) < 2 {
		t.Fatalf("cats=%d err=%v", len(cats), err)
	}
	created := mustCreate(t, svc, author, model.KindDoc, "upd", "旧")
	id := uuid.MustParse(created.ID)
	newSlug := "upd-2"
	kind := model.KindPage
	vis := model.VisibilityAuthenticated
	sum := "摘要"
	if _, err := svc.Update(context.Background(), editor(author), id, &dto.UpdateDocRequest{
		Slug: &newSlug, Kind: &kind, Visibility: &vis, Summary: &sum, CategoryID: &pid,
	}); err != nil {
		t.Fatal(err)
	}
	got, err := svc.Get(context.Background(), editor(author), id)
	if err != nil || got.Path != "/upd-2" || got.Kind != model.KindPage {
		t.Fatalf("%+v err=%v", got, err)
	}
	if _, err := svc.Update(context.Background(), editor(author), id, &dto.UpdateDocRequest{ClearCategory: true}); err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteCategory(context.Background(), editor(author), uuid.MustParse(parent.ID)); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateUnchangedMetadataDoesNotBumpVersion(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	cat, err := svc.CreateCategory(context.Background(), editor(author), &dto.CreateCategoryRequest{Name: "夹", Slug: "folder"})
	if err != nil {
		t.Fatal(err)
	}
	cid := cat.ID
	created, err := svc.Create(context.Background(), editor(author), &dto.CreateDocRequest{
		Kind: model.KindDoc, Slug: "same", Title: "同", CategoryID: &cid, Visibility: model.VisibilityAuthenticated,
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(created.ID)
	title := created.Title
	vis := model.VisibilityAuthenticated
	perm := ""
	updated, err := svc.Update(context.Background(), editor(author), id, &dto.UpdateDocRequest{
		Title: &title, Visibility: &vis, PermissionCode: &perm, CategoryID: &cid,
	})
	if err != nil || updated.Version != 1 {
		t.Fatalf("version=%d err=%v", updated.Version, err)
	}
}

func TestDeleteCategoryBlockedByChildren(t *testing.T) {
	svc, _ := newTestSvc()
	author := uuid.New()
	parent, err := svc.CreateCategory(context.Background(), editor(author), &dto.CreateCategoryRequest{Name: "根", Slug: "root-del"})
	if err != nil {
		t.Fatal(err)
	}
	pid := parent.ID
	if _, err := svc.CreateCategory(context.Background(), editor(author), &dto.CreateCategoryRequest{Name: "子", Slug: "child-del", ParentID: &pid}); err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteCategory(context.Background(), editor(author), uuid.MustParse(pid)); codeOf(err) != response.CodeKnowledgeInvalidState {
		t.Fatalf("code=%d", codeOf(err))
	}
}

func TestUpdateCategoryRejectsCycle(t *testing.T) {
	svc, _ := newTestSvc()
	author := uuid.New()
	a, err := svc.CreateCategory(context.Background(), editor(author), &dto.CreateCategoryRequest{Name: "A", Slug: "node-a"})
	if err != nil {
		t.Fatal(err)
	}
	aid := a.ID
	b, err := svc.CreateCategory(context.Background(), editor(author), &dto.CreateCategoryRequest{Name: "B", Slug: "node-b", ParentID: &aid})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateCategory(context.Background(), editor(author), uuid.MustParse(aid), &dto.UpdateCategoryRequest{ParentID: &b.ID}); err == nil {
		t.Fatal("expected cycle")
	}
}

func TestPublishedVisibilityRequiresPublishPerm(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	created, err := svc.Create(context.Background(), editor(author), &dto.CreateDocRequest{
		Kind: model.KindDoc, Slug: "members-only", Title: "手册", Visibility: model.VisibilityAuthenticated,
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(created.ID)
	if _, err := svc.Publish(context.Background(), editor(author), id); err != nil {
		t.Fatal(err)
	}
	updater := Viewer{UserID: author, CanUpdate: true}
	vis := model.VisibilityPublic
	if _, err := svc.Update(context.Background(), updater, id, &dto.UpdateDocRequest{Visibility: &vis}); codeOf(err) != response.CodeKnowledgeNoAccess {
		t.Fatalf("code=%d err=%v", codeOf(err), err)
	}
}

func TestDeleteCategoryBlockedByDocs(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	cat, err := svc.CreateCategory(context.Background(), editor(author), &dto.CreateCategoryRequest{Name: "占", Slug: "hold"})
	if err != nil {
		t.Fatal(err)
	}
	cid := cat.ID
	if _, err := svc.Create(context.Background(), editor(author), &dto.CreateDocRequest{
		Kind: model.KindDoc, Slug: "in-cat", Title: "x", CategoryID: &cid,
	}); err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteCategory(context.Background(), editor(author), uuid.MustParse(cid)); codeOf(err) != response.CodeKnowledgeInvalidState {
		t.Fatalf("code=%d", codeOf(err))
	}
}

func TestSearchRequiresQuery(t *testing.T) {
	svc, _ := newTestSvc()
	_, _, err := svc.Search(context.Background(), anon(), &dto.SearchRequest{})
	if err == nil {
		t.Fatal("expected query")
	}
	_, _, err = svc.List(context.Background(), anon(), &dto.ListDocRequest{Kind: "nope"})
	if err == nil {
		t.Fatal("expected kind")
	}
}

func TestGetVersionMissing(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	created := mustCreate(t, svc, author, model.KindDoc, "hist", "h")
	id := uuid.MustParse(created.ID)
	if _, err := svc.GetVersion(context.Background(), editor(author), id, 9); codeOf(err) != response.CodeKnowledgeNotFound {
		t.Fatalf("code=%d", codeOf(err))
	}
	if _, err := svc.Rollback(context.Background(), reader(author), id, 1); codeOf(err) != response.CodeKnowledgeNoAccess {
		t.Fatalf("code=%d", codeOf(err))
	}
}

func TestUncategorizedTree(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	created := mustCreate(t, svc, author, model.KindDoc, "loose", "散")
	if _, err := svc.Publish(context.Background(), editor(author), uuid.MustParse(created.ID)); err != nil {
		t.Fatal(err)
	}
	tree, err := svc.Tree(context.Background(), anon())
	if err != nil || len(tree) == 0 {
		t.Fatalf("tree=%+v err=%v", tree, err)
	}
}

func TestDraftHiddenFromOthers(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	created := mustCreate(t, svc, author, model.KindDoc, "hid", "藏")
	_, err := svc.Get(context.Background(), reader(uuid.New()), uuid.MustParse(created.ID))
	if codeOf(err) != response.CodeKnowledgeNotFound {
		t.Fatalf("code=%d", codeOf(err))
	}
	if _, err := svc.Get(context.Background(), anon(), uuid.MustParse(created.ID)); codeOf(err) != response.CodeKnowledgeLoginRequired {
		t.Fatalf("code=%d", codeOf(err))
	}
}

func TestInvalidVisibility(t *testing.T) {
	svc, _ := newTestSvc()
	_, err := svc.Create(context.Background(), editor(uuid.New()), &dto.CreateDocRequest{
		Kind: model.KindDoc, Slug: "badvis", Title: "x", Visibility: "everyone",
	})
	if codeOf(err) != response.CodeKnowledgeInvalidVis {
		t.Fatalf("code=%d", codeOf(err))
	}
}

func TestAuthorCanDeleteOwn(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	created := mustCreate(t, svc, author, model.KindDoc, "mine", "我的")
	v := Viewer{UserID: author}
	if err := svc.Delete(context.Background(), v, uuid.MustParse(created.ID)); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateKindChecksSlugUniqueness(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	page := mustCreate(t, svc, author, model.KindPage, "same-slug", "独立页")
	mustCreate(t, svc, author, model.KindDoc, "same-slug", "文档")
	kind := model.KindDoc
	_, err := svc.Update(context.Background(), editor(author), uuid.MustParse(page.ID), &dto.UpdateDocRequest{Kind: &kind})
	if codeOf(err) != response.CodeKnowledgeConflict {
		t.Fatalf("kind-only collision code=%d err=%v", codeOf(err), err)
	}
	kindPage := model.KindPage
	if _, err := svc.Update(context.Background(), editor(author), uuid.MustParse(page.ID), &dto.UpdateDocRequest{Kind: &kindPage}); err != nil {
		t.Fatalf("same kind should be a no-op: %v", err)
	}
}

func TestListSearchTreePaginateKeepsPublic(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "a")
	base := time.Date(2026, 9, 12, 8, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return base }
	public := mustCreate(t, svc, author, model.KindDoc, "charter-pub", "公开章程")
	if _, err := svc.Publish(context.Background(), editor(author), uuid.MustParse(public.ID)); err != nil {
		t.Fatal(err)
	}
	for i, slug := range []string{"draft-a", "draft-b", "draft-c"} {
		n := i + 1
		svc.now = func() time.Time { return base.Add(time.Duration(n) * time.Hour) }
		mustCreate(t, svc, author, model.KindDoc, slug, "草稿"+slug)
	}

	list, total, err := svc.List(context.Background(), anon(), &dto.ListDocRequest{Page: 1, PageSize: 2})
	if err != nil || total != 1 || len(list) != 1 || list[0].Slug != "charter-pub" {
		t.Fatalf("list total=%d n=%d err=%v list=%+v", total, len(list), err, list)
	}
	found, searchTotal, err := svc.Search(context.Background(), anon(), &dto.SearchRequest{Query: "草稿", Page: 1, PageSize: 2})
	if err != nil || searchTotal != 0 || len(found) != 0 {
		t.Fatalf("anon search drafts total=%d n=%d err=%v", searchTotal, len(found), err)
	}
	hit, hitTotal, err := svc.Search(context.Background(), anon(), &dto.SearchRequest{Query: "章程", Page: 1, PageSize: 1})
	if err != nil || hitTotal != 1 || len(hit) != 1 || hit[0].Slug != "charter-pub" {
		t.Fatalf("search public total=%d n=%d err=%v", hitTotal, len(hit), err)
	}
	tree, err := svc.Tree(context.Background(), anon())
	if err != nil {
		t.Fatal(err)
	}
	var saw bool
	var walk func(nodes []*dto.TreeNode)
	walk = func(nodes []*dto.TreeNode) {
		for _, n := range nodes {
			if n.Slug == "charter-pub" {
				saw = true
			}
			walk(n.Children)
		}
	}
	walk(tree)
	if !saw {
		t.Fatalf("public doc missing from tree: %+v", tree)
	}
}

func TestRoleAudienceAndRevocation(t *testing.T) {
	svc, mem := newTestSvc()
	author := uuid.New()
	mem.addUser(author, "editor")
	ctx := context.Background()
	doc, err := svc.Create(ctx, editor(author), &dto.CreateDocRequest{Kind: "doc", Slug: "role-notice", Title: "Role notice", Visibility: model.VisibilityRole, AllowedRoles: []string{"member"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.Publish(ctx, editor(author), uuid.MustParse(doc.ID)); err != nil {
		t.Fatal(err)
	}
	user := Viewer{UserID: uuid.New(), Roles: []string{"user"}}
	if _, err = svc.Get(ctx, user, uuid.MustParse(doc.ID)); err == nil {
		t.Fatal("User read a member-only document")
	}
	rows, total, err := svc.List(ctx, user, &dto.ListDocRequest{})
	if err != nil || total != 0 || len(rows) != 0 {
		t.Fatal("member-only document leaked through list", err)
	}
	user.Roles = append(user.Roles, "member")
	if _, err = svc.Get(ctx, user, uuid.MustParse(doc.ID)); err != nil {
		t.Fatal(err)
	}
	none := []string{}
	if _, err = svc.Update(ctx, editor(author), uuid.MustParse(doc.ID), &dto.UpdateDocRequest{AllowedRoles: &none}); err == nil {
		t.Fatal("empty role audience accepted")
	}
}
