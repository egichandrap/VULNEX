package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrustedProxyMiddleware_RejectsUntrustedProxyHeaders(t *testing.T) {
	// Only trust 127.0.0.1 and 10.0.0.1
	trustedCIDRs := []string{"127.0.0.1/32", "10.0.0.1/32"}
	mw, err := NewTrustedProxyMiddleware(trustedCIDRs)
	if err != nil {
		t.Fatalf("failed to create middleware: %v", err)
	}

	var capturedClientIP string
	handler := mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedClientIP = GetClientIP(r)
		w.WriteHeader(http.StatusOK)
	}))

	// Case 1: Untrusted peer (198.51.100.5) sends spoofed X-Forwarded-For header
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	req.RemoteAddr = "198.51.100.5:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.195")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Since peer 198.51.100.5 is NOT trusted, X-Forwarded-For must be ignored!
	if capturedClientIP != "198.51.100.5" {
		t.Errorf("expected client IP to be untrusted peer IP 198.51.100.5, got: %s", capturedClientIP)
	}

	// Case 2: Trusted peer (10.0.0.1) sends valid X-Forwarded-For
	req2 := httptest.NewRequest("GET", "/api/v1/test", nil)
	req2.RemoteAddr = "10.0.0.1:54321"
	req2.Header.Set("X-Forwarded-For", "203.0.113.195, 10.0.0.1")

	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	// Since peer 10.0.0.1 IS trusted, leftmost non-trusted client IP 203.0.113.195 is extracted
	if capturedClientIP != "203.0.113.195" {
		t.Errorf("expected client IP to be forwarded IP 203.0.113.195, got: %s", capturedClientIP)
	}
}

func TestTrustedProxyMiddleware_InvalidCIDRFails(t *testing.T) {
	_, err := NewTrustedProxyMiddleware([]string{"not-a-cidr"})
	if err == nil {
		t.Fatal("expected error for invalid CIDR, got nil")
	}
}
