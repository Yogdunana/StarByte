package repo

import (
	"github.com/Yogdunana/StarByte/backend/internal/knowledge/model"
	"github.com/Yogdunana/StarByte/backend/pkg/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRoleScopeSQLMatchesDetailRules(t *testing.T) {
	db := testutil.OpenPostgres(t).Begin()
	defer db.Rollback()
	require.NoError(t, db.Exec(`CREATE TEMP TABLE cms_role_test (id uuid, author_id uuid, status smallint, visibility text, permission_code text, allowed_roles jsonb)`).Error)
	author := uuid.New()
	id := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO cms_role_test VALUES (?, ?, 1, 'role', '', '["member"]')`, id, author).Error)
	d := &model.Doc{ID: id, AuthorID: author, Status: 1, Visibility: model.VisibilityRole, AllowedRoles: []string{"member"}}
	for _, s := range []ReadableScope{{}, {UserID: uuid.New(), Roles: []string{"user"}}, {UserID: uuid.New(), Roles: []string{"user", "member"}}, {UserID: author}, {Staff: true, UserID: uuid.New()}} {
		var count int64
		require.NoError(t, applyReadable(db.Table("cms_role_test AS d"), s).Count(&count).Error)
		require.Equal(t, Visible(d, s), count == 1)
	}
}
