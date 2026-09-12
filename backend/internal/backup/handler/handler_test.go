package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/backup/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() { gin.SetMode(gin.TestMode) }

type stubSvc struct {
	list    []dto.Record
	total   int64
	rec     *dto.Record
	policy  *dto.Policy
	storage *dto.StorageStats
	preview *dto.Preview
	drill   *dto.DrillResult
	err     error
}

func (s *stubSvc) List(context.Context, *dto.ListRequest) ([]dto.Record, int64, error) {
	return s.list, s.total, s.err
}
func (s *stubSvc) Get(context.Context, uuid.UUID) (*dto.Record, error) { return s.rec, s.err }
func (s *stubSvc) Create(context.Context, uuid.UUID, *dto.CreateRequest) (*dto.Record, error) {
	return s.rec, s.err
}
func (s *stubSvc) Delete(context.Context, uuid.UUID) error { return s.err }
func (s *stubSvc) Restore(context.Context, uuid.UUID, uuid.UUID, *dto.RestoreRequest) (*dto.Record, error) {
	return s.rec, s.err
}
func (s *stubSvc) DrillRestore(context.Context, uuid.UUID, uuid.UUID, *dto.DrillRequest) (*dto.DrillResult, error) {
	if s.drill != nil {
		return s.drill, s.err
	}
	return &dto.DrillResult{Queued: true, Status: "queued", TargetDBName: "starbyte_drill"}, s.err
}
func (s *stubSvc) GetDrill(context.Context, uuid.UUID) (*dto.DrillResult, error) {
	if s.drill != nil {
		return s.drill, s.err
	}
	return &dto.DrillResult{Status: "restored", Restored: true, TargetDBName: "starbyte_drill"}, s.err
}
func (s *stubSvc) WaitDrill(ctx context.Context, id uuid.UUID) (*dto.DrillResult, error) {
	return s.GetDrill(ctx, id)
}
func (s *stubSvc) Wait(context.Context, uuid.UUID) (*dto.Record, error) { return s.rec, s.err }
func (s *stubSvc) GetPolicy(context.Context) (*dto.Policy, error)       { return s.policy, s.err }
func (s *stubSvc) UpdatePolicy(context.Context, uuid.UUID, *dto.UpdatePolicyRequest) (*dto.Policy, error) {
	return s.policy, s.err
}
func (s *stubSvc) Storage(context.Context) (*dto.StorageStats, error) { return s.storage, s.err }
func (s *stubSvc) Preview(context.Context, uuid.UUID) (*dto.Preview, error) {
	return s.preview, s.err
}
func (s *stubSvc) RunScheduled(context.Context, string, func(string)) error {
	return s.err
}
func (s *stubSvc) CleanupExpired(context.Context, string, func(string)) error {
	return s.err
}
func (s *stubSvc) SyncSchedule(context.Context) error { return s.err }

func withUser(h gin.HandlerFunc, method, path string, body []byte) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, bytes.NewReader(body))
	if body != nil {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	c.Set("request_id", "rid")
	c.Set(auth.ContextKeyUserID, uuid.New().String())
	h(c)
	return w
}

func TestHandlerListOK(t *testing.T) {
	h := New(&stubSvc{list: []dto.Record{{ID: uuid.New().String()}}, total: 1})
	w := withUser(h.List, http.MethodGet, "/api/v1/system/backups?page=1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandlerDrillRestoreOK(t *testing.T) {
	id := uuid.New()
	h := New(&stubSvc{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/system/backups/"+id.String()+"/restore-drill", bytes.NewReader([]byte(
		`{"confirm":true,"confirmation":"DRILL","target_dbname":"starbyte_drill"}`,
	)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("request_id", "rid")
	c.Set(auth.ContextKeyUserID, uuid.New().String())
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.DrillRestore(c)
	assert.Equal(t, http.StatusOK, w.Code)
	var env response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.Equal(t, 0, env.Code)
}

func TestHandlerGetDrillOK(t *testing.T) {
	id := uuid.New()
	h := New(&stubSvc{drill: &dto.DrillResult{ID: id.String(), Status: "restored", Restored: true}})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/system/backups/"+id.String()+"/restore-drill", nil)
	c.Set("request_id", "rid")
	c.Set(auth.ContextKeyUserID, uuid.New().String())
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.GetDrill(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandlerGetPolicyOK(t *testing.T) {
	h := New(&stubSvc{policy: &dto.Policy{RetentionDays: 30, CronExpr: "0 30 2 * * *"}})
	w := withUser(h.GetPolicy, http.MethodGet, "/api/v1/system/backups/policies", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	var env response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.Equal(t, 0, env.Code)
}

func TestHandlerCreateError(t *testing.T) {
	h := New(&stubSvc{err: response.NewError(response.CodeBackupBusy, "忙")})
	w := withUser(h.Create, http.MethodPost, "/api/v1/system/backups", []byte(`{}`))
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestHandlerPreviewOK(t *testing.T) {
	h := New(&stubSvc{preview: &dto.Preview{Ready: true, ChecksumOK: true, GzipOK: true}})
	id := uuid.New()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/system/backups/"+id.String()+"/preview", nil)
	c.Set("request_id", "rid")
	c.Set(auth.ContextKeyUserID, uuid.New().String())
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.Preview(c)
	assert.Equal(t, http.StatusOK, w.Code)
}
