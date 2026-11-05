package application

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/vulnex/vulnex/installer/internal/application/ports"
	"github.com/vulnex/vulnex/installer/internal/domain"
)

// MaintenanceUseCase manages backup, restore, update, and uninstall workflows.
type MaintenanceUseCase struct {
	runtime   ports.RuntimePort
	stateRepo ports.StateRepository
}

func NewMaintenanceUseCase(runtime ports.RuntimePort, stateRepo ports.StateRepository) *MaintenanceUseCase {
	return &MaintenanceUseCase{
		runtime:   runtime,
		stateRepo: stateRepo,
	}
}

func (uc *MaintenanceUseCase) Backup(ctx context.Context, dataDir string, outputPath string) (string, error) {
	composePath := filepath.Join(dataDir, "docker-compose.yml")
	dumpSQL, err := uc.runtime.Exec(ctx, composePath, "postgres", []string{"pg_dump", "-U", "vulnex", "vulnex"})
	if err != nil {
		return "", domain.WrapError(domain.ExitBackupRestoreFailed, "database dump failed", err)
	}

	cleanOut := filepath.Clean(outputPath)
	if err := os.MkdirAll(cleanOut, 0700); err != nil {
		return "", domain.WrapError(domain.ExitGeneralFailure, "failed to create backup directory", err)
	}

	timestamp := time.Now().UTC().Format("20060102-150405")
	destFile := filepath.Join(cleanOut, fmt.Sprintf("vulnex-backup-%s.sql", timestamp))
	if err := os.WriteFile(destFile, []byte(dumpSQL), 0600); err != nil {
		return "", domain.WrapError(domain.ExitBackupRestoreFailed, "failed to write backup file", err)
	}

	return destFile, nil
}

func (uc *MaintenanceUseCase) Restore(ctx context.Context, dataDir string, backupFile string) error {
	cleanFile := filepath.Clean(backupFile)
	data, err := os.ReadFile(cleanFile)
	if err != nil {
		return domain.WrapError(domain.ExitBackupRestoreFailed, "failed to read backup file", err)
	}

	composePath := filepath.Join(dataDir, "docker-compose.yml")
	_, err = uc.runtime.Exec(ctx, composePath, "postgres", []string{"psql", "-U", "vulnex", "-d", "vulnex", "-c", string(data)})
	if err != nil {
		return domain.WrapError(domain.ExitBackupRestoreFailed, "failed to restore database dump", err)
	}

	return nil
}

func (uc *MaintenanceUseCase) Update(ctx context.Context, dataDir string, targetVersion string) error {
	composePath := filepath.Join(dataDir, "docker-compose.yml")
	envFile := filepath.Join(dataDir, ".env")

	// Pull new images
	if err := uc.runtime.Pull(ctx, composePath, envFile); err != nil {
		return domain.WrapError(domain.ExitUpgradeFailed, "failed to pull updated images", err)
	}

	// Recreate containers
	if err := uc.runtime.Up(ctx, composePath, envFile); err != nil {
		return domain.WrapError(domain.ExitUpgradeFailed, "failed to recreate updated containers", err)
	}

	state, err := uc.stateRepo.LoadState(ctx, dataDir)
	if err == nil {
		state.Version = targetVersion
		state.UpdatedAt = time.Now().UTC()
		_ = uc.stateRepo.SaveState(ctx, dataDir, state)
	}

	return nil
}

func (uc *MaintenanceUseCase) Uninstall(ctx context.Context, dataDir string, deleteData bool) error {
	cleanDir := filepath.Clean(dataDir)
	composePath := filepath.Join(cleanDir, "docker-compose.yml")

	// 1. Tear down stack
	_ = uc.runtime.Down(ctx, composePath, deleteData)

	// 2. Data retention enforcement
	if deleteData {
		if err := os.RemoveAll(cleanDir); err != nil {
			return domain.WrapError(domain.ExitGeneralFailure, "failed to remove data directory", err)
		}
	} else {
		state, err := uc.stateRepo.LoadState(ctx, cleanDir)
		if err == nil {
			state.Status = domain.InstallationStatusStopped
			_ = uc.stateRepo.SaveState(ctx, cleanDir, state)
		}
	}

	return nil
}
