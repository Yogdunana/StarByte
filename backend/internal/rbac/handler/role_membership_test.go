package handler

import (
	"context"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	userModel "github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/pkg/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

type membershipCache struct {
	rbacService.PermissionCacheService
	invalidated []uuid.UUID
}

func (c *membershipCache) InvalidateUserPermissions(_ context.Context, id uuid.UUID) error {
	c.invalidated = append(c.invalidated, id)
	return nil
}
func TestRoleMembershipCustomRoleAndSystemProtection(t *testing.T) {
	db := testutil.OpenPostgres(t)
	tx := db.Begin()
	require.NoError(t, tx.Error)
	defer tx.Rollback()
	role := model.Role{ID: uuid.New(), Name: "Custom", Code: "test_" + uuid.NewString()}
	require.NoError(t, tx.Create(&role).Error)
	user := userModel.User{ID: uuid.New(), Username: "membership_" + uuid.NewString(), PasswordHash: "not-a-login-hash"}
	require.NoError(t, tx.Create(&user).Error)
	cache := &membershipCache{}
	r := testutil.NewEngine()
	r.PUT("/roles/:id/users/:user_id", roleMembership(tx, cache, true))
	r.DELETE("/roles/:id/users/:user_id", roleMembership(tx, cache, false))
	path := "/roles/" + role.ID.String() + "/users/" + user.ID.String()
	for i := 0; i < 2; i++ {
		w := testutil.JSONRequest(t, r, "PUT", path, nil, nil)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	}
	var count int64
	require.NoError(t, tx.Model(&model.UserRole{}).Where("user_id = ? AND role_id = ?", user.ID, role.ID).Count(&count).Error)
	require.EqualValues(t, 1, count)
	w := testutil.JSONRequest(t, r, "DELETE", path, nil, nil)
	require.Equal(t, http.StatusOK, w.Code)
	require.Len(t, cache.invalidated, 3)
	require.NoError(t, tx.Model(&role).Update("is_system", true).Error)
	w = testutil.JSONRequest(t, r, "PUT", path, nil, nil)
	require.Equal(t, http.StatusForbidden, w.Code)
	w = testutil.JSONRequest(t, r, "DELETE", path, nil, nil)
	require.Equal(t, http.StatusForbidden, w.Code)
	require.Len(t, cache.invalidated, 3)
}
