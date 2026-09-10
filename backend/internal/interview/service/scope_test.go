package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
)

func TestRewriteInterviewScope_Self(t *testing.T) {
	uid := uuid.New()
	got := rewriteInterviewScope(&rbacModel.DataScopeCondition{Query: "1 = 0", IsSelf: true}, uid)
	require.Contains(t, got.Query, "applicant_id")
	require.Equal(t, uid, got.Args[0])
}

func TestCanAccessInterview_Self(t *testing.T) {
	uid := uuid.New()
	scope := &rbacModel.DataScopeCondition{Query: "1 = 0", IsSelf: true}
	require.True(t, canAccessInterview(scope, uid, nil, uid))
	require.False(t, canAccessInterview(scope, uuid.New(), nil, uid))
}
