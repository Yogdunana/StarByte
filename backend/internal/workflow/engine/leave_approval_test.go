package engine

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
)

func leaveApprovalEngine(t *testing.T, tasks *mockTaskRepo) (*FlowEngine, *storingInstRepo) {
	t.Helper()
	defID, verID := uuid.New(), uuid.New()
	def := &model.FlowDefinition{ID: defID, Key: LeaveDefinitionKey, Status: 1}
	ver := &model.FlowDefinitionVersion{ID: verID, DefinitionID: defID, BpmnData: LeaveApprovalBPMN(), Status: 1}
	insts := &storingInstRepo{insts: map[uuid.UUID]*model.FlowInstance{}}
	e := NewFlowEngine(&mockDefRepo{def: def, version: ver}, insts, tasks, newMockVarRepo(), nil, &mockRegistryForTest{handlers: map[string]NodeHandler{
		"start": stubStartHandler{}, "end": stubEndHandler{}, "approval": &waitingTestNode{},
	}}, NewExpressionEngine(), events.NewEventBus(), nil)
	e.businessTransaction = true
	return e, insts
}

func TestLeaveApprovalStartMinisterPresidentComplete(t *testing.T) {
	minister, president, applicant := uuid.New(), uuid.New(), uuid.New()
	tasks := newMockTaskRepo()
	e, _ := leaveApprovalEngine(t, tasks)

	inst, err := e.Start(context.Background(), LeaveDefinitionKey, uuid.New().String(), LeaveBusinessType, applicant, map[string]interface{}{
		"applicant": applicant.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, inst)

	nodeID, done, err := e.RunningApprovalNode(context.Background(), inst.ID)
	require.NoError(t, err)
	require.False(t, done)
	require.Equal(t, "minister", nodeID)

	addApprovalTask(tasks, inst.ID, "minister", minister)
	require.NoError(t, e.CompleteLeaveApproval(context.Background(), inst.ID, minister, ActionApprove, "dept ok"))

	nodeID, done, err = e.RunningApprovalNode(context.Background(), inst.ID)
	require.NoError(t, err)
	require.False(t, done)
	require.Equal(t, "president", nodeID)

	addApprovalTask(tasks, inst.ID, "president", president)
	require.NoError(t, e.CompleteLeaveApproval(context.Background(), inst.ID, president, ActionApprove, "org ok"))

	_, done, err = e.RunningApprovalNode(context.Background(), inst.ID)
	require.NoError(t, err)
	require.True(t, done)
}

func TestLeaveApprovalRejectTerminates(t *testing.T) {
	minister, applicant := uuid.New(), uuid.New()
	tasks := newMockTaskRepo()
	e, insts := leaveApprovalEngine(t, tasks)

	inst, err := e.Start(context.Background(), LeaveDefinitionKey, uuid.New().String(), LeaveBusinessType, applicant, nil)
	require.NoError(t, err)
	addApprovalTask(tasks, inst.ID, "minister", minister)
	require.NoError(t, e.CompleteLeaveApproval(context.Background(), inst.ID, minister, ActionReject, "no"))

	fresh, err := insts.GetByID(context.Background(), inst.ID)
	require.NoError(t, err)
	require.NotEqual(t, 0, fresh.Status)
}

func TestLeaveApprovalRequiresBusinessTransaction(t *testing.T) {
	e, _ := leaveApprovalEngine(t, newMockTaskRepo())
	e.businessTransaction = false
	err := e.CompleteLeaveApproval(context.Background(), uuid.New(), uuid.New(), ActionApprove, "x")
	require.Error(t, err)
}
