package domain_test

import (
	"testing"

	"github.com/vulnex/vulnex/installer/internal/domain"
)

func TestInstallationState_LifecycleTransitions(t *testing.T) {
	state := domain.NewInstallationState(domain.DeploymentModeLocal, "1.0.0", "./vulnex-data")

	if state.Status != domain.InstallationStatusPending {
		t.Errorf("expected initial status pending, got %s", state.Status)
	}
	if state.Phase != domain.DeploymentPhasePreflight {
		t.Errorf("expected initial phase preflight, got %s", state.Phase)
	}

	state.AdvancePhase(domain.DeploymentPhaseConfiguration)
	if state.Phase != domain.DeploymentPhaseConfiguration {
		t.Errorf("expected phase configuration, got %s", state.Phase)
	}

	state.MarkReady()
	if state.Status != domain.InstallationStatusReady {
		t.Errorf("expected status ready, got %s", state.Status)
	}
	if state.Phase != domain.DeploymentPhaseCommitted {
		t.Errorf("expected phase committed, got %s", state.Phase)
	}
	if state.UpdatedAt.Before(state.InstalledAt) {
		t.Errorf("updatedAt %v must not be before installedAt %v", state.UpdatedAt, state.InstalledAt)
	}

	state.MarkFailed("Docker compose failed")
	if state.Status != domain.InstallationStatusFailed {
		t.Errorf("expected status failed, got %s", state.Status)
	}
	if state.LastError != "Docker compose failed" {
		t.Errorf("expected last error recorded, got %s", state.LastError)
	}
}

func TestExitCodes(t *testing.T) {
	err := domain.NewError(domain.ExitInvalidConfig, "config is invalid")
	if err.ExitCode() != domain.ExitInvalidConfig {
		t.Errorf("expected exit code %d, got %d", domain.ExitInvalidConfig, err.ExitCode())
	}
	if err.Error() != "config is invalid" {
		t.Errorf("expected message 'config is invalid', got '%s'", err.Error())
	}
}
