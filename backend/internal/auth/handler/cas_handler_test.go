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
	// lastState 用指针记录 handler 实际传给 service 的 state，
	// 用于在测试里区分"取自 Cookie"还是"取自 URL query"。
	lastState *string
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
func (s casStubService) KickSession(context.Context, string) error           { return nil }
func (s casStubService) KickUserSessions(context.Context, string) error      { return nil }
func (s casStubService) RevokeAllUserSessions(context.Context, string) error { return nil }
func (s casStubService) CASStatus() dto.CASStatusResponse {
	return dto.CASStatusResponse{Enabled: s.enabled}
}
func (s casStubService) BuildCASLoginURL(context.Context, string, string) (*dto.CASLoginStart, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &dto.CASLoginStart{Location: s.login, State: "st-cookie", Service: "http://10.0.0.8/api/v1/auth/cas/callback"}, nil
}
func (s casStubService) CompleteCASCallback(_ context.Context, _ string, state string, _ string, _ string, _ string) (string, error) {
	if s.lastState != nil {
		*s.lastState = state
	}
	return s.cb, s.err
}
func (s casStubService) ExchangeCASCode(context.Context, string) (*dto.CASExchangeResponse, error) {
	return s.ex, s.err
}
func (s casStubService) RegisterWithCASToken(context.Context, *dto.CASRegisterRequest, string, string) (*dto.CASExchangeResponse, error) {
	return s.ex, s.err
}
func (s casStubService) VerifyEmail(context.Context, string) error { return s.err }
func (s casStubService) ResendVerification(context.Context, string, string) error {
	return s.err
}

func TestCASLogin_Redirect(t *testing.T) {
	h := NewAuthHandler(casStubService{
		enabled: true,
		login:   "https://authserver.smbu.edu.cn/authserver/login?service=x",
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/cas/login?redirect=/tasks", nil)
	req.Host = "10.0.0.8"
	c.Request = req
	h.CASLogin(c)
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, casReferrerPolicy, w.Header().Get("Referrer-Policy"))
	assert.Contains(t, w.Header().Get("Location"), "authserver.smbu.edu.cn")
	assert.Contains(t, w.Header().Get("Set-Cookie"), casStateCookie)
}

func TestCASLogin_RedirectSetsNoReferrer(t *testing.T) {
	h := NewAuthHandler(casStubService{
		enabled: true,
		login:   "https://authserver.smbu.edu.cn/authserver/login?service=http%3A%2F%2F10.0.0.8%2Fapi%2Fv1%2Fauth%2Fcas%2Fcallback",
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/cas/login", nil)
	req.Host = "10.0.0.8"
	c.Request = req
	h.CASLogin(c)
	require.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "no-referrer", w.Header().Get("Referrer-Policy"))
	assert.Contains(t, w.Header().Get("Location"), "https://authserver.smbu.edu.cn/authserver/login?service=")
	assert.Contains(t, w.Header().Get("Location"), "10.0.0.8")
}

func TestCASCallback_Redirect(t *testing.T) {
	var got string
	h := NewAuthHandler(casStubService{cb: "http://10.0.0.8/login/cas?code=abc", lastState: &got})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/cas/callback?ticket=ST-1", nil)
	// state 必须由 Cookie 提供：CAS 的登录地址里不含 state，正常回调不会出现在 query 里。
	req.AddCookie(&http.Cookie{Name: casStateCookie, Value: "s"})
	c.Request = req
	h.CASCallback(c)
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, casReferrerPolicy, w.Header().Get("Referrer-Policy"))
	assert.Contains(t, w.Header().Get("Location"), "/login/cas?code=abc")
	assert.Equal(t, "s", got)
}

// TestCASCallback_IgnoresQueryState 是这次修复的回归用例：
// 攻击者可以把自己的 state 拼到回调 URL 上发给受害者（登录 CSRF / 会话固定，CWE-352），
// 所以即使 query 里有 state，也必须以 Cookie 为准。
func TestCASCallback_IgnoresQueryState(t *testing.T) {
	var got string
	h := NewAuthHandler(casStubService{cb: "http://10.0.0.8/login/cas?code=abc", lastState: &got})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/cas/callback?ticket=ST-1&state=attacker-state", nil)
	req.AddCookie(&http.Cookie{Name: casStateCookie, Value: "victim-state"})
	c.Request = req
	h.CASCallback(c)
	require.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "victim-state", got, "query 里的 state 必须被忽略")
}

// TestCASCallback_RequiresStateCookie：没有 Cookie 时不该继续换票，
// 而应像 service 层其他 CAS 失败一样回跳前端登录页并带上原因。
func TestCASCallback_RequiresStateCookie(t *testing.T) {
	var got string
	h := NewAuthHandler(casStubService{cb: "http://10.0.0.8/login/cas?code=abc", lastState: &got})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/cas/callback?ticket=ST-1&state=attacker-state", nil)
	h.CASCallback(c)
	require.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login?cas_error=missing_state")
	assert.Equal(t, "", got, "缺少登录态 Cookie 时不得调用 CompleteCASCallback")
	// 顺手清掉可能残留的 Cookie
	assert.Contains(t, w.Header().Get("Set-Cookie"), casStateCookie)
}

func TestRequestPublicOrigin_IgnoresForwardedHost(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/cas/login", nil)
	c.Request.Host = "10.0.0.8"
	c.Request.Header.Set("X-Forwarded-Proto", "http")
	c.Request.Header.Set("X-Forwarded-Host", "evil.example")
	assert.Equal(t, "http://10.0.0.8", requestPublicOrigin(c))
}

func TestRequestPublicOrigin_RejectsNonHTTPProto(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/cas/login", nil)
	c.Request.Host = "10.100.13.17"
	c.Request.Header.Set("X-Forwarded-Proto", "javascript")
	assert.Equal(t, "http://10.100.13.17", requestPublicOrigin(c))
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

func TestCASRegister_OK(t *testing.T) {
	h := NewAuthHandler(casStubService{ex: &dto.CASExchangeResponse{
		LoginResponse: dto.LoginResponse{AccessToken: "tok", RefreshToken: "rt"},
		Redirect:      "/dashboard",
	}})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("request_id", "rid")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/cas/register", strings.NewReader(`{"token":"abc","username":"alice","password":"Passw0rd1","gender":1}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.CASRegister(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseEnvelope(t, w)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, casReferrerPolicy, w.Header().Get("Referrer-Policy"))
}
