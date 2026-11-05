package domain

import "context"

// AssetRepository defines persistence operations for the Asset aggregate.
type AssetRepository interface {
	Save(ctx context.Context, asset *Asset) error
	FindByID(ctx context.Context, id AssetID) (*Asset, error)
	FindByCanonicalValue(ctx context.Context, projectID ProjectID, assetType AssetType, canonicalValue string) (*Asset, error)
	ListByProject(ctx context.Context, projectID ProjectID) ([]*Asset, error)
	Update(ctx context.Context, asset *Asset) error
}
