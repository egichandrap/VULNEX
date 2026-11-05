package application

import (
	"context"
	"fmt"

	"github.com/vulnex/vulnex/backend/internal/workspace/domain"
)

type CreateWorkspaceCommand struct {
	Name        string
	OwnerUserID string
}

type CreateWorkspaceResult struct {
	WorkspaceID string
	ID          domain.WorkspaceID
}

// CreateWorkspaceUseCase creates a new tenant workspace with the caller as initial owner.
type CreateWorkspaceUseCase struct {
	repo domain.WorkspaceRepository
}

func NewCreateWorkspaceUseCase(repo domain.WorkspaceRepository) *CreateWorkspaceUseCase {
	return &CreateWorkspaceUseCase{repo: repo}
}

func (uc *CreateWorkspaceUseCase) Execute(ctx context.Context, cmd CreateWorkspaceCommand) (*CreateWorkspaceResult, error) {
	ownerID, err := domain.NewUserID(cmd.OwnerUserID)
	if err != nil {
		ownerID, err = domain.GenerateUserID()
		if err != nil {
			return nil, err
		}
	}

	ws, err := domain.NewWorkspace(cmd.Name, ownerID)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, ws); err != nil {
		return nil, fmt.Errorf("failed to save workspace: %w", err)
	}

	return &CreateWorkspaceResult{
		WorkspaceID: ws.ID().String(),
		ID:          ws.ID(),
	}, nil
}

type UpdateWorkspacePolicyCommand struct {
	WorkspaceID             string
	ActorUserID             string
	AllowPrivateNetworkScan bool
	Justification           string
}

// UpdateWorkspacePolicyUseCase handles administrator toggling of workspace security policies (ADR-015).
type UpdateWorkspacePolicyUseCase struct {
	repo domain.WorkspaceRepository
}

func NewUpdateWorkspacePolicyUseCase(repo domain.WorkspaceRepository) *UpdateWorkspacePolicyUseCase {
	return &UpdateWorkspacePolicyUseCase{repo: repo}
}

func (uc *UpdateWorkspacePolicyUseCase) Execute(ctx context.Context, cmd UpdateWorkspacePolicyCommand) error {
	wsID, err := domain.NewWorkspaceID(cmd.WorkspaceID)
	if err != nil {
		return err
	}

	actorID, err := domain.NewUserID(cmd.ActorUserID)
	if err != nil {
		return err
	}

	ws, err := uc.repo.FindByID(ctx, wsID)
	if err != nil || ws == nil {
		return fmt.Errorf("workspace not found: %w", err)
	}

	if cmd.AllowPrivateNetworkScan {
		if err := ws.EnablePrivateNetworkScanning(actorID, cmd.Justification); err != nil {
			return err
		}
	} else {
		ws.DisablePrivateNetworkScanning()
	}

	return uc.repo.Update(ctx, ws)
}
