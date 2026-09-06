package rbac

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorIsAndCode(t *testing.T) {
	err := NewRoleNotFoundError()
	assert.Equal(t, ErrCodeRoleNotFound, err.Code())
	assert.Equal(t, "角色不存在", err.Message())
	assert.Equal(t, "角色不存在", err.Error())
	assert.True(t, errors.Is(err, NewRoleNotFoundError()))
	assert.False(t, errors.Is(err, NewPermissionNotFoundError()))
	assert.False(t, errors.Is(err, errors.New("other")))
}

func TestErrorConstructors(t *testing.T) {
	cases := []struct {
		err  *Error
		code int
	}{
		{NewRoleCodeExistsError("admin"), ErrCodeRoleCodeExists},
		{NewRoleInUseError(), ErrCodeRoleInUse},
		{NewSystemRoleNoDeleteError(), ErrCodeSystemRoleNoDelete},
		{NewSystemRoleNoEditError(), ErrCodeSystemRoleNoEdit},
		{NewPermissionNotFoundError(), ErrCodePermissionNotFound},
		{NewPermissionCodeExistsError("user:read"), ErrCodePermissionCodeExists},
		{NewSystemPermissionNoDeleteError(), ErrCodeSystemPermissionNoDelete},
		{NewPermissionHasChildrenError(), ErrCodePermissionHasChildren},
		{NewSystemPermissionNoEditError(), ErrCodeSystemPermissionNoEdit},
		{NewPermissionDisabledError("x"), ErrCodePermissionDisabled},
		{NewDeptNotFoundError(), ErrCodeDeptNotFound},
		{NewDeptCodeExistsError("tech"), ErrCodeDeptCodeExists},
		{NewDeptHasChildrenError(), ErrCodeDeptHasChildren},
		{NewDeptInUseError(), ErrCodeDeptInUse},
		{NewPositionNotFoundError(), ErrCodePositionNotFound},
		{NewPositionCodeExistsError("intern"), ErrCodePositionCodeExists},
		{NewPositionInUseError(), ErrCodePositionInUse},
		{NewInvalidDataScopeError("nope"), ErrCodeInvalidDataScope},
		{NewPermissionDeniedError(), ErrCodePermissionDenied},
	}
	for _, tc := range cases {
		require.Equal(t, tc.code, tc.err.Code())
		require.NotEmpty(t, tc.err.Message())
	}
}
