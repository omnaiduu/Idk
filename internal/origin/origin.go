// Package origin holds the localhost CORS / MCP Origin allowlist from PLAN.md.
package origin

import (
	"net/url"
	"os"
	"strings"
)

const (
	Localhost8741 = "http://localhost:8741"
	Loopback8741  = "http://127.0.0.1:8741"
)

// Allowed is the locked browser Origin set. Empty Origin (curl, Cursor MCP) is
// handled by the caller — this map is for non-empty Origin headers only.
var Allowed = map[string]bool{
	Loopback8741:  true,
	Localhost8741: true,
}

// TunnelEnabled reports whether EXTRA_ORIGINS / ALLOW_TUNNEL is on.
// Default remains localhost-only (PLAN.md). Tunnel mode is for short-lived tests.
func TunnelEnabled() bool {
	return os.Getenv("ALLOW_TUNNEL") == "1" || strings.TrimSpace(os.Getenv("EXTRA_ORIGINS")) != ""
}

// Allow reports whether a non-empty Origin is on the allowlist.
func Allow(origin string) bool {
	if Allowed[origin] {
		return true
	}
	for _, extra := range strings.Split(os.Getenv("EXTRA_ORIGINS"), ",") {
		extra = strings.TrimSpace(extra)
		if extra != "" && extra == origin {
			return true
		}
	}
	if os.Getenv("ALLOW_TUNNEL") != "1" {
		return false
	}
	u, err := url.Parse(origin)
	if err != nil || u.Scheme != "https" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return strings.HasSuffix(host, ".ngrok-free.app") ||
		strings.HasSuffix(host, ".ngrok.app") ||
		strings.HasSuffix(host, ".ngrok.io") ||
		strings.HasSuffix(host, ".ngrok-free.dev")
}
