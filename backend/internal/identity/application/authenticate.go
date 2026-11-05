package application

import (
	"context"
	"errors"

	"github.com/vulnex/vulnex/backend/internal/identity/domain"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserSuspended      = errors.New("user account is suspended")
)

type AuthenticateUserCommand struct {
	Email    string
	Password string
}

type AuthenticateUserResult struct {
	UserID      string
	Email       string
	DisplayName string
}

// AuthenticateUserUseCase handles credential verification and login flow.
type AuthenticateUserUseCase struct {
	userRepo domain.UserRepository
	hasher   domain.PasswordHasher
}

func NewAuthenticateUserUseCase(userRepo domain.UserRepository, hasher domain.PasswordHasher) *AuthenticateUserUseCase {
	return &AuthenticateUserUseCase{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

func (uc *AuthenticateUserUseCase) Execute(ctx context.Context, cmd AuthenticateUserCommand) (*AuthenticateUserResult, error) {
	email, err := domain.NewEmail(cmd.Email)
	if err != nil {
		// Generic error to avoid timing and format leak
		return nil, ErrInvalidCredentials
	}

	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		// User does not exist -> return generic credentials error
		return nil, ErrInvalidCredentials
	}

	if !uc.hasher.Compare(cmd.Password, user.PasswordHash()) {
		return nil, ErrInvalidCredentials
	}

	if user.Status() == domain.UserStatusSuspended {
		return nil, ErrUserSuspended
	}

	return &AuthenticateUserResult{
		UserID:      user.ID().String(),
		Email:       user.Email().String(),
		DisplayName: user.DisplayName(),
	}, nil
}
