package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyEmail_MissingToken(t *testing.T) {
	h := NewAuthHandler(&sessionStub{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.VerifyEmail(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVerifyEmail_OK(t *testing.T) {
	h := NewAuthHandler(&sessionStub{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/verify-email?token=abc", nil)
	h.VerifyEmail(c)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 0, resp.Code)
}

func TestResendVerification_OK(t *testing.T) {
	h := NewAuthHandler(&sessionStub{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/resend-verification", strings.NewReader(`{"email":"a@b.c"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Host = "10.100.13.17"
	h.ResendVerification(c)
	assert.Equal(t, http.StatusOK, w.Code)
}
