package service

import (
	"testing"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRewriteScopeSelfVsFailClosed(t *testing.T) {
	viewer := uuid.New()
	owner := viewer
	other := uuid.New()
	dept := uuid.New()

	self := rewriteScope(&rbacModel.DataScopeCondition{Query: "1 = 0", IsSelf: true}, viewer)
	assert.True(t, self.IsSelf)
	assert.Equal(t, "r.user_id = ?", self.Query)
	assert.True(t, canAccess(&rbacModel.DataScopeCondition{Query: "1 = 0", IsSelf: true}, owner, &dept, viewer))
	assert.False(t, canAccess(&rbacModel.DataScopeCondition{Query: "1 = 0", IsSelf: true}, other, &dept, viewer))

	deny := &rbacModel.DataScopeCondition{Query: "1 = 0"}
	kept := rewriteScope(deny, viewer)
	assert.False(t, kept.IsSelf)
	assert.Equal(t, "1 = 0", kept.Query)
	assert.False(t, canAccess(deny, owner, &dept, viewer))
}

func TestCanAccessDepartment(t *testing.T) {
	viewer := uuid.New()
	owner := uuid.New()
	dept := uuid.New()
	scope := &rbacModel.DataScopeCondition{Query: "department_id IN ?", Args: []interface{}{[]uuid.UUID{dept}}}
	assert.True(t, canAccess(scope, owner, &dept, viewer))
	other := uuid.New()
	assert.False(t, canAccess(scope, owner, &other, viewer))
}
