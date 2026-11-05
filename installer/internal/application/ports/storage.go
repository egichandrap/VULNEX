package ports

import (
	"context"

	"github.com/vulnex/vulnex/installer/internal/domain"
)

// ConfigRepository defines persistence operations for deployment configuration.
type ConfigRepository interface {
	LoadConfig(ctx context.Context, path string) (*domain.InstallConfig, error)
	SaveConfig(ctx context.Context, path string, cfg *domain.InstallConfig) error
}

// StateRepository defines persistence operations for durable installation state.
type StateRepository interface {
	LoadState(ctx context.Context, dataDir string) (*domain.InstallationState, error)
	SaveState(ctx context.Context, dataDir string, state *domain.InstallationState) error
}

// FileLockPort enforces concurrency isolation for mutating commands.
type FileLockPort interface {
	AcquireLock(ctx context.Context, dataDir string) (unlock func() error, err error)
}
