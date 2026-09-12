package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/backup/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

func TestRegisterRoutes_RequiresBackupRead(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uid := uuid.New()
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(auth.ContextKeyUserID, uid.String())
		c.Next()
	})
	RegisterRoutes(r.Group("/api/v1"), New(&stubSvc{
		policy:  &dto.Policy{RetentionDays: 30},
		storage: &dto.StorageStats{Count: 0},
		list:    []dto.Record{},
	}), stubCache{perms: []string{"backup:read"}})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/system/backups/policies", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/system/backups/"+uuid.New().String()+"/preview", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	denied := gin.New()
	denied.Use(func(c *gin.Context) {
		c.Set(auth.ContextKeyUserID, uid.String())
		c.Next()
	})
	RegisterRoutes(denied.Group("/api/v1"), New(&stubSvc{policy: &dto.Policy{}}), stubCache{
		perms: []string{"monitor:read"},
	})
	w = httptest.NewRecorder()
	denied.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/system/backups/policies", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
}
