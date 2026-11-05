package application_test

import (
	"context"
	"testing"

	"github.com/vulnex/vulnex/installer/internal/application"
	"github.com/vulnex/vulnex/installer/internal/domain"
)

func TestDashboardUseCase_URLResolution(t *testing.T) {
	stateRepo := &mockStateRepo{
		state: &domain.InstallationState{
			Mode:    domain.DeploymentModeLocal,
			Status:  domain.InstallationStatusReady,
			DataDir: "/tmp/vulnex",
		},
	}
	configRepo := newMockConfigRepo()

	uc := application.NewDashboardUseCase(stateRepo, configRepo)

	url, err := uc.ResolveURL(context.Background(), "/tmp/vulnex")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if url != "http://localhost:3000" {
		t.Errorf("expected localhost URL, got %s", url)
	}

	// Server mode
	stateRepo.state.Mode = domain.DeploymentModeServer
	configRepo.configs["/tmp/vulnex/config.json"] = &domain.InstallConfig{
		Mode:       domain.DeploymentModeServer,
		DomainName: "vulnex.corp.internal",
	}

	serverURL, err := uc.ResolveURL(context.Background(), "/tmp/vulnex")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if serverURL != "https://vulnex.corp.internal" {
		t.Errorf("expected https://vulnex.corp.internal, got %s", serverURL)
	}
}
