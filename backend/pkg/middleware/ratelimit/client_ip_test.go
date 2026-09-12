package ratelimit

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFrontendProxyPreservesClientIPWithoutTrustingSpoofedPeers(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "172.30.77.2/32")
	r := gin.New()
	require.NoError(t, r.SetTrustedProxies(TrustedProxiesFromEnv()))
	r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, c.ClientIP()) })
	for _, tc := range []struct{ peer, forwarded, want string }{
		{"172.30.77.2:4567", "198.51.100.9", "198.51.100.9"},
		{"172.30.77.2:4567", "1.2.3.4, 198.51.100.9", "198.51.100.9"},
		{"198.51.100.9:4567", "1.2.3.4", "198.51.100.9"},
		{"172.30.77.4:4567", "1.2.3.4", "172.30.77.4"},
	} {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = tc.peer
		req.Header.Set("X-Forwarded-For", tc.forwarded)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		require.Equal(t, tc.want, w.Body.String())
	}
}
