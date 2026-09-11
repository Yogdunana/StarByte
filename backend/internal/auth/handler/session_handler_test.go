package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/auth/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sessionStub struct {
	list *dto.SessionListResponse
	user *dto.UserSessionsResponse
	err  error
}

func (s *sessionStub) Login(context.Context, *dto.LoginRequest, string, string) (*dto.LoginResponse, error) {
	return nil, nil
}
func (s *sessionStub) RefreshToken(context.Context, *dto.RefreshTokenRequest, string, string) (*dto.RefreshResponse, error) {
	return nil, nil
}
func (s *sessionStub) Logout(context.Context, string, string, string) error { return nil }
func (s *sessionStub) GetCurrentUser(context.Context, string) (*dto.UserInfo, error) {
	return nil, nil
}
func (s *sessionStub) ChangePassword(context.Context, string, *dto.ChangePasswordRequest) error {
	return nil
}
func (s *sessionStub) ListSessions(context.Context, string, string) (*dto.SessionListResponse, error) {
	return s.list, s.err
}
func (s *sessionStub) GetUserSessions(context.Context, string) (*dto.UserSessionsResponse, error) {
	return s.user, s.err
}
func (s *sessionStub) KickSession(context.Context, string) error      { return s.err }
func (s *sessionStub) KickUserSessions(context.Context, string) error { return s.err }
func (s *sessionStub) CASStatus() dto.CASStatusResponse {
	return dto.CASStatusResponse{Enabled: false}
}
func (s *sessionStub) BuildCASLoginURL(context.Context, string, string) (*dto.CASLoginStart, error) {
	return nil, nil
}
func (s *sessionStub) CompleteCASCallback(context.Context, string, string, string, string, string) (string, error) {
	return "", nil
}
func (s *sessionStub) ExchangeCASCode(context.Context, string) (*dto.CASExchangeResponse, error) {
	return nil, nil
}
func (s *sessionStub) RegisterWithCASToken(context.Context, *dto.CASRegisterRequest, string, string) (*dto.CASExchangeResponse, error) {
	return nil, nil
}

func TestListSessions_OK(t *testing.T) {
	h := NewAuthHandler(&sessionStub{list: &dto.SessionListResponse{List: []dto.SessionView{{TokenID: "j1"}}, Total: 1}})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/sessions", nil)
	c.Set("request_id", "rid")
	h.ListSessions(c)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 0, resp.Code)
}

func TestKickSession_Missing(t *testing.T) {
	h := NewAuthHandler(&sessionStub{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/auth/sessions/user", nil)
	c.Params = gin.Params{{Key: "token", Value: "user"}}
	c.Set("request_id", "rid")
	h.KickSession(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestKickSession_NotFound(t *testing.T) {
	h := NewAuthHandler(&sessionStub{err: response.NewError(response.CodeSessionNotFound, "会话不存在或已失效")})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/auth/sessions/j1", nil)
	c.Params = gin.Params{{Key: "token", Value: "j1"}}
	c.Set("request_id", "rid")
	h.KickSession(c)
	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, response.CodeSessionNotFound, resp.Code)
}

func TestGetUserSessions_OK(t *testing.T) {
	h := NewAuthHandler(&sessionStub{user: &dto.UserSessionsResponse{UserID: "u1", Sessions: []dto.SessionView{}}})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/sessions/u1", nil)
	c.Params = gin.Params{{Key: "user_id", Value: "u1"}}
	c.Set("request_id", "rid")
	h.GetUserSessions(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestKickUserSessions_OK(t *testing.T) {
	h := NewAuthHandler(&sessionStub{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/auth/sessions/user/u1", nil)
	c.Params = gin.Params{{Key: "user_id", Value: "u1"}}
	c.Set("request_id", "rid")
	h.KickUserSessions(c)
	assert.Equal(t, http.StatusOK, w.Code)
}
