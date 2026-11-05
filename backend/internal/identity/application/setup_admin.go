package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/vulnex/vulnex/backend/internal/identity/domain"
)

var (
	ErrAdminAlreadyConfigured = errors.New("initial administrator has already been bootstrapped")
	ErrEmptyPassword          = errors.New("password cannot be empty")
)

type SetupInitialAdminCommand struct {
	RawToken    string
	Email       string
	Password    string
	DisplayName string
}

type SetupInitialAdminResult struct {
	UserID string
	Email  string
}

// SetupInitialAdminUseCase implements the one-time admin bootstrap flow (ADR-013).
type SetupInitialAdminUseCase struct {
	userRepo  domain.UserRepository
	tokenRepo domain.SetupTokenRepository
	hasher    domain.PasswordHasher
}

func NewSetupInitialAdminUseCase(
	userRepo domain.UserRepository,
	tokenRepo domain.SetupTokenRepository,
	hasher domain.PasswordHasher,
) *SetupInitialAdminUseCase {
	return &SetupInitialAdminUseCase{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		hasher:    hasher,
	}
}

func (uc *SetupInitialAdminUseCase) Execute(ctx context.Context, cmd SetupInitialAdminCommand) (*SetupInitialAdminResult, error) {
	// Invariant: Bootstrap only allowed if 0 users exist
	count, err := uc.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user count: %w", err)
	}
	if count > 0 {
		return nil, ErrAdminAlreadyConfigured
	}

	if cmd.Password == "" {
		return nil, ErrEmptyPassword
	}

	// 1. Compute hash of provided raw token to find it
	hashBytes := sha256.Sum256([]byte(cmd.RawToken))
	providedHash := hex.EncodeToString(hashBytes[:])

	token, err := uc.tokenRepo.FindByHash(ctx, providedHash)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup setup token: %w", err)
	}
	if token == nil {
		return nil, domain.ErrInvalidSetupToken
	}

	// 2. Claim token (validates expiry and non-claimed state)
	now := time.Now().UTC()
	if err := token.Claim(cmd.RawToken, now); err != nil {
		return nil, err
	}

	// 3. Create user
	email, err := domain.NewEmail(cmd.Email)
	if err != nil {
		return nil, err
	}

	passwordHash, err := uc.hasher.Hash(cmd.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := domain.NewUser(email, passwordHash, cmd.DisplayName)
	if err != nil {
		return nil, err
	}

	// 4. Persist in repositories
	if err := uc.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save admin user: %w", err)
	}

	if err := uc.tokenRepo.Update(ctx, token); err != nil {
		return nil, fmt.Errorf("failed to update claimed setup token: %w", err)
	}

	return &SetupInitialAdminResult{
		UserID: user.ID().String(),
		Email:  user.Email().String(),
	}, nil
}
