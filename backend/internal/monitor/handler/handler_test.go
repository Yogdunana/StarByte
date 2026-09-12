package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/monitor/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() { gin.SetMode(gin.TestMode) }

type stubSvc struct {
	server *dto.ServerStatus
	app    *dto.AppHealth
	db     *dto.DatabaseStatus
	redis  *dto.RedisStatus
	api    *dto.APIStats
	slow   *dto.SlowQueries
	err    error
}

func (s *stubSvc) Server(context.Context) (*dto.ServerStatus, error)     { return s.server, s.err }
func (s *stubSvc) App(context.Context) (*dto.AppHealth, error)           { return s.app, s.err }
func (s *stubSvc) Database(context.Context) (*dto.DatabaseStatus, error) { return s.db, s.err }
func (s *stubSvc) Redis(context.Context) (*dto.RedisStatus, error)       { return s.redis, s.err }
func (s *stubSvc) APIStats(context.Context) (*dto.APIStats, error)       { return s.api, s.err }
func (s *stubSvc) SlowQueries(context.Context) (*dto.SlowQueries, error) { return s.slow, s.err }
func (s *stubSvc) Snapshot(context.Context) *dto.LiveSnapshot {
	return &dto.LiveSnapshot{Server: s.server, App: s.app, Database: s.db, Redis: s.redis, API: s.api}
}

func doGET(h gin.HandlerFunc, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, path, nil)
	c.Set("request_id", "rid")
	h(c)
	return w
}

func TestHandlerOK(t *testing.T) {
	h := New(&stubSvc{
		server: &dto.ServerStatus{CPUPercent: 10},
		app:    &dto.AppHealth{Goroutines: 8},
		db:     &dto.DatabaseStatus{Available: true},
		redis:  &dto.RedisStatus{Available: true},
		api:    &dto.APIStats{Available: true, RequestTotal: 3},
		slow:   &dto.SlowQueries{Available: true, Source: "in_memory"},
	})
	cases := []struct {
		fn   gin.HandlerFunc
		path string
	}{
		{h.Server, "/api/v1/monitor/server"},
		{h.App, "/api/v1/monitor/app"},
		{h.Database, "/api/v1/monitor/database"},
		{h.Redis, "/api/v1/monitor/redis"},
		{h.APIStats, "/api/v1/monitor/api-stats"},
		{h.SlowQueries, "/api/v1/monitor/slow-queries"},
	}
	for _, tc := range cases {
		w := doGET(tc.fn, tc.path)
		assert.Equal(t, http.StatusOK, w.Code, tc.path)
		var env response.Response
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
		assert.Equal(t, 0, env.Code)
	}
}

func TestHandlerError(t *testing.T) {
	h := New(&stubSvc{err: response.NewError(response.CodeMonitorCollectFail, "采集失败")})
	w := doGET(h.Server, "/api/v1/monitor/server")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
