package service

import (
	"context"
	"testing"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/search/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/Yogdunana/StarByte/backend/pkg/search"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchService_Resources(t *testing.T) {
	svc := NewSearchService(nil)
	list := svc.Resources()
	require.GreaterOrEqual(t, len(list), 4)
	codes := map[string]bool{}
	for _, r := range list {
		codes[r.Code] = true
		assert.NotEmpty(t, r.Name)
		assert.NotEmpty(t, r.Fields)
	}
	assert.True(t, codes["users"])
	assert.True(t, codes["tasks"])
	assert.True(t, codes["audit_logs"])
	assert.True(t, codes["member_applications"])
}

func TestSearchService_UnknownResourceAndKeyword(t *testing.T) {
	svc := NewSearchService(nil)
	_, err := svc.Query(context.Background(), dto.QueryRequest{Resource: "nope"}, nil, uuid.Nil)
	require.Error(t, err)
	var app *response.AppError
	require.ErrorAs(t, err, &app)
	assert.Equal(t, response.CodeSearchUnknownResource, app.Code)

	long := make([]rune, 201)
	for i := range long {
		long[i] = '招'
	}
	_, err = svc.Query(context.Background(), dto.QueryRequest{Resource: "tasks", Keyword: string(long)}, nil, uuid.Nil)
	require.ErrorAs(t, err, &app)
	assert.Equal(t, response.CodeSearchInvalidQuery, app.Code)
}

func TestWrapSearchErr(t *testing.T) {
	cases := []struct {
		in   error
		code int
	}{
		{search.ErrUnknownField, response.CodeSearchUnknownField},
		{search.ErrInvalidOp, response.CodeSearchInvalidOp},
		{search.ErrInvalidCursor, response.CodeSearchInvalidCursor},
		{search.ErrInvalidAgg, response.CodeSearchInvalidAgg},
		{search.ErrDeepPagination, response.CodeSearchDeepPage},
		{search.ErrInvalidQuery, response.CodeSearchInvalidQuery},
		{search.ErrInvalidIdent, response.CodeSearchInvalidQuery},
		{assert.AnError, response.CodeSearchInvalidQuery},
	}
	for _, tc := range cases {
		err := wrapSearchErr(tc.in)
		var app *response.AppError
		require.ErrorAs(t, err, &app)
		assert.Equal(t, tc.code, app.Code)
	}
}

func TestCatalogFTSTrusted(t *testing.T) {
	for _, sch := range catalogs() {
		_, err := search.Compile(sch, search.Query{Page: 1, PageSize: 5, Keyword: "招新"})
		require.NoError(t, err, sch.Code)
	}
}

func TestApplyScope_departmentAndAudit(t *testing.T) {
	svc := NewSearchService(nil)
	tasks, ok := svc.Lookup("tasks")
	require.True(t, ok)
	assert.Equal(t, "task", tasks.RBACResource)

	deptID := uuid.New()
	viewer := uuid.New()
	scoped := applyScope(tasks, &rbacModel.DataScopeCondition{
		Query: "department_id = ?",
		Args:  []interface{}{deptID},
	}, viewer)
	assert.Contains(t, scoped.ExtraWhere, `t."department_id" = ?`)
	require.Len(t, scoped.ExtraArgs, 1)

	audit, ok := svc.Lookup("audit_logs")
	require.True(t, ok)
	assert.Equal(t, "audit", audit.RBACResource)
	closed := applyScope(audit, &rbacModel.DataScopeCondition{
		Query: "department_id = ?",
		Args:  []interface{}{deptID},
	}, viewer)
	assert.Equal(t, "1 = 0", closed.ExtraWhere)

	open := applyScope(tasks, &rbacModel.DataScopeCondition{}, viewer)
	assert.Equal(t, tasks.ExtraWhere, open.ExtraWhere)
}
