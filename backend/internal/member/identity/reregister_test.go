package identity

import (
	"context"
	"testing"

	authsvc "github.com/Yogdunana/StarByte/backend/internal/auth/service"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	memberrepo "github.com/Yogdunana/StarByte/backend/internal/member/repo"
	rbacrepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	usermodel "github.com/Yogdunana/StarByte/backend/internal/user/model"
	userrepo "github.com/Yogdunana/StarByte/backend/internal/user/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCASRegisterAfterAdminDeleteIsNotMember(t *testing.T) {
	tx := testutil.OpenPostgres(t).Begin()
	defer tx.Rollback()
	ctx := context.Background()
	users := userrepo.NewUserRepo(tx)
	profiles := memberrepo.NewProfileRepo(tx)
	lookup := NewLookup(profiles)

	require.NoError(t, tx.Exec(
		`INSERT INTO roles(id,name,code,status) VALUES (?,'User','user',0),(?,'Member','member',0) ON CONFLICT(code) DO NOTHING`,
		uuid.New(), uuid.New(),
	).Error)

	old := &usermodel.User{ID: uuid.New(), Username: "old-" + uuid.NewString()[:8], PasswordHash: "x", Status: 0}
	require.NoError(t, users.Create(ctx, nil, old))
	studentNo := "1120" + uuid.NewString()[:6]
	require.NoError(t, profiles.Create(ctx, &model.MemberProfile{
		ID: uuid.New(), UserID: old.ID, RealName: "段茗尧", StudentNo: studentNo,
		MemberType: model.MemberTypeMember, Status: model.ProfileActive,
	}))
	require.NoError(t, users.CreateIdentity(ctx, &usermodel.UserIdentity{
		UserID: old.ID, IdentityType: "cas", IdentityValue: studentNo, IsPrimary: true,
	}))
	require.NoError(t, tx.Exec(
		`INSERT INTO user_roles(id,user_id,role_id) SELECT ?, ?, id FROM roles WHERE code='member'`,
		uuid.New(), old.ID,
	).Error)
	require.NoError(t, users.Delete(ctx, old.ID))

	fresh := &usermodel.User{ID: uuid.New(), Username: "new-" + uuid.NewString()[:8], PasswordHash: "x", Status: 0}
	require.NoError(t, users.Create(ctx, nil, fresh))
	require.NoError(t, lookup.EnsureStudentNo(ctx, fresh.ID, studentNo, "段茗尧"))
	require.NoError(t, authsvc.NewRoleAssigner(tx, rbacrepo.NewRoleRepo(tx), "member").AssignDefault(ctx, fresh.ID))

	got, err := profiles.GetByUserID(ctx, fresh.ID)
	require.NoError(t, err)
	require.Nil(t, got)

	var roles []string
	require.NoError(t, tx.Table("user_roles ur").Select("r.code").Joins("JOIN roles r ON r.id=ur.role_id").Where("ur.user_id=?", fresh.ID).Scan(&roles).Error)
	require.Equal(t, []string{"user"}, roles)
}
