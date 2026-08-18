// Package origin holds the localhost CORS / MCP Origin allowlist from PLAN.md.
package origin

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

// Allow reports whether a non-empty Origin is on the allowlist.
func Allow(origin string) bool {
	return Allowed[origin]
}
