package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/rbac"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubPermissionService struct {
	tree []dto.PermissionTreeResponse
	item *dto.PermissionResponse
	err  error
}

func (s *stubPermissionService) Create(_ context.Context, _ *dto.CreatePermissionRequest) (*dto.PermissionResponse, error) {
	return s.item, s.err
}
func (s *stubPermissionService) GetByID(_ context.Context, _ uuid.UUID) (*dto.PermissionResponse, error) {
	return s.item, s.err
}
func (s *stubPermissionService) GetTree(_ context.Context) ([]dto.PermissionTreeResponse, error) {
	return s.tree, s.err
}
func (s *stubPermissionService) Update(_ context.Context, _ uuid.UUID, _ *dto.UpdatePermissionRequest) (*dto.PermissionResponse, error) {
	return s.item, s.err
}
func (s *stubPermissionService) Delete(_ context.Context, _ uuid.UUID) error {
	return s.err
}

func TestPermissionHandler_GetTree(t *testing.T) {
	h := NewPermissionHandler(&stubPermissionService{
		tree: []dto.PermissionTreeResponse{{ID: uuid.New().String(), Name: "根", Code: "root"}},
	})
	r := testutil.NewEngine()
	r.GET("/system/permissions", h.GetTree)
	w := testutil.JSONRequest(t, r, http.MethodGet, "/system/permissions", nil, nil)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"root"`)
}

func TestPermissionHandler_GetByID_Invalid(t *testing.T) {
	h := NewPermissionHandler(&stubPermissionService{})
	r := testutil.NewEngine()
	r.GET("/system/permissions/:id", h.GetByID)
	w := testutil.JSONRequest(t, r, http.MethodGet, "/system/permissions/not-a-uuid", nil, nil)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPermissionHandler_CreateAndDelete(t *testing.T) {
	id := uuid.New()
	h := NewPermissionHandler(&stubPermissionService{
		item: &dto.PermissionResponse{ID: id.String(), Name: "写", Code: "user:write"},
	})
	r := testutil.NewEngine()
	r.POST("/system/permissions", h.Create)
	r.DELETE("/system/permissions/:id", h.Delete)

	w := testutil.JSONRequest(t, r, http.MethodPost, "/system/permissions", map[string]string{
		"name": "写", "code": "user:write", "type": "api",
	}, nil)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"user:write"`)

	w = testutil.JSONRequest(t, r, http.MethodDelete, "/system/permissions/"+id.String(), nil, nil)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestPermissionHandler_Create_Duplicate(t *testing.T) {
	h := NewPermissionHandler(&stubPermissionService{err: rbac.NewPermissionCodeExistsError("user:write")})
	r := testutil.NewEngine()
	r.POST("/system/permissions", h.Create)
	w := testutil.JSONRequest(t, r, http.MethodPost, "/system/permissions", map[string]string{
		"name": "写", "code": "user:write", "type": "api",
	}, nil)
	assert.NotEqual(t, http.StatusOK, w.Code)
}
