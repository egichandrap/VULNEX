package application_test

import (
	"context"
	"testing"

	"github.com/vulnex/vulnex/installer/internal/application"
	"github.com/vulnex/vulnex/installer/internal/domain"
)

func TestLifecycleUseCase(t *testing.T) {
	tempDir := t.TempDir()
	stateRepo := &mockStateRepo{
		state: domain.NewInstallationState(domain.DeploymentModeLocal, "1.0.0", tempDir),
	}
	runtime := &mockRuntime{}

	uc := application.NewLifecycleUseCase(runtime, stateRepo)
	ctx := context.Background()

	// 1. Status
	state, services, err := uc.Status(ctx, tempDir)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if state == nil || len(services) == 0 {
		t.Errorf("expected state and services in Status")
	}

	// 2. Start
	if err := uc.Start(ctx, tempDir); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if !runtime.upped {
		t.Error("expected runtime.Up on Start")
	}

	// 3. Stop
	if err := uc.Stop(ctx, tempDir); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if !runtime.downed {
		t.Error("expected runtime.Down on Stop")
	}

	// 4. Restart
	runtime.upped = false
	runtime.downed = false
	if err := uc.Restart(ctx, tempDir); err != nil {
		t.Fatalf("Restart failed: %v", err)
	}
	if !runtime.downed || !runtime.upped {
		t.Error("expected Down and Up on Restart")
	}
}

func TestMaintenanceUseCase_Uninstall(t *testing.T) {
	tempDir := t.TempDir()
	stateRepo := &mockStateRepo{
		state: domain.NewInstallationState(domain.DeploymentModeLocal, "1.0.0", tempDir),
	}
	runtime := &mockRuntime{}
	uc := application.NewMaintenanceUseCase(runtime, stateRepo)
	ctx := context.Background()

	// Uninstall with keep-data
	if err := uc.Uninstall(ctx, tempDir, false); err != nil {
		t.Fatalf("Uninstall keep-data failed: %v", err)
	}
	if stateRepo.state.Status != domain.InstallationStatusStopped {
		t.Errorf("expected stopped status, got %s", stateRepo.state.Status)
	}
}
