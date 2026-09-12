package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/announcement/dto"
	"github.com/Yogdunana/StarByte/backend/internal/announcement/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func mustCreate(t *testing.T, svc Service, author uuid.UUID, title, category string) *dto.AnnouncementResponse {
	t.Helper()
	resp, err := svc.Create(context.Background(), Viewer{UserID: author}, &dto.CreateAnnouncementRequest{
		Title:    title,
		Content:  "正文 **" + title + "**",
		Category: category,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	return resp
}

func parseID(t *testing.T, id string) uuid.UUID {
	t.Helper()
	v, err := uuid.Parse(id)
	if err != nil {
		t.Fatalf("parse id: %v", err)
	}
	return v
}

func codeOf(err error) int {
	if app, ok := err.(*response.AppError); ok {
		return app.Code
	}
	return 0
}

func TestCreate_Success(t *testing.T) {
	svc, _, _ := newTestSvc()
	author := uuid.New()
	resp := mustCreate(t, svc, author, "协会周报", model.CategoryAssociation)
	if resp.Title != "协会周报" || resp.Status != model.StatusDraft {
		t.Fatalf("got title=%s status=%d", resp.Title, resp.Status)
	}
	if resp.Author.ID != author.String() {
		t.Fatalf("author mismatch")
	}
	if resp.ContentType != model.ContentMarkdown {
		t.Fatalf("content_type = %s", resp.ContentType)
	}
}

func TestCreate_InvalidCategory(t *testing.T) {
	svc, _, _ := newTestSvc()
	_, err := svc.Create(context.Background(), Viewer{UserID: uuid.New()}, &dto.CreateAnnouncementRequest{
		Title: "x", Category: "unknown",
	})
	if codeOf(err) != response.CodeAnnouncementInvalidCat {
		t.Fatalf("code = %d, want %d, err=%v", codeOf(err), response.CodeAnnouncementInvalidCat, err)
	}
}

func TestCreate_EmptyTitle(t *testing.T) {
	svc, _, _ := newTestSvc()
	_, err := svc.Create(context.Background(), Viewer{UserID: uuid.New()}, &dto.CreateAnnouncementRequest{
		Title: "  ", Category: model.CategorySystem,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGet_NotFound(t *testing.T) {
	svc, _, _ := newTestSvc()
	_, err := svc.Get(context.Background(), Viewer{UserID: uuid.New()}, uuid.New())
	if codeOf(err) != response.CodeAnnouncementNotFound {
		t.Fatalf("code = %d, err=%v", codeOf(err), err)
	}
}

func TestGet_DraftHiddenFromOthers(t *testing.T) {
	svc, _, _ := newTestSvc()
	author := uuid.New()
	resp := mustCreate(t, svc, author, "草稿", model.CategorySystem)
	id := parseID(t, resp.ID)
	_, err := svc.Get(context.Background(), Viewer{UserID: uuid.New(), Staff: false}, id)
	if codeOf(err) != response.CodeAnnouncementNoAccess {
		t.Fatalf("code = %d, err=%v", codeOf(err), err)
	}
	got, err := svc.Get(context.Background(), Viewer{UserID: author}, id)
	if err != nil || got.Title != "草稿" {
		t.Fatalf("author should see draft: %v", err)
	}
}

func TestUpdate_Draft(t *testing.T) {
	svc, _, _ := newTestSvc()
	author := uuid.New()
	resp := mustCreate(t, svc, author, "旧标题", model.CategoryActivity)
	id := parseID(t, resp.ID)
	title := "新标题"
	cat := model.CategoryPersonnel
	updated, err := svc.Update(context.Background(), Viewer{UserID: author}, id, &dto.UpdateAnnouncementRequest{
		Title:    &title,
		Category: &cat,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Title != title || updated.Category != cat {
		t.Fatalf("got %+v", updated)
	}
}

func TestUpdate_ArchivedForbidden(t *testing.T) {
	svc, _, _ := newTestSvc()
	author := uuid.New()
	staff := Viewer{UserID: author, Staff: true}
	resp := mustCreate(t, svc, author, "归档前", model.CategorySystem)
	id := parseID(t, resp.ID)
	if _, err := svc.Publish(context.Background(), staff, id); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Archive(context.Background(), staff, id); err != nil {
		t.Fatal(err)
	}
	title := "不可改"
	err := error(nil)
	_, err = svc.Update(context.Background(), staff, id, &dto.UpdateAnnouncementRequest{Title: &title})
	if codeOf(err) != response.CodeAnnouncementInvalidState {
		t.Fatalf("code = %d, err=%v", codeOf(err), err)
	}
}

func TestDelete_DraftAndHidden(t *testing.T) {
	svc, _, _ := newTestSvc()
	author := uuid.New()
	resp := mustCreate(t, svc, author, "将删", model.CategorySystem)
	id := parseID(t, resp.ID)
	if err := svc.Delete(context.Background(), Viewer{UserID: author}, id); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Get(context.Background(), Viewer{UserID: author, Staff: true}, id); err == nil {
		t.Fatal("deleted should be hidden")
	}
}

func TestPublish_NotifyAndReadFlow(t *testing.T) {
	svc, repo, n := newTestSvc()
	author := uuid.New()
	reader := uuid.New()
	repo.addUser(author, "作者")
	repo.addUser(reader, "读者")

	resp := mustCreate(t, svc, author, "开学通知", model.CategoryAssociation)
	id := parseID(t, resp.ID)
	staff := Viewer{UserID: author, Staff: true}
	pub, err := svc.Publish(context.Background(), staff, id)
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if pub.Status != model.StatusPublished || pub.PublishedAt == "" {
		t.Fatalf("published fields: %+v", pub)
	}
	if len(n.calls) != 1 || n.calls[0].template != tplAnnouncementPublished {
		t.Fatalf("notify calls = %+v", n.calls)
	}
	if len(n.calls[0].users) != 2 {
		t.Fatalf("recipients = %d", len(n.calls[0].users))
	}

	unread, err := svc.UnreadCount(context.Background(), reader)
	if err != nil || unread.Count != 1 {
		t.Fatalf("unread = %+v err=%v", unread, err)
	}
	if err := svc.MarkRead(context.Background(), reader, id); err != nil {
		t.Fatal(err)
	}
	if err := svc.MarkRead(context.Background(), reader, id); err != nil {
		t.Fatal(err)
	}
	unread, err = svc.UnreadCount(context.Background(), reader)
	if err != nil || unread.Count != 0 {
		t.Fatalf("unread after read = %+v err=%v", unread, err)
	}
	st, err := svc.ReadStatus(context.Background(), Viewer{UserID: author, CanManage: true}, id)
	if err != nil {
		t.Fatal(err)
	}
	if st.ReadCount != 1 || len(st.Readers) != 1 || st.Readers[0].User.Name != "读者" {
		t.Fatalf("read status = %+v", st)
	}
}

func TestPublish_NotDraft(t *testing.T) {
	svc, _, _ := newTestSvc()
	author := uuid.New()
	staff := Viewer{UserID: author, Staff: true}
	resp := mustCreate(t, svc, author, "已发", model.CategorySystem)
	id := parseID(t, resp.ID)
	if _, err := svc.Publish(context.Background(), staff, id); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Publish(context.Background(), staff, id)
	if codeOf(err) != response.CodeAnnouncementInvalidState {
		t.Fatalf("code = %d, err=%v", codeOf(err), err)
	}
}

func TestMarkRead_DraftForbidden(t *testing.T) {
	svc, _, _ := newTestSvc()
	author := uuid.New()
	resp := mustCreate(t, svc, author, "草稿", model.CategorySystem)
	err := svc.MarkRead(context.Background(), author, parseID(t, resp.ID))
	if codeOf(err) != response.CodeAnnouncementInvalidState {
		t.Fatalf("code = %d, err=%v", codeOf(err), err)
	}
}

func TestPinToggleAndArchive(t *testing.T) {
	svc, _, _ := newTestSvc()
	author := uuid.New()
	staff := Viewer{UserID: author, Staff: true}
	resp := mustCreate(t, svc, author, "置顶", model.CategorySystem)
	id := parseID(t, resp.ID)
	if _, err := svc.Publish(context.Background(), staff, id); err != nil {
		t.Fatal(err)
	}
	pinned, err := svc.Pin(context.Background(), staff, id, nil)
	if err != nil || !pinned.Pinned {
		t.Fatalf("pin toggle: %+v err=%v", pinned, err)
	}
	want := false
	unpinned, err := svc.Pin(context.Background(), staff, id, &want)
	if err != nil || unpinned.Pinned {
		t.Fatalf("unpin: %+v err=%v", unpinned, err)
	}
	archived, err := svc.Archive(context.Background(), staff, id)
	if err != nil || archived.Status != model.StatusArchived {
		t.Fatalf("archive: %+v err=%v", archived, err)
	}
}

func TestPin_RequiresStaff(t *testing.T) {
	svc, _, _ := newTestSvc()
	author := uuid.New()
	resp := mustCreate(t, svc, author, "x", model.CategorySystem)
	_, err := svc.Pin(context.Background(), Viewer{UserID: author}, parseID(t, resp.ID), nil)
	if codeOf(err) != response.CodeAnnouncementNoAccess {
		t.Fatalf("code = %d, err=%v", codeOf(err), err)
	}
}

func TestList_FiltersAndVisibility(t *testing.T) {
	svc, _, _ := newTestSvc()
	author := uuid.New()
	other := uuid.New()
	staff := Viewer{UserID: author, Staff: true}
	d1 := mustCreate(t, svc, author, "协会草稿", model.CategoryAssociation)
	p1 := mustCreate(t, svc, author, "活动已发", model.CategoryActivity)
	if _, err := svc.Publish(context.Background(), staff, parseID(t, p1.ID)); err != nil {
		t.Fatal(err)
	}

	member := Viewer{UserID: other}
	list, total, err := svc.List(context.Background(), member, &dto.ListAnnouncementRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(list) != 1 || list[0].ID != p1.ID {
		t.Fatalf("member list = %+v total=%d", list, total)
	}

	staffList, staffTotal, err := svc.List(context.Background(), staff, &dto.ListAnnouncementRequest{})
	if err != nil || staffTotal != 2 {
		t.Fatalf("staff list total=%d err=%v items=%d", staffTotal, err, len(staffList))
	}

	catList, catTotal, err := svc.List(context.Background(), member, &dto.ListAnnouncementRequest{Category: model.CategoryActivity})
	if err != nil || catTotal != 1 || catList[0].Category != model.CategoryActivity {
		t.Fatalf("category filter: %+v total=%d err=%v", catList, catTotal, err)
	}

	_, kwTotal, err := svc.List(context.Background(), member, &dto.ListAnnouncementRequest{Keyword: "活动"})
	if err != nil || kwTotal != 1 {
		t.Fatalf("keyword filter total=%d err=%v", kwTotal, err)
	}

	draft := int16(model.StatusDraft)
	ownDrafts, n, err := svc.List(context.Background(), Viewer{UserID: author}, &dto.ListAnnouncementRequest{Status: &draft})
	if err != nil || n != 1 || ownDrafts[0].ID != d1.ID {
		t.Fatalf("own drafts = %+v n=%d err=%v", ownDrafts, n, err)
	}
}

func TestDispatchDuePublishes(t *testing.T) {
	svc, repo, n := newTestSvc()
	author := uuid.New()
	repo.addUser(author, "作者")
	soon := time.Date(2026, 9, 12, 11, 0, 0, 0, time.UTC)
	future := time.Date(2026, 9, 13, 10, 0, 0, 0, time.UTC)
	publisher := Viewer{UserID: author, CanPublish: true}
	due, err := svc.Create(context.Background(), publisher, &dto.CreateAnnouncementRequest{
		Title: "到点发布", Category: model.CategorySystem, ScheduledAt: &soon,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Create(context.Background(), publisher, &dto.CreateAnnouncementRequest{
		Title: "未到点", Category: model.CategorySystem, ScheduledAt: &future,
	}); err != nil {
		t.Fatal(err)
	}
	svc.now = func() time.Time { return time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC) }
	if err := svc.DispatchDuePublishes(context.Background(), "", nil); err != nil {
		t.Fatal(err)
	}
	got, err := svc.Get(context.Background(), Viewer{UserID: author, Staff: true}, parseID(t, due.ID))
	if err != nil || got.Status != model.StatusPublished {
		t.Fatalf("due should publish: %+v err=%v", got, err)
	}
	if len(n.calls) != 1 {
		t.Fatalf("notify calls = %d", len(n.calls))
	}
	if got.ScheduledAt != "" {
		t.Fatalf("scheduled_at should clear on publish, got %q", got.ScheduledAt)
	}
}

func TestPublish_ClearsScheduledAtAndAllowsEdit(t *testing.T) {
	svc, _, _ := newTestSvc()
	author := uuid.New()
	staff := Viewer{UserID: author, Staff: true}
	when := time.Date(2026, 9, 12, 11, 0, 0, 0, time.UTC)
	resp, err := svc.Create(context.Background(), Viewer{UserID: author, CanPublish: true}, &dto.CreateAnnouncementRequest{
		Title: "定时稿", Category: model.CategorySystem, ScheduledAt: &when,
	})
	if err != nil {
		t.Fatal(err)
	}
	id := parseID(t, resp.ID)
	pub, err := svc.Publish(context.Background(), staff, id)
	if err != nil {
		t.Fatal(err)
	}
	if pub.ScheduledAt != "" {
		t.Fatalf("publish should clear scheduled_at, got %q", pub.ScheduledAt)
	}
	title := "发布后可改"
	updated, err := svc.Update(context.Background(), staff, id, &dto.UpdateAnnouncementRequest{
		Title:       &title,
		ScheduledAt: &when,
	})
	if err != nil {
		t.Fatalf("published update should succeed: %v", err)
	}
	if updated.Title != title {
		t.Fatalf("title = %q", updated.Title)
	}
	if updated.ScheduledAt != "" {
		t.Fatalf("update must not restore scheduled_at on published, got %q", updated.ScheduledAt)
	}
}

func TestCreate_ScheduleRequiresPublish(t *testing.T) {
	svc, _, _ := newTestSvc()
	when := time.Date(2026, 9, 12, 11, 0, 0, 0, time.UTC)
	_, err := svc.Create(context.Background(), Viewer{UserID: uuid.New()}, &dto.CreateAnnouncementRequest{
		Title: "越权定时", Category: model.CategorySystem, ScheduledAt: &when,
	})
	if codeOf(err) != response.CodeAnnouncementNoAccess {
		t.Fatalf("code = %d, err=%v", codeOf(err), err)
	}
}

func TestCreate_ScheduleRejectsPast(t *testing.T) {
	svc, _, _ := newTestSvc()
	past := time.Date(2026, 9, 12, 9, 0, 0, 0, time.UTC)
	_, err := svc.Create(context.Background(), Viewer{UserID: uuid.New(), CanPublish: true}, &dto.CreateAnnouncementRequest{
		Title: "过去时间", Category: model.CategorySystem, ScheduledAt: &past,
	})
	if err == nil {
		t.Fatal("expected past scheduled_at to fail")
	}
	if codeOf(err) != response.CodeBadRequest {
		t.Fatalf("code = %d, err=%v", codeOf(err), err)
	}
}

func TestUpdate_UpdateOnlyPreservesScheduledAt(t *testing.T) {
	svc, _, _ := newTestSvc()
	author := uuid.New()
	when := time.Date(2026, 9, 12, 11, 0, 0, 0, time.UTC)
	resp, err := svc.Create(context.Background(), Viewer{UserID: author, CanPublish: true}, &dto.CreateAnnouncementRequest{
		Title: "定时草稿", Category: model.CategorySystem, ScheduledAt: &when,
	})
	if err != nil {
		t.Fatal(err)
	}
	title := "副部长只改标题"
	updated, err := svc.Update(context.Background(), Viewer{UserID: author}, parseID(t, resp.ID), &dto.UpdateAnnouncementRequest{
		Title:      &title,
		ClearSched: true,
	})
	if err != nil {
		t.Fatalf("update-only edit should succeed: %v", err)
	}
	if updated.Title != title {
		t.Fatalf("title = %q", updated.Title)
	}
	if updated.ScheduledAt == "" {
		t.Fatal("update-only role must not clear scheduled_at")
	}
}

func TestUpdate_ScheduleRequiresPublish(t *testing.T) {
	svc, _, _ := newTestSvc()
	author := uuid.New()
	resp := mustCreate(t, svc, author, "草稿", model.CategorySystem)
	when := time.Date(2026, 9, 12, 11, 0, 0, 0, time.UTC)
	_, err := svc.Update(context.Background(), Viewer{UserID: author}, parseID(t, resp.ID), &dto.UpdateAnnouncementRequest{
		ScheduledAt: &when,
	})
	if codeOf(err) != response.CodeAnnouncementNoAccess {
		t.Fatalf("code = %d, err=%v", codeOf(err), err)
	}
}

func TestReadStatus_RequiresManage(t *testing.T) {
	svc, _, _ := newTestSvc()
	author := uuid.New()
	resp := mustCreate(t, svc, author, "已发", model.CategorySystem)
	id := parseID(t, resp.ID)
	if _, err := svc.Publish(context.Background(), Viewer{UserID: author, Staff: true}, id); err != nil {
		t.Fatal(err)
	}
	_, err := svc.ReadStatus(context.Background(), Viewer{UserID: author}, id)
	if codeOf(err) != response.CodeAnnouncementNoAccess {
		t.Fatalf("code = %d, err=%v", codeOf(err), err)
	}
}

func TestErrorCodesInAnnouncementRange(t *testing.T) {
	codes := []int{
		response.CodeAnnouncementNotFound,
		response.CodeAnnouncementInvalidState,
		response.CodeAnnouncementNoAccess,
		response.CodeAnnouncementInvalidCat,
	}
	for _, c := range codes {
		if c < 28000 || c > 28999 {
			t.Fatalf("code %d outside 28000-28999", c)
		}
		if c >= 10000 && c <= 10499 {
			t.Fatalf("code %d collides with issue-listed internship range", c)
		}
	}
}

func TestList_InvalidCategory(t *testing.T) {
	svc, _, _ := newTestSvc()
	_, _, err := svc.List(context.Background(), Viewer{UserID: uuid.New()}, &dto.ListAnnouncementRequest{Category: "nope"})
	if codeOf(err) != response.CodeAnnouncementInvalidCat {
		t.Fatalf("code = %d, err=%v", codeOf(err), err)
	}
}
