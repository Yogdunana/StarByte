package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
)

func TestRewriteScope_SelfBecomesUserID(t *testing.T) {
	uid := uuid.New()
	out := rewriteScope(&rbacModel.DataScopeCondition{Query: "1 = 0", IsSelf: true}, "a", uid)
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
	assert.True(t, canAccessRecord(&rbacModel.DataScopeCondition{Query: "1 = 0", IsSelf: true}, owner, nil, viewer))
	assert.False(t, canAccessRecord(&rbacModel.DataScopeCondition{Query: "1 = 0", IsSelf: true}, uuid.New(), nil, viewer))
}

func TestCanAccessRecord_All(t *testing.T) {
	assert.True(t, canAccessRecord(&rbacModel.DataScopeCondition{}, uuid.New(), nil, uuid.New()))
}

func TestScopeFailsClosed(t *testing.T) {
	owner, dept := uuid.New(), uuid.New()
	for _, scope := range []*rbacModel.DataScopeCondition{nil, {Query: "1 = 0"}, {Query: "unrecognized IN ?", Args: []interface{}{[]uuid.UUID{dept}}}} {
		assert.False(t, canAccessRecord(scope, owner, &dept, owner))
	}
	assert.Equal(t, "1 = 0", rewriteScope(nil, "a", owner).Query)
	assert.Equal(t, "1 = 0", rewriteScope(&rbacModel.DataScopeCondition{Query: "1 = 0"}, "a", owner).Query)
}
