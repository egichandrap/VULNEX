package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/vulnex/vulnex/backend/internal/authorization/domain"
	platformPostgres "github.com/vulnex/vulnex/backend/internal/platform/postgres"
)

func TestAuthorizationRepository_Save_ExecutesExpectedSQL(t *testing.T) {
	var capturedQueries []string

	mockDB := &platformPostgres.MockDBTX{
		ExecFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
			capturedQueries = append(capturedQueries, query)
			return nil, nil
		},
	}

	repo := NewAuthorizationRepository(mockDB)
	authID, _ := domain.GenerateAuthorizationID()
	projectID, _ := domain.GenerateProjectID()
	approverID, _ := domain.GenerateUserID()

	auth, err := domain.NewAuthorization(
		authID,
		projectID,
		"Acme Cyber Audit",
		approverID,
		"s3://evidence/ref-01.pdf",
		"sha256-abcdef0123456789",
		time.Now().UTC(),
		time.Now().UTC().Add(48*time.Hour),
		[]string{"network_port_scan", "web_baseline"},
	)
	if err != nil {
		t.Fatalf("unexpected authorization creation error: %v", err)
	}

	auth.AddScopeRule(domain.ScopeRule{
		RuleType:        domain.RuleTypeInclude,
		TargetType:      domain.ScopeTargetDomain,
		CanonicalTarget: "api.vulnex.local",
	})

	err = repo.Save(context.Background(), auth)
	if err != nil {
		t.Fatalf("unexpected repo save error: %v", err)
	}

	if len(capturedQueries) != 2 {
		t.Fatalf("expected 2 queries (auth and scope rule), got %d", len(capturedQueries))
	}

	if !strings.Contains(capturedQueries[0], "INSERT INTO authorizations") {
		t.Errorf("expected INSERT INTO authorizations, got: %s", capturedQueries[0])
	}
	if !strings.Contains(capturedQueries[1], "INSERT INTO authorization_scopes") {
		t.Errorf("expected INSERT INTO authorization_scopes, got: %s", capturedQueries[1])
	}
}
