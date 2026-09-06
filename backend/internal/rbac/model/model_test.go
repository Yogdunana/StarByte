package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataScopeConditionIsEmpty(t *testing.T) {
	var nilCond *DataScopeCondition
	assert.True(t, nilCond.IsEmpty())
	assert.True(t, (&DataScopeCondition{}).IsEmpty())
	assert.False(t, (&DataScopeCondition{Query: "dept_id = ?"}).IsEmpty())
}

func TestDataScopeConstants(t *testing.T) {
	assert.Equal(t, "all", DataScopeAll)
	assert.Equal(t, "department", DataScopeDepartment)
	assert.Equal(t, "department_and_sub", DataScopeDepartmentAndSub)
	assert.Equal(t, "self", DataScopeSelf)
	assert.Equal(t, "custom", DataScopeCustom)
}

func TestPermissionTypeParseAndString(t *testing.T) {
	assert.Equal(t, "menu", PermissionTypeMenu.String())
	assert.Equal(t, "button", PermissionTypeButton.String())
	assert.Equal(t, "api", PermissionTypeAPI.String())
	assert.Equal(t, "menu", PermissionType(0).String())

	assert.Equal(t, PermissionTypeMenu, ParsePermissionType("menu"))
	assert.Equal(t, PermissionTypeButton, ParsePermissionType("button"))
	assert.Equal(t, PermissionTypeAPI, ParsePermissionType("api"))
	assert.Equal(t, PermissionTypeMenu, ParsePermissionType("other"))
}

func TestTableNames(t *testing.T) {
	assert.Equal(t, "permissions", Permission{}.TableName())
	assert.Equal(t, "roles", Role{}.TableName())
	assert.Equal(t, "role_permissions", RolePermission{}.TableName())
	assert.Equal(t, "user_roles", UserRole{}.TableName())
	assert.Equal(t, "role_data_scopes", RoleDataScope{}.TableName())
	assert.Equal(t, "departments", Department{}.TableName())
	assert.Equal(t, "positions", Position{}.TableName())
}

func TestStatusConstants(t *testing.T) {
	assert.Equal(t, 0, PermissionStatusEnabled)
	assert.Equal(t, 1, PermissionStatusDisabled)
	assert.Equal(t, 0, RoleStatusEnabled)
	assert.Equal(t, 1, RoleStatusDisabled)
	assert.Equal(t, 0, DepartmentStatusEnabled)
	assert.Equal(t, 1, DepartmentStatusDisabled)
	assert.Equal(t, 0, PositionStatusEnabled)
	assert.Equal(t, 1, PositionStatusDisabled)
}
