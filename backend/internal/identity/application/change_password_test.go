package application

import (
	"context"
	"testing"

	"github.com/vulnex/vulnex/backend/internal/identity/domain"
)

func TestChangePasswordUseCase_Success(t *testing.T) {
	ctx := context.Background()
	userRepo := newFakeUserRepository()
	hasher := fakePasswordHasher{}

	email, _ := domain.NewEmail("admin@vulnex.local")
	hashedPassword, _ := hasher.Hash("admin")
	user, _ := domain.NewUser(email, hashedPassword, "Admin")
	_ = userRepo.Save(ctx, user)

	useCase := NewChangePasswordUseCase(userRepo, hasher)

	err := useCase.Execute(ctx, ChangePasswordCommand{
		UserID:          user.ID().String(),
		CurrentPassword: "admin",
		NewPassword:     "NewSecurePassword123!",
	})
	if err != nil {
		t.Fatalf("ChangePasswordUseCase failed: %v", err)
	}

	updatedUser, _ := userRepo.FindByID(ctx, user.ID())
	if !hasher.Compare("NewSecurePassword123!", updatedUser.PasswordHash()) {
		t.Errorf("Password was not updated to new hash")
	}
}

func TestChangePasswordUseCase_IncorrectCurrentPassword(t *testing.T) {
	ctx := context.Background()
	userRepo := newFakeUserRepository()
	hasher := fakePasswordHasher{}

	email, _ := domain.NewEmail("admin@vulnex.local")
	hashedPassword, _ := hasher.Hash("admin")
	user, _ := domain.NewUser(email, hashedPassword, "Admin")
	_ = userRepo.Save(ctx, user)

	useCase := NewChangePasswordUseCase(userRepo, hasher)

	err := useCase.Execute(ctx, ChangePasswordCommand{
		UserID:          user.ID().String(),
		CurrentPassword: "wrongpassword",
		NewPassword:     "NewSecurePassword123!",
	})
	if err != ErrIncorrectCurrentPassword {
		t.Errorf("expected ErrIncorrectCurrentPassword, got %v", err)
	}
}
