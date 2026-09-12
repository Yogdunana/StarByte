package repo

import (
	"context"
	"github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/pkg/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProfileUpdateKeepsIdentityAndSyncsMemberName(t *testing.T) {
	db := testutil.OpenPostgres(t)
	tx := db.Begin()
	require.NoError(t, tx.Error)
	defer tx.Rollback()
	id := uuid.New()
	u := &model.User{ID: id, Username: "profile_" + id.String(), PasswordHash: "not-a-login-hash", RealName: "Old", Email: "old@example.test", Status: 2}
	require.NoError(t, tx.Create(u).Error)
	require.NoError(t, tx.Exec("INSERT INTO member_profiles(id,user_id,real_name,student_no) VALUES (?,?,?,?)", uuid.New(), id, "Old", "student-"+id.String()[:20]).Error)
	require.NoError(t, NewUserRepo(tx).UpdateProfile(context.Background(), id, map[string]interface{}{"real_name": "New", "email": ""}))
	var got model.User
	require.NoError(t, tx.First(&got, "id = ?", id).Error)
	require.Equal(t, "New", got.RealName)
	require.Empty(t, got.Email)
	require.Equal(t, u.Username, got.Username)
	require.Equal(t, 2, got.Status)
	var p struct{ RealName, StudentNo string }
	require.NoError(t, tx.Table("member_profiles").Where("user_id = ?", id).First(&p).Error)
	require.Equal(t, "New", p.RealName)
	require.Equal(t, "student-"+id.String()[:20], p.StudentNo)
}
