package service

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/internal/task/repo"
	"github.com/Yogdunana/StarByte/backend/internal/workflow"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
)

// Explicit opt-in: all fixture rows roll back. Never run against another DB.
func TestTaskWorkflowPostgresLifecycle(t *testing.T) {
	dsn := os.Getenv("STARBYTE_TASK_TEST_DSN")
	if dsn == "" {
		t.Skip("set STARBYTE_TASK_TEST_DSN for isolated Postgres integration")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	var databaseName string
	if err := db.Raw("SELECT current_database()").Scan(&databaseName).Error; err != nil {
		t.Fatal(err)
	}
	if databaseName != "starbyte_preview" {
		t.Fatal("task integration tests only support starbyte_preview")
	}
	rollback := errors.New("rollback integration fixtures")
	err = db.Transaction(func(tx *gorm.DB) error {
		ctx := context.Background()
		owner, executor, reviewer, acceptor := uuid.New(), uuid.New(), uuid.New(), uuid.New()
		for _, id := range []uuid.UUID{owner, executor, reviewer, acceptor} {
			if err := tx.Exec("INSERT INTO users(id,username,password_hash,real_name,status) VALUES (?,?,'not-a-login-hash',?,0)", id, "task_it_"+id.String(), "任务事务测试").Error; err != nil {
				return err
			}
		}
		bus := events.NewEventBus()
		wf := workflow.Init(tx, bus, zap.NewNop())
		if err := wf.RegisterBusinessApprover(engine.TaskBusinessType, NewTaskApprover(tx)); err != nil {
			return err
		}
		svc := NewTaskService(repo.NewTaskRepo(tx), repo.NewLogRepo(tx), repo.NewCommentRepo(tx), repo.NewAttachmentRepo(tx), nil, nil, nil, tx).(*taskService)
		svc.SetWorkflowEngine(wf.Engine)
		create := func(assigned bool) *dto.TaskResponse {
			req := &dto.CreateTaskRequest{Title: "真实事务流程", Workflow: &dto.WorkflowConfig{ReviewerID: reviewer.String(), AcceptorID: acceptor.String()}}
			if assigned {
				req.AssigneeID = executor.String()
			}
			row, err := svc.Create(ctx, owner, req)
			if err != nil {
				t.Fatal(err)
			}
			return row
		}
		task := create(false)
		id := uuid.MustParse(task.ID)
		if task.WorkflowStage != "assignment" {
			t.Fatal("unassigned task skipped assignment")
		}
		if _, err := svc.Assign(ctx, id, owner, executor.String()); err != nil {
			return err
		}
		if _, err := svc.ChangeStatus(ctx, id, owner, &dto.StatusRequest{Status: 1}); err == nil {
			t.Fatal("creator impersonated executor")
		}
		if _, err := svc.ChangeStatus(ctx, id, executor, &dto.StatusRequest{Status: 1}); err != nil {
			return err
		}
		if _, err := svc.ChangeStatus(ctx, id, executor, &dto.StatusRequest{Status: 2}); err == nil {
			t.Fatal("direct completion bypassed signatures")
		}
		act := func(actor uuid.UUID, action string) *dto.WorkflowResponse {
			state, err := svc.GetWorkflow(ctx, id, actor)
			if err != nil {
				t.Fatal(err)
			}
			out, err := svc.ActWorkflow(ctx, id, actor, &dto.WorkflowActionRequest{Action: action, Comment: "实际交付与审核意见", Revision: state.Revision})
			if err != nil {
				t.Fatal(err)
			}
			return out
		}
		state := act(executor, "submit")
		if state.Stage != "review" {
			t.Fatal("submission skipped review")
		}
		if _, err := svc.Update(ctx, id, owner, &dto.UpdateTaskRequest{Description: new(string)}); err == nil {
			t.Fatal("review allowed editing delivery")
		}
		if _, err := svc.GetWorkflow(ctx, id, uuid.New()); err == nil {
			t.Fatal("unrelated viewer admitted")
		}
		if _, err := svc.ActWorkflow(ctx, id, executor, &dto.WorkflowActionRequest{Action: "approve", Comment: "not the reviewer", Revision: state.Revision}); err == nil {
			t.Fatal("executor signed review")
		}
		act(reviewer, "return")
		act(executor, "submit")
		state = act(reviewer, "approve")
		if state.Stage != "acceptance" {
			t.Fatal("review skipped acceptance")
		}
		state = act(acceptor, "approve")
		if state.Stage != "completed" || state.CanApprove || len(state.History) != 5 {
			t.Fatalf("wrong completion projection: %+v", state)
		}
		stored, err := repo.NewTaskRepo(tx).GetByID(ctx, id)
		if err != nil {
			return err
		}
		if stored.Status != model.StatusDone || stored.Progress != 100 || stored.CompletedAt == nil {
			t.Fatal("task and workflow completion diverged")
		}
		// New assigned task and cancellation must close its real pending engine task.
		second := create(true)
		secondID := uuid.MustParse(second.ID)
		if _, err := svc.ChangeStatus(ctx, secondID, owner, &dto.StatusRequest{Status: 3, Comment: "取消原因"}); err != nil {
			return err
		}
		cancelled, err := svc.GetWorkflow(ctx, secondID, owner)
		if err != nil {
			return err
		}
		if cancelled.Stage != "cancelled" {
			t.Fatal("cancelled task kept active stage")
		}
		var pending int64
		if err := tx.Table("flow_tasks").Where("instance_id=? AND status=0", second.WorkflowInstanceID).Count(&pending).Error; err != nil {
			return err
		}
		if pending != 0 {
			t.Fatal("cancel left orphan pending approval")
		}
		// Invalid config rolls back the newly inserted task and workflow together.
		_, err = svc.Create(ctx, owner, &dto.CreateTaskRequest{Title: "invalid signer", AssigneeID: executor.String(), Workflow: &dto.WorkflowConfig{ReviewerID: executor.String(), AcceptorID: acceptor.String()}})
		if err == nil || !strings.Contains(err.Error(), "不能") {
			t.Fatal("self review accepted")
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatal(err)
	}
}
