package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	wfmodel "github.com/Yogdunana/StarByte/backend/internal/workflow/model"
)

type memTransfers struct {
	items      map[uuid.UUID]*model.TaskTransfer
	pending    map[uuid.UUID]*model.TaskTransfer
	signatures map[uuid.UUID][]model.TransferSignature
	actors     map[uuid.UUID]*model.TransferActor
	depts      map[uuid.UUID]*model.TransferDepartment
}

func newMemTransfers() *memTransfers {
	return &memTransfers{
		items: map[uuid.UUID]*model.TaskTransfer{}, pending: map[uuid.UUID]*model.TaskTransfer{},
		signatures: map[uuid.UUID][]model.TransferSignature{}, actors: map[uuid.UUID]*model.TransferActor{},
		depts: map[uuid.UUID]*model.TransferDepartment{},
	}
}
func (m *memTransfers) Create(_ context.Context, t *model.TaskTransfer) error {
	cp := *t
	m.items[t.ID] = &cp
	if t.Status == "pending" {
		m.pending[t.TaskID] = &cp
	}
	return nil
}
func (m *memTransfers) Save(_ context.Context, t *model.TaskTransfer) error {
	cp := *t
	m.items[t.ID] = &cp
	if t.Status == "pending" {
		m.pending[t.TaskID] = &cp
	} else {
		delete(m.pending, t.TaskID)
	}
	return nil
}
func (m *memTransfers) Get(_ context.Context, id uuid.UUID) (*model.TaskTransfer, error) {
	if m.items[id] == nil {
		return nil, nil
	}
	cp := *m.items[id]
	return &cp, nil
}
func (m *memTransfers) Lock(ctx context.Context, id uuid.UUID) (*model.TaskTransfer, error) {
	return m.Get(ctx, id)
}
func (m *memTransfers) Pending(_ context.Context, id uuid.UUID) (*model.TaskTransfer, error) {
	if m.pending[id] == nil {
		return nil, nil
	}
	cp := *m.pending[id]
	return &cp, nil
}
func (m *memTransfers) Signatures(_ context.Context, id uuid.UUID) ([]model.TransferSignature, error) {
	return append([]model.TransferSignature{}, m.signatures[id]...), nil
}
func (m *memTransfers) Sign(_ context.Context, s *model.TransferSignature) error {
	m.signatures[s.TransferID] = append(m.signatures[s.TransferID], *s)
	return nil
}
func (m *memTransfers) Actors(context.Context) ([]model.TransferActor, error) {
	out := make([]model.TransferActor, 0, len(m.actors))
	for _, a := range m.actors {
		out = append(out, *a)
	}
	return out, nil
}
func (m *memTransfers) Actor(_ context.Context, id uuid.UUID) (*model.TransferActor, error) {
	return m.actors[id], nil
}
func (m *memTransfers) Department(_ context.Context, id uuid.UUID) (*model.TransferDepartment, error) {
	return m.depts[id], nil
}

func TestSameDepartmentDelegateTransfersExecution(t *testing.T) {
	svc, tasks, stub, ids := workflowFixture(t)
	store := newMemTransfers()
	svc.transfers = store
	dept := uuid.New()
	store.depts[dept] = &model.TransferDepartment{ID: dept, Name: "同组"}
	peer := uuid.New()
	tasks.users[peer] = &model.NamedUser{ID: peer, Username: "peer", DepartmentID: &dept}
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{Title: "Delegate", AssigneeID: ids[1].String(), DepartmentID: dept.String(), Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()}})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	out, err := svc.RequestHandover(ctx, id, ids[1], &dto.HandoverRequest{TargetID: peer.String(), Reason: "请接手联调", Revision: 1})
	if err != nil {
		t.Fatal(err)
	}
	if out.Kind != "internal" || out.Status != "completed" || tasks.items[id].AssigneeID == nil || *tasks.items[id].AssigneeID != peer {
		t.Fatalf("delegate not applied: %+v assignee=%v", out, tasks.items[id].AssigneeID)
	}
	if stub.next != peer || len(stub.calls) == 0 || stub.calls[len(stub.calls)-1].action != "delegate" {
		t.Fatal("engine handover skipped")
	}
	state, err := svc.GetWorkflow(ctx, id, ids[0])
	if err != nil || !state.CanDelegate && tasks.items[id].AssigneeID != nil && *tasks.items[id].AssigneeID == ids[1] {
		t.Fatal(err)
	}
}

func TestCrossDepartmentHandoverWaitsForSignatures(t *testing.T) {
	svc, tasks, stub, ids := workflowFixture(t)
	store := newMemTransfers()
	svc.transfers = store
	src, dst, center := uuid.New(), uuid.New(), uuid.New()
	store.depts[src] = &model.TransferDepartment{ID: src, ParentID: &center, Name: "前端"}
	store.depts[dst] = &model.TransferDepartment{ID: dst, ParentID: &center, Name: "后端"}
	peer := uuid.New()
	minister := uuid.New()
	tasks.users[peer] = &model.NamedUser{ID: peer, Username: "peer", DepartmentID: &dst}
	store.actors[ids[1]] = &model.TransferActor{ID: ids[1], DepartmentID: &src, Roles: []string{}}
	store.actors[minister] = &model.TransferActor{ID: minister, DepartmentID: &src, Roles: []string{"minister"}}
	store.actors[uuid.New()] = &model.TransferActor{}
	otherMinister := uuid.New()
	store.actors[otherMinister] = &model.TransferActor{ID: otherMinister, DepartmentID: &dst, Roles: []string{"minister"}}
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{
		Title: "Handover", AssigneeID: ids[1].String(), DepartmentID: src.String(),
		Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()},
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	out, err := svc.RequestHandover(ctx, id, ids[1], &dto.HandoverRequest{TargetID: peer.String(), Reason: "跨组接手", Revision: 1})
	if err != nil {
		t.Fatal(err)
	}
	if out.Kind != "department" || out.Status != "pending" || tasks.items[id].AssigneeID == nil || *tasks.items[id].AssigneeID != ids[1] {
		t.Fatalf("signed handover applied early: %+v", out)
	}
	if _, err := svc.RequestHandover(ctx, id, ids[1], &dto.HandoverRequest{TargetID: peer.String(), Reason: "重复申请", Revision: 2}); err == nil {
		t.Fatal("duplicate pending handover accepted")
	}
	if _, err := svc.DecideHandover(ctx, id, ids[1], &dto.HandoverDecision{Requirement: "source_minister", Decision: "approve", Comment: "自己签", Revision: 1}); err == nil {
		t.Fatal("initiator signed own handover")
	}
	stub.done = false
	first, err := svc.DecideHandover(ctx, id, minister, &dto.HandoverDecision{Requirement: "source_minister", Decision: "approve", Comment: "同意转出", Revision: 1})
	if err != nil || first.Status != "pending" {
		t.Fatalf("first signature: %v %+v", err, first)
	}
	stub.done = true
	final, err := svc.DecideHandover(ctx, id, otherMinister, &dto.HandoverDecision{Requirement: "target_minister", Decision: "approve", Comment: "同意转入", Revision: 2})
	if err != nil || final.Status != "completed" || tasks.items[id].AssigneeID == nil || *tasks.items[id].AssigneeID != peer {
		t.Fatalf("completed handover: %v %+v assignee=%v", err, final, tasks.items[id].AssigneeID)
	}
}

func TestHandoverRejectKeepsOriginalAssignee(t *testing.T) {
	svc, tasks, stub, ids := workflowFixture(t)
	store := newMemTransfers()
	svc.transfers = store
	src, dst, center := uuid.New(), uuid.New(), uuid.New()
	store.depts[src] = &model.TransferDepartment{ID: src, ParentID: &center, Name: "A"}
	store.depts[dst] = &model.TransferDepartment{ID: dst, ParentID: &center, Name: "B"}
	peer := uuid.New()
	minister := uuid.New()
	tasks.users[peer] = &model.NamedUser{ID: peer, Username: "peer", DepartmentID: &dst}
	store.actors[minister] = &model.TransferActor{ID: minister, DepartmentID: &src, Roles: []string{"minister"}}
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{
		Title: "Reject", AssigneeID: ids[1].String(), DepartmentID: src.String(),
		Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()},
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	if _, err := svc.RequestHandover(ctx, id, ids[1], &dto.HandoverRequest{TargetID: peer.String(), Reason: "跨组", Revision: 1}); err != nil {
		t.Fatal(err)
	}
	out, err := svc.DecideHandover(ctx, id, minister, &dto.HandoverDecision{Requirement: "source_minister", Decision: "reject", Comment: "人手不够", Revision: 1})
	if err != nil || out.Status != "rejected" || *tasks.items[id].AssigneeID != ids[1] || !stub.terminated {
		t.Fatalf("reject handover: %v %+v assignee=%v", err, out, tasks.items[id].AssigneeID)
	}
}

func TestTransferPolicyAuthority(t *testing.T) {
	src, dst, center := uuid.New(), uuid.New(), uuid.New()
	row := &model.TaskTransfer{
		InitiatorID: uuid.New(), FromUserID: uuid.New(), ToUserID: uuid.New(),
		SourceDepartmentID: src, TargetDepartmentID: dst, SourceCenterID: center, TargetCenterID: center,
		SupervisorRole: "minister",
	}
	minister := &model.TransferActor{ID: uuid.New(), DepartmentID: &src, Roles: []string{"minister"}}
	if actual, waived := transferAuthority(minister, row, "source_minister"); actual != "minister" || waived {
		t.Fatalf("source minister: %s %v", actual, waived)
	}
	if actual, _ := transferAuthority(minister, row, "target_minister"); actual != "" {
		t.Fatal("source minister signed target")
	}
	director := &model.TransferActor{ID: uuid.New(), DepartmentID: &center, Roles: []string{"center_director"}}
	if actual, waived := transferAuthority(director, row, "source_minister"); actual != "center" || !waived {
		t.Fatalf("center waiver: %s %v", actual, waived)
	}
}

func TestClassifyHandoverKinds(t *testing.T) {
	svc, _, _, ids := workflowFixture(t)
	store := newMemTransfers()
	svc.transfers = store
	src, dst, other, c1, c2 := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	store.depts[src] = &model.TransferDepartment{ID: src, ParentID: &c1, Name: "A"}
	store.depts[dst] = &model.TransferDepartment{ID: dst, ParentID: &c1, Name: "B"}
	store.depts[other] = &model.TransferDepartment{ID: other, ParentID: &c2, Name: "C"}
	task := &model.Task{DepartmentID: &src, AssigneeID: &ids[1]}
	kind, _, _, _, _, _, err := svc.classifyHandover(context.Background(), task, &model.NamedUser{ID: uuid.New(), DepartmentID: &src})
	if err != nil || kind != "internal" {
		t.Fatalf("internal: %s %v", kind, err)
	}
	kind, _, _, _, _, _, err = svc.classifyHandover(context.Background(), task, &model.NamedUser{ID: uuid.New(), DepartmentID: &dst})
	if err != nil || kind != "department" {
		t.Fatalf("department: %s %v", kind, err)
	}
	kind, _, _, _, _, _, err = svc.classifyHandover(context.Background(), task, &model.NamedUser{ID: uuid.New(), DepartmentID: &other})
	if err != nil || kind != "center" {
		t.Fatalf("center: %s %v", kind, err)
	}
	if _, _, _, _, _, _, err = svc.classifyHandover(context.Background(), task, &model.NamedUser{ID: uuid.New()}); err == nil {
		t.Fatal("missing target department treated as internal")
	}
	if _, _, _, _, _, _, err = svc.classifyHandover(context.Background(), &model.Task{AssigneeID: &ids[1]}, &model.NamedUser{ID: uuid.New(), DepartmentID: &src}); err == nil {
		t.Fatal("missing source department treated as internal")
	}
	store.actors[ids[1]] = &model.TransferActor{ID: ids[1], DepartmentID: &dst}
	stale := &model.Task{DepartmentID: &src, AssigneeID: &ids[1]}
	kind, sourceDept, _, _, _, _, err := svc.classifyHandover(context.Background(), stale, &model.NamedUser{ID: uuid.New(), DepartmentID: &dst})
	if err != nil || kind != "internal" || sourceDept != dst {
		t.Fatalf("stale task department hid current assignee team: %s %s %v", kind, sourceDept, err)
	}
}

func TestGetAndDecideTransferByID(t *testing.T) {
	svc, tasks, stub, ids := workflowFixture(t)
	store := newMemTransfers()
	svc.transfers = store
	src, dst, center := uuid.New(), uuid.New(), uuid.New()
	store.depts[src] = &model.TransferDepartment{ID: src, ParentID: &center, Name: "A"}
	store.depts[dst] = &model.TransferDepartment{ID: dst, ParentID: &center, Name: "B"}
	peer := uuid.New()
	minister := uuid.New()
	tasks.users[peer] = &model.NamedUser{ID: peer, Username: "peer", DepartmentID: &dst}
	store.actors[minister] = &model.TransferActor{ID: minister, DepartmentID: &src, Roles: []string{"minister"}}
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{
		Title: "ByID", AssigneeID: ids[1].String(), DepartmentID: src.String(),
		Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()},
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	created, err := svc.RequestHandover(ctx, id, ids[1], &dto.HandoverRequest{TargetID: peer.String(), Reason: "跨组", Revision: 1})
	if err != nil {
		t.Fatal(err)
	}
	got, err := svc.GetHandover(ctx, id, ids[1])
	if err != nil || got.ID != created.ID {
		t.Fatalf("get handover: %v %+v", err, got)
	}
	transferID := uuid.MustParse(created.ID)
	viaID, err := svc.GetTransfer(ctx, transferID, minister)
	if err != nil || viaID.ID != created.ID || len(viaID.CanSign) == 0 {
		t.Fatalf("get transfer: %v %+v", err, viaID)
	}
	if _, err := svc.GetHandover(ctx, id, uuid.New()); err == nil {
		t.Fatal("stranger read handover")
	}
	out, err := svc.DecideTransfer(ctx, transferID, minister, &dto.HandoverDecision{Requirement: "source_minister", Decision: "reject", Comment: "人手不够", Revision: 1})
	if err != nil || out.Status != "rejected" || !stub.terminated {
		t.Fatalf("decide transfer: %v %+v", err, out)
	}
	afterReject, err := svc.GetTransfer(ctx, transferID, minister)
	if err != nil || afterReject.Status != "rejected" {
		t.Fatalf("signer lost rejected transfer: %v %+v", err, afterReject)
	}
	if _, err := svc.GetTransfer(ctx, transferID, uuid.New()); err == nil {
		t.Fatal("stranger read rejected transfer")
	}
}

func TestHandoverGuardsRejectSelfReviewerAndWrongStage(t *testing.T) {
	svc, tasks, _, ids := workflowFixture(t)
	svc.transfers = newMemTransfers()
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{Title: "Guards", Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()}})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	if _, err := svc.RequestHandover(ctx, id, ids[0], &dto.HandoverRequest{TargetID: ids[1].String(), Reason: "未分配", Revision: 1}); err == nil {
		t.Fatal("publisher delegated unassigned task")
	}
	if _, err := svc.Assign(ctx, id, ids[0], ids[1].String()); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RequestHandover(ctx, id, ids[1], &dto.HandoverRequest{TargetID: ids[1].String(), Reason: "自己", Revision: 2}); err == nil {
		t.Fatal("self delegate accepted")
	}
	if _, err := svc.RequestHandover(ctx, id, ids[1], &dto.HandoverRequest{TargetID: ids[2].String(), Reason: "审核人", Revision: 2}); err == nil {
		t.Fatal("reviewer became executor")
	}
	if tasks.items[id].AssigneeID == nil || *tasks.items[id].AssigneeID != ids[1] {
		t.Fatal("assignee lost after failed handover")
	}
}

func TestTransferApproverRejectsMismatchedBusiness(t *testing.T) {
	resolver := NewTransferApprover(nil)
	if resolver.ForTransaction(nil) == nil {
		t.Fatal("missing transactional resolver")
	}
	if _, err := resolver.Resolve(context.Background(), &wfmodel.FlowInstance{BusinessType: engine.TaskBusinessType}, &engine.FlowNode{}); err == nil {
		t.Fatal("collaboration instance used transfer policy")
	}
	if _, err := resolver.Resolve(context.Background(), &wfmodel.FlowInstance{BusinessType: engine.TaskTransferBusinessType, BusinessKey: "not-a-uuid"}, &engine.FlowNode{}); err == nil {
		t.Fatal("invalid transfer key accepted")
	}
}

func TestWorkflowTransferUsesHandoverInsteadOfSilentBlock(t *testing.T) {
	svc, tasks, _, ids := workflowFixture(t)
	store := newMemTransfers()
	svc.transfers = store
	dept := uuid.New()
	store.depts[dept] = &model.TransferDepartment{ID: dept, Name: "同组"}
	peer := uuid.New()
	tasks.users[peer] = &model.NamedUser{ID: peer, Username: "peer", DepartmentID: &dept}
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{Title: "Transfer API", AssigneeID: ids[1].String(), DepartmentID: dept.String(), Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()}})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	out, err := svc.Transfer(ctx, id, ids[1], &dto.TransferRequest{NewAssigneeID: peer.String(), Reason: "请接手"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Assignee == nil || out.Assignee.ID != peer.String() {
		t.Fatalf("transfer API did not delegate: %+v", out)
	}
}

type handoverNotifyCapture struct{ templates []string }

func (c *handoverNotifyCapture) Send(_ context.Context, _ []uuid.UUID, template string, _ map[string]interface{}) error {
	c.templates = append(c.templates, template)
	return nil
}

func TestPendingHandoverBlocksSubmitAndIgnoresStaleTransferID(t *testing.T) {
	svc, tasks, _, ids := workflowFixture(t)
	store := newMemTransfers()
	svc.transfers = store
	src, dst, center := uuid.New(), uuid.New(), uuid.New()
	store.depts[src] = &model.TransferDepartment{ID: src, ParentID: &center, Name: "A"}
	store.depts[dst] = &model.TransferDepartment{ID: dst, ParentID: &center, Name: "B"}
	peer, minister := uuid.New(), uuid.New()
	tasks.users[peer] = &model.NamedUser{ID: peer, Username: "peer", DepartmentID: &dst}
	store.actors[minister] = &model.TransferActor{ID: minister, DepartmentID: &src, Roles: []string{"minister"}}
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{
		Title: "PendingSubmit", AssigneeID: ids[1].String(), DepartmentID: src.String(),
		Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()},
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	if _, err := svc.ActWorkflow(ctx, id, ids[1], &dto.WorkflowActionRequest{Action: "start", Comment: "开始", Revision: 1}); err != nil {
		t.Fatal(err)
	}
	created, err := svc.RequestHandover(ctx, id, ids[1], &dto.HandoverRequest{TargetID: peer.String(), Reason: "跨组", Revision: 2})
	if err != nil {
		t.Fatal(err)
	}
	state, err := svc.GetWorkflow(ctx, id, ids[1])
	if err != nil || state.CanSubmit || state.CanStart || state.CanDelegate {
		t.Fatalf("pending handover still executable: %+v %v", state, err)
	}
	if _, err := svc.ActWorkflow(ctx, id, ids[1], &dto.WorkflowActionRequest{Action: "submit", Comment: "交付说明", Revision: state.Revision}); err == nil {
		t.Fatal("submit accepted during pending handover")
	}
	oldID := uuid.MustParse(created.ID)
	if _, err := svc.DecideTransfer(ctx, oldID, minister, &dto.HandoverDecision{Requirement: "source_minister", Decision: "reject", Comment: "先拒", Revision: 1}); err != nil {
		t.Fatal(err)
	}
	next, err := svc.RequestHandover(ctx, id, ids[1], &dto.HandoverRequest{TargetID: peer.String(), Reason: "再转一次", Revision: tasks.items[id].WorkflowRevision})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DecideTransfer(ctx, oldID, minister, &dto.HandoverDecision{Requirement: "source_minister", Decision: "approve", Comment: "旧单仍想签", Revision: 1}); err == nil {
		t.Fatal("stale transfer id signed the later handover")
	}
	pending, err := store.Pending(ctx, id)
	if err != nil || pending == nil || pending.ID.String() != next.ID {
		t.Fatalf("later handover lost: %v %+v", err, pending)
	}
}

func TestHandoverRequestDoesNotNotifyAsTransferred(t *testing.T) {
	svc, tasks, stub, ids := workflowFixture(t)
	store := newMemTransfers()
	svc.transfers = store
	capture := &handoverNotifyCapture{}
	svc.notify = capture
	flushNotify := func() {
		if svc.afterCommit == nil {
			return
		}
		pending := *svc.afterCommit
		*svc.afterCommit = nil
		for _, fn := range pending {
			fn()
		}
	}
	src, dst, center := uuid.New(), uuid.New(), uuid.New()
	store.depts[src] = &model.TransferDepartment{ID: src, ParentID: &center, Name: "A"}
	store.depts[dst] = &model.TransferDepartment{ID: dst, ParentID: &center, Name: "B"}
	peer, minister, other := uuid.New(), uuid.New(), uuid.New()
	tasks.users[peer] = &model.NamedUser{ID: peer, Username: "peer", DepartmentID: &dst}
	store.actors[minister] = &model.TransferActor{ID: minister, DepartmentID: &src, Roles: []string{"minister"}}
	store.actors[other] = &model.TransferActor{ID: other, DepartmentID: &dst, Roles: []string{"minister"}}
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{
		Title: "Notify", AssigneeID: ids[1].String(), DepartmentID: src.String(),
		Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()},
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	if _, err := svc.RequestHandover(ctx, id, ids[1], &dto.HandoverRequest{TargetID: peer.String(), Reason: "跨组", Revision: 1}); err != nil {
		t.Fatal(err)
	}
	flushNotify()
	for _, tpl := range capture.templates {
		if tpl == tplTaskTransferred {
			t.Fatal("request notified as already transferred")
		}
	}
	if _, err := svc.DecideHandover(ctx, id, minister, &dto.HandoverDecision{Requirement: "source_minister", Decision: "approve", Comment: "同意转出", Revision: 1}); err != nil {
		t.Fatal(err)
	}
	stub.done = true
	if _, err := svc.DecideHandover(ctx, id, other, &dto.HandoverDecision{Requirement: "target_minister", Decision: "approve", Comment: "同意转入", Revision: 2}); err != nil {
		t.Fatal(err)
	}
	flushNotify()
	found := false
	for _, tpl := range capture.templates {
		if tpl == tplTaskTransferred {
			found = true
		}
	}
	if !found {
		t.Fatal("completed handover did not notify transfer")
	}
}

func TestSignedHandoverCancelsWhenExecutionAlreadyLeft(t *testing.T) {
	svc, tasks, stub, ids := workflowFixture(t)
	store := newMemTransfers()
	svc.transfers = store
	src, dst, center := uuid.New(), uuid.New(), uuid.New()
	store.depts[src] = &model.TransferDepartment{ID: src, ParentID: &center, Name: "A"}
	store.depts[dst] = &model.TransferDepartment{ID: dst, ParentID: &center, Name: "B"}
	peer, minister := uuid.New(), uuid.New()
	tasks.users[peer] = &model.NamedUser{ID: peer, Username: "peer", DepartmentID: &dst}
	store.actors[minister] = &model.TransferActor{ID: minister, DepartmentID: &src, Roles: []string{"minister"}}
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{
		Title: "LeftExecution", AssigneeID: ids[1].String(), DepartmentID: src.String(),
		Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()},
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	if _, err := svc.RequestHandover(ctx, id, ids[1], &dto.HandoverRequest{TargetID: peer.String(), Reason: "跨组", Revision: 1}); err != nil {
		t.Fatal(err)
	}
	tasks.items[id].WorkflowStage = "review"
	stub.done = true
	out, err := svc.DecideHandover(ctx, id, minister, &dto.HandoverDecision{Requirement: "source_minister", Decision: "approve", Comment: "最后一签", Revision: 1})
	if err != nil || out.Status != "cancelled" || *tasks.items[id].AssigneeID != ids[1] {
		t.Fatalf("expected cancel not stuck complete: %v %+v assignee=%v", err, out, tasks.items[id].AssigneeID)
	}
	if pending, _ := store.Pending(ctx, id); pending != nil {
		t.Fatal("cancelled handover still pending")
	}
}

func TestMissingDepartmentCannotSkipSignatures(t *testing.T) {
	svc, tasks, _, ids := workflowFixture(t)
	svc.transfers = newMemTransfers()
	peer := uuid.New()
	tasks.users[peer] = &model.NamedUser{ID: peer, Username: "peer"}
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{Title: "NoDept", AssigneeID: ids[1].String(), Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RequestHandover(ctx, uuid.MustParse(row.ID), ids[1], &dto.HandoverRequest{TargetID: peer.String(), Reason: "无部门仍想直接交", Revision: 1}); err == nil {
		t.Fatal("missing departments skipped signatures")
	}
	if tasks.items[uuid.MustParse(row.ID)].AssigneeID == nil || *tasks.items[uuid.MustParse(row.ID)].AssigneeID != ids[1] {
		t.Fatal("assignee changed without signatures")
	}
}

func TestCompletedHandoverMovesDepartmentAndKeepsSignerAccess(t *testing.T) {
	svc, tasks, stub, ids := workflowFixture(t)
	store := newMemTransfers()
	svc.transfers = store
	src, dst, center := uuid.New(), uuid.New(), uuid.New()
	store.depts[src] = &model.TransferDepartment{ID: src, ParentID: &center, Name: "前端"}
	store.depts[dst] = &model.TransferDepartment{ID: dst, ParentID: &center, Name: "后端"}
	peer, teammate, minister, otherMinister := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	tasks.users[peer] = &model.NamedUser{ID: peer, Username: "peer", DepartmentID: &dst}
	tasks.users[teammate] = &model.NamedUser{ID: teammate, Username: "teammate", DepartmentID: &dst}
	tasks.users[ids[1]].DepartmentID = &src
	store.actors[ids[1]] = &model.TransferActor{ID: ids[1], DepartmentID: &src}
	store.actors[peer] = &model.TransferActor{ID: peer, DepartmentID: &dst}
	store.actors[minister] = &model.TransferActor{ID: minister, DepartmentID: &src, Roles: []string{"minister"}}
	store.actors[otherMinister] = &model.TransferActor{ID: otherMinister, DepartmentID: &dst, Roles: []string{"minister"}}
	ctx := context.Background()
	row, err := svc.Create(ctx, ids[0], &dto.CreateTaskRequest{
		Title: "DeptMove", AssigneeID: ids[1].String(), DepartmentID: src.String(),
		Workflow: &dto.WorkflowConfig{ReviewerID: ids[2].String(), AcceptorID: ids[3].String()},
	})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(row.ID)
	created, err := svc.RequestHandover(ctx, id, ids[1], &dto.HandoverRequest{TargetID: peer.String(), Reason: "跨组接手", Revision: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DecideHandover(ctx, id, minister, &dto.HandoverDecision{Requirement: "source_minister", Decision: "approve", Comment: "同意转出", Revision: 1}); err != nil {
		t.Fatal(err)
	}
	stub.done = true
	final, err := svc.DecideHandover(ctx, id, otherMinister, &dto.HandoverDecision{Requirement: "target_minister", Decision: "approve", Comment: "同意转入", Revision: 2})
	if err != nil || final.Status != "completed" {
		t.Fatalf("completed handover: %v %+v", err, final)
	}
	if tasks.items[id].DepartmentID == nil || *tasks.items[id].DepartmentID != dst {
		t.Fatalf("task stayed in source department: %v", tasks.items[id].DepartmentID)
	}
	transferID := uuid.MustParse(created.ID)
	for _, signer := range []uuid.UUID{minister, otherMinister} {
		got, err := svc.GetTransfer(ctx, transferID, signer)
		if err != nil || got.Status != "completed" {
			t.Fatalf("signer lost completed transfer: actor=%s err=%v %+v", signer, err, got)
		}
	}
	sameTeam, err := svc.RequestHandover(ctx, id, peer, &dto.HandoverRequest{TargetID: teammate.String(), Reason: "同组再交", Revision: tasks.items[id].WorkflowRevision})
	if err != nil || sameTeam.Kind != "internal" || sameTeam.Status != "completed" {
		t.Fatalf("same-team after transfer still required signatures: %v %+v", err, sameTeam)
	}
	back, err := svc.RequestHandover(ctx, id, teammate, &dto.HandoverRequest{TargetID: ids[1].String(), Reason: "转回原组", Revision: tasks.items[id].WorkflowRevision})
	if err != nil || back.Kind != "department" || back.Status != "pending" {
		t.Fatalf("return to original department skipped signatures: %v %+v", err, back)
	}
}
