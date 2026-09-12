package engine

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

type storingInstRepo struct {
	insts map[uuid.UUID]*model.FlowInstance
}

func (m *storingInstRepo) Create(_ context.Context, _ *gorm.DB, inst *model.FlowInstance) error {
	if m.insts == nil {
		m.insts = map[uuid.UUID]*model.FlowInstance{}
	}
	m.insts[inst.ID] = inst
	return nil
}
func (m *storingInstRepo) GetByID(_ context.Context, id uuid.UUID) (*model.FlowInstance, error) {
	return m.insts[id], nil
}
func (m *storingInstRepo) Update(_ context.Context, _ *gorm.DB, inst *model.FlowInstance) error {
	m.insts[inst.ID] = inst
	return nil
}
func (m *storingInstRepo) List(context.Context, int, int, *int, *uuid.UUID, *uuid.UUID) ([]model.FlowInstance, int64, error) {
	return nil, 0, nil
}

type recordingApproval struct {
	waitingTestNode
	enteredIDs []string
}

func (n *recordingApproval) OnEnter(ctx context.Context, inst *model.FlowInstance, node *FlowNode, vars map[string]interface{}) error {
	n.enteredIDs = append(n.enteredIDs, node.ID)
	return n.waitingTestNode.OnEnter(ctx, inst, node, vars)
}

func memberApplicationEngine(t *testing.T, tasks *mockTaskRepo) (*FlowEngine, *storingInstRepo) {
	t.Helper()
	return memberApplicationEngineWith(t, tasks, &waitingTestNode{})
}

func memberApplicationEngineWith(t *testing.T, tasks *mockTaskRepo, approval NodeHandler) (*FlowEngine, *storingInstRepo) {
	t.Helper()
	defID, verID := uuid.New(), uuid.New()
	def := &model.FlowDefinition{ID: defID, Key: MemberApplicationDefinitionKey, Status: 1}
	ver := &model.FlowDefinitionVersion{ID: verID, DefinitionID: defID, BpmnData: MemberApplicationBPMN(), Status: 1}
	insts := &storingInstRepo{insts: map[uuid.UUID]*model.FlowInstance{}}
	e := NewFlowEngine(&mockDefRepo{def: def, version: ver}, insts, tasks, newMockVarRepo(), nil, &mockRegistryForTest{handlers: map[string]NodeHandler{
		"start": stubStartHandler{}, "end": stubEndHandler{}, "approval": approval,
	}}, NewExpressionEngine(), events.NewEventBus(), nil)
	e.businessTransaction = true
	return e, insts
}

func addApprovalTask(tasks *mockTaskRepo, instanceID uuid.UUID, nodeID string, assignee uuid.UUID) {
	id := uuid.New()
	activation := uuid.New()
	tasks.tasks[id] = &model.FlowTask{
		ID: id, InstanceID: instanceID, NodeID: nodeID, NodeName: nodeID, TaskType: "approval",
		ActivationID: &activation, AssigneeID: &assignee, Status: 0,
	}
}

func TestMemberApplicationStartOfficerMinisterPresidentComplete(t *testing.T) {
	officer, minister, president, applicant := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	tasks := newMockTaskRepo()
	e, insts := memberApplicationEngine(t, tasks)

	inst, err := e.Start(context.Background(), MemberApplicationDefinitionKey, uuid.New().String(), "member_application", applicant, map[string]interface{}{
		"applicant": applicant.String(), "apply_type": int16(2),
	})
	require.NoError(t, err)
	require.Equal(t, 0, inst.Status)
	nodeID, done, err := e.RunningApprovalNode(context.Background(), inst.ID)
	require.NoError(t, err)
	require.False(t, done)
	require.Equal(t, "officer", nodeID)

	addApprovalTask(tasks, inst.ID, "officer", officer)
	require.NoError(t, e.CompleteApplicationApproval(context.Background(), inst.ID, officer, ActionApprove, "干事同意"))
	nodeID, done, err = e.RunningApprovalNode(context.Background(), inst.ID)
	require.NoError(t, err)
	require.False(t, done)
	require.Equal(t, "minister", nodeID)

	addApprovalTask(tasks, inst.ID, "minister", minister)
	require.NoError(t, e.CompleteApplicationApproval(context.Background(), inst.ID, minister, ActionApprove, "部长同意"))
	nodeID, done, err = e.RunningApprovalNode(context.Background(), inst.ID)
	require.NoError(t, err)
	require.False(t, done)
	require.Equal(t, "president", nodeID)

	addApprovalTask(tasks, inst.ID, "president", president)
	require.NoError(t, e.CompleteApplicationApproval(context.Background(), inst.ID, president, ActionApprove, "社长同意"))
	_, done, err = e.RunningApprovalNode(context.Background(), inst.ID)
	require.NoError(t, err)
	require.True(t, done)
	require.Equal(t, 1, insts.insts[inst.ID].Status)
}

func TestMemberApplicationRejectStopsChain(t *testing.T) {
	minister, applicant := uuid.New(), uuid.New()
	tasks := newMockTaskRepo()
	e, insts := memberApplicationEngine(t, tasks)

	inst, err := e.Start(context.Background(), MemberApplicationDefinitionKey, uuid.New().String(), "member_application", applicant, nil)
	require.NoError(t, err)
	nodeID, _, err := e.RunningApprovalNode(context.Background(), inst.ID)
	require.NoError(t, err)
	addApprovalTask(tasks, inst.ID, nodeID, minister)
	require.NoError(t, e.CompleteApplicationApproval(context.Background(), inst.ID, minister, ActionReject, "材料不符"))
	require.Equal(t, 2, insts.insts[inst.ID].Status)
}

func TestMemberApplicationRejectTerminatesOtherPendingTasks(t *testing.T) {
	first, second, applicant := uuid.New(), uuid.New(), uuid.New()
	tasks := newMockTaskRepo()
	e, insts := memberApplicationEngine(t, tasks)

	inst, err := e.Start(context.Background(), MemberApplicationDefinitionKey, uuid.New().String(), "member_application", applicant, nil)
	require.NoError(t, err)
	nodeID, _, err := e.RunningApprovalNode(context.Background(), inst.ID)
	require.NoError(t, err)
	addApprovalTask(tasks, inst.ID, nodeID, first)
	addApprovalTask(tasks, inst.ID, nodeID, second)
	require.NoError(t, e.CompleteApplicationApproval(context.Background(), inst.ID, first, ActionReject, "材料不符"))
	require.Equal(t, 2, insts.insts[inst.ID].Status)

	var canceled, rejected int
	for _, task := range tasks.tasks {
		if task.InstanceID != inst.ID {
			continue
		}
		if task.AssigneeID != nil && *task.AssigneeID == first {
			require.Equal(t, 2, task.Status)
			rejected++
		}
		if task.AssigneeID != nil && *task.AssigneeID == second {
			require.Equal(t, 5, task.Status)
			require.Equal(t, "cancel", task.Action)
			canceled++
		}
	}
	require.Equal(t, 1, rejected)
	require.Equal(t, 1, canceled)
	require.Error(t, e.CompleteApplicationApproval(context.Background(), inst.ID, second, ActionApprove, "不能救活"))
}

func TestStartSkipMinisterDoesNotEnterMinister(t *testing.T) {
	applicant := uuid.New()
	tasks := newMockTaskRepo()
	approval := &recordingApproval{}
	e, _ := memberApplicationEngineWith(t, tasks, approval)

	inst, err := e.Start(context.Background(), MemberApplicationDefinitionKey, uuid.New().String(), "member_application", applicant, map[string]interface{}{
		"applicant": applicant.String(), "apply_type": int16(1), SkipOfficerVariable: true, SkipMinisterVariable: true,
	})
	require.NoError(t, err)
	require.NotContains(t, approval.enteredIDs, "officer")
	require.NotContains(t, approval.enteredIDs, "minister")
	require.Equal(t, []string{"president"}, approval.enteredIDs)
	nodeID, done, err := e.RunningApprovalNode(context.Background(), inst.ID)
	require.NoError(t, err)
	require.False(t, done)
	require.Equal(t, "president", nodeID)
	require.Empty(t, tasks.tasks)
}

func TestSkipMinisterApprovalFlag(t *testing.T) {
	require.False(t, skipMinisterApproval(&FlowNode{ID: "minister"}, nil))
	require.False(t, skipMinisterApproval(&FlowNode{ID: "president"}, map[string]interface{}{SkipMinisterVariable: true}))
	require.True(t, skipMinisterApproval(&FlowNode{ID: "minister"}, map[string]interface{}{SkipMinisterVariable: true}))
	require.False(t, skipMinisterApproval(&FlowNode{ID: "minister"}, map[string]interface{}{}))
}

func TestSkipApplicationApprovalUsesSkipWhen(t *testing.T) {
	node := &FlowNode{ID: "extra", Type: "approval", Config: map[string]interface{}{"skipWhen": "skip_extra"}}
	require.False(t, skipApplicationApproval(node, nil))
	require.True(t, skipApplicationApproval(node, map[string]interface{}{"skip_extra": true}))
	require.True(t, skipApplicationApproval(&FlowNode{ID: "officer", Type: "approval"}, map[string]interface{}{SkipOfficerVariable: true}))
}

func TestStartSkipOfficerDoesNotEnterOfficer(t *testing.T) {
	applicant := uuid.New()
	tasks := newMockTaskRepo()
	approval := &recordingApproval{}
	e, _ := memberApplicationEngineWith(t, tasks, approval)

	inst, err := e.Start(context.Background(), MemberApplicationDefinitionKey, uuid.New().String(), "member_application", applicant, map[string]interface{}{
		"applicant": applicant.String(), "apply_type": int16(1), SkipOfficerVariable: true,
	})
	require.NoError(t, err)
	require.NotContains(t, approval.enteredIDs, "officer")
	require.Equal(t, []string{"minister"}, approval.enteredIDs)
	nodeID, done, err := e.RunningApprovalNode(context.Background(), inst.ID)
	require.NoError(t, err)
	require.False(t, done)
	require.Equal(t, "minister", nodeID)
}

func TestApplicationProgressMarksCurrentAndSkipped(t *testing.T) {
	applicant := uuid.New()
	tasks := newMockTaskRepo()
	e, _ := memberApplicationEngine(t, tasks)
	inst, err := e.Start(context.Background(), MemberApplicationDefinitionKey, uuid.New().String(), "member_application", applicant, map[string]interface{}{
		SkipOfficerVariable: true,
	})
	require.NoError(t, err)
	progress, err := e.ApplicationProgress(context.Background(), inst.ID)
	require.NoError(t, err)
	require.False(t, progress.Completed)
	require.Equal(t, []string{"minister"}, progress.CurrentNodeIDs)
	states := map[string]string{}
	for _, step := range progress.Steps {
		states[step.ID] = step.State
	}
	require.Equal(t, "done", states["start"])
	require.Equal(t, "skipped", states["officer"])
	require.Equal(t, "current", states["minister"])
	require.Equal(t, "pending", states["president"])
}

func TestIsEmptyAssigneeError(t *testing.T) {
	require.False(t, isEmptyAssigneeError(nil))
	require.True(t, isEmptyAssigneeError(response.NewAppError(response.CodeWorkflowInvalidNode, "审批节点没有处理人")))
	require.False(t, isEmptyAssigneeError(response.NewAppError(response.CodeWorkflowInvalidNode, "其它错误")))
}

func TestNodeAllowsTransferAndSkipIfEmpty(t *testing.T) {
	require.True(t, nodeAllowsTransfer(nil))
	require.True(t, nodeAllowsTransfer(&FlowNode{Config: map[string]interface{}{}}))
	require.False(t, nodeAllowsTransfer(&FlowNode{Config: map[string]interface{}{"allowTransfer": false}}))
	require.True(t, skipIfEmptyNode(&FlowNode{Config: map[string]interface{}{"skipIfEmpty": true}}))
	require.False(t, skipIfEmptyNode(&FlowNode{Config: map[string]interface{}{}}))
}

func TestTransferApplicationApprovalGuards(t *testing.T) {
	e, _ := memberApplicationEngine(t, newMockTaskRepo())
	err := e.TransferApplicationApproval(context.Background(), uuid.New(), uuid.New(), uuid.New(), "")
	require.Error(t, err)
	e.businessTransaction = false
	err = e.TransferApplicationApproval(context.Background(), uuid.New(), uuid.New(), uuid.New(), "")
	require.Error(t, err)
	e.businessTransaction = true
	from := uuid.New()
	require.Error(t, e.TransferApplicationApproval(context.Background(), uuid.New(), from, from, ""))
}

func TestSelectExistingApprovalTaskDoesNotMint(t *testing.T) {
	assignee, delegate, applicant := uuid.New(), uuid.New(), uuid.New()
	tasks := newMockTaskRepo()
	e, _ := memberApplicationEngine(t, tasks)
	inst, err := e.Start(context.Background(), MemberApplicationDefinitionKey, uuid.New().String(), "member_application", applicant, map[string]interface{}{
		"applicant": applicant.String(), "apply_type": int16(2),
	})
	require.NoError(t, err)
	nodeID, _, err := e.RunningApprovalNode(context.Background(), inst.ID)
	require.NoError(t, err)
	addApprovalTask(tasks, inst.ID, nodeID, assignee)
	created := len(tasks.created)

	own, err := e.selectExistingApprovalTask(context.Background(), inst.ID, nodeID, assignee)
	require.NoError(t, err)
	require.Equal(t, assignee, *own.AssigneeID)
	require.Equal(t, created, len(tasks.created))

	handed, err := e.selectExistingApprovalTask(context.Background(), inst.ID, nodeID, delegate)
	require.NoError(t, err)
	require.Equal(t, assignee, *handed.AssigneeID)
	require.Equal(t, own.ID, handed.ID)
	require.Equal(t, created, len(tasks.created))

	minted, err := e.selectApprovalTask(context.Background(), inst.ID, nodeID, delegate)
	require.NoError(t, err)
	require.Equal(t, delegate, *minted.AssigneeID)
	require.NotEqual(t, own.ID, minted.ID)
	require.Equal(t, created+1, len(tasks.created))
}

func TestApprovalRoleCodeFromGraph(t *testing.T) {
	applicant := uuid.New()
	e, _ := memberApplicationEngine(t, newMockTaskRepo())
	inst, err := e.Start(context.Background(), MemberApplicationDefinitionKey, uuid.New().String(), "member_application", applicant, map[string]interface{}{
		"applicant": applicant.String(), "apply_type": int16(2),
	})
	require.NoError(t, err)
	role, err := e.ApprovalRoleCode(context.Background(), inst.ID, "officer")
	require.NoError(t, err)
	require.Equal(t, "officer", role)
	role, err = e.ApprovalRoleCode(context.Background(), inst.ID, "president")
	require.NoError(t, err)
	require.Equal(t, "president", role)
	require.Equal(t, "hr", ResolveApprovalRole("custom", "hr"))
	require.Empty(t, ResolveApprovalRole("custom", ""))
	roles, err := e.ApplicationApprovalRoles(context.Background(), inst.ID)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"officer", "minister", "president"}, roles)
}

func TestIsLastApplicationApprovalDefaultSpine(t *testing.T) {
	applicant := uuid.New()
	e, _ := memberApplicationEngine(t, newMockTaskRepo())
	inst, err := e.Start(context.Background(), MemberApplicationDefinitionKey, uuid.New().String(), "member_application", applicant, map[string]interface{}{
		"applicant": applicant.String(), "apply_type": int16(2),
	})
	require.NoError(t, err)
	last, err := e.IsLastApplicationApproval(context.Background(), inst.ID, "officer")
	require.NoError(t, err)
	require.False(t, last)

	final, err := e.Start(context.Background(), MemberApplicationDefinitionKey, uuid.New().String(), "member_application", applicant, map[string]interface{}{
		SkipOfficerVariable: true, SkipMinisterVariable: true,
	})
	require.NoError(t, err)
	last, err = e.IsLastApplicationApproval(context.Background(), final.ID, "president")
	require.NoError(t, err)
	require.True(t, last)
}

func TestApplicationProgressCompleted(t *testing.T) {
	applicant := uuid.New()
	tasks := newMockTaskRepo()
	e, insts := memberApplicationEngine(t, tasks)
	inst, err := e.Start(context.Background(), MemberApplicationDefinitionKey, uuid.New().String(), "member_application", applicant, map[string]interface{}{
		SkipOfficerVariable: true, SkipMinisterVariable: true,
	})
	require.NoError(t, err)
	addApprovalTask(tasks, inst.ID, "president", applicant)
	require.NoError(t, e.CompleteApplicationApproval(context.Background(), inst.ID, applicant, ActionApprove, "终审"))
	require.Equal(t, 1, insts.insts[inst.ID].Status)
	progress, err := e.ApplicationProgress(context.Background(), inst.ID)
	require.NoError(t, err)
	require.True(t, progress.Completed)
	for _, step := range progress.Steps {
		if step.ID == "officer" || step.ID == "minister" {
			require.Equal(t, "skipped", step.State)
		}
	}
}

func TestValidateMemberApplicationRejectsUnknownType(t *testing.T) {
	graph, err := ParseGraph(MemberApplicationBPMN())
	require.NoError(t, err)
	graph.Nodes["timer"] = &FlowNode{ID: "timer", Type: "timer"}
	require.Error(t, validateMemberApplicationGraph(graph))
	require.Error(t, validateMemberApplicationGraph(nil))
}

func TestMemberApplicationReachesEnd(t *testing.T) {
	graph, err := ParseGraph(MemberApplicationBPMN())
	require.NoError(t, err)
	require.True(t, memberApplicationReachesEnd(graph, "start"))
	require.False(t, memberApplicationReachesEnd(graph, "missing"))
}
