package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestID_PropagatesToRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/id", func(c *gin.Context) {
		assert.Equal(t, "abc-123", c.GetString("request_id"))
		assert.Equal(t, "abc-123", logger.RequestIDFrom(c.Request.Context()))
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/id", nil)
	req.Header.Set("X-Request-Id", "abc-123")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "abc-123", w.Header().Get("X-Request-Id"))
}
