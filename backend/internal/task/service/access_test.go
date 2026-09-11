package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	rbac "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
)

func TestTaskMutationObjectScopeAndMentions(t *testing.T) {
	svc, tasks, _, notify := newTestSvc()
	ctx := context.Background()
	owner, outsider, target, dept := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	tasks.users[target] = &model.NamedUser{ID: target, Username: "target"}
	tasks.users[outsider] = &model.NamedUser{ID: outsider, Username: "outsider"}
	row, err := svc.Create(ctx, owner, &dto.CreateTaskRequest{Title: "Private task", DepartmentID: dept.String()})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	denied := model.WithViewer(ctx, model.Viewer{ID: outsider, Scope: &rbac.DataScopeCondition{Query: "1 = 0", IsSelf: true}})
	if _, err := svc.Assign(denied, id, outsider, target.String()); err == nil {
		t.Fatal("out-of-scope assignment accepted")
	}
	if _, err := svc.AddComment(denied, id, outsider, &dto.CommentRequest{Content: "outsider"}); err == nil {
		t.Fatal("out-of-scope comment accepted")
	}
	if _, err := svc.Get(ctx, outsider, id, nil); err == nil {
		t.Fatal("missing scope granted unrestricted read")
	}
	allowed := model.WithViewer(ctx, model.Viewer{ID: outsider, Scope: &rbac.DataScopeCondition{Query: "department_id = ?", Args: []interface{}{dept}}})
	if _, err := svc.Assign(allowed, id, outsider, target.String()); err != nil {
		t.Fatal(err)
	}
	before := notify.calls
	capture := &mentionCapture{}
	svc.notify = capture
	c, err := svc.AddComment(ctx, id, owner, &dto.CommentRequest{Content: "@outsider @target", Mentions: []string{outsider.String()}})
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Mentions) != 2 || len(capture.messages) != 2 || notify.calls != before {
		t.Fatalf("mention notifications missing: %+v", c)
	}
	if strings.Contains(capture.messages[outsider], "Private task") || strings.Contains(capture.messages[outsider], "@outsider") {
		t.Fatal("unrelated recipient received private text")
	}
}

func TestTaskDescendantsRespectVisibility(t *testing.T) {
	svc, tasks, _, _ := newTestSvc()
	ctx := context.Background()
	owner, viewer := uuid.New(), uuid.New()
	tasks.users[viewer] = &model.NamedUser{ID: viewer}
	parent, _ := svc.Create(ctx, owner, &dto.CreateTaskRequest{Title: "Parent", AssigneeID: viewer.String()})
	_, err := svc.Create(ctx, owner, &dto.CreateTaskRequest{Title: "Private child", ParentID: parent.ID})
	if err != nil {
		t.Fatal(err)
	}
	got, err := svc.Get(ctx, viewer, uuid.MustParse(parent.ID), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Children) != 0 {
		t.Fatal("parent participation disclosed private children")
	}
}

type failingTaskLogs struct{ *memLogs }

func (f *failingTaskLogs) Create(context.Context, *model.TaskLog) error {
	return errors.New("injected task log failure")
}
func TestTaskLogFailureIsNotSuccess(t *testing.T) {
	svc, _, _, _ := newTestSvc()
	svc.logs = &failingTaskLogs{newMemLogs()}
	if _, err := svc.Create(context.Background(), uuid.New(), &dto.CreateTaskRequest{Title: "Fail"}); err == nil {
		t.Fatal("log failure reported success")
	}
}

func TestTaskInputAndIdempotentStatus(t *testing.T) {
	svc, _, logs, _ := newTestSvc()
	ctx := context.Background()
	owner := uuid.New()
	for _, req := range []*dto.CreateTaskRequest{{Title: "  "}, {Title: "X", AssigneeID: "not-a-uuid"}, {Title: "X", Priority: 9}} {
		if _, err := svc.Create(ctx, owner, req); err == nil {
			t.Fatalf("invalid input accepted: %+v", req)
		}
	}
	row, err := svc.Create(ctx, owner, &dto.CreateTaskRequest{Title: "X"})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	before, _ := logs.ListByTask(ctx, id)
	if _, err := svc.ChangeStatus(ctx, id, owner, &dto.StatusRequest{Status: 0}); err != nil {
		t.Fatal(err)
	}
	after, _ := logs.ListByTask(ctx, id)
	if len(after) != len(before) {
		t.Fatal("same status produced a duplicate transition")
	}
}

type mentionCapture struct{ messages map[uuid.UUID]string }

func (m *mentionCapture) Send(_ context.Context, ids []uuid.UUID, _ string, vars map[string]interface{}) error {
	if m.messages == nil {
		m.messages = map[uuid.UUID]string{}
	}
	for _, id := range ids {
		m.messages[id] = vars["title"].(string) + vars["message"].(string)
	}
	return nil
}
