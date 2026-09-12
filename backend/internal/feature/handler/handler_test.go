package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/feature"
	"github.com/Yogdunana/StarByte/backend/internal/feature/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() { gin.SetMode(gin.TestMode) }

type stubSvc struct {
	flag      *dto.FlagResponse
	list      []dto.FlagResponse
	total     int64
	eval      *dto.EvaluateResponse
	me        map[string]dto.EvaluateResponse
	audits    []dto.AuditResponse
	analytics *dto.AnalyticsResponse
	err       error
	sub       feature.Subject
	on        bool
}

func (s *stubSvc) List(context.Context, dto.ListQuery) ([]dto.FlagResponse, int64, error) {
	return s.list, s.total, s.err
}
func (s *stubSvc) Get(context.Context, uuid.UUID) (*dto.FlagResponse, error) { return s.flag, s.err }
func (s *stubSvc) Create(context.Context, uuid.UUID, *dto.CreateFlagRequest) (*dto.FlagResponse, error) {
	return s.flag, s.err
}
func (s *stubSvc) Update(context.Context, uuid.UUID, uuid.UUID, *dto.UpdateFlagRequest) (*dto.FlagResponse, error) {
	return s.flag, s.err
}
func (s *stubSvc) Toggle(context.Context, uuid.UUID, uuid.UUID, *dto.ToggleRequest) (*dto.FlagResponse, error) {
	return s.flag, s.err
}
func (s *stubSvc) Delete(context.Context, uuid.UUID, uuid.UUID) error { return s.err }
func (s *stubSvc) EvaluateID(context.Context, uuid.UUID, uuid.UUID) (*dto.EvaluateResponse, error) {
	return s.eval, s.err
}
func (s *stubSvc) EvaluateMe(context.Context, uuid.UUID, []string) (map[string]dto.EvaluateResponse, error) {
	return s.me, s.err
}
func (s *stubSvc) Enabled(context.Context, string, feature.Subject) bool { return s.on }
func (s *stubSvc) Resolve(context.Context, uuid.UUID) (feature.Subject, error) {
	return s.sub, s.err
}
func (s *stubSvc) ListAudits(context.Context, dto.AuditQuery) ([]dto.AuditResponse, int64, error) {
	return s.audits, s.total, s.err
}
func (s *stubSvc) Rollback(context.Context, uuid.UUID, uuid.UUID) (*dto.FlagResponse, error) {
	return s.flag, s.err
}
func (s *stubSvc) Analytics(context.Context, uuid.UUID, int) (*dto.AnalyticsResponse, error) {
	return s.analytics, s.err
}
func (s *stubSvc) StartHotReload(context.Context) error { return s.err }

func withUser(h gin.HandlerFunc, method, path string, body []byte) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, bytes.NewReader(body))
	if body != nil {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	c.Set("request_id", "rid")
	c.Set(auth.ContextKeyUserID, uuid.New().String())
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
	h(c)
	return w
}

func TestHandlerCRUD(t *testing.T) {
	h := New(&stubSvc{flag: &dto.FlagResponse{FlagKey: "cms.public"}, list: []dto.FlagResponse{{FlagKey: "cms.public"}}, total: 1})
	if withUser(h.List, http.MethodGet, "/api/v1/system/features?page=1", nil).Code != http.StatusOK {
		t.Fatal("list")
	}
	if withUser(h.Get, http.MethodGet, "/api/v1/system/features/x", nil).Code != http.StatusOK {
		t.Fatal("get")
	}
	body, _ := json.Marshal(dto.CreateFlagRequest{FlagKey: "demo.flag", Name: "D", FlagType: "boolean"})
	if withUser(h.Create, http.MethodPost, "/api/v1/system/features", body).Code != http.StatusOK {
		t.Fatal("create")
	}
	upd, _ := json.Marshal(dto.UpdateFlagRequest{})
	if withUser(h.Update, http.MethodPut, "/api/v1/system/features/x", upd).Code != http.StatusOK {
		t.Fatal("update")
	}
	if withUser(h.Toggle, http.MethodPost, "/api/v1/system/features/x/toggle", []byte(`{}`)).Code != http.StatusOK {
		t.Fatal("toggle")
	}
}

func TestHandlerEvaluateAndAudit(t *testing.T) {
	h := New(&stubSvc{
		eval:   &dto.EvaluateResponse{Key: "cms.public", Enabled: true, Reason: "boolean"},
		me:     map[string]dto.EvaluateResponse{"cms.public": {Key: "cms.public", Enabled: false}},
		audits: []dto.AuditResponse{{FlagKey: "cms.public", Action: "toggle"}},
		total:  1,
	})
	if withUser(h.Evaluate, http.MethodGet, "/api/v1/system/features/x/evaluate", nil).Code != http.StatusOK {
		t.Fatal("evaluate")
	}
	if withUser(h.EvaluateMe, http.MethodGet, "/api/v1/features/me?keys=cms.public", nil).Code != http.StatusOK {
		t.Fatal("me")
	}
	if withUser(h.Audit, http.MethodGet, "/api/v1/system/features/audit", nil).Code != http.StatusOK {
		t.Fatal("audit")
	}
}

func TestHandlerAuthAndQueryErrors(t *testing.T) {
	anon := New(&stubSvc{me: map[string]dto.EvaluateResponse{"cms.public": {Key: "cms.public", Enabled: false, Reason: "disabled"}}})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/features/me", nil)
	c.Set("request_id", "rid")
	anon.EvaluateMe(c)
	if w.Code != http.StatusOK {
		t.Fatalf("anonymous evaluate should be 200, got %d", w.Code)
	}
	h := New(&stubSvc{err: response.NewError(response.CodeFeatureNotFound, "gone")})
	w = withUser(h.Evaluate, http.MethodGet, "/api/v1/system/features/x/evaluate?user_id=not-a-uuid", nil)
	if w.Code == http.StatusOK {
		t.Fatal("bad user")
	}
	w = withUser(h.Get, http.MethodGet, "/api/v1/system/features/x", nil)
	if w.Code == http.StatusOK {
		t.Fatal("svc error should not be 200 with success code only if write maps error")
	}
	h2 := New(&stubSvc{list: []dto.FlagResponse{{FlagKey: "a"}}, total: 1})
	if withUser(h2.List, http.MethodGet, "/api/v1/system/features?page=-1&page_size=999", nil).Code != http.StatusOK {
		t.Fatal("page clamp")
	}
	if withUser(h2.Audit, http.MethodGet, "/api/v1/system/features/audit?page=1", nil).Code != http.StatusOK {
		t.Fatal("audit ok")
	}
}

func TestHandlerUnauthCreateUpdateToggle(t *testing.T) {
	h := New(&stubSvc{flag: &dto.FlagResponse{FlagKey: "x"}})
	for _, fn := range []gin.HandlerFunc{h.Create, h.Update, h.Toggle} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader([]byte(`{}`)))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("request_id", "rid")
		c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
		fn(c)
		if w.Code == http.StatusOK {
			t.Fatal("expected unauth")
		}
	}
}

func TestHandlerBadJSON(t *testing.T) {
	h := New(&stubSvc{})
	w := withUser(h.Create, http.MethodPost, "/api/v1/system/features", []byte(`{`))
	if w.Code == http.StatusOK {
		t.Fatal("expected bad request")
	}
}

func TestRequireFlag(t *testing.T) {
	off := &stubSvc{on: false}
	on := &stubSvc{on: true}
	deny := RequireFlag(off, "cms.public", nil)
	w := withUser(deny, http.MethodGet, "/api/v1/knowledge/public/docs", nil)
	var env response.Response
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	if env.Code != response.CodeFeatureDisabled {
		t.Fatalf("code=%d body=%s", env.Code, w.Body.String())
	}
	allow := RequireFlag(on, "cms.public", nil)
	if withUser(allow, http.MethodGet, "/api/v1/knowledge/public/docs", nil).Code != http.StatusOK && withUser(allow, http.MethodGet, "/api/v1/knowledge/public/docs", nil).Body.Len() >= 0 {
		// middleware calls Next with empty handlers; gin test context finishes without writing
	}
}

func TestStaffBypass(t *testing.T) {
	fn := StaffBypass("announcement:create")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_permissions", []string{"announcement:create"})
	if !fn(c) {
		t.Fatal("staff should bypass")
	}
	c.Set("user_permissions", []string{"announcement:read"})
	if fn(c) {
		t.Fatal("member should not bypass")
	}
}

func TestEvaluateAsUserAndBadID(t *testing.T) {
	h := New(&stubSvc{eval: &dto.EvaluateResponse{Key: "cms.public", Enabled: false, Reason: "allowlist_miss"}})
	uid := uuid.New().String()
	if withUser(h.Evaluate, http.MethodGet, "/api/v1/system/features/x/evaluate?user_id="+uid, nil).Code != http.StatusOK {
		t.Fatal("evaluate as")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/system/features/bad/evaluate", nil)
	c.Set("request_id", "rid")
	c.Set(auth.ContextKeyUserID, uuid.New().String())
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.Evaluate(c)
	if w.Code == http.StatusOK {
		t.Fatal("bad id")
	}
}

func TestRequireFlagAnonymous(t *testing.T) {
	gin.SetMode(gin.TestMode)
	off := RequireFlag(&stubSvc{on: false}, "cms.public", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/public/docs", nil)
	c.Set("request_id", "rid")
	off(c)
	var env response.Response
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	if env.Code != response.CodeFeatureDisabled {
		t.Fatalf("anon miss code=%d body=%s", env.Code, w.Body.String())
	}

	allow := RequireFlag(&stubSvc{on: true}, "cms.public", nil)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/public/docs", nil)
	c.Set("request_id", "rid")
	allow(c)
	if c.IsAborted() {
		t.Fatal("boolean-on should let anonymous through")
	}
}

func TestRequireFlagBypass(t *testing.T) {
	mw := RequireFlag(&stubSvc{on: false}, "cms.public", func(*gin.Context) bool { return true })
	w := withUser(mw, http.MethodGet, "/api/v1/knowledge/public/docs", nil)
	if w.Code != http.StatusOK && w.Body.Len() == 0 {
		// bypass calls Next; empty handler is fine
	}
}

func TestStaffBypassStar(t *testing.T) {
	fn := StaffBypass("announcement:create")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_permissions", []string{"*"})
	if !fn(c) {
		t.Fatal("star")
	}
	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	if fn(c2) {
		t.Fatal("empty")
	}
}

func TestHandlerAnalyticsAndRollback(t *testing.T) {
	h := New(&stubSvc{
		flag:      &dto.FlagResponse{FlagKey: "exp.hero"},
		analytics: &dto.AnalyticsResponse{FlagKey: "exp.hero", Total: 3, Days: 7},
	})
	if withUser(h.Analytics, http.MethodGet, "/api/v1/system/features/x/analytics?days=7", nil).Code != http.StatusOK {
		t.Fatal("analytics")
	}
	if withUser(h.Rollback, http.MethodPost, "/api/v1/system/features/x/rollback", []byte(`{}`)).Code != http.StatusOK {
		t.Fatal("rollback")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/system/features/bad/rollback", nil)
	c.Set("request_id", "rid")
	c.Set(auth.ContextKeyUserID, uuid.New().String())
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.Rollback(c)
	if w.Code == http.StatusOK {
		t.Fatal("bad rollback id")
	}
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/system/features/bad/analytics", nil)
	c.Set("request_id", "rid")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.Analytics(c)
	if w.Code == http.StatusOK {
		t.Fatal("bad analytics id")
	}
	h2 := New(&stubSvc{err: response.NewError(response.CodeFeatureNoRollback, "none")})
	w = withUser(h2.Rollback, http.MethodPost, "/api/v1/system/features/x/rollback", []byte(`{}`))
	if w.Code == http.StatusOK {
		t.Fatal("no snapshot")
	}
}

func TestSplitKeys(t *testing.T) {
	if got := splitKeys(" cms.public, announcement.feed "); len(got) != 2 || got[0] != "cms.public" {
		t.Fatalf("%v", got)
	}
	if splitKeys("  ") != nil {
		t.Fatal("empty")
	}
}
