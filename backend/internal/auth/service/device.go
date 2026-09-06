package service

import "strings"

// ParseUserAgent extracts a coarse browser / OS / device class from UA.
// No extra dependency: enough for the session admin list.
func ParseUserAgent(ua string) (browser, osName, device string) {
	if strings.TrimSpace(ua) == "" {
		return "Unknown", "Unknown", "Unknown"
	}
	l := strings.ToLower(ua)
	device = "Desktop"
	switch {
	case strings.Contains(l, "ipad") || strings.Contains(l, "tablet"):
		device = "Tablet"
	case strings.Contains(l, "mobi") || strings.Contains(l, "android") || strings.Contains(l, "iphone"):
		device = "Mobile"
	}

	switch {
	case strings.Contains(l, "edg/") || strings.Contains(l, "edge/"):
		browser = "Edge"
	case strings.Contains(l, "opr/") || strings.Contains(l, "opera"):
		browser = "Opera"
	case strings.Contains(l, "chrome/") && !strings.Contains(l, "edg"):
		browser = "Chrome"
	case strings.Contains(l, "firefox/") || strings.Contains(l, "fxios"):
		browser = "Firefox"
	case strings.Contains(l, "safari/") && !strings.Contains(l, "chrome") && !strings.Contains(l, "android"):
		browser = "Safari"
	default:
		browser = "Unknown"
	}

	switch {
	case strings.Contains(l, "windows"):
		osName = "Windows"
	case strings.Contains(l, "android"):
		osName = "Android"
	case strings.Contains(l, "iphone") || strings.Contains(l, "ipad") || strings.Contains(l, "ios"):
		osName = "iOS"
	case strings.Contains(l, "mac os") || strings.Contains(l, "macintosh"):
		osName = "macOS"
	case strings.Contains(l, "linux"):
		osName = "Linux"
	default:
		osName = "Unknown"
	}
	return browser, osName, device
}
