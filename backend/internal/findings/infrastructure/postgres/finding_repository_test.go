package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/vulnex/vulnex/backend/internal/findings/domain"
	platformPostgres "github.com/vulnex/vulnex/backend/internal/platform/postgres"
)

func TestFindingRepository_Save_ExecutesExpectedSQL(t *testing.T) {
	var capturedQuery string
	var capturedArgs []any

	mockDB := &platformPostgres.MockDBTX{
		ExecFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
			capturedQuery = query
			capturedArgs = args
			return nil, nil
		},
	}

	repo := NewFindingRepository(mockDB)
	findingID, _ := domain.GenerateFindingID()
	wsID, _ := domain.GenerateWorkspaceID()
	projID, _ := domain.GenerateProjectID()
	scanID, _ := domain.GenerateScanID()
	assetID, _ := domain.GenerateAssetID()
	fingerprint := domain.CalculateFingerprint(projID.String(), "target.vulnex.local", "nmap", "port-80", "tcp/80")

	finding, err := domain.NewFinding(
		findingID,
		wsID,
		projID,
		scanID,
		assetID,
		fingerprint,
		"Open HTTP Port",
		"Insecure HTTP port 80 detected",
		domain.SeverityMedium,
		domain.ConfidenceHigh,
		"nmap",
		time.Now().UTC(),
		5.0,
		"v1.0.0",
	)
	if err != nil {
		t.Fatalf("unexpected finding creation error: %v", err)
	}

	err = repo.Save(context.Background(), finding)
	if err != nil {
		t.Fatalf("unexpected repo save error: %v", err)
	}

	if !strings.Contains(capturedQuery, "INSERT INTO findings") {
		t.Errorf("expected INSERT INTO findings, got: %s", capturedQuery)
	}
	if len(capturedArgs) != 16 {
		t.Fatalf("expected 16 bound arguments, got %d", len(capturedArgs))
	}
	if capturedArgs[6] != "Open HTTP Port" {
		t.Errorf("expected title 'Open HTTP Port', got %v", capturedArgs[6])
	}
}
