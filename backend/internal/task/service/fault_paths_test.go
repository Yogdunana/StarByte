package service

import (
	"context"
	"errors"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/google/uuid"

	filedto "github.com/Yogdunana/StarByte/backend/internal/file/dto"
	rbac "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
)

func TestCapabilitiesDoNotBorrowAnotherActionsScope(t *testing.T) {
	owner, assignee, outsider, dept := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	row := &model.Task{CreatorID: owner, AssigneeID: &assignee, DepartmentID: &dept}
	all := &rbac.DataScopeCondition{}
	self := &rbac.DataScopeCondition{IsSelf: true}
	v := model.Viewer{ID: outsider, Scope: all, Allowed: map[string]bool{"task:assign": true, "task:update": true, "task:transfer": true, "task:comment": true, "task:delete": true, "task:create": true}, Scopes: map[string]*rbac.DataScopeCondition{"task:assign": all, "task:comment": self}}
	out := &dto.TaskResponse{}
	taskCapabilities(model.WithViewer(context.Background(), v), row, out)
	if !out.CanAssign || out.CanUpdate || out.CanTransfer || out.CanComment || out.CanDelete || out.CanUrge {
		t.Fatalf("broadened capabilities: %+v", out)
	}
	v.ID = owner
	taskCapabilities(model.WithViewer(context.Background(), v), row, out)
	if !out.CanUpdate || !out.CanComment || !out.CanUrge {
		t.Fatalf("owner missing permitted actions: %+v", out)
	}
	row.Status = model.StatusDoing
	taskCapabilities(model.WithViewer(context.Background(), v), row, out)
	if out.CanDelete {
		t.Fatal("doing task advertised deletion")
	}
	row.Status = model.StatusDone
	taskCapabilities(model.WithViewer(context.Background(), v), row, out)
	if out.CanAssign || out.CanUpdate || out.CanComment || out.CanTransfer || out.CanUrge {
		t.Fatal("closed task advertised mutation")
	}
}

func TestTaskDepartmentScopeFailsClosed(t *testing.T) {
	owner, viewer, dept := uuid.New(), uuid.New(), uuid.New()
	row := &model.Task{CreatorID: owner, DepartmentID: &dept}
	for _, tc := range []struct {
		scope *rbac.DataScopeCondition
		want  bool
	}{
		{nil, false}, {&rbac.DataScopeCondition{IsSelf: true}, false},
		{&rbac.DataScopeCondition{}, true},
		{&rbac.DataScopeCondition{Query: "department_id IN ?", Args: []interface{}{[]uuid.UUID{uuid.New(), dept}}}, true},
		{&rbac.DataScopeCondition{Query: "department_id IN ?", Args: []interface{}{[]uuid.UUID{uuid.New()}}}, false},
		{&rbac.DataScopeCondition{Query: "unexpected = ?", Args: []interface{}{dept}}, false},
	} {
		if got := canViewTask(row, viewer, tc.scope); got != tc.want {
			t.Fatalf("scope %+v: %v", tc.scope, got)
		}
	}
	if canViewTask(row, uuid.Nil, &rbac.DataScopeCondition{}) {
		t.Fatal("nil identity admitted")
	}
	if rewriteTaskScope(nil, uuid.Nil).Query != "1 = 0" {
		t.Fatal("nil identity list admitted")
	}
}

type brokenMentions struct {
	*memTasks
	namesFail, idFail bool
}

func (r *brokenMentions) FindUsersByUsername(ctx context.Context, names []string) ([]model.NamedUser, error) {
	if r.namesFail {
		return nil, errors.New("lookup unavailable")
	}
	return r.memTasks.FindUsersByUsername(ctx, names)
}
func (r *brokenMentions) GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error) {
	if r.idFail {
		return nil, errors.New("lookup unavailable")
	}
	return r.memTasks.GetUser(ctx, id)
}
func TestMentionLookupFailureDoesNotSilentlyDropRecipients(t *testing.T) {
	for _, namesFail := range []bool{true, false} {
		svc, tasks, logs, notify := newTestSvc()
		owner := uuid.New()
		ctx := context.Background()
		row, err := svc.Create(ctx, owner, &dto.CreateTaskRequest{Title: "X"})
		if err != nil {
			t.Fatal(err)
		}
		svc.tasks = &brokenMentions{tasks, namesFail, !namesFail}
		before := len(logs.items)
		_, err = svc.AddComment(ctx, uuid.MustParse(row.ID), owner, &dto.CommentRequest{Content: "@someone", Mentions: []string{uuid.New().String()}})
		if err == nil || !strings.Contains(err.Error(), "lookup unavailable") {
			t.Fatalf("missing lookup error: %v", err)
		}
		if len(svc.comments.(*memComments).items) != 0 || len(logs.items) != before || notify.calls != 0 {
			t.Fatal("failed resolution left partial comment or notification")
		}
	}
}

type raceUploadBridge struct {
	mockBridge
	afterUpload func()
	deleted     bool
	cleanupErr  error
	fileID      uuid.UUID
}

func (b *raceUploadBridge) Upload(ctx context.Context, id uuid.UUID, h *multipart.FileHeader, cat string, pub bool) (*filedto.FileUploadResponse, error) {
	b.afterUpload()
	return &filedto.FileUploadResponse{ID: b.fileID.String(), OriginalName: h.Filename}, nil
}
func (b *raceUploadBridge) Delete(ctx context.Context, id, user uuid.UUID) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	b.deleted = id == b.fileID
	return b.cleanupErr
}
func TestUploadRechecksOwnershipAndCleansNewObject(t *testing.T) {
	for _, cleanupFails := range []bool{false, true} {
		svc, tasks, _, _ := newTestSvc()
		owner := uuid.New()
		ctx, cancel := context.WithCancel(context.Background())
		row, err := svc.Create(ctx, owner, &dto.CreateTaskRequest{Title: "F"})
		if err != nil {
			t.Fatal(err)
		}
		id := uuid.MustParse(row.ID)
		b := &raceUploadBridge{fileID: uuid.New()}
		b.afterUpload = func() { tasks.items[id].Status = model.StatusDone; cancel() }
		if cleanupFails {
			b.cleanupErr = errors.New("storage cleanup unavailable")
		}
		svc.bridge = b
		_, err = svc.UploadAttachment(ctx, id, owner, &multipart.FileHeader{Filename: "f.pdf"})
		if err == nil || !b.deleted {
			t.Fatalf("upload race not compensated: %v", err)
		}
		if cleanupFails && !strings.Contains(err.Error(), "storage cleanup unavailable") {
			t.Fatal("cleanup failure hidden")
		}
		if len(svc.files.(*memFiles).items) != 0 {
			t.Fatal("closed task gained attachment")
		}
	}
}
