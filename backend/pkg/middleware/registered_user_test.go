package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
)

func TestRegisteredUserBoundary(t *testing.T) {
	for _, tc := range []struct {
		roles        []string
		method, path string
		want         int
	}{
		{[]string{"user"}, "GET", "/api/v1/user/me", 200},
		{nil, "GET", "/api/v1/user/me", 200},
		{[]string{"user"}, "POST", "/api/v1/member/applications", 200},
		{[]string{"user"}, "GET", "/api/v1/users", 403},
		{[]string{"user"}, "POST", "/api/v1/task", 403},
		{[]string{"user", "member"}, "GET", "/api/v1/task", 200},
	} {
		r := gin.New()
		r.Use(func(c *gin.Context) { c.Set("current_roles", tc.roles) }, RegisteredUserAccess())
		r.Handle(tc.method, tc.path, func(c *gin.Context) { c.Status(200) })
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
		require.Equal(t, tc.want, w.Code, tc.path)
	}
}

func TestBaseAccountDiscardsStaleElevatedPermissions(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("current_roles", []string{"user"})
	perms, super := baseAccountPermissions(c, []string{"*"}, true)
	require.False(t, super)
	require.Equal(t, []string{"announcement:read"}, perms)
	c.Set("current_roles", []string{"user", "president"})
	perms, super = baseAccountPermissions(c, []string{"member:approve"}, false)
	require.False(t, super)
	require.Equal(t, []string{"member:approve"}, perms)
}
