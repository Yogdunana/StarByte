package engine

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
)

func timeoutEngine(insts *mockInstRepo, tasks *mockTaskRepo) *FlowEngine {
	return NewFlowEngine(&mockDefRepo{}, insts, tasks, newMockVarRepo(), nil, nil, NewExpressionEngine(), events.NewEventBus(), nil)
}

func TestEscalateOverdueApprovalHandsNodeToReviewer(t *testing.T) {
	instID, applicant, absent, president := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	tasks := newMockTaskRepo()
	stale := &model.FlowTask{ID: uuid.New(), InstanceID: instID, NodeID: "committee", NodeName: "常委会会签", TaskType: "approval", AssigneeID: &absent, Status: 0}
	tasks.tasks[stale.ID] = stale
	insts := &mockInstRepo{inst: &model.FlowInstance{ID: instID, BusinessType: "member_application", BusinessKey: "app-1", InitiatorID: applicant, Status: 0}}
	e := timeoutEngine(insts, tasks)

	n, err := e.EscalateOverdueApproval(context.Background(), OverdueApprovalTodo{
		InstanceID: instID, NodeID: "committee", NodeName: "常委会会签",
	}, []uuid.UUID{president}, uuid.Nil)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Len(t, tasks.created, 1)

	created := tasks.created[0]
	require.Equal(t, president, *created.AssigneeID)
	require.Equal(t, "committee", created.NodeID)
	require.NotNil(t, created.ActivationID)
	require.Nil(t, created.DueDate, "a fresh due date would escalate the same stall forever")
	require.Equal(t, 5, tasks.tasks[stale.ID].Status, "the unanswered task is closed first")
	require.NotEqual(t, stale.ActivationID, created.ActivationID, "countersign quorum must be recounted")
}

func TestEscalateOverdueApprovalSkipsDecidedNodes(t *testing.T) {
	instID, approver, next := uuid.New(), uuid.New(), uuid.New()
	tasks := newMockTaskRepo()
	decided := &model.FlowTask{ID: uuid.New(), InstanceID: instID, NodeID: "committee", TaskType: "approval", AssigneeID: &approver, Status: 1}
	tasks.tasks[decided.ID] = decided
	insts := &mockInstRepo{inst: &model.FlowInstance{ID: instID, BusinessType: "member_application", Status: 0}}
	e := timeoutEngine(insts, tasks)

	n, err := e.EscalateOverdueApproval(context.Background(), OverdueApprovalTodo{InstanceID: instID, NodeID: "committee"}, []uuid.UUID{next}, uuid.Nil)
	require.NoError(t, err)
	require.Equal(t, 0, n)
	require.Empty(t, tasks.created, "a decided quorum still resolves on its own")
}

func TestEscalateOverdueApprovalGuards(t *testing.T) {
	instID, assignee := uuid.New(), uuid.New()
	tasks := newMockTaskRepo()
	pending := &model.FlowTask{ID: uuid.New(), InstanceID: instID, NodeID: "committee", TaskType: "approval", AssigneeID: &assignee, Status: 0}
	tasks.tasks[pending.ID] = pending

	t.Run("no reviewer means nothing changes", func(t *testing.T) {
		insts := &mockInstRepo{inst: &model.FlowInstance{ID: instID, BusinessType: "member_application", Status: 0}}
		e := timeoutEngine(insts, tasks)
		n, err := e.EscalateOverdueApproval(context.Background(), OverdueApprovalTodo{InstanceID: instID, NodeID: "committee"}, nil, uuid.Nil)
		require.NoError(t, err)
		require.Equal(t, 0, n)
		require.Empty(t, tasks.created)
	})

	t.Run("finished instance is untouched", func(t *testing.T) {
		insts := &mockInstRepo{inst: &model.FlowInstance{ID: instID, BusinessType: "member_application", Status: 1}}
		e := timeoutEngine(insts, tasks)
		n, err := e.EscalateOverdueApproval(context.Background(), OverdueApprovalTodo{InstanceID: instID, NodeID: "committee"}, []uuid.UUID{uuid.New()}, uuid.Nil)
		require.NoError(t, err)
		require.Equal(t, 0, n)
		require.Empty(t, tasks.created)
	})

	t.Run("unknown instance errors", func(t *testing.T) {
		e := timeoutEngine(&mockInstRepo{}, tasks)
		_, err := e.EscalateOverdueApproval(context.Background(), OverdueApprovalTodo{InstanceID: uuid.New(), NodeID: "committee"}, nil, uuid.Nil)
		require.Error(t, err)
	})
}

// ListOverdueApprovalTodos must not be limited to collaboration tasks, which is
// why membership dues were never picked up before.
func TestListOverdueApprovalTodosWithoutDBIsNoop(t *testing.T) {
	e := timeoutEngine(&mockInstRepo{}, newMockTaskRepo())
	out, err := e.ListOverdueApprovalTodos(context.Background(), time.Now(), "member_application")
	require.NoError(t, err)
	require.Empty(t, out)
}
