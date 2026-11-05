package application

import (
	"context"
	"errors"

	"github.com/vulnex/vulnex/backend/internal/identity/domain"
)

var (
	ErrUserNotFound             = errors.New("user not found")
	ErrIncorrectCurrentPassword = errors.New("current password is incorrect")
	ErrPasswordEmpty            = errors.New("new password cannot be empty")
)

type ChangePasswordCommand struct {
	UserID          string
	CurrentPassword string
	NewPassword     string
}

type ChangePasswordUseCase struct {
	userRepo domain.UserRepository
	hasher   domain.PasswordHasher
}

func NewChangePasswordUseCase(userRepo domain.UserRepository, hasher domain.PasswordHasher) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

func (uc *ChangePasswordUseCase) Execute(ctx context.Context, cmd ChangePasswordCommand) error {
	if cmd.NewPassword == "" {
		return ErrPasswordEmpty
	}

	userID, err := domain.NewUserID(cmd.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	if !uc.hasher.Compare(cmd.CurrentPassword, user.PasswordHash()) {
		return ErrIncorrectCurrentPassword
	}

	newHash, err := uc.hasher.Hash(cmd.NewPassword)
	if err != nil {
		return err
	}

	if err := user.ChangePassword(newHash); err != nil {
		return err
	}

	return uc.userRepo.Save(ctx, user)
}
