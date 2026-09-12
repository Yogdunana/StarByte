package repo

import (
	"context"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPermissionSelectionPreservesExistingDataScope(t *testing.T) {
	tx := testutil.OpenPostgres(t).Begin()
	require.NoError(t, tx.Error)
	defer tx.Rollback()
	role := model.Role{ID: uuid.New(), Name: "Scope", Code: "scope_" + uuid.NewString()}
	require.NoError(t, tx.Create(&role).Error)
	perms := []model.Permission{{ID: uuid.New(), Name: "Read", Code: "read_" + uuid.NewString(), Type: model.PermissionTypeAPI}, {ID: uuid.New(), Name: "Write", Code: "write_" + uuid.NewString(), Type: model.PermissionTypeAPI}}
	require.NoError(t, tx.Create(&perms).Error)
	grant := model.RolePermission{ID: uuid.New(), RoleID: role.ID, PermissionID: perms[0].ID, DataScope: model.DataScopeAll}
	require.NoError(t, tx.Create(&grant).Error)
	require.NoError(t, NewRoleRepo(tx).AssignPermissions(context.Background(), tx, role.ID, []uuid.UUID{perms[0].ID, perms[1].ID}, ""))
	var grants []model.RolePermission
	require.NoError(t, tx.Where("role_id = ?", role.ID).Find(&grants).Error)
	require.Len(t, grants, 2)
	for _, g := range grants {
		if g.PermissionID == perms[0].ID {
			require.Equal(t, grant.ID, g.ID)
			require.Equal(t, model.DataScopeAll, g.DataScope)
		} else {
			require.Equal(t, model.DataScopeDepartment, g.DataScope)
		}
	}
}
