package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckWSOrigin(t *testing.T) {
	prodAllowed := map[string]bool{
		"https://starbyte.smbu.edu.cn": true,
		"https://starbyte.com":         true,
	}

	tests := []struct {
		name    string
		allowed map[string]bool
		origin  string
		host    string
		want    bool
	}{
		{
			name:    "intranet IP origin matches Host",
			allowed: prodAllowed,
			origin:  "http://10.100.13.17",
			host:    "10.100.13.17",
			want:    true,
		},
		{
			name:    "intranet IP origin matches Host ignoring proxy port",
			allowed: prodAllowed,
			origin:  "http://10.100.13.17",
			host:    "10.100.13.17:8080",
			want:    true,
		},
		{
			name:    "configured CORS origin",
			allowed: prodAllowed,
			origin:  "https://starbyte.smbu.edu.cn",
			host:    "backend.internal",
			want:    true,
		},
		{
			name:    "empty origin allowed",
			allowed: prodAllowed,
			origin:  "",
			host:    "10.100.13.17",
			want:    true,
		},
		{
			name:    "dev empty allowlist permits any origin",
			allowed: map[string]bool{},
			origin:  "https://evil.example",
			host:    "10.100.13.17",
			want:    true,
		},
		{
			name:    "cross-site origin rejected",
			allowed: prodAllowed,
			origin:  "https://evil.example",
			host:    "10.100.13.17",
			want:    false,
		},
		{
			name:    "malformed origin rejected",
			allowed: prodAllowed,
			origin:  "://not-a-url",
			host:    "10.100.13.17",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/ws/notifications", nil)
			r.Host = tt.host
			if tt.origin != "" {
				r.Header.Set("Origin", tt.origin)
			}
			assert.Equal(t, tt.want, checkWSOrigin(r, tt.allowed))
		})
	}
}

func TestWSHandlerCheckOrigin_IPMatchesHost(t *testing.T) {
	h := NewWSHandler(nil, nil, []string{"https://starbyte.smbu.edu.cn"})
	r := httptest.NewRequest(http.MethodGet, "/ws/notifications", nil)
	r.Host = "10.100.13.17"
	r.Header.Set("Origin", "http://10.100.13.17")
	assert.True(t, h.upgrader.CheckOrigin(r))
}
