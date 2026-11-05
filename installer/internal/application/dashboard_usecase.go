package application

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/vulnex/vulnex/installer/internal/application/ports"
	"github.com/vulnex/vulnex/installer/internal/domain"
)

// DashboardUseCase resolves the dashboard URL and launches the web browser.
type DashboardUseCase struct {
	stateRepo  ports.StateRepository
	configRepo ports.ConfigRepository
}

func NewDashboardUseCase(stateRepo ports.StateRepository, configRepo ports.ConfigRepository) *DashboardUseCase {
	return &DashboardUseCase{
		stateRepo:  stateRepo,
		configRepo: configRepo,
	}
}

func (uc *DashboardUseCase) ResolveURL(ctx context.Context, dataDir string) (string, error) {
	state, err := uc.stateRepo.LoadState(ctx, dataDir)
	if err == nil && state.Mode == domain.DeploymentModeServer {
		configPath := state.ConfigPath
		if configPath == "" {
			configPath = filepath.Join(dataDir, "config.json")
		}
		cfg, cfgErr := uc.configRepo.LoadConfig(ctx, configPath)
		if cfgErr == nil && cfg.DomainName != "" {
			return fmt.Sprintf("https://%s", cfg.DomainName), nil
		}
	}

	return "http://localhost:3000", nil
}

func (uc *DashboardUseCase) OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported OS for automatic browser launch: %s", runtime.GOOS)
	}
	return cmd.Start()
}
