package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/cache/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() { gin.SetMode(gin.TestMode) }

type stubSvc struct {
	stats *dto.Stats
	del   *dto.DeleteResult
	warm  *dto.WarmupResult
	err   error
}

func (s *stubSvc) Stats(context.Context, string) (*dto.Stats, error) { return s.stats, s.err }
func (s *stubSvc) DeleteKey(context.Context, string) error           { return s.err }
func (s *stubSvc) DeletePattern(context.Context, string) (*dto.DeleteResult, error) {
	return s.del, s.err
}
func (s *stubSvc) Warmup(context.Context, *dto.WarmupRequest) (*dto.WarmupResult, error) {
	return s.warm, s.err
}
func (s *stubSvc) Close() {}

func doReq(h gin.HandlerFunc, method, path string, body any, params gin.Params) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	c.Request = httptest.NewRequest(method, path, &buf)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = params
	c.Set("request_id", "rid")
	h(c)
	return w
}

func TestCacheHandler_StatsOK(t *testing.T) {
	h := NewCacheHandler(&stubSvc{stats: &dto.Stats{Healthy: true, Pattern: "*"}})
	w := doReq(h.Stats, http.MethodGet, "/api/v1/system/cache/stats", nil, nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCacheHandler_DeleteKey(t *testing.T) {
	h := NewCacheHandler(&stubSvc{})
	w := doReq(h.DeleteKey, http.MethodDelete, "/api/v1/system/cache/demo:a", nil, gin.Params{{Key: "key", Value: "demo:a"}})
	assert.Equal(t, http.StatusOK, w.Code)

	w = doReq(h.DeleteKey, http.MethodDelete, "/api/v1/system/cache/", nil, gin.Params{{Key: "key", Value: ""}})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	h = NewCacheHandler(&stubSvc{err: response.NewError(response.CodeCacheKeyNotFound, "缓存键不存在")})
	w = doReq(h.DeleteKey, http.MethodDelete, "/api/v1/system/cache/x", nil, gin.Params{{Key: "key", Value: "x"}})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCacheHandler_DeletePatternAndWarmup(t *testing.T) {
	h := NewCacheHandler(&stubSvc{del: &dto.DeleteResult{Deleted: 2, Keys: []string{"a", "b"}}})
	w := doReq(h.DeletePattern, http.MethodDelete, "/api/v1/system/cache/pattern/demo:*", nil, gin.Params{{Key: "pattern", Value: "demo:*"}})
	assert.Equal(t, http.StatusOK, w.Code)

	w = doReq(h.DeletePattern, http.MethodDelete, "/api/v1/system/cache/pattern/", nil, gin.Params{{Key: "pattern", Value: ""}})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	h = NewCacheHandler(&stubSvc{warm: &dto.WarmupResult{Loaded: 1, Keys: []string{"k"}}})
	w = doReq(h.Warmup, http.MethodPost, "/api/v1/system/cache/warmup", dto.WarmupRequest{
		Entries: []dto.WarmupEntry{{Key: "k", Value: "v"}},
	}, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/system/cache/warmup", bytes.NewBufferString("{"))
	c.Set("request_id", "rid")
	h = NewCacheHandler(&stubSvc{err: response.NewError(response.CodeCacheWarmupFail, "fail")})
	w = doReq(h.Warmup, http.MethodPost, "/api/v1/system/cache/warmup", dto.WarmupRequest{
		Entries: []dto.WarmupEntry{{Key: "k", Value: "v"}},
	}, nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	h = NewCacheHandler(&stubSvc{err: response.NewError(response.CodeCacheInvalidPattern, "wide")})
	w = doReq(h.DeletePattern, http.MethodDelete, "/api/v1/system/cache/pattern/*", nil, gin.Params{{Key: "pattern", Value: "*"}})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCacheHandler_StatsError(t *testing.T) {
	h := NewCacheHandler(&stubSvc{err: response.NewError(response.CodeCacheRedisDown, "down")})
	w := doReq(h.Stats, http.MethodGet, "/api/v1/system/cache/stats", nil, nil)
	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, response.CodeCacheRedisDown, resp.Code)
}
