package repo

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/pkg/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetByIdentity_KeepsUserPrimaryKey(t *testing.T) {
	db := testutil.OpenPostgres(t)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.UserIdentity{}))

	userID := uuid.New()
	identID := uuid.New()
	require.NotEqual(t, userID, identID)

	user := &model.User{
		ID:           userID,
		Username:     "cas-join-" + userID.String()[:8],
		PasswordHash: "x",
		CreatedAt:    time.Now().Add(-time.Hour),
	}
	require.NoError(t, db.Create(user).Error)
	t.Cleanup(func() {
		_ = db.Where("id = ?", identID).Delete(&model.UserIdentity{}).Error
		_ = db.Unscoped().Where("id = ?", userID).Delete(&model.User{}).Error
	})

	require.NoError(t, db.Create(&model.UserIdentity{
		ID:            identID,
		UserID:        userID,
		IdentityType:  "cas",
		IdentityValue: "20210001-" + identID.String()[:8],
		IsPrimary:     true,
		CreatedAt:     time.Now(),
	}).Error)

	got, err := NewUserRepo(db).GetByIdentity(context.Background(), "cas", "20210001-"+identID.String()[:8])
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, userID, got.ID)
	require.NotEqual(t, identID, got.ID)
}
