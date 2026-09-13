package handler

import (
	"context"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	userModel "github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/pkg/testutil"
	"github.com/gin-gonic/gin"
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

func TestScopedOfficeAppointments(t *testing.T) {
	tx := testutil.OpenPostgres(t).Begin()
	require.NoError(t, tx.Error)
	defer tx.Rollback()
	user := userModel.User{ID: uuid.New(), Username: "office_" + uuid.NewString(), PasswordHash: "x"}
	require.NoError(t, tx.Create(&user).Error)
	center := model.Department{ID: uuid.New(), Code: "center_" + uuid.NewString(), Name: "Center"}
	require.NoError(t, tx.Create(&center).Error)
	dept := model.Department{ID: uuid.New(), Code: "dept_" + uuid.NewString(), Name: "Department", ParentID: &center.ID}
	require.NoError(t, tx.Create(&dept).Error)
	var minister model.Role
	require.NoError(t, tx.Exec("INSERT INTO roles(id,name,code,status,is_system) VALUES (?, 'Minister','minister',0,true) ON CONFLICT(code) DO UPDATE SET status=0", uuid.New()).Error)
	require.NoError(t, tx.Where("code='minister'").First(&minister).Error)
	cache := &membershipCache{}
	r := testutil.NewEngine()
	super := false
	r.Use(func(c *gin.Context) { c.Set("is_super_admin", super); c.Next() })
	r.PUT("/roles/:id/users/:user_id", roleMembership(tx, cache, true))
	r.DELETE("/roles/:id/users/:user_id", roleMembership(tx, cache, false))
	path := "/roles/" + minister.ID.String() + "/users/" + user.ID.String()
	body := map[string]interface{}{"department_ids": []uuid.UUID{dept.ID}}
	require.Equal(t, http.StatusForbidden, testutil.JSONRequest(t, r, "PUT", path, body, nil).Code)
	super = true
	require.Equal(t, http.StatusBadRequest, testutil.JSONRequest(t, r, "PUT", path, nil, nil).Code)
	require.Equal(t, http.StatusBadRequest, testutil.JSONRequest(t, r, "PUT", path, map[string]interface{}{"department_ids": []uuid.UUID{center.ID}}, nil).Code)
	response := testutil.JSONRequest(t, r, "PUT", path, body, nil)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	scopes, err := rbacRepo.LoadRoleDepartments(context.Background(), tx, user.ID)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{dept.ID}, scopes["minister"])
	require.NoError(t, tx.Model(&dept).Update("status", 1).Error)
	scopes, err = rbacRepo.LoadRoleDepartments(context.Background(), tx, user.ID)
	require.NoError(t, err)
	require.Contains(t, scopes, "minister")
	require.Empty(t, scopes["minister"])
	require.Equal(t, http.StatusOK, testutil.JSONRequest(t, r, "DELETE", path, nil, nil).Code)
	var count int64
	require.NoError(t, tx.Table("user_role_departments").Where("department_id=?", dept.ID).Count(&count).Error)
	require.Zero(t, count)
}
