package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseUserAgent(t *testing.T) {
	cases := []struct {
		ua      string
		browser string
		os      string
		device  string
	}{
		{"", "Unknown", "Unknown", "Unknown"},
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
			"Chrome", "Windows", "Desktop",
		},
		{
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
			"Safari", "macOS", "Desktop",
		},
		{
			"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			"Safari", "iOS", "Mobile",
		},
		{
			"Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Mobile Safari/537.36",
			"Chrome", "Android", "Mobile",
		},
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36 Edg/128.0.0.0",
			"Edge", "Windows", "Desktop",
		},
		{
			"Mozilla/5.0 (X11; Linux x86_64; rv:129.0) Gecko/20100101 Firefox/129.0",
			"Firefox", "Linux", "Desktop",
		},
		{
			"Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			"Safari", "iOS", "Tablet",
		},
	}
	for _, tc := range cases {
		b, o, d := ParseUserAgent(tc.ua)
		assert.Equal(t, tc.browser, b, tc.ua)
		assert.Equal(t, tc.os, o, tc.ua)
		assert.Equal(t, tc.device, d, tc.ua)
	}
}
