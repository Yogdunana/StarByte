package service

import (
	"testing"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRewriteInterviewScope_Self(t *testing.T) {
	uid := uuid.New()
	got := rewriteInterviewScope(&rbacModel.DataScopeCondition{Query: "1 = 0"}, uid)
	require.Contains(t, got.Query, "applicant_id")
	require.Equal(t, uid, got.Args[0])
}

func TestCanAccessInterview_Self(t *testing.T) {
	uid := uuid.New()
	scope := &rbacModel.DataScopeCondition{Query: "1 = 0"}
	require.True(t, canAccessInterview(scope, uid, nil, uid))
	require.False(t, canAccessInterview(scope, uuid.New(), nil, uid))
}
