package application

import (
	"context"
	"path/filepath"

	"github.com/vulnex/vulnex/installer/internal/application/ports"
	"github.com/vulnex/vulnex/installer/internal/domain"
)

// LifecycleUseCase coordinates stack start, stop, restart, status, and logs.
type LifecycleUseCase struct {
	runtime   ports.RuntimePort
	stateRepo ports.StateRepository
}

func NewLifecycleUseCase(runtime ports.RuntimePort, stateRepo ports.StateRepository) *LifecycleUseCase {
	return &LifecycleUseCase{
		runtime:   runtime,
		stateRepo: stateRepo,
	}
}

func (uc *LifecycleUseCase) Status(ctx context.Context, dataDir string) (*domain.InstallationState, map[string]string, error) {
	state, err := uc.stateRepo.LoadState(ctx, dataDir)
	if err != nil {
		return nil, nil, err
	}

	composePath := filepath.Join(dataDir, "docker-compose.yml")
	services, err := uc.runtime.Ps(ctx, composePath)
	if err == nil {
		state.SetServices(services)
		_ = uc.stateRepo.SaveState(ctx, dataDir, state)
	}

	return state, services, nil
}

func (uc *LifecycleUseCase) Start(ctx context.Context, dataDir string) error {
	composePath := filepath.Join(dataDir, "docker-compose.yml")
	envFile := filepath.Join(dataDir, ".env")
	if err := uc.runtime.Up(ctx, composePath, envFile); err != nil {
		return err
	}

	state, err := uc.stateRepo.LoadState(ctx, dataDir)
	if err == nil {
		state.Status = domain.InstallationStatusReady
		_ = uc.stateRepo.SaveState(ctx, dataDir, state)
	}
	return nil
}

func (uc *LifecycleUseCase) Stop(ctx context.Context, dataDir string) error {
	composePath := filepath.Join(dataDir, "docker-compose.yml")
	if err := uc.runtime.Down(ctx, composePath, false); err != nil {
		return err
	}

	state, err := uc.stateRepo.LoadState(ctx, dataDir)
	if err == nil {
		state.Status = domain.InstallationStatusStopped
		_ = uc.stateRepo.SaveState(ctx, dataDir, state)
	}
	return nil
}

func (uc *LifecycleUseCase) Restart(ctx context.Context, dataDir string) error {
	if err := uc.Stop(ctx, dataDir); err != nil {
		return err
	}
	return uc.Start(ctx, dataDir)
}

func (uc *LifecycleUseCase) Logs(ctx context.Context, dataDir string, service string, follow bool) error {
	composePath := filepath.Join(dataDir, "docker-compose.yml")
	return uc.runtime.Logs(ctx, composePath, service, follow)
}
