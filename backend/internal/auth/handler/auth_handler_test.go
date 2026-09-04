package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newReservedContext(method, path string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(method, path, nil)
	c.Request = req
	c.Set("request_id", "test-request-id")
	return c, w
}

func parseEnvelope(t *testing.T, w *httptest.ResponseRecorder) response.Response {
	t.Helper()
	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	return resp
}

func TestWechatQRCode_NotImplementedEnvelope(t *testing.T) {
	h := NewAuthHandler(nil)
	c, w := newReservedContext(http.MethodPost, "/api/v1/auth/wechat/qrcode")
	h.WechatQRCode(c)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
	resp := parseEnvelope(t, w)
	assert.Equal(t, response.CodeNotImplemented, resp.Code)
	assert.NotEqual(t, response.CodeNotFound, resp.Code)
	assert.Equal(t, "test-request-id", resp.RequestID)
	assert.Contains(t, resp.Message, "微信扫码登录")
}

func TestWechatCallback_NotImplementedEnvelope(t *testing.T) {
	h := NewAuthHandler(nil)
	c, w := newReservedContext(http.MethodPost, "/api/v1/auth/wechat/callback")
	h.WechatCallback(c)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
	resp := parseEnvelope(t, w)
	assert.Equal(t, response.CodeNotImplemented, resp.Code)
	assert.Equal(t, "test-request-id", resp.RequestID)
}

func TestOAuthLogin_NotImplementedEnvelope(t *testing.T) {
	h := NewAuthHandler(nil)
	c, w := newReservedContext(http.MethodPost, "/api/v1/auth/oauth/github")
	c.Params = gin.Params{{Key: "provider", Value: "github"}}
	h.OAuthLogin(c)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
	resp := parseEnvelope(t, w)
	assert.Equal(t, response.CodeNotImplemented, resp.Code)
	assert.NotEqual(t, response.CodeNotFound, resp.Code)
	assert.Equal(t, "test-request-id", resp.RequestID)
	assert.Contains(t, resp.Message, "github")
}
