package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
)

func TestEscalateOverdueAssignmentAutoPicks(t *testing.T) {
	svc, tasks, stub, ids := workflowFixture(t)
	svc.transfers = newMemTransfers()
	dept := uuid.New()
	selected := uuid.New()
	tasks.users[selected] = &model.NamedUser{ID: selected, Username: "picked"}
	svc.assignments = &assignmentStub{picked: &model.NamedUser{ID: selected}}
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{
		Title: "Open", DepartmentID: dept.String(),
		Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()},
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	stub.overdue = []engine.OverdueCollaborationTodo{{
		InstanceID: *tasks.items[id].WorkflowInstanceID, BusinessKey: id.String(), Stage: "assignment", DueDate: time.Now().Add(-time.Hour),
	}}
	n, err := svc.EscalateOverdueWorkflows(ctx)
	if err != nil || n != 1 {
		t.Fatalf("escalate assignment: n=%d err=%v", n, err)
	}
	if tasks.items[id].AssigneeID == nil || *tasks.items[id].AssigneeID != selected || stub.stage != "execution" {
		t.Fatalf("assignment not claimed: %+v stage=%s", tasks.items[id], stub.stage)
	}
	n, err = svc.EscalateOverdueWorkflows(ctx)
	if err != nil || n != 1 {
		t.Fatalf("second pass should still run but skip: n=%d err=%v", n, err)
	}
	if tasks.items[id].AssigneeID == nil || *tasks.items[id].AssigneeID != selected {
		t.Fatal("second escalate reassigned")
	}
}

func TestEscalateOverdueExecutionReassigns(t *testing.T) {
	svc, tasks, stub, ids := workflowFixture(t)
	svc.transfers = newMemTransfers()
	dept := uuid.New()
	next := uuid.New()
	tasks.users[next] = &model.NamedUser{ID: next, Username: "next"}
	svc.assignments = &assignmentStub{picked: &model.NamedUser{ID: next}}
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{
		Title: "Late", AssigneeID: ids[1].String(), DepartmentID: dept.String(),
		Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()},
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	stub.overdue = []engine.OverdueCollaborationTodo{{
		InstanceID: *tasks.items[id].WorkflowInstanceID, BusinessKey: id.String(), Stage: "execution",
		AssigneeID: &ids[1], DueDate: time.Now().Add(-time.Hour),
	}}
	n, err := svc.EscalateOverdueWorkflows(ctx)
	if err != nil || n != 1 {
		t.Fatalf("escalate execution: n=%d err=%v", n, err)
	}
	if tasks.items[id].AssigneeID == nil || *tasks.items[id].AssigneeID != next || stub.next != next {
		t.Fatalf("execution not reassigned: %+v next=%s", tasks.items[id].AssigneeID, stub.next)
	}
}

func TestEscalateOverdueReviewFallsBackToCreator(t *testing.T) {
	svc, tasks, stub, ids := workflowFixture(t)
	svc.transfers = newMemTransfers()
	svc.assignments = &assignmentStub{}
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{
		Title: "Review late", AssigneeID: ids[1].String(),
		Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()},
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	tasks.items[id].WorkflowStage = "review"
	stub.stage = "review"
	stub.overdue = []engine.OverdueCollaborationTodo{{
		InstanceID: *tasks.items[id].WorkflowInstanceID, BusinessKey: id.String(), Stage: "review",
		AssigneeID: &ids[2], DueDate: time.Now().Add(-time.Hour),
	}}
	n, err := svc.EscalateOverdueWorkflows(ctx)
	if err != nil || n != 1 {
		t.Fatalf("escalate review: n=%d err=%v", n, err)
	}
	if tasks.items[id].ReviewerID == nil || *tasks.items[id].ReviewerID != ids[0] {
		t.Fatalf("review not escalated to creator: %v", tasks.items[id].ReviewerID)
	}
}

func TestPickEscalationTargetUsesLiveDepartment(t *testing.T) {
	svc, _, _, ids := workflowFixture(t)
	src, dst := uuid.New(), uuid.New()
	next := uuid.New()
	stub := &assignmentStub{picked: &model.NamedUser{ID: next}}
	svc.assignments = stub
	encoded := `{"mode":"department","department_id":"` + src.String() + `"}`
	task := &model.Task{DepartmentID: &dst, AssignmentPolicy: encoded, ReviewerID: &ids[2], AcceptorID: &ids[3]}
	got, err := svc.pickEscalationTarget(context.Background(), task, nil)
	if err != nil || got != next {
		t.Fatalf("pick: %s %v", got, err)
	}
	if stub.policy.DepartmentID != dst || stub.policy.Mode != "department" {
		t.Fatalf("stale assignment department used: %+v", stub.policy)
	}
}
