package service

import (
	"testing"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRewriteScope_SelfBecomesUserID(t *testing.T) {
	uid := uuid.New()
	out := rewriteScope(&rbacModel.DataScopeCondition{Query: "1 = 0"}, "a", uid)
	assert.Equal(t, "a.user_id = ?", out.Query)
	assert.Equal(t, uid, out.Args[0])
}

func TestRewriteScope_DepartmentAlias(t *testing.T) {
	dept := uuid.New()
	out := rewriteScope(&rbacModel.DataScopeCondition{Query: "department_id = ?", Args: []interface{}{dept}}, "p", uuid.New())
	assert.Equal(t, "p.department_id = ?", out.Query)
}

func TestCanAccessRecord_Self(t *testing.T) {
	owner := uuid.New()
	viewer := owner
	assert.True(t, canAccessRecord(&rbacModel.DataScopeCondition{Query: "1 = 0"}, owner, nil, viewer))
	assert.False(t, canAccessRecord(&rbacModel.DataScopeCondition{Query: "1 = 0"}, uuid.New(), nil, viewer))
}

func TestCanAccessRecord_All(t *testing.T) {
	assert.True(t, canAccessRecord(nil, uuid.New(), nil, uuid.New()))
}
