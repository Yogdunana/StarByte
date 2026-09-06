package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/stats/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() { gin.SetMode(gin.TestMode) }

type stubSvc struct {
	providers []dto.ProviderInfo
	result    *dto.StatsResult
	overview  *dto.OverviewResponse
	bytes     []byte
	filename  string
	err       error
}

func (s *stubSvc) ListProviders() []dto.ProviderInfo { return s.providers }
func (s *stubSvc) GetStats(context.Context, string, *dto.StatsQuery) (*dto.StatsResult, error) {
	return s.result, s.err
}
func (s *stubSvc) Overview(context.Context, uuid.UUID) (*dto.OverviewResponse, error) {
	return s.overview, s.err
}
func (s *stubSvc) Export(context.Context, string, string, *dto.StatsQuery) ([]byte, string, error) {
	return s.bytes, s.filename, s.err
}

func doGET(h gin.HandlerFunc, path string, params gin.Params) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, path, nil)
	c.Params = params
	c.Set("request_id", "rid")
	c.Set("user_id", uuid.New().String())
	h(c)
	return w
}

func TestProviders(t *testing.T) {
	h := NewStatsHandler(&stubSvc{providers: []dto.ProviderInfo{{Name: "member-distribution", DisplayName: "会员分布统计"}}})
	w := doGET(h.Providers, "/api/v1/stats/providers", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOverview(t *testing.T) {
	h := NewStatsHandler(&stubSvc{overview: &dto.OverviewResponse{TotalMembers: 10}})
	w := doGET(h.Overview, "/api/v1/stats/overview", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 0, resp.Code)
}

func TestGet_OK(t *testing.T) {
	h := NewStatsHandler(&stubSvc{result: &dto.StatsResult{Provider: "member-distribution"}})
	w := doGET(h.Get, "/api/v1/stats/member-distribution", gin.Params{{Key: "provider", Value: "member-distribution"}})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGet_InvalidDepartment(t *testing.T) {
	h := NewStatsHandler(&stubSvc{result: &dto.StatsResult{Provider: "x"}})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/stats/member-distribution?department_id=not-uuid", nil)
	c.Params = gin.Params{{Key: "provider", Value: "member-distribution"}}
	c.Set("request_id", "rid")
	h.Get(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGet_InvalidDate(t *testing.T) {
	h := NewStatsHandler(&stubSvc{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/stats/member-distribution?start_date=bogus", nil)
	c.Params = gin.Params{{Key: "provider", Value: "member-distribution"}}
	c.Set("request_id", "rid")
	h.Get(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGet_ServiceError(t *testing.T) {
	h := NewStatsHandler(&stubSvc{err: response.NewError(response.CodeStatsProviderNotFound, "统计提供者不存在")})
	w := doGET(h.Get, "/api/v1/stats/missing", gin.Params{{Key: "provider", Value: "missing"}})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExport_OK(t *testing.T) {
	h := NewStatsHandler(&stubSvc{bytes: []byte("a,b"), filename: "stats_member-distribution.csv"})
	w := doGET(h.Export, "/api/v1/stats/export/member-distribution?format=csv", gin.Params{{Key: "provider", Value: "member-distribution"}})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Disposition"), "stats_member-distribution.csv")
}

func TestExport_Error(t *testing.T) {
	h := NewStatsHandler(&stubSvc{err: response.NewError(response.CodeStatsExportFormat, "导出格式不支持")})
	w := doGET(h.Export, "/api/v1/stats/export/member-distribution?format=pdf", gin.Params{{Key: "provider", Value: "member-distribution"}})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterRoutes(t *testing.T) {
	assert.NotPanics(t, func() {
		RegisterRoutes(gin.New().Group("/api/v1"), NewStatsHandler(&stubSvc{}), nil, nil, nil)
	})
}

func TestServeNamed(t *testing.T) {
	h := NewStatsHandler(&stubSvc{result: &dto.StatsResult{Provider: "task-progress"}})
	w := doGET(h.serveNamed("task-progress"), "/api/v1/stats/task-progress", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestApplyScopeAllAndDenied(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)
	q := applyScope(c, &dto.StatsQuery{})
	assert.True(t, q.AllScope)
	assert.False(t, q.Denied)

	deptID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	c.Set("data_scope_condition", &rbacModel.DataScopeCondition{
		Query: "department_id = ?",
		Args:  []interface{}{deptID},
	})
	q = applyScope(c, &dto.StatsQuery{})
	assert.False(t, q.AllScope)
	assert.Equal(t, []uuid.UUID{deptID}, q.ScopeDeptIDs)

	c.Set("data_scope_condition", &rbacModel.DataScopeCondition{Query: "1 = 0"})
	q = applyScope(c, nil)
	assert.True(t, q.Denied)
	assert.False(t, q.AllScope)
}

func TestExtractScopeDeptIDs(t *testing.T) {
	d1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	d2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	ids := extractScopeDeptIDs([]interface{}{d1, &d2, []uuid.UUID{d1}, d1.String(), "not-uuid"})
	assert.Equal(t, []uuid.UUID{d1, d2, d1, d1}, ids)
}

func TestBindQueryDates(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/x?start_date=2026-01-01&end_date=2026-01-31&group_by=department&granularity=month", nil)
	q, err := bindQuery(c)
	require.NoError(t, err)
	require.NotNil(t, q.StartDate)
	require.NotNil(t, q.EndDate)
	assert.Equal(t, "department", q.GroupBy)
	assert.Equal(t, "month", q.Granularity)
}
