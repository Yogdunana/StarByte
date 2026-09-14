package httpx

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeOrigin(t *testing.T) {
	assert.Equal(t, "http://10.0.0.8", SanitizeOrigin("http://10.0.0.8/login"))
	assert.Equal(t, "https://app.example:443", SanitizeOrigin("https://app.example:443/x"))
	assert.Equal(t, "", SanitizeOrigin("javascript:alert(1)"))
	assert.Equal(t, "", SanitizeOrigin("//evil.example"))
	assert.Equal(t, "", SanitizeOrigin("https://user:pass@evil.example"))
	assert.Equal(t, "", SanitizeOrigin(""))
}

func TestMailOriginIgnoresRequestHost(t *testing.T) {
	assert.Equal(t, "https://app.example", MailOrigin("https://app.example/login", "http://attacker.example"))
	assert.Equal(t, "https://app.example", MailOrigin("https://app.example", "http://10.100.13.17"))
	assert.Equal(t, "http://127.0.0.1", MailOrigin("", "http://10.100.13.17/x"))
	assert.Equal(t, "http://127.0.0.1", MailOrigin("", "http://localhost"))
	assert.Equal(t, "http://127.0.0.1", MailOrigin("", "https://attacker.example"))
	assert.Equal(t, "http://127.0.0.1", MailOrigin("javascript:alert(1)", "https://evil.example"))
}

func TestRequestOriginAllowlistsProto(t *testing.T) {
	originOf := func(proto, host string) string {
		req := httptest.NewRequest(http.MethodGet, "http://"+host+"/", nil)
		req.Host = host
		if proto != "" {
			req.Header.Set("X-Forwarded-Proto", proto)
		}
		return RequestOrigin(req)
	}
	assert.Equal(t, "https://app.example", originOf("https, http", "app.example"))
	assert.Equal(t, "http://app.example", originOf("http", "app.example"))
	assert.Equal(t, "http://app.example", originOf("https://evil.example", "app.example"))
	assert.Equal(t, "http://app.example", originOf("javascript", "app.example"))
	assert.Equal(t, "", RequestOrigin(nil))

	httpsReq := httptest.NewRequest(http.MethodGet, "https://secure.example/", nil)
	httpsReq.Host = "secure.example"
	httpsReq.TLS = &tls.ConnectionState{}
	assert.Equal(t, "https://secure.example", RequestOrigin(httpsReq))
}
