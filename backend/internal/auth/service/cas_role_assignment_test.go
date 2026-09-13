package service

import (
	"context"
	rbacrepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	usermodel "github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/pkg/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCASCannotAutoGrantMembershipFromLegacyConfig(t *testing.T) {
	tx := testutil.OpenPostgres(t).Begin()
	defer tx.Rollback()
	require.NoError(t, tx.Exec(`INSERT INTO roles(id,name,code,status) VALUES (?,'User','user',0) ON CONFLICT(code) DO NOTHING`, uuid.New()).Error)
	u := &usermodel.User{ID: uuid.New(), Username: "cas-role-" + uuid.NewString()[:12], PasswordHash: "not-a-password"}
	require.NoError(t, tx.Create(u).Error)
	a := NewRoleAssigner(tx, rbacrepo.NewRoleRepo(tx), "member")
	require.NoError(t, a.AssignDefault(context.Background(), u.ID))
	var roles []string
	require.NoError(t, tx.Table("user_roles ur").Select("r.code").Joins("JOIN roles r ON r.id=ur.role_id").Where("ur.user_id=?", u.ID).Scan(&roles).Error)
	require.Equal(t, []string{"user"}, roles)
}
