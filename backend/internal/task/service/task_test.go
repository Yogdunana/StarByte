package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func newTestSvc() (*taskService, *memTasks, *memLogs, *memNotify) {
	tasks := newMemTasks()
	logs := newMemLogs()
	notify := &memNotify{}
	svc := NewTaskService(tasks, logs, newMemComments(), newMemFiles(), notify, nil, nil).(*taskService)
	return svc, tasks, logs, notify
}

func TestCreateAssignStatusUrgeComment(t *testing.T) {
	svc, tasks, logs, notify := newTestSvc()
	ctx := context.Background()
	creator := uuid.New()
	assignee := uuid.New()
	tasks.users[assignee] = &model.NamedUser{ID: assignee, Username: "alice", RealName: "爱丽丝"}

	created, err := svc.Create(ctx, creator, &dto.CreateTaskRequest{
		Title: "实现登录", Priority: 2, AssigneeID: assignee.String(), Tags: []string{"frontend", "auth"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Status != 0 || created.Assignee == nil || created.Assignee.ID != assignee.String() {
		t.Fatalf("create %+v", created)
	}
	id := uuid.MustParse(created.ID)

	if _, err := svc.ChangeStatus(ctx, id, creator, &dto.StatusRequest{Status: 1, Comment: "开工"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ChangeStatus(ctx, id, creator, &dto.StatusRequest{Status: 2}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ChangeStatus(ctx, id, creator, &dto.StatusRequest{Status: 1}); err == nil {
		t.Fatal("expected closed")
	} else if ae, ok := err.(*response.AppError); !ok || ae.Code != response.CodeTaskClosed {
		t.Fatalf("want 9007 got %v", err)
	}

	other := uuid.New()
	tasks.users[other] = &model.NamedUser{ID: other, Username: "bob"}
	if _, err := svc.Create(ctx, creator, &dto.CreateTaskRequest{Title: "B", AssigneeID: assignee.String()}); err != nil {
		t.Fatal(err)
	}
	_ = logs
	_ = notify
}

func TestStateMachineAndTransfer(t *testing.T) {
	svc, tasks, _, notify := newTestSvc()
	ctx := context.Background()
	creator := uuid.New()
	a1 := uuid.New()
	a2 := uuid.New()
	tasks.users[a1] = &model.NamedUser{ID: a1, Username: "u1"}
	tasks.users[a2] = &model.NamedUser{ID: a2, Username: "u2"}

	row, err := svc.Create(ctx, creator, &dto.CreateTaskRequest{Title: "T", AssigneeID: a1.String()})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	if _, err := svc.ChangeStatus(ctx, id, creator, &dto.StatusRequest{Status: 4}); err == nil {
		t.Fatal("pending cannot hold")
	}
	if _, err := svc.ChangeStatus(ctx, id, creator, &dto.StatusRequest{Status: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ChangeStatus(ctx, id, creator, &dto.StatusRequest{Status: 4}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ChangeStatus(ctx, id, creator, &dto.StatusRequest{Status: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Transfer(ctx, id, creator, &dto.TransferRequest{NewAssigneeID: a2.String(), Reason: "忙不过来"}); err != nil {
		t.Fatal(err)
	}
	got, _ := svc.Get(ctx, id)
	if got.Assignee == nil || got.Assignee.ID != a2.String() {
		t.Fatalf("transfer %+v", got.Assignee)
	}
	if err := svc.Urge(ctx, id, creator, "尽快"); err != nil {
		t.Fatal(err)
	}
	if notify.calls < 2 {
		t.Fatalf("notify calls=%d", notify.calls)
	}
	if _, err := svc.Assign(ctx, id, creator, uuid.New().String()); err == nil {
		t.Fatal("missing user")
	}
}

func TestCommentMentionAndMyLists(t *testing.T) {
	svc, tasks, _, _ := newTestSvc()
	ctx := context.Background()
	creator := uuid.New()
	assignee := uuid.New()
	tasks.users[assignee] = &model.NamedUser{ID: assignee, Username: "alice"}
	row, err := svc.Create(ctx, creator, &dto.CreateTaskRequest{Title: "C", AssigneeID: assignee.String()})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	cmt, err := svc.AddComment(ctx, id, creator, &dto.CommentRequest{Content: "hello @alice"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cmt.Mentions) != 1 || cmt.Mentions[0] != assignee.String() {
		t.Fatalf("mentions %+v", cmt.Mentions)
	}
	list, err := svc.ListComments(ctx, id)
	if err != nil || len(list) != 1 {
		t.Fatalf("list comments %v %d", err, len(list))
	}
	todo, n, err := svc.ListMy(ctx, assignee, "todo", &dto.MyTaskRequest{})
	if err != nil || n != 1 || len(todo) != 1 {
		t.Fatalf("todo %v %d", err, n)
	}
	created, n, err := svc.ListMy(ctx, creator, "created", &dto.MyTaskRequest{})
	if err != nil || n != 1 || created[0].ID != row.ID {
		t.Fatalf("created %v %d", err, n)
	}
	past := time.Now().Add(-2 * time.Hour)
	_, err = svc.Create(ctx, creator, &dto.CreateTaskRequest{Title: "late", AssigneeID: assignee.String(), DueDate: &past})
	if err != nil {
		t.Fatal(err)
	}
	over, n, err := svc.ListMy(ctx, assignee, "overdue", &dto.MyTaskRequest{})
	if err != nil || n != 1 || over[0].Title != "late" {
		t.Fatalf("overdue %v %d %+v", err, n, over)
	}
	stats, err := svc.Stats(ctx, &dto.StatsRequest{})
	if err != nil || stats.Total < 2 || stats.Overdue < 1 {
		t.Fatalf("stats %+v %v", stats, err)
	}
}

func TestRemindDueAndOverdue(t *testing.T) {
	svc, tasks, _, notify := newTestSvc()
	ctx := context.Background()
	creator := uuid.New()
	assignee := uuid.New()
	soon := time.Now().Add(2 * time.Hour)
	past := time.Now().Add(-time.Hour)
	t1 := &model.Task{ID: uuid.New(), Title: "soon", CreatorID: creator, AssigneeID: &assignee, DueDate: &soon, Status: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	t2 := &model.Task{ID: uuid.New(), Title: "late", CreatorID: creator, AssigneeID: &assignee, DueDate: &past, Status: 0, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	_ = tasks.Create(ctx, t1)
	_ = tasks.Create(ctx, t2)
	n, err := svc.RemindDueAndOverdue(ctx)
	if err != nil || n != 2 {
		t.Fatalf("remind n=%d err=%v", n, err)
	}
	if notify.calls != 2 {
		t.Fatalf("notify=%d", notify.calls)
	}
	n, err = svc.RemindDueAndOverdue(ctx)
	if err != nil || n != 0 {
		t.Fatalf("second remind n=%d err=%v", n, err)
	}
}

func TestNoAccessAndNotFound(t *testing.T) {
	svc, _, _, _ := newTestSvc()
	ctx := context.Background()
	if _, err := svc.Get(ctx, uuid.New()); err == nil {
		t.Fatal("expected not found")
	}
	creator := uuid.New()
	row, err := svc.Create(ctx, creator, &dto.CreateTaskRequest{Title: "X"})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	other := uuid.New()
	if _, err := svc.Update(ctx, id, other, &dto.UpdateTaskRequest{}); err == nil {
		t.Fatal("expected no access")
	}
	if err := svc.Urge(ctx, id, other, "x"); err == nil {
		t.Fatal("urge no access")
	}
}
