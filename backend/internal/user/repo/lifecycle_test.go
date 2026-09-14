package repo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	membermodel "github.com/Yogdunana/StarByte/backend/internal/member/model"
	memberrepo "github.com/Yogdunana/StarByte/backend/internal/member/repo"
	"github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/Yogdunana/StarByte/backend/pkg/testutil"
)

func TestSoftDeleteReleasesStudentNoAndIdentity(t *testing.T) {
	tx := testutil.OpenPostgres(t).Begin()
	defer tx.Rollback()
	ctx := context.Background()
	users := NewUserRepo(tx)
	profiles := memberrepo.NewProfileRepo(tx)

	u1 := &model.User{ID: uuid.New(), Username: "old-" + uuid.NewString()[:8], PasswordHash: "x", Status: 0}
	require.NoError(t, users.Create(ctx, nil, u1))
	studentNo := "1120" + uuid.NewString()[:6]
	require.NoError(t, profiles.Create(ctx, &membermodel.MemberProfile{
		ID: uuid.New(), UserID: u1.ID, RealName: "段茗尧", StudentNo: studentNo,
		MemberType: membermodel.MemberTypeMember, Status: membermodel.ProfileActive,
	}))
	require.NoError(t, users.CreateIdentity(ctx, &model.UserIdentity{
		UserID: u1.ID, IdentityType: "cas", IdentityValue: studentNo, IsPrimary: true,
	}))
	require.NoError(t, users.Delete(ctx, u1.ID))

	var student string
	require.NoError(t, tx.Raw("SELECT student_no FROM member_profiles WHERE user_id = ?", u1.ID).Scan(&student).Error)
	require.Empty(t, student)
	var identCount int64
	require.NoError(t, tx.Table("user_identities").Where("user_id = ?", u1.ID).Count(&identCount).Error)
	require.Zero(t, identCount)

	u2 := &model.User{ID: uuid.New(), Username: "new-" + uuid.NewString()[:8], PasswordHash: "x", Status: 0}
	require.NoError(t, users.Create(ctx, nil, u2))
	require.NoError(t, profiles.Create(ctx, &membermodel.MemberProfile{
		ID: uuid.New(), UserID: u2.ID, RealName: "段茗尧", StudentNo: studentNo,
		MemberType: membermodel.MemberTypeMember, Status: membermodel.ProfileActive,
	}))
	got, err := profiles.GetByStudentNo(ctx, studentNo, nil)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, u2.ID, got.UserID)
}

func TestCreateIdentityConflictMapsTo409(t *testing.T) {
	tx := testutil.OpenPostgres(t).Begin()
	defer tx.Rollback()
	ctx := context.Background()
	users := NewUserRepo(tx)
	a := &model.User{ID: uuid.New(), Username: "ida-" + uuid.NewString()[:8], PasswordHash: "x"}
	b := &model.User{ID: uuid.New(), Username: "idb-" + uuid.NewString()[:8], PasswordHash: "x"}
	require.NoError(t, users.Create(ctx, nil, a))
	require.NoError(t, users.Create(ctx, nil, b))
	casUser := "cas-" + uuid.NewString()[:8]
	require.NoError(t, users.CreateIdentity(ctx, &model.UserIdentity{
		UserID: a.ID, IdentityType: "cas", IdentityValue: casUser,
	}))
	err := users.CreateIdentity(ctx, &model.UserIdentity{
		UserID: b.ID, IdentityType: "cas", IdentityValue: casUser,
	})
	require.Error(t, err)
	appErr, ok := err.(*response.AppError)
	require.True(t, ok)
	require.Equal(t, response.CodeUserExists, appErr.Code)
	require.Equal(t, 409, appErr.HTTPStatus)
}

func TestHardDeleteClearsAuditAndFlowFKs(t *testing.T) {
	tx := testutil.OpenPostgres(t).Begin()
	defer tx.Rollback()
	ctx := context.Background()
	users := NewUserRepo(tx)
	u := &model.User{ID: uuid.New(), Username: "hd-" + uuid.NewString()[:8], PasswordHash: "x"}
	require.NoError(t, users.Create(ctx, nil, u))
	auditID := uuid.New()
	require.NoError(t, tx.Exec(
		`INSERT INTO audit_logs (id, user_id, operation) VALUES (?, ?, 'login')`,
		auditID, u.ID,
	).Error)

	defID := uuid.New()
	verID := uuid.New()
	instID := uuid.New()
	require.NoError(t, tx.Exec(
		`INSERT INTO flow_definitions (id, key, name) VALUES (?, ?, 'lifecycle')`,
		defID, "lc-"+uuid.NewString()[:8],
	).Error)
	require.NoError(t, tx.Exec(
		`INSERT INTO flow_definition_versions (id, definition_id, version, bpmn_data, status)
		 VALUES (?, ?, 1, '{}'::jsonb, 1)`,
		verID, defID,
	).Error)
	require.NoError(t, tx.Exec(
		`INSERT INTO flow_instances (id, definition_id, definition_version_id, initiator_id, status)
		 VALUES (?, ?, ?, ?, 0)`,
		instID, defID, verID, u.ID,
	).Error)

	require.NoError(t, users.HardDelete(ctx, u.ID))
	var auditUser *uuid.UUID
	require.NoError(t, tx.Raw("SELECT user_id FROM audit_logs WHERE id = ?", auditID).Scan(&auditUser).Error)
	require.Nil(t, auditUser)
	var initiator *uuid.UUID
	require.NoError(t, tx.Raw("SELECT initiator_id FROM flow_instances WHERE id = ?", instID).Scan(&initiator).Error)
	require.Nil(t, initiator)
}
