package repo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	usermodel "github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/pkg/testutil"
)

func TestCreateProfileWritesEmptyJSONArraysNotNull(t *testing.T) {
	tx := testutil.OpenPostgres(t).Begin()
	defer tx.Rollback()
	userID := uuid.New()
	require.NoError(t, tx.Create(&usermodel.User{
		ID: userID, Username: "prof-" + uuid.NewString()[:8], PasswordHash: "x", Status: 0,
	}).Error)
	p := &model.MemberProfile{
		ID:         uuid.New(),
		UserID:     userID,
		RealName:   "招新",
		StudentNo:  "1120" + uuid.NewString()[:6],
		MemberType: model.MemberTypeMember,
		Status:     model.ProfileActive,
	}
	require.NoError(t, NewProfileRepo(tx).Create(context.Background(), p))
	var skills, projects string
	require.NoError(t, tx.Raw("SELECT skills::text, projects::text FROM member_profiles WHERE id = ?", p.ID).
		Row().Scan(&skills, &projects))
	require.Equal(t, "[]", skills)
	require.Equal(t, "[]", projects)
}
