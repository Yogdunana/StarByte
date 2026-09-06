package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveDataScope_failClosedAndSuperAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	got, err := ResolveDataScope(c, nil, nil, nil, "task")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "1 = 0", got.Query)

	got, err = ResolveDataScope(c, nil, nil, nil, "")
	require.NoError(t, err)
	assert.Equal(t, "1 = 0", got.Query)

	c.Set("is_super_admin", true)
	c.Set(auth.ContextKeyUserID, "not-a-uuid")
	got, err = ResolveDataScope(c, nil, nil, nil, "task")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "", got.Query)
}
