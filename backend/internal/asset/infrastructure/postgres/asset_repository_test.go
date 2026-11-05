package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/vulnex/vulnex/backend/internal/asset/domain"
	platformPostgres "github.com/vulnex/vulnex/backend/internal/platform/postgres"
)

func TestAssetRepository_Save_ExecutesExpectedSQL(t *testing.T) {
	var capturedQuery string
	var capturedArgs []any

	mockDB := &platformPostgres.MockDBTX{
		ExecFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
			capturedQuery = query
			capturedArgs = args
			return nil, nil
		},
	}

	repo := NewAssetRepository(mockDB)
	projectID, _ := domain.NewProjectID("00000000-0000-0000-0000-000000000002")
	canonicalVal := domain.ReconstituteCanonicalAssetValue("target.vulnex.local")

	asset, err := domain.NewAsset(
		projectID,
		domain.AssetTypeDomain,
		canonicalVal,
		"Target Domain",
		domain.CriticalityHigh,
	)
	if err != nil {
		t.Fatalf("unexpected asset creation error: %v", err)
	}

	err = repo.Save(context.Background(), asset)
	if err != nil {
		t.Fatalf("unexpected repo save error: %v", err)
	}

	if !strings.Contains(capturedQuery, "INSERT INTO assets") {
		t.Errorf("expected INSERT INTO assets, got: %s", capturedQuery)
	}
	if len(capturedArgs) != 9 {
		t.Fatalf("expected 9 bound arguments, got %d", len(capturedArgs))
	}
	if capturedArgs[3] != "target.vulnex.local" {
		t.Errorf("expected canonical value 'target.vulnex.local', got %v", capturedArgs[3])
	}
}
