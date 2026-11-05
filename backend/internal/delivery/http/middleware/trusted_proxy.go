package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
)

type contextKey string

const clientIPKey contextKey = "vulnex_client_ip"

// TrustedProxyMiddleware ensures X-Forwarded-For is only parsed from explicitly trusted CIDRs (ADR-014).
type TrustedProxyMiddleware struct {
	trustedNets []*net.IPNet
}

func NewTrustedProxyMiddleware(cidrs []string) (*TrustedProxyMiddleware, error) {
	var nets []*net.IPNet
	for _, cidr := range cidrs {
		trimmed := strings.TrimSpace(cidr)
		if trimmed == "" {
			continue
		}
		if !strings.Contains(trimmed, "/") {
			if strings.Contains(trimmed, ":") {
				trimmed += "/128"
			} else {
				trimmed += "/32"
			}
		}
		_, ipNet, err := net.ParseCIDR(trimmed)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy CIDR %q: %w", cidr, err)
		}
		nets = append(nets, ipNet)
	}

	return &TrustedProxyMiddleware{trustedNets: nets}, nil
}

func (m *TrustedProxyMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := m.resolveClientIP(r)
		ctx := context.WithValue(r.Context(), clientIPKey, clientIP)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *TrustedProxyMiddleware) resolveClientIP(r *http.Request) string {
	peerHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		peerHost = r.RemoteAddr
	}

	peerIP := net.ParseIP(peerHost)
	if peerIP == nil {
		return peerHost
	}

	// Verify if the direct peer is a trusted proxy
	isTrusted := false
	for _, trustedNet := range m.trustedNets {
		if trustedNet.Contains(peerIP) {
			isTrusted = true
			break
		}
	}

	if !isTrusted {
		// Peer is NOT trusted: ignore X-Forwarded-For to prevent spoofing
		return peerHost
	}

	// Peer IS trusted: inspect X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff == "" {
		return peerHost
	}

	parts := strings.Split(xff, ",")
	for _, part := range parts {
		candidate := strings.TrimSpace(part)
		if candidateIP := net.ParseIP(candidate); candidateIP != nil {
			// Return the first valid forwarded IP
			return candidateIP.String()
		}
	}

	return peerHost
}

// GetClientIP extracts the resolved and validated client IP from request context.
func GetClientIP(r *http.Request) string {
	if val, ok := r.Context().Value(clientIPKey).(string); ok && val != "" {
		return val
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
