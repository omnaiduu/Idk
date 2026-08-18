package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"tiny-gpu-bench/internal/origin"
)

func TestCORSAllowsLocalhost(t *testing.T) {
	h := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", origin.Loopback8741)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("code %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != origin.Loopback8741 {
		t.Fatalf("missing allow origin")
	}
}

func TestCORSRejectsForeign(t *testing.T) {
	h := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	}))
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "http://evil.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("code %d", rec.Code)
	}
}

func TestSPADoesNotJoinAbsoluteURLPath(t *testing.T) {
	dir := t.TempDir()
	h := spaHandler(dir)
	req := httptest.NewRequest(http.MethodGet, "/etc/passwd", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	// Missing index.html → 404 from ServeFile, not a dump of /etc/passwd.
	if rec.Body.String() == "root:" || rec.Code == http.StatusOK && rec.Body.Len() > 0 && rec.Header().Get("Content-Type") == "text/plain" {
		t.Fatalf("unexpected body %q", rec.Body.String())
	}
}
