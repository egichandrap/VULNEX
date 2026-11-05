package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/vulnex/vulnex/backend/internal/platform/postgres"
	"github.com/vulnex/vulnex/backend/internal/workspace/domain"
)

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
)

// WorkspaceRepository implements domain.WorkspaceRepository backed by PostgreSQL.
type WorkspaceRepository struct {
	db postgres.DBTX
}

func NewWorkspaceRepository(db postgres.DBTX) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

func (r *WorkspaceRepository) Save(ctx context.Context, ws *domain.Workspace) error {
	query := `
		INSERT INTO workspaces (id, name, owner_user_id, allow_private_network_scan, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		ws.ID().String(),
		ws.Name(),
		ws.OwnerUserID().String(),
		ws.AllowPrivateNetworkScan(),
		string(ws.Status()),
		ws.CreatedAt(),
		ws.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("failed to save workspace: %w", err)
	}

	// Save members
	for _, m := range ws.Members() {
		memberQuery := `
			INSERT INTO workspace_members (id, workspace_id, user_id, role, created_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (workspace_id, user_id) DO UPDATE SET role = EXCLUDED.role
		`
		_, err := r.db.ExecContext(
			ctx,
			memberQuery,
			m.ID(),
			ws.ID().String(),
			m.UserID().String(),
			string(m.Role()),
			m.CreatedAt(),
		)
		if err != nil {
			return fmt.Errorf("failed to save workspace member: %w", err)
		}
	}

	return nil
}

func (r *WorkspaceRepository) FindByID(ctx context.Context, id domain.WorkspaceID) (*domain.Workspace, error) {
	query := `
		SELECT id, name, owner_user_id, allow_private_network_scan, status, created_at, updated_at
		FROM workspaces
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id.String())

	var (
		idStr                   string
		name                    string
		ownerIDStr              string
		allowPrivateNetworkScan bool
		statusStr               string
		createdAt               time.Time
		updatedAt               time.Time
	)

	err := row.Scan(&idStr, &name, &ownerIDStr, &allowPrivateNetworkScan, &statusStr, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWorkspaceNotFound
		}
		return nil, fmt.Errorf("failed to scan workspace: %w", err)
	}

	wsID, err := domain.NewWorkspaceID(idStr)
	if err != nil {
		return nil, fmt.Errorf("corrupted workspace ID: %w", err)
	}

	ownerID, err := domain.NewUserID(ownerIDStr)
	if err != nil {
		return nil, fmt.Errorf("corrupted owner ID: %w", err)
	}

	members, err := r.loadMembers(ctx, wsID)
	if err != nil {
		return nil, err
	}

	return domain.ReconstituteWorkspace(
		wsID,
		name,
		ownerID,
		allowPrivateNetworkScan,
		domain.WorkspaceStatus(statusStr),
		members,
		createdAt,
		updatedAt,
	), nil
}

func (r *WorkspaceRepository) FindByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Workspace, error) {
	query := `
		SELECT w.id, w.name, w.owner_user_id, w.allow_private_network_scan, w.status, w.created_at, w.updated_at
		FROM workspaces w
		JOIN workspace_members wm ON w.id = wm.workspace_id
		WHERE wm.user_id = $1
		ORDER BY w.name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query workspaces by user: %w", err)
	}
	defer rows.Close()

	var workspaces []*domain.Workspace
	for rows.Next() {
		var (
			idStr                   string
			name                    string
			ownerIDStr              string
			allowPrivateNetworkScan bool
			statusStr               string
			createdAt               time.Time
			updatedAt               time.Time
		)

		if err := rows.Scan(&idStr, &name, &ownerIDStr, &allowPrivateNetworkScan, &statusStr, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan workspace row: %w", err)
		}

		wsID, err := domain.NewWorkspaceID(idStr)
		if err != nil {
			return nil, fmt.Errorf("corrupted workspace ID: %w", err)
		}

		ownerID, err := domain.NewUserID(ownerIDStr)
		if err != nil {
			return nil, fmt.Errorf("corrupted owner ID: %w", err)
		}

		workspaces = append(workspaces, domain.ReconstituteWorkspace(
			wsID,
			name,
			ownerID,
			allowPrivateNetworkScan,
			domain.WorkspaceStatus(statusStr),
			nil, // Members loaded on demand per aggregate
			createdAt,
			updatedAt,
		))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return workspaces, nil
}

func (r *WorkspaceRepository) Update(ctx context.Context, ws *domain.Workspace) error {
	query := `
		UPDATE workspaces
		SET name = $1, allow_private_network_scan = $2, status = $3, updated_at = $4
		WHERE id = $5
	`
	res, err := r.db.ExecContext(
		ctx,
		query,
		ws.Name(),
		ws.AllowPrivateNetworkScan(),
		string(ws.Status()),
		ws.UpdatedAt(),
		ws.ID().String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrWorkspaceNotFound
	}

	return nil
}

func (r *WorkspaceRepository) loadMembers(ctx context.Context, wsID domain.WorkspaceID) ([]domain.WorkspaceMember, error) {
	query := `
		SELECT id, user_id, role, created_at
		FROM workspace_members
		WHERE workspace_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, wsID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace members: %w", err)
	}
	defer rows.Close()

	var members []domain.WorkspaceMember
	for rows.Next() {
		var (
			idStr     string
			userIDStr string
			roleStr   string
			createdAt time.Time
		)

		if err := rows.Scan(&idStr, &userIDStr, &roleStr, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan workspace member: %w", err)
		}

		uID, err := domain.NewUserID(userIDStr)
		if err != nil {
			return nil, fmt.Errorf("corrupted member user ID: %w", err)
		}

		members = append(members, domain.ReconstituteWorkspaceMember(
			idStr,
			uID,
			domain.Role(roleStr),
			createdAt,
		))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("member iteration error: %w", err)
	}

	return members, nil
}
