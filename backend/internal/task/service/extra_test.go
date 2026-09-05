package service

import (
	"context"
	"io"
	"mime/multipart"
	"strings"
	"testing"
	"time"

	filedto "github.com/Yogdunana/StarByte/backend/internal/file/dto"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/google/uuid"
)

type mockBridge struct{}

func (m *mockBridge) Upload(_ context.Context, _ uuid.UUID, header *multipart.FileHeader, _ string, _ bool) (*filedto.FileUploadResponse, error) {
	return &filedto.FileUploadResponse{
		ID: uuid.New().String(), OriginalName: header.Filename, Filename: header.Filename,
		FileSize: header.Size, MimeType: "text/plain",
	}, nil
}
func (m *mockBridge) GetByID(_ context.Context, id uuid.UUID) (*filedto.FileDetailResponse, error) {
	return &filedto.FileDetailResponse{ID: id.String(), Path: "obj/f.txt", OriginalName: "f.txt"}, nil
}
func (m *mockBridge) Delete(_ context.Context, _, _ uuid.UUID) error { return nil }

type mockStore struct{}

func (m *mockStore) Download(_ context.Context, _ string) (io.ReadCloser, string, error) {
	return io.NopCloser(strings.NewReader("hi")), "text/plain", nil
}

func TestUpdateDeleteListParent(t *testing.T) {
	svc, tasks, _, _ := newTestSvc()
	ctx := context.Background()
	creator := uuid.New()
	parent, err := svc.Create(ctx, creator, &dto.CreateTaskRequest{Title: "P", Tags: []string{"core"}})
	if err != nil {
		t.Fatal(err)
	}
	pid := uuid.MustParse(parent.ID)
	title := "P2"
	desc := "d"
	prio := int16(3)
	due := time.Now().Add(48 * time.Hour)
	upd, err := svc.Update(ctx, pid, creator, &dto.UpdateTaskRequest{
		Title: &title, Description: &desc, Priority: &prio, DueDate: &due, Tags: []string{"a"},
	})
	if err != nil || upd.Title != "P2" || upd.Priority != 3 || len(upd.Tags) != 1 {
		t.Fatalf("update %+v %v", upd, err)
	}
	child, err := svc.Create(ctx, creator, &dto.CreateTaskRequest{Title: "C", ParentID: parent.ID})
	if err != nil {
		t.Fatal(err)
	}
	got, err := svc.Get(ctx, pid)
	if err != nil || len(got.Children) != 1 || got.Children[0].ID != child.ID {
		t.Fatalf("children %+v %v", got, err)
	}
	if _, err := svc.Create(ctx, creator, &dto.CreateTaskRequest{Title: "bad", ParentID: child.ID}); err == nil {
		t.Fatal("nested parent")
	}
	list, n, err := svc.List(ctx, creator, &dto.ListTaskRequest{Page: 1, PageSize: 10}, nil)
	if err != nil || n < 2 || len(list) < 2 {
		t.Fatalf("list %v %d", err, n)
	}
	scope := &rbacModel.DataScopeCondition{Query: "1 = 0"}
	_, _, err = svc.List(ctx, creator, &dto.ListTaskRequest{}, scope)
	if err != nil {
		t.Fatal(err)
	}
	scope2 := &rbacModel.DataScopeCondition{Query: "department_id = ?", Args: []interface{}{uuid.New()}}
	rewritten := rewriteTaskScope(scope2, creator)
	if !strings.Contains(rewritten.Query, "t.department_id") {
		t.Fatalf("scope %s", rewritten.Query)
	}
	if err := svc.Delete(ctx, uuid.MustParse(child.ID), creator); err != nil {
		t.Fatal(err)
	}
	doing, _ := svc.Create(ctx, creator, &dto.CreateTaskRequest{Title: "doing"})
	did := uuid.MustParse(doing.ID)
	row, _ := tasks.GetByID(ctx, did)
	row.Status = model.StatusDoing
	_ = tasks.Update(ctx, row)
	if err := svc.Delete(ctx, did, creator); err == nil {
		t.Fatal("cannot delete doing")
	}
}

func TestCommentUpdateDeleteAndLogs(t *testing.T) {
	svc, _, _, _ := newTestSvc()
	ctx := context.Background()
	creator := uuid.New()
	row, _ := svc.Create(ctx, creator, &dto.CreateTaskRequest{Title: "X"})
	id := uuid.MustParse(row.ID)
	cmt, err := svc.AddComment(ctx, id, creator, &dto.CommentRequest{Content: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	upd, err := svc.UpdateComment(ctx, id, uuid.MustParse(cmt.ID), creator, "hello @nobody")
	if err != nil || upd.Content != "hello @nobody" {
		t.Fatalf("upd %+v %v", upd, err)
	}
	other := uuid.New()
	if _, err := svc.UpdateComment(ctx, id, uuid.MustParse(cmt.ID), other, "x"); err == nil {
		t.Fatal("other cannot edit")
	}
	if err := svc.DeleteComment(ctx, id, uuid.MustParse(cmt.ID), other); err == nil {
		t.Fatal("other cannot delete")
	}
	if err := svc.DeleteComment(ctx, id, uuid.MustParse(cmt.ID), creator); err != nil {
		t.Fatal(err)
	}
	logs, err := svc.ListLogs(ctx, id)
	if err != nil || len(logs) == 0 {
		t.Fatalf("logs %v %d", err, len(logs))
	}
}

func TestAttachments(t *testing.T) {
	tasks := newMemTasks()
	files := newMemFiles()
	svc := NewTaskService(tasks, newMemLogs(), newMemComments(), files, &memNotify{}, &mockBridge{}, &mockStore{}).(*taskService)
	ctx := context.Background()
	creator := uuid.New()
	row, _ := svc.Create(ctx, creator, &dto.CreateTaskRequest{Title: "F"})
	id := uuid.MustParse(row.ID)
	header := &multipart.FileHeader{Filename: "a.txt", Size: 2}
	att, err := svc.UploadAttachment(ctx, id, creator, header)
	if err != nil {
		t.Fatal(err)
	}
	// mem store needs a path for download
	named, _ := files.GetNamed(ctx, uuid.MustParse(att.ID))
	named.FilePath = "obj/f.txt"
	named.FileName = "a.txt"
	list, err := svc.ListAttachments(ctx, id)
	if err != nil || len(list) != 1 {
		t.Fatalf("list att %v %d", err, len(list))
	}
	rc, name, _, err := svc.DownloadAttachment(ctx, id, uuid.MustParse(att.ID))
	if err != nil {
		t.Fatal(err)
	}
	_ = rc.Close()
	if name == "" {
		t.Fatal("empty name")
	}
	if err := svc.DeleteAttachment(ctx, id, uuid.MustParse(att.ID), creator); err != nil {
		t.Fatal(err)
	}
	bare := NewTaskService(tasks, newMemLogs(), newMemComments(), files, nil, nil, nil).(*taskService)
	if _, err := bare.UploadAttachment(ctx, id, creator, header); err == nil {
		t.Fatal("nil bridge")
	}
}

func TestAssignClosedAndHelpers(t *testing.T) {
	svc, tasks, _, _ := newTestSvc()
	ctx := context.Background()
	creator := uuid.New()
	u := uuid.New()
	tasks.users[u] = &model.NamedUser{ID: u, Username: "z", RealName: "张"}
	row, _ := svc.Create(ctx, creator, &dto.CreateTaskRequest{Title: "A"})
	id := uuid.MustParse(row.ID)
	if _, err := svc.Assign(ctx, id, creator, u.String()); err != nil {
		t.Fatal(err)
	}
	if err := svc.Urge(ctx, id, creator, "go"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ChangeStatus(ctx, id, creator, &dto.StatusRequest{Status: 3}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Assign(ctx, id, creator, u.String()); err == nil {
		t.Fatal("closed assign")
	}
	if displayName(&model.NamedUser{Username: "u"}) != "u" {
		t.Fatal("display")
	}
	if displayName(&model.NamedUser{RealName: "R", Username: "u"}) != "R" {
		t.Fatal("real")
	}
	if parseUUIDPtr("bad") != nil || parseUUIDPtr(u.String()) == nil {
		t.Fatal("parse")
	}
	if uuidPtrString(nil) != "" {
		t.Fatal("ptr")
	}
	if encodeTags(nil) != "[]" {
		t.Fatal("tags")
	}
}

func TestNotifyAdapterNil(t *testing.T) {
	if NewNotifier(nil) != nil {
		t.Fatal("nil notify")
	}
	n := reminderTargets(&model.Task{CreatorID: uuid.New()})
	if len(n) != 1 {
		t.Fatalf("targets %d", n)
	}
}
