package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/vulnex/vulnex/backend/internal/identity/domain"
)

// In-memory fake implementations for test isolation
type fakeUserRepository struct {
	users map[string]*domain.User
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{users: make(map[string]*domain.User)}
}

func (f *fakeUserRepository) Save(ctx context.Context, u *domain.User) error {
	f.users[u.ID().String()] = u
	return nil
}

func (f *fakeUserRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	u, ok := f.users[id.String()]
	if !ok {
		return nil, domain.ErrInvalidEmail // or not found
	}
	return u, nil
}

func (f *fakeUserRepository) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	for _, u := range f.users {
		if u.Email().String() == email.String() {
			return u, nil
		}
	}
	return nil, nil
}

func (f *fakeUserRepository) Count(ctx context.Context) (int, error) {
	return len(f.users), nil
}

type fakeSetupTokenRepository struct {
	tokens map[string]*domain.SetupToken
}

func newFakeSetupTokenRepository() *fakeSetupTokenRepository {
	return &fakeSetupTokenRepository{tokens: make(map[string]*domain.SetupToken)}
}

func (f *fakeSetupTokenRepository) Save(ctx context.Context, token *domain.SetupToken) error {
	f.tokens[token.TokenHash()] = token
	return nil
}

func (f *fakeSetupTokenRepository) FindByHash(ctx context.Context, tokenHash string) (*domain.SetupToken, error) {
	t, ok := f.tokens[tokenHash]
	if !ok {
		return nil, nil
	}
	return t, nil
}

func (f *fakeSetupTokenRepository) Update(ctx context.Context, token *domain.SetupToken) error {
	f.tokens[token.TokenHash()] = token
	return nil
}

type fakePasswordHasher struct{}

func (f fakePasswordHasher) Hash(plain string) (string, error) {
	sum := sha256.Sum256([]byte("hashed:" + plain))
	return hex.EncodeToString(sum[:]), nil
}

func (f fakePasswordHasher) Compare(plain, hash string) bool {
	sum := sha256.Sum256([]byte("hashed:" + plain))
	return hex.EncodeToString(sum[:]) == hash
}

func TestSetupInitialAdminUseCase_Success(t *testing.T) {
	ctx := context.Background()
	userRepo := newFakeUserRepository()
	tokenRepo := newFakeSetupTokenRepository()
	hasher := fakePasswordHasher{}

	// Pre-seed an active setup token
	token, rawSecret, err := domain.GenerateSetupToken(24 * time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	_ = tokenRepo.Save(ctx, token)

	useCase := NewSetupInitialAdminUseCase(userRepo, tokenRepo, hasher)

	cmd := SetupInitialAdminCommand{
		RawToken:    rawSecret,
		Email:       "admin@vulnex.local",
		Password:    "SuperSecureAdminPassphrase123!",
		DisplayName: "Lead Security Architect",
	}

	result, err := useCase.Execute(ctx, cmd)
	if err != nil {
		t.Fatalf("SetupInitialAdminUseCase failed: %v", err)
	}

	if result.UserID == "" {
		t.Errorf("Expected result to have valid UserID")
	}

	// Verify user is persisted
	count, _ := userRepo.Count(ctx)
	if count != 1 {
		t.Errorf("Expected 1 user in repo, got %d", count)
	}

	// Verify token is claimed
	claimedToken, _ := tokenRepo.FindByHash(ctx, token.TokenHash())
	if !claimedToken.IsClaimed() {
		t.Errorf("Setup token should be marked claimed")
	}

	// Invariant: Second attempt must fail with ErrAdminAlreadyConfigured
	_, err = useCase.Execute(ctx, cmd)
	if err != ErrAdminAlreadyConfigured {
		t.Errorf("Expected ErrAdminAlreadyConfigured on second run, got: %v", err)
	}
}

func TestSetupInitialAdminUseCase_InvalidTokenFails(t *testing.T) {
	ctx := context.Background()
	userRepo := newFakeUserRepository()
	tokenRepo := newFakeSetupTokenRepository()
	hasher := fakePasswordHasher{}

	useCase := NewSetupInitialAdminUseCase(userRepo, tokenRepo, hasher)

	cmd := SetupInitialAdminCommand{
		RawToken:    "invalid-bogus-token",
		Email:       "admin@vulnex.local",
		Password:    "password123",
		DisplayName: "Admin",
	}

	_, err := useCase.Execute(ctx, cmd)
	if err != domain.ErrInvalidSetupToken {
		t.Errorf("Expected ErrInvalidSetupToken, got: %v", err)
	}
}
