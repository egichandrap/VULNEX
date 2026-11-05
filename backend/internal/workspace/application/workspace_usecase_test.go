package application

import (
	"context"
	"testing"

	"github.com/vulnex/vulnex/backend/internal/workspace/domain"
)

type fakeWorkspaceRepository struct {
	workspaces map[string]*domain.Workspace
}

func newFakeWorkspaceRepository() *fakeWorkspaceRepository {
	return &fakeWorkspaceRepository{workspaces: make(map[string]*domain.Workspace)}
}

func (f *fakeWorkspaceRepository) Save(ctx context.Context, ws *domain.Workspace) error {
	f.workspaces[ws.ID().String()] = ws
	return nil
}

func (f *fakeWorkspaceRepository) FindByID(ctx context.Context, id domain.WorkspaceID) (*domain.Workspace, error) {
	ws, ok := f.workspaces[id.String()]
	if !ok {
		return nil, domain.ErrMemberNotFound
	}
	return ws, nil
}

func (f *fakeWorkspaceRepository) FindByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Workspace, error) {
	var res []*domain.Workspace
	for _, ws := range f.workspaces {
		for _, m := range ws.Members() {
			if m.UserID().String() == userID.String() {
				res = append(res, ws)
				break
			}
		}
	}
	return res, nil
}

func (f *fakeWorkspaceRepository) Update(ctx context.Context, ws *domain.Workspace) error {
	f.workspaces[ws.ID().String()] = ws
	return nil
}

func TestCreateWorkspaceUseCase_Success(t *testing.T) {
	ctx := context.Background()
	repo := newFakeWorkspaceRepository()
	useCase := NewCreateWorkspaceUseCase(repo)

	ownerID, _ := domain.GenerateUserID()

	res, err := useCase.Execute(ctx, CreateWorkspaceCommand{
		Name:        "Corporate Security Office",
		OwnerUserID: ownerID.String(),
	})

	if err != nil {
		t.Fatalf("CreateWorkspaceUseCase failed: %v", err)
	}

	if res.WorkspaceID == "" {
		t.Errorf("Expected valid WorkspaceID")
	}

	ws, _ := repo.FindByID(ctx, res.ID)
	if ws.Name() != "Corporate Security Office" {
		t.Errorf("Workspace name mismatch: %s", ws.Name())
	}
}

func TestUpdateWorkspacePolicyUseCase_OwnerCanEnablePrivateScanning(t *testing.T) {
	ctx := context.Background()
	repo := newFakeWorkspaceRepository()
	ownerID, _ := domain.GenerateUserID()
	ws, _ := domain.NewWorkspace("Core WS", ownerID)
	_ = repo.Save(ctx, ws)

	useCase := NewUpdateWorkspacePolicyUseCase(repo)

	// Owner enables private network scan
	err := useCase.Execute(ctx, UpdateWorkspacePolicyCommand{
		WorkspaceID:             ws.ID().String(),
		ActorUserID:             ownerID.String(),
		AllowPrivateNetworkScan: true,
		Justification:           "Testing internal RFC 1918 staging subnets per authorization doc #12",
	})

	if err != nil {
		t.Fatalf("UpdateWorkspacePolicyUseCase failed: %v", err)
	}

	updatedWS, _ := repo.FindByID(ctx, ws.ID())
	if !updatedWS.AllowPrivateNetworkScan() {
		t.Errorf("Expected AllowPrivateNetworkScan to be true")
	}

	// Non-admin member cannot enable private scanning
	nonAdminID, _ := domain.GenerateUserID()
	_ = updatedWS.AddMember(nonAdminID, domain.RoleMember)
	_ = repo.Update(ctx, updatedWS)

	err = useCase.Execute(ctx, UpdateWorkspacePolicyCommand{
		WorkspaceID:             ws.ID().String(),
		ActorUserID:             nonAdminID.String(),
		AllowPrivateNetworkScan: true,
		Justification:           "Attempting unauthorized toggle",
	})
	if err != domain.ErrUnauthorizedPolicyChange {
		t.Errorf("Expected ErrUnauthorizedPolicyChange, got: %v", err)
	}
}
