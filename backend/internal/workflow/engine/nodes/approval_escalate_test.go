package nodes

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	wfmodel "github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
)

type stubApproverRepo struct {
	holders map[string][]uuid.UUID
	calls   []string
}

func (s *stubApproverRepo) Search(context.Context, string, uuid.UUID) ([]model.ApproverOption, error) {
	return nil, nil
}
func (s *stubApproverRepo) ActiveUsers(_ context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	return ids, nil
}
func (s *stubApproverRepo) ByRole(context.Context, uuid.UUID) ([]uuid.UUID, error) { return nil, nil }
func (s *stubApproverRepo) ByRoleCode(_ context.Context, code string, _ *uuid.UUID) ([]uuid.UUID, error) {
	s.calls = append(s.calls, code)
	return s.holders[code], nil
}
func (s *stubApproverRepo) DepartmentLeaders(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func committeeNode(extra map[string]interface{}) *engine.FlowNode {
	config := map[string]interface{}{
		"assigneeStrategy": "role", "roleCode": "standing_committee", "approvalType": "all", "dueDays": float64(1),
	}
	for k, v := range extra {
		config[k] = v
	}
	return &engine.FlowNode{ID: "committee", Type: "approval", Label: "常委会会签", Config: config}
}

func TestApprovalNodeEscalatesUnstaffedCountersignNode(t *testing.T) {
	president := uuid.New()
	approvers := &stubApproverRepo{holders: map[string][]uuid.UUID{"standing_committee": {}, "president": {president}}}
	taskRepo := &mockTaskRepo{}
	n := &ApprovalNode{TaskRepo: taskRepo, EventBus: events.NewEventBus(), Approvers: approvers}
	inst := &wfmodel.FlowInstance{ID: uuid.New(), InitiatorID: uuid.New(), BusinessType: "member_application"}

	require.NoError(t, n.OnEnter(context.Background(), inst, committeeNode(nil), map[string]interface{}{}))
	require.Equal(t, []string{"standing_committee", "president"}, approvers.calls)
	require.NotNil(t, taskRepo.createdTask)
	require.Equal(t, president, *taskRepo.createdTask.AssigneeID)
	require.NotNil(t, taskRepo.createdTask.DueDate, "escalated task keeps the node due date")
}

func TestApprovalNodeEscalatesWhenDepartmentScopeMissing(t *testing.T) {
	centerDirector := uuid.New()
	approvers := &stubApproverRepo{holders: map[string][]uuid.UUID{"minister": {}, "center_director": {centerDirector}}}
	taskRepo := &mockTaskRepo{}
	n := &ApprovalNode{TaskRepo: taskRepo, EventBus: events.NewEventBus(), Approvers: approvers}
	inst := &wfmodel.FlowInstance{ID: uuid.New(), InitiatorID: uuid.New(), BusinessType: "member_application"}
	node := &engine.FlowNode{ID: "department_review", Type: "approval", Label: "部门初审", Config: map[string]interface{}{
		"assigneeStrategy": "role", "roleCode": "minister", "approvalType": "any",
		"departmentScope": true, "requireDepartment": true, "fallbackRoleCode": "center_director",
	}}

	require.NoError(t, n.OnEnter(context.Background(), inst, node, map[string]interface{}{}))
	require.Equal(t, []string{"center_director"}, approvers.calls)
	require.NotNil(t, taskRepo.createdTask)
	require.Equal(t, centerDirector, *taskRepo.createdTask.AssigneeID)
}

func TestApprovalNodeWithoutFallbackStillReportsEmpty(t *testing.T) {
	approvers := &stubApproverRepo{holders: map[string][]uuid.UUID{"president": {}}}
	taskRepo := &mockTaskRepo{}
	n := &ApprovalNode{TaskRepo: taskRepo, EventBus: events.NewEventBus(), Approvers: approvers}
	inst := &wfmodel.FlowInstance{ID: uuid.New(), InitiatorID: uuid.New(), BusinessType: "member_application"}

	err := n.OnEnter(context.Background(), inst, committeeNode(nil), map[string]interface{}{})
	require.Error(t, err)
	// Nothing escalated the skip decision: nothing stayed behind either, so the
	// engine can still apply the node policy.
	require.True(t, engine.IsEmptyAssignee(err))
	require.Nil(t, taskRepo.createdTask)
}

func TestApprovalNodeFailPolicyDoesNotEscalate(t *testing.T) {
	approvers := &stubApproverRepo{holders: map[string][]uuid.UUID{"president": {uuid.New()}}}
	taskRepo := &mockTaskRepo{}
	n := &ApprovalNode{TaskRepo: taskRepo, EventBus: events.NewEventBus(), Approvers: approvers}
	inst := &wfmodel.FlowInstance{ID: uuid.New(), InitiatorID: uuid.New(), BusinessType: "member_application"}

	err := n.OnEnter(context.Background(), inst, committeeNode(map[string]interface{}{"onEmpty": "fail"}), map[string]interface{}{})
	require.Error(t, err)
	require.Equal(t, []string{"standing_committee"}, approvers.calls, "onEmpty=fail must not reach for a fallback role")
	require.Nil(t, taskRepo.createdTask)
}

func TestApprovalNodeNonRoleStrategyKeepsFailing(t *testing.T) {
	approvers := &stubApproverRepo{holders: map[string][]uuid.UUID{"president": {uuid.New()}}}
	taskRepo := &mockTaskRepo{}
	n := &ApprovalNode{TaskRepo: taskRepo, EventBus: events.NewEventBus(), Approvers: approvers}
	inst := &wfmodel.FlowInstance{ID: uuid.New(), InitiatorID: uuid.New(), BusinessType: "collaboration_task"}
	node := &engine.FlowNode{ID: "review", Type: "approval", Label: "验收", Config: map[string]interface{}{
		"assigneeStrategy": "business_role", "businessType": engine.TaskBusinessType, "taskStage": "review",
	}}

	err := n.OnEnter(context.Background(), inst, node, map[string]interface{}{})
	require.Error(t, err)
	require.False(t, engine.IsEmptyAssignee(err))
	require.Nil(t, taskRepo.createdTask)
}

func TestApprovalNodeSkipPolicyWithoutFallbackStaysSilent(t *testing.T) {
	approvers := &stubApproverRepo{holders: map[string][]uuid.UUID{"president": {uuid.New()}}}
	taskRepo := &mockTaskRepo{}
	n := &ApprovalNode{TaskRepo: taskRepo, EventBus: events.NewEventBus(), Approvers: approvers}
	inst := &wfmodel.FlowInstance{ID: uuid.New(), InitiatorID: uuid.New(), BusinessType: "member_application"}
	node := &engine.FlowNode{ID: "department_review", Type: "approval", Label: "部门初审", Config: map[string]interface{}{
		"assigneeStrategy": "role", "roleCode": "minister", "approvalType": "any", "departmentScope": true,
	}}

	err := n.OnEnter(context.Background(), inst, node, map[string]interface{}{})
	require.Error(t, err)
	require.True(t, engine.IsEmptyAssignee(err))
	require.Equal(t, []string{"minister"}, approvers.calls, "no fallbackRoleCode means the node keeps skipping quietly")
	require.Nil(t, taskRepo.createdTask)
}

var _ repo.ApproverRepo = (*stubApproverRepo)(nil)
