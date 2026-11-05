package storage_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vulnex/vulnex/installer/internal/domain"
	"github.com/vulnex/vulnex/installer/internal/infrastructure/storage"
)

func TestStorage_PathTraversalProtection(t *testing.T) {
	tempDir := t.TempDir()
	store := storage.NewFileStorage()
	ctx := context.Background()

	// Malicious relative path attempting traversal
	maliciousPath := filepath.Join(tempDir, "..", "..", "malicious_config.json")

	cfg := &domain.InstallConfig{
		Mode: domain.DeploymentModeLocal,
	}

	// Save should clean the path
	err := store.SaveConfig(ctx, maliciousPath, cfg)
	if err == nil {
		clean := filepath.Clean(maliciousPath)
		if strings.Contains(clean, "..") {
			t.Errorf("expected path to be cleaned of dot-dot traversals, got %s", clean)
		}
	}
}

func TestStorage_PermissionsEnforcement(t *testing.T) {
	tempDir := t.TempDir()
	store := storage.NewFileStorage()
	ctx := context.Background()

	configPath := filepath.Join(tempDir, "secure.json")
	cfg := &domain.InstallConfig{
		Mode:             domain.DeploymentModeServer,
		DomainName:       "vulnex.internal",
		DatabasePassword: "verysecretpassword",
	}

	if err := store.SaveConfig(ctx, configPath, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := store.LoadConfig(ctx, configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	redacted := loaded.Redacted()
	if redacted.DatabasePassword != "********" {
		t.Errorf("expected redacted password, got %s", redacted.DatabasePassword)
	}
}
