package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/leave/dto"
	"github.com/Yogdunana/StarByte/backend/internal/leave/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubCache struct {
	perms []string
	super bool
}

func (s stubCache) GetUserPermissions(context.Context, uuid.UUID) ([]string, error) {
	return s.perms, nil
}
func (s stubCache) InvalidateUserPermissions(context.Context, uuid.UUID) error { return nil }
func (s stubCache) InvalidateRolePermissions(context.Context, uuid.UUID) error { return nil }
func (s stubCache) IsSuperAdmin(context.Context, uuid.UUID) (bool, error)      { return s.super, nil }
func (s stubCache) GetUserPermissionsAndSuperAdmin(context.Context, uuid.UUID) ([]string, bool, error) {
	return s.perms, s.super, nil
}
func (s stubCache) GetUserRoleCodes(context.Context, uuid.UUID) ([]string, error) { return nil, nil }

type fakeSvc struct {
	lastViewer service.Viewer
	types      []dto.LeaveTypeResponse
}

func (f *fakeSvc) ListTypes(context.Context) ([]dto.LeaveTypeResponse, error) {
	return f.types, nil
}
func (f *fakeSvc) Submit(_ context.Context, v service.Viewer, _ *dto.SubmitLeaveRequest) (*dto.LeaveApplicationResponse, error) {
	f.lastViewer = v
	return &dto.LeaveApplicationResponse{ID: "app-1", Applicant: dto.Person{ID: v.UserID.String()}}, nil
}
func (f *fakeSvc) Get(_ context.Context, v service.Viewer, id uuid.UUID) (*dto.LeaveApplicationResponse, error) {
	f.lastViewer = v
	return &dto.LeaveApplicationResponse{ID: id.String()}, nil
}
func (f *fakeSvc) ListMine(_ context.Context, v service.Viewer, _ *dto.ListLeaveRequest) ([]*dto.LeaveApplicationResponse, int64, error) {
	f.lastViewer = v
	return []*dto.LeaveApplicationResponse{{ID: "mine"}}, 1, nil
}
func (f *fakeSvc) ListAll(_ context.Context, v service.Viewer, _ *dto.ListLeaveRequest) ([]*dto.LeaveApplicationResponse, int64, error) {
	f.lastViewer = v
	return []*dto.LeaveApplicationResponse{{ID: "all"}}, 1, nil
}
func (f *fakeSvc) Approve(_ context.Context, v service.Viewer, _ uuid.UUID, _ string) error {
	f.lastViewer = v
	return nil
}
func (f *fakeSvc) Reject(_ context.Context, v service.Viewer, _ uuid.UUID, _ string) error {
	f.lastViewer = v
	return nil
}
func (f *fakeSvc) Balances(_ context.Context, v service.Viewer, _ string, _ int) ([]*dto.LeaveBalanceResponse, error) {
	f.lastViewer = v
	return nil, nil
}
func (f *fakeSvc) Stats(_ context.Context, v service.Viewer) (*dto.LeaveStatsResponse, error) {
	f.lastViewer = v
	return &dto.LeaveStatsResponse{Total: 2}, nil
}

func setupLeaveRouter(t *testing.T, svc *fakeSvc, perms []string) (*gin.Engine, uuid.UUID) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	uid := uuid.New()
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(auth.ContextKeyUserID, uid.String())
		c.Next()
	})
	api := r.Group("/api/v1")
	RegisterRoutes(api, New(svc), stubCache{perms: perms})
	return r, uid
}

func TestRoutes_TypesAndStatsNotCapturedByID(t *testing.T) {
	svc := &fakeSvc{types: []dto.LeaveTypeResponse{{Code: "annual", Name: "年假"}}}
	r, _ := setupLeaveRouter(t, svc, []string{"leave:read", "leave:approve"})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/leave/types", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("types status %d body=%s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/leave/stats", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("stats status %d body=%s", w.Code, w.Body.String())
	}
	var env struct {
		Data dto.LeaveStatsResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.Total != 2 {
		t.Fatalf("stats total %d", env.Data.Total)
	}
}

func TestRoutes_ListMineUsesJWTUser(t *testing.T) {
	svc := &fakeSvc{}
	r, uid := setupLeaveRouter(t, svc, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/leave/my?user_id="+uuid.New().String(), nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body=%s", w.Code, w.Body.String())
	}
	if svc.lastViewer.UserID != uid {
		t.Fatalf("viewer %s want %s", svc.lastViewer.UserID, uid)
	}
}

func TestRoutes_ApproveRequiresPermission(t *testing.T) {
	svc := &fakeSvc{}
	r, _ := setupLeaveRouter(t, svc, nil)
	id := uuid.New()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/leave/"+id.String()+"/approve", nil)
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Fatal("approve without leave:approve should be forbidden")
	}

	r, uid := setupLeaveRouter(t, svc, []string{"leave:approve"})
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/v1/leave/"+id.String()+"/approve", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("approve status %d body=%s", w.Code, w.Body.String())
	}
	if svc.lastViewer.UserID != uid || !svc.lastViewer.CanApprove {
		t.Fatalf("viewer %+v", svc.lastViewer)
	}
}
