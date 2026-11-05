package application

import (
	"context"

	"github.com/vulnex/vulnex/installer/internal/application/ports"
	"github.com/vulnex/vulnex/installer/internal/domain"
)

// ValidateConfigUseCase verifies syntax and semantic invariants of a configuration file.
type ValidateConfigUseCase struct {
	repo ports.ConfigRepository
}

func NewValidateConfigUseCase(repo ports.ConfigRepository) *ValidateConfigUseCase {
	return &ValidateConfigUseCase{repo: repo}
}

func (uc *ValidateConfigUseCase) Execute(ctx context.Context, configPath string) error {
	cfg, err := uc.repo.LoadConfig(ctx, configPath)
	if err != nil {
		return err
	}
	return cfg.Validate()
}

// ShowConfigUseCase loads configuration and returns a sanitized copy with secrets redacted.
type ShowConfigUseCase struct {
	repo ports.ConfigRepository
}

func NewShowConfigUseCase(repo ports.ConfigRepository) *ShowConfigUseCase {
	return &ShowConfigUseCase{repo: repo}
}

func (uc *ShowConfigUseCase) Execute(ctx context.Context, configPath string) (*domain.InstallConfig, error) {
	cfg, err := uc.repo.LoadConfig(ctx, configPath)
	if err != nil {
		return nil, err
	}
	return cfg.Redacted(), nil
}
