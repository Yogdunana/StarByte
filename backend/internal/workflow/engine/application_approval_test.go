package engine

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
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

func memberApplicationEngine(t *testing.T, tasks *mockTaskRepo) (*FlowEngine, *storingInstRepo) {
	t.Helper()
	defID, verID := uuid.New(), uuid.New()
	def := &model.FlowDefinition{ID: defID, Key: MemberApplicationDefinitionKey, Status: 1}
	ver := &model.FlowDefinitionVersion{ID: verID, DefinitionID: defID, BpmnData: MemberApplicationBPMN(), Status: 1}
	insts := &storingInstRepo{insts: map[uuid.UUID]*model.FlowInstance{}}
	approval := &waitingTestNode{}
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

func TestMemberApplicationStartMinisterPresidentComplete(t *testing.T) {
	minister, president, applicant := uuid.New(), uuid.New(), uuid.New()
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
	addApprovalTask(tasks, inst.ID, "minister", minister)
	require.NoError(t, e.CompleteApplicationApproval(context.Background(), inst.ID, minister, ActionReject, "材料不符"))
	require.Equal(t, 2, insts.insts[inst.ID].Status)
}
