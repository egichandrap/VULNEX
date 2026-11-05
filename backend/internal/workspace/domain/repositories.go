package domain

import "context"

// WorkspaceRepository defines persistence operations for the Workspace aggregate.
type WorkspaceRepository interface {
	Save(ctx context.Context, ws *Workspace) error
	FindByID(ctx context.Context, id WorkspaceID) (*Workspace, error)
	FindByUserID(ctx context.Context, userID UserID) ([]*Workspace, error)
	Update(ctx context.Context, ws *Workspace) error
}
