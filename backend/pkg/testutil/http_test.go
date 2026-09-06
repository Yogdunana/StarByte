package testutil

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestJSONRequest(t *testing.T) {
	r := NewEngine()
	r.POST("/echo", func(c *gin.Context) {
		var body map[string]string
		_ = c.ShouldBindJSON(&body)
		c.JSON(http.StatusOK, body)
	})
	w := JSONRequest(t, r, http.MethodPost, "/echo", map[string]string{"k": "v"}, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"k":"v"`)
}

func TestJSONRequest_Headers(t *testing.T) {
	r := NewEngine()
	r.GET("/h", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"x": c.GetHeader("X-Test")})
	})
	w := JSONRequest(t, r, http.MethodGet, "/h", nil, map[string]string{"X-Test": "yes"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"yes"`)
}
