package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/vulnex/vulnex/backend/internal/identity/domain"
	"github.com/vulnex/vulnex/backend/internal/platform/postgres"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// UserRepository implements domain.UserRepository backed by PostgreSQL.
type UserRepository struct {
	db postgres.DBTX
}

func NewUserRepository(db postgres.DBTX) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Save(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email_normalized, password_hash, display_name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			display_name = EXCLUDED.display_name,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID().String(),
		user.Email().String(),
		user.PasswordHash(),
		user.DisplayName(),
		string(user.Status()),
		user.CreatedAt(),
		user.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	query := `
		SELECT id, email_normalized, password_hash, display_name, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id.String())
	return r.scanUser(row)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	query := `
		SELECT id, email_normalized, password_hash, display_name, status, created_at, updated_at
		FROM users
		WHERE email_normalized = $1
	`
	row := r.db.QueryRowContext(ctx, query, email.String())
	return r.scanUser(row)
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

func (r *UserRepository) scanUser(row *sql.Row) (*domain.User, error) {
	var (
		idStr        string
		emailStr     string
		passwordHash string
		displayName  string
		statusStr    string
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := row.Scan(&idStr, &emailStr, &passwordHash, &displayName, &statusStr, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	userID, err := domain.NewUserID(idStr)
	if err != nil {
		return nil, fmt.Errorf("corrupted user ID in DB: %w", err)
	}

	email, err := domain.NewEmail(emailStr)
	if err != nil {
		return nil, fmt.Errorf("corrupted email in DB: %w", err)
	}

	return domain.ReconstituteUser(
		userID,
		email,
		passwordHash,
		displayName,
		domain.UserStatus(statusStr),
		createdAt,
		updatedAt,
	), nil
}
