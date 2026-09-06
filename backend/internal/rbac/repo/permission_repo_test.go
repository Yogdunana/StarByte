package repo

import (
	"context"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPermissionRepo_CRUD(t *testing.T) {
	db := testutil.OpenPostgres(t)
	r := NewPermissionRepo(db)
	ctx := context.Background()
	code := "test:perm:" + uuid.NewString()[:8]
	parent := &model.Permission{
		ID: uuid.New(), Name: "parent", Code: code + ":p", Type: model.PermissionTypeMenu,
	}
	require.NoError(t, r.Create(ctx, nil, parent))
	t.Cleanup(func() { _ = r.Delete(ctx, parent.ID) })

	child := &model.Permission{
		ID: uuid.New(), Name: "child", Code: code + ":c", Type: model.PermissionTypeAPI, ParentID: &parent.ID,
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		return r.Create(ctx, tx, child)
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = r.Delete(ctx, child.ID) })

	got, err := r.GetByID(ctx, parent.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, parent.Code, got.Code)

	missing, err := r.GetByID(ctx, uuid.New())
	require.NoError(t, err)
	assert.Nil(t, missing)

	byCode, err := r.GetByCode(ctx, parent.Code)
	require.NoError(t, err)
	require.NotNil(t, byCode)
	assert.Equal(t, parent.ID, byCode.ID)

	none, err := r.GetByCode(ctx, "no-such-"+uuid.NewString())
	require.NoError(t, err)
	assert.Nil(t, none)

	empty, err := r.GetByIDs(ctx, nil, nil)
	require.NoError(t, err)
	assert.Empty(t, empty)

	batch, err := r.GetByIDs(ctx, nil, []uuid.UUID{parent.ID, child.ID})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(batch), 2)

	list, err := r.List(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, list)

	n, err := r.CountChildren(ctx, parent.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, n, int64(1))

	parent.Name = "parent-updated"
	require.NoError(t, r.Update(ctx, nil, parent))
	got, err = r.GetByID(ctx, parent.ID)
	require.NoError(t, err)
	assert.Equal(t, "parent-updated", got.Name)

	ids, err := r.GetPermissionIDsByUserID(ctx, uuid.New())
	require.NoError(t, err)
	assert.Empty(t, ids)

	codes, err := r.GetPermissionCodesByUserID(ctx, uuid.New())
	require.NoError(t, err)
	assert.Empty(t, codes)
}
