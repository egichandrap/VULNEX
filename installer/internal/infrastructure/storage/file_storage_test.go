package storage_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/vulnex/vulnex/installer/internal/domain"
	"github.com/vulnex/vulnex/installer/internal/infrastructure/storage"
)

func TestFileStorage_SaveAndLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	store := storage.NewFileStorage()
	ctx := context.Background()

	cfg := &domain.InstallConfig{
		Mode:             domain.DeploymentModeServer,
		DomainName:       "sec.vulnex.dev",
		Ingress:          domain.IngressEmbeddedCaddy,
		DataDir:          tempDir,
		Version:          "1.0.0",
		DatabasePassword: "db-secret-password-xyz",
	}

	if err := store.SaveConfig(ctx, configPath, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file permissions (0600)
	fi, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}
	if fi.Mode().Perm() != 0600 {
		t.Errorf("expected permissions 0600, got %o", fi.Mode().Perm())
	}

	loaded, err := store.LoadConfig(ctx, configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.Mode != cfg.Mode {
		t.Errorf("expected mode %s, got %s", cfg.Mode, loaded.Mode)
	}
	if loaded.DomainName != cfg.DomainName {
		t.Errorf("expected domain %s, got %s", cfg.DomainName, loaded.DomainName)
	}
	if loaded.DatabasePassword != cfg.DatabasePassword {
		t.Errorf("expected password %s, got %s", cfg.DatabasePassword, loaded.DatabasePassword)
	}
}

func TestFileStorage_SaveAndLoadState(t *testing.T) {
	tempDir := t.TempDir()
	store := storage.NewFileStorage()
	ctx := context.Background()

	state := domain.NewInstallationState(domain.DeploymentModeLocal, "1.0.0", tempDir)
	state.AdvancePhase(domain.DeploymentPhaseDeployment)

	if err := store.SaveState(ctx, tempDir, state); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	loaded, err := store.LoadState(ctx, tempDir)
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if loaded.Mode != domain.DeploymentModeLocal {
		t.Errorf("expected mode local, got %s", loaded.Mode)
	}
	if loaded.Phase != domain.DeploymentPhaseDeployment {
		t.Errorf("expected phase deployment, got %s", loaded.Phase)
	}
}
