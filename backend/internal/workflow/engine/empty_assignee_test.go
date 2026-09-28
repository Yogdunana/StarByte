package engine

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func TestIsEmptyAssignee(t *testing.T) {
	require.False(t, IsEmptyAssignee(nil))
	require.False(t, IsEmptyAssignee(errors.New("boom")))
	// A plain AppError is a real failure, not a vacancy.
	require.False(t, IsEmptyAssignee(response.NewAppError(response.CodeWorkflowInvalidNode, "审批节点没有处理人")))

	empty := NewEmptyAssignee("committee", "standing_committee", "no_role_holder", "审批节点没有处理人（常委会会签）")
	require.True(t, IsEmptyAssignee(empty))
	require.True(t, IsEmptyAssignee(errors.Join(errors.New("wrap"), empty)))

	// Existing AppError consumers keep working through Unwrap.
	var appErr *response.AppError
	require.True(t, errors.As(empty, &appErr))
	require.Equal(t, response.CodeWorkflowInvalidNode, appErr.Code)
	require.Equal(t, "审批节点没有处理人（常委会会签）", appErr.Message)

	detail := EmptyAssigneeOf(errors.Join(errors.New("wrap"), empty))
	require.NotNil(t, detail)
	require.Equal(t, "committee", detail.NodeID)
	require.Equal(t, "standing_committee", detail.Role)
	require.Equal(t, "no_role_holder", detail.Reason)
	require.Nil(t, EmptyAssigneeOf(nil))
}

func TestEmptyPolicyOf(t *testing.T) {
	// Role nodes that used to be skipped keep skipping.
	require.Equal(t, EmptySkip, EmptyPolicyOf(&FlowNode{ID: "department_review", Config: map[string]interface{}{"assigneeStrategy": "role", "roleCode": "minister"}}))
	require.Equal(t, EmptySkip, EmptyPolicyOf(&FlowNode{ID: "center_review", Config: map[string]interface{}{"assigneeStrategy": "role", "roleCode": "center_director"}}))
	require.Equal(t, EmptySkip, EmptyPolicyOf(&FlowNode{ID: "officer", Config: map[string]interface{}{"assigneeStrategy": "role", "roleCode": "officer"}}))
	require.Equal(t, EmptySkip, EmptyPolicyOf(&FlowNode{ID: "minister", Config: map[string]interface{}{"assigneeStrategy": "role", "roleCode": "minister"}}))

	// A countersign node with no luck used to tear the approval down.
	require.Equal(t, EmptyEscalate, EmptyPolicyOf(&FlowNode{ID: "committee", Config: map[string]interface{}{"assigneeStrategy": "role", "roleCode": "standing_committee"}}))
	require.Equal(t, EmptyEscalate, EmptyPolicyOf(&FlowNode{ID: "custom"}))

	// Explicit opt-out wins for role nodes.
	require.Equal(t, EmptyFail, EmptyPolicyOf(&FlowNode{ID: "department_review", Config: map[string]interface{}{"assigneeStrategy": "role", "skipIfEmpty": false}}))
	require.Equal(t, EmptyFail, EmptyPolicyOf(&FlowNode{ID: "committee", Config: map[string]interface{}{"assigneeStrategy": "role", "onEmpty": "fail"}}))
	require.Equal(t, EmptySkip, EmptyPolicyOf(&FlowNode{ID: "committee", Config: map[string]interface{}{"assigneeStrategy": "role", "onEmpty": "SKIP"}}))

	// Non-role strategies are configuration errors, so they keep failing loudly.
	require.Equal(t, EmptyFail, EmptyPolicyOf(&FlowNode{ID: "review", Config: map[string]interface{}{"assigneeStrategy": "business_role"}}))
	require.Equal(t, EmptyFail, EmptyPolicyOf(&FlowNode{ID: "review", Config: map[string]interface{}{"assigneeStrategy": "static"}}))
	require.Equal(t, EmptyFail, EmptyPolicyOf(&FlowNode{ID: "review", Config: map[string]interface{}{"assigneeStrategy": "dept_leader"}}))
}

func TestFallbackRoleCode(t *testing.T) {
	require.Equal(t, DefaultFallbackRoleCode, FallbackRoleCode(nil))
	require.Equal(t, DefaultFallbackRoleCode, FallbackRoleCode(&FlowNode{ID: "committee"}))
	require.Equal(t, "center_director", FallbackRoleCode(&FlowNode{ID: "department_review", Config: map[string]interface{}{"fallbackRoleCode": " center_director "}}))
}
