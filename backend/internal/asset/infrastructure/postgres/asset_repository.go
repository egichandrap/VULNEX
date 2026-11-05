package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/vulnex/vulnex/backend/internal/asset/domain"
	"github.com/vulnex/vulnex/backend/internal/platform/postgres"
)

var (
	ErrAssetNotFound = errors.New("asset not found")
)

// AssetRepository implements domain.AssetRepository backed by PostgreSQL.
type AssetRepository struct {
	db postgres.DBTX
}

func NewAssetRepository(db postgres.DBTX) *AssetRepository {
	return &AssetRepository{db: db}
}

func (r *AssetRepository) Save(ctx context.Context, asset *domain.Asset) error {
	query := `
		INSERT INTO assets (id, project_id, asset_type, canonical_value, display_value, criticality, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		asset.ID().String(),
		asset.ProjectID().String(),
		string(asset.AssetType()),
		asset.CanonicalValue().String(),
		asset.DisplayValue(),
		string(asset.Criticality()),
		string(asset.Status()),
		asset.CreatedAt(),
		asset.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("failed to save asset: %w", err)
	}
	return nil
}

func (r *AssetRepository) FindByID(ctx context.Context, id domain.AssetID) (*domain.Asset, error) {
	query := `
		SELECT id, project_id, asset_type, canonical_value, display_value, criticality, status, created_at, updated_at
		FROM assets
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id.String())
	return r.scanAsset(row)
}

func (r *AssetRepository) FindByCanonicalValue(
	ctx context.Context,
	projectID domain.ProjectID,
	assetType domain.AssetType,
	canonicalValue string,
) (*domain.Asset, error) {
	query := `
		SELECT id, project_id, asset_type, canonical_value, display_value, criticality, status, created_at, updated_at
		FROM assets
		WHERE project_id = $1 AND asset_type = $2 AND canonical_value = $3
	`
	row := r.db.QueryRowContext(ctx, query, projectID.String(), string(assetType), canonicalValue)
	return r.scanAsset(row)
}

func (r *AssetRepository) ListByProject(ctx context.Context, projectID domain.ProjectID) ([]*domain.Asset, error) {
	query := `
		SELECT id, project_id, asset_type, canonical_value, display_value, criticality, status, created_at, updated_at
		FROM assets
		WHERE project_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, projectID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query assets: %w", err)
	}
	defer rows.Close()

	var assets []*domain.Asset
	for rows.Next() {
		var (
			idStr        string
			projIDStr    string
			assetTypeStr string
			canonicalStr string
			displayVal   string
			critStr      string
			statusStr    string
			createdAt    time.Time
			updatedAt    time.Time
		)

		if err := rows.Scan(&idStr, &projIDStr, &assetTypeStr, &canonicalStr, &displayVal, &critStr, &statusStr, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan asset row: %w", err)
		}

		assetID, err := domain.NewAssetID(idStr)
		if err != nil {
			return nil, fmt.Errorf("corrupted asset ID: %w", err)
		}

		projID, err := domain.NewProjectID(projIDStr)
		if err != nil {
			return nil, fmt.Errorf("corrupted project ID: %w", err)
		}

		canonicalVal := domain.ReconstituteCanonicalAssetValue(canonicalStr)

		assets = append(assets, domain.ReconstituteAsset(
			assetID,
			projID,
			domain.AssetType(assetTypeStr),
			canonicalVal,
			displayVal,
			domain.AssetCriticality(critStr),
			domain.AssetStatus(statusStr),
			createdAt,
			updatedAt,
		))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("asset row iteration error: %w", err)
	}

	return assets, nil
}

func (r *AssetRepository) Update(ctx context.Context, asset *domain.Asset) error {
	query := `
		UPDATE assets
		SET display_value = $1, criticality = $2, status = $3, updated_at = $4
		WHERE id = $5
	`
	res, err := r.db.ExecContext(
		ctx,
		query,
		asset.DisplayValue(),
		string(asset.Criticality()),
		string(asset.Status()),
		asset.UpdatedAt(),
		asset.ID().String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update asset: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrAssetNotFound
	}

	return nil
}

func (r *AssetRepository) scanAsset(row *sql.Row) (*domain.Asset, error) {
	var (
		idStr        string
		projIDStr    string
		assetTypeStr string
		canonicalStr string
		displayVal   string
		critStr      string
		statusStr    string
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := row.Scan(&idStr, &projIDStr, &assetTypeStr, &canonicalStr, &displayVal, &critStr, &statusStr, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAssetNotFound
		}
		return nil, fmt.Errorf("failed to scan asset: %w", err)
	}

	assetID, err := domain.NewAssetID(idStr)
	if err != nil {
		return nil, fmt.Errorf("corrupted asset ID: %w", err)
	}

	projID, err := domain.NewProjectID(projIDStr)
	if err != nil {
		return nil, fmt.Errorf("corrupted project ID: %w", err)
	}

	canonicalVal := domain.ReconstituteCanonicalAssetValue(canonicalStr)

	return domain.ReconstituteAsset(
		assetID,
		projID,
		domain.AssetType(assetTypeStr),
		canonicalVal,
		displayVal,
		domain.AssetCriticality(critStr),
		domain.AssetStatus(statusStr),
		createdAt,
		updatedAt,
	), nil
}
