package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	rbac "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
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
	next       uuid.UUID
	overdue    []engine.OverdueCollaborationTodo
	done       bool
}

func (w *workflowStub) Start(context.Context, string, string, string, uuid.UUID, map[string]interface{}) (*wfmodel.FlowInstance, error) {
	w.stage = "assignment"
	return &wfmodel.FlowInstance{ID: uuid.New()}, w.fail
}
func (w *workflowStub) BusinessStage(context.Context, uuid.UUID) (string, bool, error) {
	if w.done {
		return "", true, w.fail
	}
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
	} else if action == "reject" {
		w.terminated = true
		w.stage = "rejected"
	} else {
		w.stage = map[string]string{"assignment": "execution", "execution": "review", "review": "acceptance", "acceptance": "completed"}[stage]
	}
	return nil
}
func (w *workflowStub) CompleteTaskApproval(_ context.Context, _ uuid.UUID, actor uuid.UUID, action engine.TaskAction, comment string) error {
	if w.fail != nil {
		return w.fail
	}
	w.calls = append(w.calls, checkpointCall{w.stage, string(action), actor})
	if action == engine.ActionReject {
		w.terminated = true
		w.stage = "rejected"
	} else {
		w.stage = map[string]string{"assignment": "execution", "execution": "review", "review": "acceptance", "acceptance": "completed"}[w.stage]
	}
	return nil
}
func (w *workflowStub) ClaimTaskAssignment(_ context.Context, _ uuid.UUID, actor uuid.UUID, _ string) error {
	if w.fail != nil {
		return w.fail
	}
	w.calls = append(w.calls, checkpointCall{"assignment", "claim", actor})
	w.stage = "execution"
	return nil
}
func (w *workflowStub) InstanceTerminated(context.Context, uuid.UUID) (bool, error) {
	return w.terminated, w.fail
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

func TestWorkflowRejectTerminatesTaskAndInstance(t *testing.T) {
	svc, tasks, stub, ids := workflowFixture(t)
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{Title: "Deliverable", AssigneeID: ids[1].String(), Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()}})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	act := func(actor uuid.UUID, action string) *dto.WorkflowResponse {
		state, err := svc.GetWorkflow(ctx, id, actor)
		if err != nil {
			t.Fatal(err)
		}
		next, err := svc.ActWorkflow(ctx, id, actor, &dto.WorkflowActionRequest{Action: action, Comment: "拒绝原因", Revision: state.Revision})
		if err != nil {
			t.Fatal(err)
		}
		return next
	}
	if _, err := svc.ActWorkflow(ctx, id, ids[1], &dto.WorkflowActionRequest{Action: "start", Comment: "开始执行", Revision: 1}); err != nil {
		t.Fatal(err)
	}
	act(ids[1], "submit")
	if _, err := svc.ActWorkflow(ctx, id, ids[1], &dto.WorkflowActionRequest{Action: "reject", Comment: "执行人不能拒绝", Revision: 2}); err == nil {
		t.Fatal("executor rejected review")
	}
	state := act(ids[2], "reject")
	if state.Stage != "rejected" || !stub.terminated || tasks.items[id].Status != model.StatusCancelled {
		t.Fatalf("reject did not terminate: %+v status=%d terminated=%v", state, tasks.items[id].Status, stub.terminated)
	}
	if _, err := svc.ActWorkflow(ctx, id, ids[3], &dto.WorkflowActionRequest{Action: "approve", Comment: "不能救活", Revision: state.Revision}); err == nil {
		t.Fatal("acceptor revived rejected task")
	}
}

func TestWorkflowClaimAdvancesAssignment(t *testing.T) {
	svc, tasks, stub, ids := workflowFixture(t)
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{Title: "Open task", Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()}})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	if row.WorkflowStage != "assignment" {
		t.Fatal("expected assignment stage")
	}
	claimer := model.WithViewer(ctx, model.Viewer{ID: ids[1], Scope: &rbac.DataScopeCondition{}})
	if _, err := svc.ActWorkflow(claimer, id, ids[2], &dto.WorkflowActionRequest{Action: "claim", Comment: "审核人不能认领自己的交付", Revision: 1}); err == nil {
		t.Fatal("reviewer claimed own review")
	}
	state, err := svc.ActWorkflow(claimer, id, ids[1], &dto.WorkflowActionRequest{Action: "claim", Comment: "我来认领", Revision: 1})
	if err != nil {
		t.Fatal(err)
	}
	if state.Stage != "execution" || tasks.items[id].AssigneeID == nil || *tasks.items[id].AssigneeID != ids[1] || len(stub.calls) == 0 {
		t.Fatalf("claim did not advance: %+v assignee=%v calls=%v", state, tasks.items[id].AssigneeID, stub.calls)
	}
}

func TestWorkflowClaimHonorsTaskReadScopeOnPersonalViewer(t *testing.T) {
	svc, tasks, stub, ids := workflowFixture(t)
	dept, other, peer, stranger := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	tasks.users[peer] = &model.NamedUser{ID: peer, Username: "peer"}
	tasks.users[stranger] = &model.NamedUser{ID: stranger, Username: "stranger"}
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{Title: "Dept task", DepartmentID: dept.String(), Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()}})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	personal := func(id uuid.UUID, read *rbac.DataScopeCondition) context.Context {
		return model.WithViewer(ctx, model.Viewer{
			ID:     id,
			Scope:  &rbac.DataScopeCondition{Query: "1 = 0", IsSelf: true},
			Scopes: map[string]*rbac.DataScopeCondition{"task:read": read},
		})
	}
	deptRead := &rbac.DataScopeCondition{Query: "department_id = ?", Args: []interface{}{dept}}
	selfRead := &rbac.DataScopeCondition{Query: "1 = 0", IsSelf: true}
	otherRead := &rbac.DataScopeCondition{Query: "department_id = ?", Args: []interface{}{other}}
	if _, err := svc.GetWorkflow(personal(stranger, selfRead), id, stranger); err == nil {
		t.Fatal("self-scope outsider opened workflow")
	}
	if _, err := svc.ActWorkflow(personal(stranger, selfRead), id, stranger, &dto.WorkflowActionRequest{Action: "claim", Comment: "我来认领", Revision: 1}); err == nil {
		t.Fatal("self-scope outsider claimed")
	}
	if _, err := svc.GetWorkflow(personal(stranger, otherRead), id, stranger); err == nil {
		t.Fatal("other-department viewer opened workflow")
	}
	if _, err := svc.GetWorkflow(personal(peer, deptRead), id, peer); err != nil {
		t.Fatal(err)
	}
	state, err := svc.ActWorkflow(personal(peer, deptRead), id, peer, &dto.WorkflowActionRequest{Action: "claim", Comment: "我来认领", Revision: 1})
	if err != nil {
		t.Fatal(err)
	}
	if state.Stage != "execution" || tasks.items[id].AssigneeID == nil || *tasks.items[id].AssigneeID != peer || len(stub.calls) == 0 {
		t.Fatalf("department peer could not claim: %+v assignee=%v calls=%v", state, tasks.items[id].AssigneeID, stub.calls)
	}
}

func (w *workflowStub) CompleteTaskTransferApproval(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, bool) error {
	return w.fail
}
func (w *workflowStub) ReassignTaskExecution(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, string) error {
	return w.fail
}
func (w *workflowStub) ReassignStage(_ context.Context, _ uuid.UUID, stage string, _, next, operator uuid.UUID, _ string) error {
	if w.fail != nil {
		return w.fail
	}
	w.calls = append(w.calls, checkpointCall{stage, "reassign", operator})
	w.next = next
	return nil
}
func (w *workflowStub) TransferTaskExecution(_ context.Context, _ uuid.UUID, from, to uuid.UUID, _ string) error {
	if w.fail != nil {
		return w.fail
	}
	w.calls = append(w.calls, checkpointCall{"execution", "delegate", from})
	w.next = to
	return nil
}
func (w *workflowStub) ListOverdueCollaborationTodos(context.Context, time.Time) ([]engine.OverdueCollaborationTodo, error) {
	return w.overdue, w.fail
}
