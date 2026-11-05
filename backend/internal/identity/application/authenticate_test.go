package application

import (
	"context"
	"testing"

	"github.com/vulnex/vulnex/backend/internal/identity/domain"
)

func TestAuthenticateUserUseCase_Success(t *testing.T) {
	ctx := context.Background()
	userRepo := newFakeUserRepository()
	hasher := fakePasswordHasher{}

	email, _ := domain.NewEmail("analyst@vulnex.local")
	hashedPassword, _ := hasher.Hash("CorrectPassword123!")
	user, _ := domain.NewUser(email, hashedPassword, "Security Analyst")
	_ = userRepo.Save(ctx, user)

	useCase := NewAuthenticateUserUseCase(userRepo, hasher)

	result, err := useCase.Execute(ctx, AuthenticateUserCommand{
		Email:    "analyst@vulnex.local",
		Password: "CorrectPassword123!",
	})

	if err != nil {
		t.Fatalf("AuthenticateUserUseCase failed: %v", err)
	}

	if result.UserID != user.ID().String() {
		t.Errorf("UserID mismatch: got %s, want %s", result.UserID, user.ID().String())
	}
	if result.Email != "analyst@vulnex.local" {
		t.Errorf("Email mismatch: got %s", result.Email)
	}
}

func TestAuthenticateUserUseCase_InvalidCredentials(t *testing.T) {
	ctx := context.Background()
	userRepo := newFakeUserRepository()
	hasher := fakePasswordHasher{}

	email, _ := domain.NewEmail("analyst@vulnex.local")
	hashedPassword, _ := hasher.Hash("CorrectPassword123!")
	user, _ := domain.NewUser(email, hashedPassword, "Security Analyst")
	_ = userRepo.Save(ctx, user)

	useCase := NewAuthenticateUserUseCase(userRepo, hasher)

	// Case 1: Wrong password
	_, err := useCase.Execute(ctx, AuthenticateUserCommand{
		Email:    "analyst@vulnex.local",
		Password: "WrongPassword!",
	})
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials for wrong password, got: %v", err)
	}

	// Case 2: Non-existent email (must not reveal user absence)
	_, err = useCase.Execute(ctx, AuthenticateUserCommand{
		Email:    "nonexistent@vulnex.local",
		Password: "SomePassword!",
	})
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials for non-existent user, got: %v", err)
	}
}

func TestAuthenticateUserUseCase_SuspendedUser(t *testing.T) {
	ctx := context.Background()
	userRepo := newFakeUserRepository()
	hasher := fakePasswordHasher{}

	email, _ := domain.NewEmail("suspended@vulnex.local")
	hashedPassword, _ := hasher.Hash("Password123!")
	user, _ := domain.NewUser(email, hashedPassword, "Suspended User")
	user.Suspend()
	_ = userRepo.Save(ctx, user)

	useCase := NewAuthenticateUserUseCase(userRepo, hasher)

	_, err := useCase.Execute(ctx, AuthenticateUserCommand{
		Email:    "suspended@vulnex.local",
		Password: "Password123!",
	})
	if err != ErrUserSuspended {
		t.Errorf("Expected ErrUserSuspended, got: %v", err)
	}
}
