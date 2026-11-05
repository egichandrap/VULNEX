package domain

import (
	"context"
	"time"
)

// AuthorizationRepository defines persistence operations for the Authorization aggregate.
type AuthorizationRepository interface {
	Save(ctx context.Context, auth *Authorization) error
	FindByID(ctx context.Context, id AuthorizationID) (*Authorization, error)
	FindByProjectID(ctx context.Context, projectID ProjectID) ([]*Authorization, error)
	IsValidForExecution(ctx context.Context, id AuthorizationID, at time.Time) (bool, error)
	Update(ctx context.Context, auth *Authorization) error
}
