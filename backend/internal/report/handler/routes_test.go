package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/report/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubPermissionCache struct {
	permissions []string
}

func (s stubPermissionCache) GetUserPermissions(context.Context, uuid.UUID) ([]string, error) {
	return s.permissions, nil
}
func (s stubPermissionCache) InvalidateUserPermissions(context.Context, uuid.UUID) error {
	return nil
}
func (s stubPermissionCache) InvalidateRolePermissions(context.Context, uuid.UUID) error {
	return nil
}
func (s stubPermissionCache) IsSuperAdmin(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}
func (s stubPermissionCache) GetUserPermissionsAndSuperAdmin(context.Context, uuid.UUID) ([]string, bool, error) {
	return s.permissions, false, nil
}
func (s stubPermissionCache) GetUserRoleCodes(context.Context, uuid.UUID) ([]string, error) {
	return nil, nil
}

func setupRoutes(svc *fakeService, permissions []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	userID := uuid.New()
	router.Use(func(c *gin.Context) {
		c.Set(auth.ContextKeyUserID, userID.String())
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, New(svc), stubPermissionCache{permissions: permissions}, nil, nil)
	return router
}

func TestRoutesRequireReportRead(t *testing.T) {
	svc := &fakeService{}
	router := setupRoutes(svc, nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil))

	assert.Equal(t, http.StatusForbidden, response.Code)
	assert.Nil(t, svc.lastRequest)
}

func TestRoutesAllowReportReadAndFailClosedWhenScopeCannotLoad(t *testing.T) {
	svc := &fakeService{list: []*dto.ReportResponse{}, page: 1, pageSize: 20}
	router := setupRoutes(svc, []string{"report:read"})
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil))

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.NotNil(t, svc.lastRequest)
	require.NotNil(t, svc.lastDataScope)
	assert.Equal(t, "1 = 0", svc.lastDataScope.Query)
	assert.False(t, svc.lastDataScope.IsSelf)
}
