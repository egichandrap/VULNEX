package middleware

import (
	"context"
	"net/http"
	"strings"
)

const userIDContextKey contextKey = "vulnex_user_id"

// Authenticator verifies user credentials/tokens.
type Authenticator interface {
	ValidateToken(token string) (userID string, err error)
}

type AuthMiddleware struct {
	auth Authenticator
}

func NewAuthMiddleware(auth Authenticator) *AuthMiddleware {
	return &AuthMiddleware{auth: auth}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"missing or invalid authorization header"}}`, http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		userID, err := m.auth.ValidateToken(token)
		if err != nil || userID == "" {
			http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"invalid or expired token"}}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID extracts the authenticated user ID from the request context.
func GetUserID(r *http.Request) string {
	if val, ok := r.Context().Value(userIDContextKey).(string); ok {
		return val
	}
	return ""
}
