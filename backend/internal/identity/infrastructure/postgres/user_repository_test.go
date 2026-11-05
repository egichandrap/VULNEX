package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/vulnex/vulnex/backend/internal/identity/domain"
	platformPostgres "github.com/vulnex/vulnex/backend/internal/platform/postgres"
)

func TestUserRepository_Save_ExecutesExpectedSQL(t *testing.T) {
	var capturedQuery string
	var capturedArgs []any

	mockDB := &platformPostgres.MockDBTX{
		ExecFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
			capturedQuery = query
			capturedArgs = args
			return nil, nil
		},
	}

	repo := NewUserRepository(mockDB)
	email, _ := domain.NewEmail("security.lead@example.com")
	user, err := domain.NewUser(email, "$2a$10$hashedpw", "Security Lead")
	if err != nil {
		t.Fatalf("unexpected user creation error: %v", err)
	}

	err = repo.Save(context.Background(), user)
	if err != nil {
		t.Fatalf("unexpected repo save error: %v", err)
	}

	if !strings.Contains(capturedQuery, "INSERT INTO users") {
		t.Errorf("expected INSERT query, got: %s", capturedQuery)
	}

	if len(capturedArgs) != 7 {
		t.Fatalf("expected 7 bound arguments, got %d", len(capturedArgs))
	}

	if capturedArgs[1] != "security.lead@example.com" {
		t.Errorf("expected email param 'security.lead@example.com', got %v", capturedArgs[1])
	}
	if capturedArgs[3] != "Security Lead" {
		t.Errorf("expected display name param 'Security Lead', got %v", capturedArgs[3])
	}
}

func TestSetupTokenRepository_Save_ExecutesExpectedSQL(t *testing.T) {
	var capturedQuery string
	var capturedArgs []any

	mockDB := &platformPostgres.MockDBTX{
		ExecFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
			capturedQuery = query
			capturedArgs = args
			return nil, nil
		},
	}

	repo := NewSetupTokenRepository(mockDB)
	token, _, err := domain.GenerateSetupToken(30 * time.Minute)
	if err != nil {
		t.Fatalf("unexpected token generation error: %v", err)
	}

	err = repo.Save(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected repo save error: %v", err)
	}

	if !strings.Contains(capturedQuery, "INSERT INTO setup_tokens") {
		t.Errorf("expected INSERT query, got: %s", capturedQuery)
	}

	if len(capturedArgs) != 5 {
		t.Fatalf("expected 5 bound arguments, got %d", len(capturedArgs))
	}
}
