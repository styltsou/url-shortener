package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	handler := RateLimit(RateLimitConfig{
		Requests: 2,
		Window:   time.Minute,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.0.2.1:1234"
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("request %d status = %d, want %d", i+1, rec.Code, http.StatusNoContent)
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("limited status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatalf("Retry-After header was not set")
	}
	if rec.Header().Get("X-RateLimit-Limit") != "2" || rec.Header().Get("X-RateLimit-Remaining") != "0" || rec.Header().Get("X-RateLimit-Reset") == "" {
		t.Fatalf("rate limit headers missing or incorrect: %v", rec.Header())
	}
}

func TestRateLimitIgnoresForgedForwardingHeaders(t *testing.T) {
	var hits int
	handler := RateLimit(RateLimitConfig{
		Requests: 1,
		Window:   time.Minute,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusNoContent)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	req1.RemoteAddr = "198.51.100.10:1234"
	req1.Header.Set("X-Forwarded-For", "192.0.2.1, 198.51.100.1")
	handler.ServeHTTP(httptest.NewRecorder(), req1)

	rec := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	req2.RemoteAddr = "198.51.100.10:5678"
	req2.Header.Set("X-Forwarded-For", "203.0.113.99")
	req2.Header.Set("X-Real-IP", "203.0.113.98")
	handler.ServeHTTP(rec, req2)

	if hits != 1 {
		t.Fatalf("hits = %d, want 1", hits)
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimitTrustedProxyChain(t *testing.T) {
	handler := RateLimit(RateLimitConfig{
		Requests:       1,
		Window:         time.Minute,
		TrustedProxies: []netip.Prefix{netip.MustParsePrefix("192.0.2.0/24")},
	})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	request := func(chain string) int {
		req := httptest.NewRequest(http.MethodGet, "/code", nil)
		req.RemoteAddr = "192.0.2.2:443"
		req.Header.Set("X-Forwarded-For", chain)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}
	if got := request("198.51.100.1, 192.0.2.1"); got != http.StatusNoContent {
		t.Fatalf("first request status = %d", got)
	}
	// A client can prepend a forged address, but cannot change the first
	// untrusted hop selected from the right side of the chain.
	if got := request("203.0.113.99, 198.51.100.1, 192.0.2.1"); got != http.StatusTooManyRequests {
		t.Fatalf("forged chain status = %d, want 429", got)
	}
	if got := request("198.51.100.2, 192.0.2.1"); got != http.StatusNoContent {
		t.Fatalf("different client status = %d", got)
	}
}

func TestRateLimitSeparatesRedirectsFromAPI(t *testing.T) {
	handler := RateLimit(RateLimitConfig{Requests: 1, Window: time.Minute})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	for _, test := range []struct {
		path string
		want int
	}{
		{"/code", http.StatusNoContent},
		{"/api/v1/links", http.StatusNoContent},
		{"/code", http.StatusTooManyRequests},
	} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, test.path, nil))
		if rec.Code != test.want {
			t.Fatalf("%s status = %d, want %d", test.path, rec.Code, test.want)
		}
	}
}
