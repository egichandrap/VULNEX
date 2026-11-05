package domain

import "context"

// UserRepository defines persistence operations for the User aggregate.
type UserRepository interface {
	Save(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id UserID) (*User, error)
	FindByEmail(ctx context.Context, email Email) (*User, error)
	Count(ctx context.Context) (int, error)
}

// SetupTokenRepository defines persistence operations for SetupToken.
type SetupTokenRepository interface {
	Save(ctx context.Context, token *SetupToken) error
	FindByHash(ctx context.Context, tokenHash string) (*SetupToken, error)
	Update(ctx context.Context, token *SetupToken) error
}

// PasswordHasher abstracts password hashing and verification.
type PasswordHasher interface {
	Hash(plainPassword string) (string, error)
	Compare(plainPassword, passwordHash string) bool
}
