package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	platformPostgres "github.com/vulnex/vulnex/backend/internal/platform/postgres"
	"github.com/vulnex/vulnex/backend/internal/scanning/domain"
)

func TestScanRepository_Save_ExecutesExpectedSQL(t *testing.T) {
	var capturedQuery string
	var capturedArgs []any

	mockDB := &platformPostgres.MockDBTX{
		ExecFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
			capturedQuery = query
			capturedArgs = args
			return nil, nil
		},
	}

	repo := NewScanRepository(mockDB)
	scanID, _ := domain.GenerateScanID()
	wsID, _ := domain.GenerateWorkspaceID()
	projID, _ := domain.GenerateProjectID()
	authID, _ := domain.GenerateAuthorizationID()
	profID, _ := domain.GenerateProfileVersionID()
	userID, _ := domain.GenerateUserID()

	scan, err := domain.NewScan(
		scanID,
		wsID,
		projID,
		authID,
		profID,
		userID,
		"idem-test-key-1234",
		[]string{"api.vulnex.local"},
	)
	if err != nil {
		t.Fatalf("unexpected scan creation error: %v", err)
	}

	err = repo.Save(context.Background(), scan)
	if err != nil {
		t.Fatalf("unexpected repo save error: %v", err)
	}

	if !strings.Contains(capturedQuery, "INSERT INTO scans") {
		t.Errorf("expected INSERT INTO scans, got: %s", capturedQuery)
	}
	if len(capturedArgs) != 13 {
		t.Fatalf("expected 13 bound arguments, got %d", len(capturedArgs))
	}
}

func TestScanJobQueue_EnqueueAndClaimSQL(t *testing.T) {
	var capturedQuery string
	var capturedArgs []any

	mockDB := &platformPostgres.MockDBTX{
		ExecFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
			capturedQuery = query
			capturedArgs = args
			return nil, nil
		},
	}

	queue := NewScanJobQueue(mockDB)
	jobID, _ := domain.GenerateJobID()
	scanID, _ := domain.GenerateScanID()

	job := &domain.ScanJob{
		ID:          jobID,
		ScanID:      scanID,
		Status:      domain.ScanStatusQueued,
		Attempt:     0,
		AvailableAt: time.Now().UTC(),
	}

	err := queue.Enqueue(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected enqueue error: %v", err)
	}

	if !strings.Contains(capturedQuery, "INSERT INTO scan_jobs") {
		t.Errorf("expected INSERT INTO scan_jobs, got: %s", capturedQuery)
	}
	if len(capturedArgs) != 5 {
		t.Fatalf("expected 5 bound arguments, got %d", len(capturedArgs))
	}
}
