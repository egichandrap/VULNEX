package domain

import "context"

// FindingRepository defines persistence operations for the Finding aggregate.
type FindingRepository interface {
	Save(ctx context.Context, finding *Finding) error
	FindByID(ctx context.Context, id FindingID) (*Finding, error)
	FindByFingerprint(ctx context.Context, projectID ProjectID, fingerprint FindingFingerprint) (*Finding, error)
	ListByProject(ctx context.Context, projectID ProjectID) ([]*Finding, error)
	Update(ctx context.Context, finding *Finding) error
}
