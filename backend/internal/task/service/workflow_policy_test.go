package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	wfmodel "github.com/Yogdunana/StarByte/backend/internal/workflow/model"
)

type checkpointCall struct {
	stage, action string
	actor         uuid.UUID
}
type workflowStub struct {
	stage      string
	calls      []checkpointCall
	fail       error
	terminated bool
}

func (w *workflowStub) Start(context.Context, string, string, string, uuid.UUID, map[string]interface{}) (*wfmodel.FlowInstance, error) {
	w.stage = "assignment"
	return &wfmodel.FlowInstance{ID: uuid.New()}, w.fail
}
func (w *workflowStub) BusinessStage(context.Context, uuid.UUID) (string, bool, error) {
	return w.stage, w.stage == "completed", w.fail
}
func (w *workflowStub) TaskCheckpoint(_ context.Context, _ uuid.UUID, stage string, actor uuid.UUID, action, comment string) error {
	if w.fail != nil {
		return w.fail
	}
	if stage != w.stage {
		return errors.New("checkpoint mismatch")
	}
	w.calls = append(w.calls, checkpointCall{stage, action, actor})
	if action == "return" {
		w.stage = "execution"
	} else {
		w.stage = map[string]string{"assignment": "execution", "execution": "review", "review": "acceptance", "acceptance": "completed"}[stage]
	}
	return nil
}
func (w *workflowStub) Terminate(context.Context, uuid.UUID, uuid.UUID, string) error {
	if w.fail != nil {
		return w.fail
	}
	w.terminated = true
	return nil
}

func workflowFixture(t *testing.T) (*taskService, *memTasks, *workflowStub, []uuid.UUID) {
	t.Helper()
	svc, tasks, _, _ := newTestSvc()
	stub := &workflowStub{}
	svc.flow = stub
	pending := []func(){}
	svc.afterCommit = &pending
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	for i, id := range ids {
		tasks.users[id] = &model.NamedUser{ID: id, Username: []string{"owner", "executor", "reviewer", "acceptor"}[i]}
	}
	return svc, tasks, stub, ids
}
func TestWorkflowPolicyRequiresSeparateDeliveryAndSignatures(t *testing.T) {
	svc, tasks, stub, ids := workflowFixture(t)
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{Title: "Deliverable", AssigneeID: ids[1].String(), Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()}})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	if row.Status == model.StatusDone || row.WorkflowStage != "execution" {
		t.Fatal("creation completed the task")
	}
	if _, err := svc.ActWorkflow(ctx, id, ids[1], &dto.WorkflowActionRequest{Action: "start", Comment: "开始执行", Revision: 1}); err != nil {
		t.Fatal(err)
	}
	state, err := svc.GetWorkflow(ctx, id, ids[1])
	if err != nil {
		t.Fatal(err)
	}
	if !state.CanSubmit || state.CanApprove || state.Reviewer.Name != "reviewer" || state.Acceptor.Name != "acceptor" {
		t.Fatalf("wrong role projection: %+v", state)
	}
	act := func(actor uuid.UUID, action string) *dto.WorkflowResponse {
		state, err := svc.GetWorkflow(ctx, id, actor)
		if err != nil {
			t.Fatal(err)
		}
		next, err := svc.ActWorkflow(ctx, id, actor, &dto.WorkflowActionRequest{Action: action, Comment: "真实交付说明", Revision: state.Revision})
		if err != nil {
			t.Fatal(err)
		}
		return next
	}
	act(ids[1], "pause")
	act(ids[1], "resume")
	state = act(ids[1], "submit")
	if state.Stage != "review" || tasks.items[id].Status != 1 || state.Submission != "真实交付说明" {
		t.Fatal("submission must wait for review")
	}
	review, err := svc.GetWorkflow(ctx, id, ids[2])
	if err != nil {
		t.Fatal(err)
	}
	if !review.CanApprove || !review.CanReturn || review.CanSubmit {
		t.Fatal("review capability mismatch")
	}
	state = act(ids[2], "return")
	if state.Stage != "execution" || tasks.items[id].Progress != 50 {
		t.Fatal("rework lost execution stage")
	}
	act(ids[1], "submit")
	act(ids[2], "approve")
	state = act(ids[3], "approve")
	if state.Stage != "completed" || tasks.items[id].Progress != 100 || tasks.items[id].CompletedAt == nil || len(state.History) != 8 {
		t.Fatal("completion projection/history mismatch")
	}
	if len(stub.calls) != 6 || stub.calls[len(stub.calls)-1].actor != ids[3] {
		t.Fatal("signature calls omitted or impersonated")
	}
}
func TestWorkflowPolicyRejectsStaleBlankAndIncompleteDeliveries(t *testing.T) {
	svc, tasks, stub, ids := workflowFixture(t)
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{Title: "Deliverable", AssigneeID: ids[1].String(), Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()}})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	if _, err := svc.GetWorkflow(ctx, id, uuid.New()); err == nil {
		t.Fatal("unrelated viewer admitted")
	}
	for _, req := range []*dto.WorkflowActionRequest{nil, {Action: "submit", Comment: "  ", Revision: 1}, {Action: "submit", Comment: strings.Repeat("x", 5001), Revision: 1}, {Action: "submit", Comment: "old", Revision: 9}, {Action: "approve", Comment: "early", Revision: 1}, {Action: "wrong", Comment: "invalid", Revision: 1}} {
		if _, err := svc.ActWorkflow(ctx, id, ids[1], req); err == nil {
			t.Fatalf("invalid action accepted: %+v", req)
		}
	}
	if _, err := svc.ChangeStatus(ctx, id, ids[1], &dto.StatusRequest{Status: 1}); err != nil {
		t.Fatal(err)
	}
	child := &model.Task{ID: uuid.New(), CreatorID: ids[0], ParentID: &id, Status: model.StatusDoing}
	tasks.items[child.ID] = child
	if _, err := svc.ActWorkflow(ctx, id, ids[1], &dto.WorkflowActionRequest{Action: "submit", Comment: "incomplete children", Revision: 1}); err == nil {
		t.Fatal("unfinished child bypassed delivery")
	}
	delete(tasks.items, child.ID)
	before := len(stub.calls)
	stub.fail = errors.New("workflow unavailable")
	if _, err := svc.ActWorkflow(ctx, id, ids[1], &dto.WorkflowActionRequest{Action: "submit", Comment: "failed", Revision: 1}); !errors.Is(err, stub.fail) {
		t.Fatal("engine failure hidden")
	}
	if len(stub.calls) != before {
		t.Fatal("failed engine counted signature")
	}
}
func TestWorkflowAssignmentCancellationAndConfigurationGuards(t *testing.T) {
	svc, _, stub, ids := workflowFixture(t)
	ctx := context.Background()
	config := &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()}
	for _, invalid := range []*dto.WorkflowConfig{{ReviewerID: "invalid", AcceptorID: ids[3].String()}, {ReviewerID: ids[2].String(), AcceptorID: "invalid"}, {ReviewerID: ids[1].String(), AcceptorID: ids[3].String()}} {
		if _, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{Title: "Invalid", AssigneeID: ids[1].String(), Workflow: invalid}); err == nil {
			t.Fatal("bad signature config accepted")
		}
	}
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{Title: "Assignment", Workflow: config})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	if row.WorkflowStage != "assignment" {
		t.Fatal("unassigned task skipped assignment")
	}
	if _, err := svc.Assign(ctx, id, ids[0], ids[1].String()); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ChangeStatus(ctx, id, ids[1], &dto.StatusRequest{Status: 3, Comment: "not owner"}); err == nil {
		t.Fatal("executor cancelled formal flow")
	}
	if _, err := svc.ChangeStatus(ctx, id, ids[0], &dto.StatusRequest{Status: 3, Comment: " "}); err == nil {
		t.Fatal("blank cancellation accepted")
	}
	if _, err := svc.ChangeStatus(ctx, id, ids[0], &dto.StatusRequest{Status: 3, Comment: "cancel reason"}); err != nil {
		t.Fatal(err)
	}
	state, err := svc.GetWorkflow(ctx, id, ids[0])
	if err != nil {
		t.Fatal(err)
	}
	if state.Stage != "cancelled" || !stub.terminated {
		t.Fatal("cancel bypassed engine")
	}
	svc.flow = nil
	if _, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{Title: "No engine", Workflow: config}); err == nil {
		t.Fatal("missing workflow engine reported success")
	}
}

func (w *workflowStub) CompleteTaskTransferApproval(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, bool) error {
	return w.fail
}
func (w *workflowStub) ReassignTaskExecution(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, string) error {
	return w.fail
}
