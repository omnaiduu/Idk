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
}
