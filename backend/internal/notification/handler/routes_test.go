package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type stubPermCache struct {
	perms []string
	super bool
}

func (s stubPermCache) GetUserPermissions(context.Context, uuid.UUID) ([]string, error) {
	return s.perms, nil
}
func (s stubPermCache) InvalidateUserPermissions(context.Context, uuid.UUID) error { return nil }
func (s stubPermCache) InvalidateRolePermissions(context.Context, uuid.UUID) error { return nil }
func (s stubPermCache) IsSuperAdmin(context.Context, uuid.UUID) (bool, error)      { return s.super, nil }
func (s stubPermCache) GetUserPermissionsAndSuperAdmin(context.Context, uuid.UUID) ([]string, bool, error) {
	return s.perms, s.super, nil
}
func (s stubPermCache) GetUserRoleCodes(context.Context, uuid.UUID) ([]string, error) {
	return nil, nil
}

func newNotifRouter(perms []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	uid := uuid.New()
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(auth.ContextKeyUserID, uid.String())
		c.Next()
	})
	g := r.Group("/api/v1")
	RegisterRoutes(g, g, NewNotificationHandler(nil, nil), NewTemplateHandler(nil), nil, nil, stubPermCache{perms: perms})
	return r
}

func TestRegisterRoutes_SendRequiresNotificationSend(t *testing.T) {
	body := []byte(`{"user_ids":["` + uuid.New().String() + `"],"template_code":"backup_failed","channels":["email"]}`)

	denied := newNotifRouter([]string{"notification:template:read"})
	for _, path := range []string{"/api/v1/notifications/send", "/api/v1/system/notifications/send"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		denied.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code, path)
	}
}
