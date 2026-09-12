package handler

import (
	"context"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/internal/user/dto"
	"github.com/Yogdunana/StarByte/backend/internal/user/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

type routeCache struct {
	rbacService.PermissionCacheService
	permissions []string
}

func (r routeCache) GetUserPermissionsAndSuperAdmin(context.Context, uuid.UUID) ([]string, bool, error) {
	return r.permissions, false, nil
}

type profileService struct {
	service.UserService
	caller  string
	profile *dto.UpdateProfileRequest
}

func (s *profileService) UpdateProfile(_ context.Context, caller string, profile *dto.UpdateProfileRequest) error {
	s.caller, s.profile = caller, profile
	return nil
}
func TestUserManagementRequiresPermissions(t *testing.T) {
	id := uuid.NewString()
	r := testutil.NewEngine()
	r.Use(func(c *gin.Context) { c.Set(auth.ContextKeyUserID, id); c.Next() })
	RegisterUserRoutes(r.Group(""), NewUserHandler(&profileService{}), routeCache{})
	for _, route := range []struct{ method, path string }{{"GET", "/users"}, {"GET", "/users/" + id}, {"POST", "/users"}, {"PUT", "/users/" + id}, {"DELETE", "/users/" + id}} {
		w := testutil.JSONRequest(t, r, route.method, route.path, nil, nil)
		require.Equal(t, http.StatusForbidden, w.Code, route)
	}
}
func TestProfileUsesCallerAndIgnoresPrivilegeFields(t *testing.T) {
	id := uuid.NewString()
	svc := &profileService{}
	r := testutil.NewEngine()
	r.Use(func(c *gin.Context) { c.Set(auth.ContextKeyUserID, id); c.Next() })
	RegisterUserRoutes(r.Group(""), NewUserHandler(svc), routeCache{})
	w := testutil.JSONRequest(t, r, "PUT", "/user/profile", map[string]interface{}{"real_name": "新姓名", "user_id": uuid.NewString(), "role_ids": []string{uuid.NewString()}, "status": 0, "phone": "", "email": ""}, nil)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, id, svc.caller)
	require.Equal(t, "新姓名", svc.profile.RealName)
	require.NotNil(t, svc.profile.Phone)
	require.Empty(t, *svc.profile.Phone)
}
