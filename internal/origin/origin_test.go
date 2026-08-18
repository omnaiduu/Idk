package origin

import "testing"

func TestAllow(t *testing.T) {
	if !Allow(Loopback8741) || !Allow(Localhost8741) {
		t.Fatal("localhost origins must be allowed")
	}
	if Allow("http://evil.example") || Allow("https://127.0.0.1:8741") {
		t.Fatal("foreign origins must be denied")
	}
	if Allow("") {
		t.Fatal("empty origin is not in the map; callers treat it separately")
	}
	if Allow("https://abc.ngrok-free.app") {
		t.Fatal("ngrok origin denied unless ALLOW_TUNNEL=1")
	}
}

func TestAllowTunnelNgrok(t *testing.T) {
	t.Setenv("ALLOW_TUNNEL", "1")
	if !Allow("https://abc.ngrok-free.app") {
		t.Fatal("ngrok origin should be allowed when ALLOW_TUNNEL=1")
	}
	if Allow("http://abc.ngrok-free.app") {
		t.Fatal("http ngrok should still be denied")
	}
	if Allow("https://evil.example") {
		t.Fatal("unrelated https still denied")
	}
}
