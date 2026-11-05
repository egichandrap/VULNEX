package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	platformPostgres "github.com/vulnex/vulnex/backend/internal/platform/postgres"
	"github.com/vulnex/vulnex/backend/internal/workspace/domain"
)

func TestWorkspaceRepository_Save_ExecutesExpectedSQL(t *testing.T) {
	var capturedQueries []string
	var capturedArgs [][]any

	mockDB := &platformPostgres.MockDBTX{
		ExecFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
			capturedQueries = append(capturedQueries, query)
			capturedArgs = append(capturedArgs, args)
			return nil, nil
		},
	}

	repo := NewWorkspaceRepository(mockDB)
	ownerID, _ := domain.NewUserID("00000000-0000-0000-0000-000000000001")
	ws, err := domain.NewWorkspace("SecOps Workspace", ownerID)
	if err != nil {
		t.Fatalf("unexpected workspace creation error: %v", err)
	}

	err = repo.Save(context.Background(), ws)
	if err != nil {
		t.Fatalf("unexpected repo save error: %v", err)
	}

	if len(capturedQueries) != 2 {
		t.Fatalf("expected 2 queries (workspace and owner member), got %d", len(capturedQueries))
	}

	if !strings.Contains(capturedQueries[0], "INSERT INTO workspaces") {
		t.Errorf("expected INSERT INTO workspaces, got: %s", capturedQueries[0])
	}
	if !strings.Contains(capturedQueries[1], "INSERT INTO workspace_members") {
		t.Errorf("expected INSERT INTO workspace_members, got: %s", capturedQueries[1])
	}
}
