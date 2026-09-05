package service

import (
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/internship/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRewriteScopeSelf(t *testing.T) {
	uid := uuid.New()
	got := rewriteScope(&rbacModel.DataScopeCondition{Query: "1 = 0"}, uid)
	require.Equal(t, "i.user_id = ?", got.Query)
	require.Equal(t, uid, got.Args[0])
}

func TestRewriteScopeDepartment(t *testing.T) {
	dept := uuid.New()
	got := rewriteScope(&rbacModel.DataScopeCondition{Query: "department_id = ?", Args: []interface{}{dept}}, uuid.New())
	require.Equal(t, "i.department_id = ?", got.Query)
	require.Equal(t, dept, got.Args[0])
}

func TestCanAccessRecord(t *testing.T) {
	owner := uuid.New()
	viewer := uuid.New()
	dept := uuid.New()
	require.True(t, canAccessRecord(&rbacModel.DataScopeCondition{Query: "1 = 0"}, owner, &dept, owner))
	require.False(t, canAccessRecord(&rbacModel.DataScopeCondition{Query: "1 = 0"}, owner, &dept, viewer))
	require.True(t, canAccessRecord(&rbacModel.DataScopeCondition{Query: "department_id = ?", Args: []interface{}{dept}}, owner, &dept, viewer))
	require.False(t, canAccessRecord(&rbacModel.DataScopeCondition{Query: "department_id = ?", Args: []interface{}{uuid.New()}}, owner, &dept, viewer))
	require.True(t, canAccessRecord(nil, owner, &dept, viewer))
}

func TestCanEditAndComplete(t *testing.T) {
	owner := uuid.New()
	other := uuid.New()
	dept := uuid.New()
	row := &model.Internship{UserID: owner, DepartmentID: &dept}
	cfgOn := model.InternshipConfig{AllowStudentEdit: true, AllowMinisterEdit: true}
	cfgOff := model.InternshipConfig{}
	self := &rbacModel.DataScopeCondition{Query: "1 = 0"}
	deptScope := &rbacModel.DataScopeCondition{Query: "department_id = ?", Args: []interface{}{dept}}

	require.True(t, canEdit(row, owner, self, cfgOn))
	require.False(t, canEdit(row, owner, self, cfgOff))
	require.True(t, canEdit(row, other, nil, cfgOff))
	require.True(t, canEdit(row, other, deptScope, cfgOn))
	require.False(t, canEdit(row, other, deptScope, cfgOff))
	require.True(t, canCompleteOrDelete(row, owner, self))
	require.False(t, canCompleteOrDelete(row, other, deptScope))
	require.True(t, canCompleteOrDelete(row, other, nil))
}
