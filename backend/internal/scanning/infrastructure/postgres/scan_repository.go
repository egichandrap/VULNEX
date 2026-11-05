package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/vulnex/vulnex/backend/internal/platform/postgres"
	"github.com/vulnex/vulnex/backend/internal/scanning/domain"
)

var (
	ErrScanNotFound = errors.New("scan not found")
)

// ScanRepository implements domain.ScanRepository backed by PostgreSQL.
type ScanRepository struct {
	db postgres.DBTX
}

func NewScanRepository(db postgres.DBTX) *ScanRepository {
	return &ScanRepository{db: db}
}

func (r *ScanRepository) Save(ctx context.Context, scan *domain.Scan) error {
	targetsJSON, err := json.Marshal(scan.TargetSnapshot())
	if err != nil {
		return fmt.Errorf("failed to marshal target snapshot: %w", err)
	}

	query := `
		INSERT INTO scans (
			id, workspace_id, project_id, authorization_id, scanner_profile_version_id,
			requested_by, status, idempotency_key, target_snapshot,
			started_at, completed_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err = r.db.ExecContext(
		ctx,
		query,
		scan.ID().String(),
		scan.WorkspaceID().String(),
		scan.ProjectID().String(),
		scan.AuthorizationID().String(),
		scan.ScannerProfileVersionID().String(),
		scan.RequestedBy().String(),
		string(scan.Status()),
		scan.IdempotencyKey(),
		targetsJSON,
		scan.StartedAt(),
		scan.CompletedAt(),
		scan.CreatedAt(),
		scan.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("failed to save scan: %w", err)
	}
	return nil
}

func (r *ScanRepository) FindByID(ctx context.Context, id domain.ScanID) (*domain.Scan, error) {
	query := `
		SELECT id, workspace_id, project_id, authorization_id, scanner_profile_version_id,
		       requested_by, status, idempotency_key, target_snapshot,
		       started_at, completed_at, created_at, updated_at
		FROM scans
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id.String())
	return r.scanScan(row)
}

func (r *ScanRepository) FindByWorkspace(ctx context.Context, workspaceID domain.WorkspaceID) ([]*domain.Scan, error) {
	query := `
		SELECT id, workspace_id, project_id, authorization_id, scanner_profile_version_id,
		       requested_by, status, idempotency_key, target_snapshot,
		       started_at, completed_at, created_at, updated_at
		FROM scans
		WHERE workspace_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, workspaceID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query scans: %w", err)
	}
	defer rows.Close()

	var list []*domain.Scan
	for rows.Next() {
		var (
			idStr           string
			wsIDStr         string
			projIDStr       string
			authIDStr       string
			profileVerIDStr string
			requestedByStr  string
			statusStr       string
			idempotencyKey  string
			targetsRaw      []byte
			startedAt       *time.Time
			completedAt     *time.Time
			createdAt       time.Time
			updatedAt       time.Time
		)

		err := rows.Scan(
			&idStr, &wsIDStr, &projIDStr, &authIDStr, &profileVerIDStr,
			&requestedByStr, &statusStr, &idempotencyKey, &targetsRaw,
			&startedAt, &completedAt, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan scan row: %w", err)
		}

		scanID, err := domain.NewScanID(idStr)
		if err != nil {
			return nil, err
		}
		wsID, err := domain.NewWorkspaceID(wsIDStr)
		if err != nil {
			return nil, err
		}
		projID, err := domain.NewProjectID(projIDStr)
		if err != nil {
			return nil, err
		}
		authID, err := domain.NewAuthorizationID(authIDStr)
		if err != nil {
			return nil, err
		}
		profileVerID, err := domain.NewProfileVersionID(profileVerIDStr)
		if err != nil {
			return nil, err
		}
		reqBy, err := domain.NewUserID(requestedByStr)
		if err != nil {
			return nil, err
		}

		var targets []string
		if len(targetsRaw) > 0 {
			_ = json.Unmarshal(targetsRaw, &targets)
		}

		list = append(list, domain.ReconstituteScan(
			scanID,
			wsID,
			projID,
			authID,
			profileVerID,
			reqBy,
			domain.ScanStatus(statusStr),
			idempotencyKey,
			targets,
			startedAt,
			completedAt,
			"", // failure reason
			createdAt,
			updatedAt,
		))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scans row iteration error: %w", err)
	}

	return list, nil
}

func (r *ScanRepository) Update(ctx context.Context, scan *domain.Scan) error {
	query := `
		UPDATE scans
		SET status = $1, started_at = $2, completed_at = $3, updated_at = $4
		WHERE id = $5
	`
	res, err := r.db.ExecContext(
		ctx,
		query,
		string(scan.Status()),
		scan.StartedAt(),
		scan.CompletedAt(),
		scan.UpdatedAt(),
		scan.ID().String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update scan: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrScanNotFound
	}

	return nil
}

func (r *ScanRepository) scanScan(row *sql.Row) (*domain.Scan, error) {
	var (
		idStr           string
		wsIDStr         string
		projIDStr       string
		authIDStr       string
		profileVerIDStr string
		requestedByStr  string
		statusStr       string
		idempotencyKey  string
		targetsRaw      []byte
		startedAt       *time.Time
		completedAt     *time.Time
		createdAt       time.Time
		updatedAt       time.Time
	)

	err := row.Scan(
		&idStr, &wsIDStr, &projIDStr, &authIDStr, &profileVerIDStr,
		&requestedByStr, &statusStr, &idempotencyKey, &targetsRaw,
		&startedAt, &completedAt, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrScanNotFound
		}
		return nil, fmt.Errorf("failed to scan scan aggregate: %w", err)
	}

	scanID, err := domain.NewScanID(idStr)
	if err != nil {
		return nil, err
	}
	wsID, err := domain.NewWorkspaceID(wsIDStr)
	if err != nil {
		return nil, err
	}
	projID, err := domain.NewProjectID(projIDStr)
	if err != nil {
		return nil, err
	}
	authID, err := domain.NewAuthorizationID(authIDStr)
	if err != nil {
		return nil, err
	}
	profileVerID, err := domain.NewProfileVersionID(profileVerIDStr)
	if err != nil {
		return nil, err
	}
	reqBy, err := domain.NewUserID(requestedByStr)
	if err != nil {
		return nil, err
	}

	var targets []string
	if len(targetsRaw) > 0 {
		_ = json.Unmarshal(targetsRaw, &targets)
	}

	return domain.ReconstituteScan(
		scanID,
		wsID,
		projID,
		authID,
		profileVerID,
		reqBy,
		domain.ScanStatus(statusStr),
		idempotencyKey,
		targets,
		startedAt,
		completedAt,
		"", // failure reason
		createdAt,
		updatedAt,
	), nil
}
