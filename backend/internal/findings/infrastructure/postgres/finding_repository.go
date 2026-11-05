package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/vulnex/vulnex/backend/internal/findings/domain"
	"github.com/vulnex/vulnex/backend/internal/platform/postgres"
)

var (
	ErrFindingNotFound = errors.New("finding not found")
)

// FindingRepository implements domain.FindingRepository backed by PostgreSQL.
type FindingRepository struct {
	db postgres.DBTX
}

func NewFindingRepository(db postgres.DBTX) *FindingRepository {
	return &FindingRepository{db: db}
}

func (r *FindingRepository) Save(ctx context.Context, f *domain.Finding) error {
	query := `
		INSERT INTO findings (
			id, workspace_id, project_id, scan_id, asset_id, fingerprint,
			title, description, severity, confidence, status, source_scanner,
			first_seen_at, last_seen_at, risk_model_version, risk_score
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		f.ID().String(),
		f.WorkspaceID().String(),
		f.ProjectID().String(),
		f.ScanID().String(),
		f.AssetID().String(),
		f.Fingerprint().String(),
		f.Title(),
		f.Description(),
		string(f.Severity()),
		string(f.Confidence()),
		string(f.Status()),
		f.SourceScanner(),
		f.FirstSeenAt(),
		f.LastSeenAt(),
		f.RiskModelVersion(),
		f.RiskScore(),
	)
	if err != nil {
		return fmt.Errorf("failed to save finding: %w", err)
	}
	return nil
}

func (r *FindingRepository) FindByID(ctx context.Context, id domain.FindingID) (*domain.Finding, error) {
	query := `
		SELECT id, workspace_id, project_id, scan_id, asset_id, fingerprint,
		       title, description, severity, confidence, status, source_scanner,
		       first_seen_at, last_seen_at, risk_model_version, risk_score
		FROM findings
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id.String())
	return r.scanFinding(row)
}

func (r *FindingRepository) FindByFingerprint(
	ctx context.Context,
	projectID domain.ProjectID,
	fingerprint domain.FindingFingerprint,
) (*domain.Finding, error) {
	query := `
		SELECT id, workspace_id, project_id, scan_id, asset_id, fingerprint,
		       title, description, severity, confidence, status, source_scanner,
		       first_seen_at, last_seen_at, risk_model_version, risk_score
		FROM findings
		WHERE project_id = $1 AND fingerprint = $2
	`
	row := r.db.QueryRowContext(ctx, query, projectID.String(), fingerprint.String())
	finding, err := r.scanFinding(row)
	if err != nil {
		if errors.Is(err, ErrFindingNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return finding, nil
}

func (r *FindingRepository) ListByProject(ctx context.Context, projectID domain.ProjectID) ([]*domain.Finding, error) {
	query := `
		SELECT id, workspace_id, project_id, scan_id, asset_id, fingerprint,
		       title, description, severity, confidence, status, source_scanner,
		       first_seen_at, last_seen_at, risk_model_version, risk_score
		FROM findings
		WHERE project_id = $1
		ORDER BY risk_score DESC, last_seen_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, projectID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query findings: %w", err)
	}
	defer rows.Close()

	var findings []*domain.Finding
	for rows.Next() {
		var (
			idStr          string
			wsIDStr        string
			projIDStr      string
			scanIDStr      string
			assetIDStr     string
			fingerprintStr string
			title          string
			description    string
			severityStr    string
			confidenceStr  string
			statusStr      string
			sourceScanner  string
			firstSeenAt    time.Time
			lastSeenAt     time.Time
			riskModelVer   string
			riskScore      float64
		)

		err := rows.Scan(
			&idStr, &wsIDStr, &projIDStr, &scanIDStr, &assetIDStr, &fingerprintStr,
			&title, &description, &severityStr, &confidenceStr, &statusStr, &sourceScanner,
			&firstSeenAt, &lastSeenAt, &riskModelVer, &riskScore,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan finding row: %w", err)
		}

		findingID, err := domain.NewFindingID(idStr)
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
		scanID, err := domain.NewScanID(scanIDStr)
		if err != nil {
			return nil, err
		}
		assetID, err := domain.NewAssetID(assetIDStr)
		if err != nil {
			return nil, err
		}
		fp, err := domain.NewFindingFingerprint(fingerprintStr)
		if err != nil {
			return nil, err
		}

		findings = append(findings, domain.ReconstituteFinding(
			findingID,
			wsID,
			projID,
			scanID,
			assetID,
			fp,
			title,
			description,
			domain.Severity(severityStr),
			domain.Confidence(confidenceStr),
			domain.FindingStatus(statusStr),
			sourceScanner,
			firstSeenAt,
			lastSeenAt,
			riskModelVer,
			riskScore,
			"",
		))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("findings row iteration error: %w", err)
	}

	return findings, nil
}

func (r *FindingRepository) Update(ctx context.Context, f *domain.Finding) error {
	query := `
		UPDATE findings
		SET scan_id = $1, status = $2, last_seen_at = $3, risk_score = $4
		WHERE id = $5
	`
	res, err := r.db.ExecContext(
		ctx,
		query,
		f.ScanID().String(),
		string(f.Status()),
		f.LastSeenAt(),
		f.RiskScore(),
		f.ID().String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update finding: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrFindingNotFound
	}

	return nil
}

func (r *FindingRepository) scanFinding(row *sql.Row) (*domain.Finding, error) {
	var (
		idStr          string
		wsIDStr        string
		projIDStr      string
		scanIDStr      string
		assetIDStr     string
		fingerprintStr string
		title          string
		description    string
		severityStr    string
		confidenceStr  string
		statusStr      string
		sourceScanner  string
		firstSeenAt    time.Time
		lastSeenAt     time.Time
		riskModelVer   string
		riskScore      float64
	)

	err := row.Scan(
		&idStr, &wsIDStr, &projIDStr, &scanIDStr, &assetIDStr, &fingerprintStr,
		&title, &description, &severityStr, &confidenceStr, &statusStr, &sourceScanner,
		&firstSeenAt, &lastSeenAt, &riskModelVer, &riskScore,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFindingNotFound
		}
		return nil, fmt.Errorf("failed to scan finding: %w", err)
	}

	findingID, err := domain.NewFindingID(idStr)
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
	scanID, err := domain.NewScanID(scanIDStr)
	if err != nil {
		return nil, err
	}
	assetID, err := domain.NewAssetID(assetIDStr)
	if err != nil {
		return nil, err
	}
	fp, err := domain.NewFindingFingerprint(fingerprintStr)
	if err != nil {
		return nil, err
	}

	return domain.ReconstituteFinding(
		findingID,
		wsID,
		projID,
		scanID,
		assetID,
		fp,
		title,
		description,
		domain.Severity(severityStr),
		domain.Confidence(confidenceStr),
		domain.FindingStatus(statusStr),
		sourceScanner,
		firstSeenAt,
		lastSeenAt,
		riskModelVer,
		riskScore,
		"",
	), nil
}
