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
	ErrSetupTokenNotFound = errors.New("setup token not found")
)

// SetupTokenRepository implements domain.SetupTokenRepository backed by PostgreSQL.
type SetupTokenRepository struct {
	db postgres.DBTX
}

func NewSetupTokenRepository(db postgres.DBTX) *SetupTokenRepository {
	return &SetupTokenRepository{db: db}
}

func (r *SetupTokenRepository) Save(ctx context.Context, token *domain.SetupToken) error {
	query := `
		INSERT INTO setup_tokens (id, token_hash, expires_at, used_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		token.ID().String(),
		token.TokenHash(),
		token.ExpiresAt(),
		token.UsedAt(),
		token.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("failed to save setup token: %w", err)
	}
	return nil
}

func (r *SetupTokenRepository) FindByHash(ctx context.Context, tokenHash string) (*domain.SetupToken, error) {
	query := `
		SELECT id, token_hash, expires_at, used_at, created_at
		FROM setup_tokens
		WHERE token_hash = $1
	`
	row := r.db.QueryRowContext(ctx, query, tokenHash)

	var (
		idStr     string
		hash      string
		expiresAt time.Time
		usedAt    *time.Time
		createdAt time.Time
	)

	err := row.Scan(&idStr, &hash, &expiresAt, &usedAt, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSetupTokenNotFound
		}
		return nil, fmt.Errorf("failed to scan setup token: %w", err)
	}

	tokenID, err := domain.NewSetupTokenID(idStr)
	if err != nil {
		return nil, fmt.Errorf("corrupted setup token ID in DB: %w", err)
	}
	return domain.ReconstituteSetupToken(tokenID, hash, expiresAt, usedAt, createdAt), nil
}

func (r *SetupTokenRepository) Update(ctx context.Context, token *domain.SetupToken) error {
	query := `
		UPDATE setup_tokens
		SET used_at = $1
		WHERE id = $2
	`
	res, err := r.db.ExecContext(ctx, query, token.UsedAt(), token.ID().String())
	if err != nil {
		return fmt.Errorf("failed to update setup token: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrSetupTokenNotFound
	}
	return nil
}
