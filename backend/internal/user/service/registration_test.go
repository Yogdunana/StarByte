package service

import (
	"context"
	"testing"

	rbacmodel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/user/dto"
	"github.com/Yogdunana/StarByte/backend/internal/user/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRegistrationGrantsOnlyUserAndRollsBackWithoutRole(t *testing.T) {
	tx := testutil.OpenPostgres(t).Begin()
	defer tx.Rollback()
	require.NoError(t, tx.Exec(`INSERT INTO roles(id,name,code,status) VALUES (?,'User','user',0) ON CONFLICT(code) DO NOTHING`, uuid.New()).Error)
	svc := NewUserService(tx, repo.NewUserRepo(tx), &config.JWTConfig{})
	u, err := svc.Register(context.Background(), &dto.RegisterRequest{Username: "reg-" + uuid.NewString()[:12], Password: "ValidPass123", Email: "new@example.test"})
	require.NoError(t, err)
	var roles []string
	require.NoError(t, tx.Table("user_roles ur").Select("r.code").Joins("JOIN roles r ON r.id=ur.role_id").Where("ur.user_id=?", u.ID).Scan(&roles).Error)
	require.Equal(t, []string{"user"}, roles)
	require.NoError(t, tx.Model(&rbacmodel.Role{}).Where("code=?", "user").Update("status", 1).Error)
	name := "reg-" + uuid.NewString()[:12]
	_, err = svc.Register(context.Background(), &dto.RegisterRequest{Username: name, Password: "ValidPass123"})
	require.Error(t, err)
	var count int64
	require.NoError(t, tx.Table("users").Where("username=?", name).Count(&count).Error)
	require.Zero(t, count)
}
