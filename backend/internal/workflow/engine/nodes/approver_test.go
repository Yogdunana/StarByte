package nodes

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

type roleApprovers struct {
	code string
	dept *uuid.UUID
	ids  []uuid.UUID
}

func (r *roleApprovers) Search(context.Context, string, uuid.UUID) ([]model.ApproverOption, error) {
	return nil, nil
}
func (r *roleApprovers) ActiveUsers(context.Context, []uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (r *roleApprovers) ByRole(context.Context, uuid.UUID) ([]uuid.UUID, error) { return nil, nil }
func (r *roleApprovers) ByRoleCode(_ context.Context, code string, dept *uuid.UUID) ([]uuid.UUID, error) {
	r.code, r.dept = code, dept
	return r.ids, nil
}
func (r *roleApprovers) DepartmentLeaders(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func TestResolveRoleAssigneesUsesCodeAndDepartment(t *testing.T) {
	dept := uuid.New()
	holder := uuid.New()
	store := &roleApprovers{ids: []uuid.UUID{holder}}
	node := &ApprovalNode{Approvers: store}
	ids, err := node.resolveRuntime(context.Background(), map[string]interface{}{
		"assigneeStrategy": "role", "roleCode": "minister", "departmentScope": true,
	}, uuid.New(), map[string]interface{}{"department_id": dept.String()}, nil)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{holder}, ids)
	require.Equal(t, "minister", store.code)
	require.NotNil(t, store.dept)
	require.Equal(t, dept, *store.dept)
}

func TestScopedDepartmentNilWithoutDepartment(t *testing.T) {
	require.Nil(t, scopedDepartment(map[string]interface{}{"departmentScope": true}, map[string]interface{}{}))
	require.Nil(t, scopedDepartment(map[string]interface{}{"departmentScope": true}, map[string]interface{}{"department_id": ""}))
}

func TestValidateRoleAcceptsRoleCode(t *testing.T) {
	node := &ApprovalNode{}
	err := node.Validate(&engine.FlowNode{ID: "minister", Type: "approval", Config: map[string]interface{}{
		"assigneeStrategy": "role", "roleCode": "minister", "approvalType": "any",
	}})
	require.NoError(t, err)
	err = node.Validate(&engine.FlowNode{ID: "minister", Type: "approval", Config: map[string]interface{}{
		"assigneeStrategy": "role", "approvalType": "any",
	}})
	require.Error(t, err)
	_, ok := err.(*response.AppError)
	require.True(t, ok)
}
