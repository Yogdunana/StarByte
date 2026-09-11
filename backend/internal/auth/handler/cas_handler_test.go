package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/auth/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCASStatus_NilService(t *testing.T) {
	h := NewAuthHandler(nil)
	c, w := newReservedContext(http.MethodGet, "/api/v1/auth/cas/status")
	h.CASStatus(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseEnvelope(t, w)
	raw, _ := json.Marshal(resp.Data)
	var status dto.CASStatusResponse
	require.NoError(t, json.Unmarshal(raw, &status))
	assert.False(t, status.Enabled)
}

func TestCASLogin_NilService(t *testing.T) {
	h := NewAuthHandler(nil)
	c, w := newReservedContext(http.MethodGet, "/api/v1/auth/cas/login")
	h.CASLogin(c)
	assert.Equal(t, http.StatusNotImplemented, w.Code)
	resp := parseEnvelope(t, w)
	assert.Equal(t, response.CodeNotImplemented, resp.Code)
}

type casStubService struct {
	enabled bool
	login   string
	cb      string
	ex      *dto.CASExchangeResponse
	err     error
}

func (s casStubService) Login(context.Context, *dto.LoginRequest, string, string) (*dto.LoginResponse, error) {
	return nil, nil
}
func (s casStubService) RefreshToken(context.Context, *dto.RefreshTokenRequest, string, string) (*dto.RefreshResponse, error) {
	return nil, nil
}
func (s casStubService) Logout(context.Context, string, string, string) error { return nil }
func (s casStubService) GetCurrentUser(context.Context, string) (*dto.UserInfo, error) {
	return nil, nil
}
func (s casStubService) ChangePassword(context.Context, string, *dto.ChangePasswordRequest) error {
	return nil
}
func (s casStubService) ListSessions(context.Context, string, string) (*dto.SessionListResponse, error) {
	return nil, nil
}
func (s casStubService) GetUserSessions(context.Context, string) (*dto.UserSessionsResponse, error) {
	return nil, nil
}
func (s casStubService) KickSession(context.Context, string) error      { return nil }
func (s casStubService) KickUserSessions(context.Context, string) error { return nil }
func (s casStubService) CASStatus() dto.CASStatusResponse {
	return dto.CASStatusResponse{Enabled: s.enabled}
}
func (s casStubService) BuildCASLoginURL(context.Context, string) (string, error) {
	return s.login, s.err
}
func (s casStubService) CompleteCASCallback(context.Context, string, string, string, string) (string, error) {
	return s.cb, s.err
}
func (s casStubService) ExchangeCASCode(context.Context, string) (*dto.CASExchangeResponse, error) {
	return s.ex, s.err
}

func TestCASLogin_Redirect(t *testing.T) {
	h := NewAuthHandler(casStubService{
		enabled: true,
		login:   "https://authserver.smbu.edu.cn/authserver/login?service=x",
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/cas/login?redirect=/tasks", nil)
	h.CASLogin(c)
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "authserver.smbu.edu.cn")
}

func TestCASCallback_Redirect(t *testing.T) {
	h := NewAuthHandler(casStubService{cb: "https://starbyte.smbu.edu.cn/login/cas?code=abc"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/cas/callback?ticket=ST-1&state=s", nil)
	h.CASCallback(c)
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login/cas?code=abc")
}

func TestCASExchange_OK(t *testing.T) {
	h := NewAuthHandler(casStubService{ex: &dto.CASExchangeResponse{
		LoginResponse: dto.LoginResponse{AccessToken: "tok", RefreshToken: "rt"},
		Redirect:      "/dashboard",
	}})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("request_id", "rid")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/cas/exchange", strings.NewReader(`{"code":"abc"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.CASExchange(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseEnvelope(t, w)
	assert.Equal(t, 0, resp.Code)
}
