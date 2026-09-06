package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/search/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/Yogdunana/StarByte/backend/pkg/search"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() { gin.SetMode(gin.TestMode) }

type stubSvc struct {
	res []dto.ResourceInfo
	out *search.Result
	err error
}

func (s *stubSvc) Resources() []dto.ResourceInfo { return s.res }
func (s *stubSvc) Lookup(code string) (search.Schema, bool) {
	if code == "gone" {
		return search.Schema{}, false
	}
	return search.Schema{Code: code, RBACResource: "task", ScopeColumn: "department_id"}, true
}
func (s *stubSvc) Query(context.Context, dto.QueryRequest, *rbacModel.DataScopeCondition, uuid.UUID) (*search.Result, error) {
	return s.out, s.err
}

func newHandler(svc *stubSvc) *SearchHandler {
	return NewSearchHandler(svc, nil, nil, nil)
}

func doJSON(h gin.HandlerFunc, body any) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/system/search/query", &buf)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("request_id", "rid")
	h(c)
	return w
}

func TestSearchHandler_Resources(t *testing.T) {
	h := newHandler(&stubSvc{res: []dto.ResourceInfo{{Code: "tasks", Name: "任务"}}})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/system/search/resources", nil)
	c.Set("request_id", "rid")
	h.Resources(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegisterRoutes(t *testing.T) {
	assert.NotPanics(t, func() {
		RegisterRoutes(gin.New().Group("/api/v1"), newHandler(&stubSvc{}), nil)
	})
}

func TestSearchHandler_Query(t *testing.T) {
	h := newHandler(&stubSvc{out: &search.Result{List: []map[string]any{{"title": "a"}}, Total: 1, Page: 1, PageSize: 20}})
	w := doJSON(h.Query, dto.QueryRequest{Resource: "tasks", Keyword: "招新"})
	assert.Equal(t, http.StatusOK, w.Code)

	w = doJSON(h.Query, "nope")
	assert.Equal(t, http.StatusBadRequest, w.Code)

	h = newHandler(&stubSvc{err: response.NewError(response.CodeSearchUnknownResource, "未知检索资源")})
	w = doJSON(h.Query, dto.QueryRequest{Resource: "gone"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "20001")
}
