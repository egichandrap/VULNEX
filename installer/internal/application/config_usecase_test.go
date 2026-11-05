package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vulnex/vulnex/installer/internal/application"
	"github.com/vulnex/vulnex/installer/internal/domain"
)

// mockConfigRepo implements ports.ConfigRepository in memory for unit testing.
type mockConfigRepo struct {
	configs map[string]*domain.InstallConfig
	saveErr error
	loadErr error
}

func newMockConfigRepo() *mockConfigRepo {
	return &mockConfigRepo{
		configs: make(map[string]*domain.InstallConfig),
	}
}

func (m *mockConfigRepo) LoadConfig(ctx context.Context, path string) (*domain.InstallConfig, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	cfg, ok := m.configs[path]
	if !ok {
		return nil, domain.NewError(domain.ExitInvalidConfig, "config not found")
	}
	return cfg, nil
}

func (m *mockConfigRepo) SaveConfig(ctx context.Context, path string, cfg *domain.InstallConfig) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.configs[path] = cfg
	return nil
}

func TestValidateConfigUseCase(t *testing.T) {
	repo := newMockConfigRepo()
	ctx := context.Background()

	validCfg := &domain.InstallConfig{
		Mode:    domain.DeploymentModeLocal,
		DataDir: "./vulnex-data",
	}
	repo.configs["/valid/config.json"] = validCfg

	invalidCfg := &domain.InstallConfig{
		Mode:       domain.DeploymentModeServer,
		DomainName: "", // missing domain
		DataDir:    "/var/lib/vulnex",
	}
	repo.configs["/invalid/config.json"] = invalidCfg

	uc := application.NewValidateConfigUseCase(repo)

	// Valid config
	if err := uc.Execute(ctx, "/valid/config.json"); err != nil {
		t.Fatalf("expected valid config to succeed, got %v", err)
	}

	// Invalid config
	if err := uc.Execute(ctx, "/invalid/config.json"); err == nil {
		t.Fatal("expected invalid config to fail, got nil")
	} else {
		var domErr *domain.DomainError
		if errors.As(err, &domErr) {
			if domErr.ExitCode() != domain.ExitInvalidConfig {
				t.Errorf("expected exit code %d, got %d", domain.ExitInvalidConfig, domErr.ExitCode())
			}
		}
	}
}

func TestShowConfigUseCase_RedactsSecrets(t *testing.T) {
	repo := newMockConfigRepo()
	ctx := context.Background()

	cfg := &domain.InstallConfig{
		Mode:             domain.DeploymentModeServer,
		DomainName:       "sec.vulnex.dev",
		DatabasePassword: "plain-secret-password-1234",
	}
	repo.configs["/path/to/config.json"] = cfg

	uc := application.NewShowConfigUseCase(repo)

	redacted, err := uc.Execute(ctx, "/path/to/config.json")
	if err != nil {
		t.Fatalf("ShowConfig failed: %v", err)
	}

	if redacted.DatabasePassword != "********" {
		t.Errorf("expected redacted password '********', got '%s'", redacted.DatabasePassword)
	}
}
