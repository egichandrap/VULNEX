package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vulnex/vulnex/backend/internal/delivery/http/handlers"
	identityApp "github.com/vulnex/vulnex/backend/internal/identity/application"
	identityDomain "github.com/vulnex/vulnex/backend/internal/identity/domain"
)

type dummyAuthenticator struct{}

func (d *dummyAuthenticator) ValidateToken(token string) (string, error) {
	if token == "valid-token-123" {
		return "user-0001", nil
	}
	return "", nil
}

type fakeUserRepo struct {
	users map[string]*identityDomain.User
}

func (r *fakeUserRepo) Save(ctx context.Context, u *identityDomain.User) error {
	r.users[u.ID().String()] = u
	return nil
}
func (r *fakeUserRepo) FindByID(ctx context.Context, id identityDomain.UserID) (*identityDomain.User, error) {
	u, ok := r.users[id.String()]
	if !ok {
		return nil, identityDomain.ErrUserAlreadySuspended
	}
	return u, nil
}
func (r *fakeUserRepo) FindByEmail(ctx context.Context, e identityDomain.Email) (*identityDomain.User, error) {
	for _, u := range r.users {
		if u.Email().String() == e.String() {
			return u, nil
		}
	}
	return nil, nil
}
func (r *fakeUserRepo) Count(ctx context.Context) (int, error) {
	return len(r.users), nil
}

type fakeTokenRepo struct {
	token *identityDomain.SetupToken
}

func (r *fakeTokenRepo) Save(ctx context.Context, token *identityDomain.SetupToken) error {
	r.token = token
	return nil
}
func (r *fakeTokenRepo) FindByHash(ctx context.Context, hash string) (*identityDomain.SetupToken, error) {
	return r.token, nil
}
func (r *fakeTokenRepo) Update(ctx context.Context, token *identityDomain.SetupToken) error {
	r.token = token
	return nil
}

type dummyHasher struct{}

func (h *dummyHasher) Hash(p string) (string, error) { return "hashed-" + p, nil }
func (h *dummyHasher) Compare(p, hash string) bool    { return hash == "hashed-"+p }

func TestRouter_HealthCheckAndAuthEndpoints(t *testing.T) {
	userRepo := &fakeUserRepo{users: make(map[string]*identityDomain.User)}
	tokenRepo := &fakeTokenRepo{}
	hasher := &dummyHasher{}

	token, rawSecret, _ := identityDomain.GenerateSetupToken(1 * time.Hour)
	tokenRepo.token = token

	setupAdminUC := identityApp.NewSetupInitialAdminUseCase(userRepo, tokenRepo, hasher)
	authUC := identityApp.NewAuthenticateUserUseCase(userRepo, hasher)

	authHandler := handlers.NewAuthHandler(setupAdminUC, authUC, userRepo)

	router, err := NewRouter(RouterConfig{
		TrustedProxies: []string{"127.0.0.1/32"},
		AuthHandler:    authHandler,
		Workspace:      nil,
		Asset:          nil,
		Scan:           nil,
		Finding:        nil,
		Authenticator:  &dummyAuthenticator{},
	})
	if err != nil {
		t.Fatalf("failed to create router: %v", err)
	}

	// 1. Healthz check
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK on /healthz, got %d", w.Code)
	}

	// 2. Setup Initial Admin
	setupPayload := map[string]string{
		"setup_token":  rawSecret,
		"email":        "admin@vulnex.local",
		"password":     "AdminMasterPassword123!",
		"display_name": "Root Administrator",
	}
	body, _ := json.Marshal(setupPayload)
	req2 := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on /api/v1/auth/setup, got %d: %s", w2.Code, w2.Body.String())
	}

	// 3. Login with newly created admin credentials
	loginPayload := map[string]string{
		"email":    "admin@vulnex.local",
		"password": "AdminMasterPassword123!",
	}
	loginBody, _ := json.Marshal(loginPayload)
	req3 := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginBody))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on /api/v1/auth/login, got %d: %s", w3.Code, w3.Body.String())
	}

	// 4. Access protected /api/v1/auth/me with invalid token -> 401 Unauthorized
	req4 := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req4.Header.Set("Authorization", "Bearer invalid-token")
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)

	if w4.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized on invalid token, got %d", w4.Code)
	}
}
