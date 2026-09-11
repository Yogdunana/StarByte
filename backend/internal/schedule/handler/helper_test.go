package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestOriginAllowlistsSchemeAndTrimsForwardedProto(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originOf := func(proto, host string) string {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "http://"+host+"/schedules/google/callback", nil)
		req.Host = host
		if proto != "" {
			req.Header.Set("X-Forwarded-Proto", proto)
		}
		c.Request = req
		return requestOrigin(c)
	}

	assert.Equal(t, "https://app.example", originOf("https, http", "app.example"))
	assert.Equal(t, "http://app.example", originOf("http", "app.example"))
	assert.Equal(t, "http://app.example", originOf("https://evil.example", "app.example"))
	assert.Equal(t, "http://app.example", originOf("javascript", "app.example"))
	assert.Equal(t, "", requestOrigin(nil))
}
