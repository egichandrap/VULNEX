package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vulnex/vulnex/installer/internal/application/ports"
	"github.com/vulnex/vulnex/installer/internal/domain"
)

// TemplateRenderer defines the contract for materializing deployment files.
type TemplateRenderer interface {
	Render(cfg *domain.InstallConfig, targetDir string) error
}

// InstallVulnexUseCase orchestrates idempotent deployment, bootstrap, and verification.
type InstallVulnexUseCase struct {
	runtime    ports.RuntimePort
	configRepo ports.ConfigRepository
	stateRepo  ports.StateRepository
	fileLock   ports.FileLockPort
	detector   ports.EnvironmentDetector
	renderer   TemplateRenderer
}

func NewInstallVulnexUseCase(
	runtime ports.RuntimePort,
	configRepo ports.ConfigRepository,
	stateRepo ports.StateRepository,
	fileLock ports.FileLockPort,
	detector ports.EnvironmentDetector,
	renderer TemplateRenderer,
) *InstallVulnexUseCase {
	return &InstallVulnexUseCase{
		runtime:    runtime,
		configRepo: configRepo,
		stateRepo:  stateRepo,
		fileLock:   fileLock,
		detector:   detector,
		renderer:   renderer,
	}
}

func (uc *InstallVulnexUseCase) Execute(ctx context.Context, cfg *domain.InstallConfig) (*domain.InstallationState, string, error) {
	if cfg == nil {
		return nil, "", domain.NewError(domain.ExitInvalidConfig, "nil install config provided")
	}

	if err := cfg.Validate(); err != nil {
		return nil, "", domain.WrapError(domain.ExitInvalidConfig, "invalid installation parameters", err)
	}

	if cfg.DataDir == "" {
		if cfg.Mode == domain.DeploymentModeServer {
			cfg.DataDir = "/var/lib/vulnex"
		} else {
			cfg.DataDir = "./vulnex-data"
		}
	}

	// 1. Acquire concurrency lock
	unlock, err := uc.fileLock.AcquireLock(ctx, cfg.DataDir)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = unlock() }()

	// 2. Preserve existing database password or generate secure one if absent
	if cfg.DatabasePassword == "" {
		if existingCfg, err := uc.configRepo.LoadConfig(ctx, filepath.Join(cfg.DataDir, "config.json")); err == nil && existingCfg.DatabasePassword != "" {
			cfg.DatabasePassword = existingCfg.DatabasePassword
		} else {
			passBytes := make([]byte, 16)
			if _, err := rand.Read(passBytes); err != nil {
				return nil, "", domain.WrapError(domain.ExitGeneralFailure, "failed to generate secure password", err)
			}
			cfg.DatabasePassword = hex.EncodeToString(passBytes)
		}
	}

	// 3. Pre-flight diagnostics
	if !cfg.SkipPreflight && os.Getenv("VULNEX_SKIP_PREFLIGHT") != "1" {
		docUC := NewDiagnoseEnvironmentUseCase(uc.detector)
		report, err := docUC.Execute(ctx)
		if err != nil {
			return nil, "", err
		}
		if report.HasFailures() {
			var failMsgs []string
			for _, c := range report.Checks {
				if c.Status == domain.CheckFail {
					failMsgs = append(failMsgs, fmt.Sprintf("%s: %s (Action: %s)", c.Name, c.Message, c.Remediation))
				}
			}
			return nil, "", domain.NewError(domain.ExitUnsupportedEnv, fmt.Sprintf("pre-flight diagnostics failed:\n  - %s", strings.Join(failMsgs, "\n  - ")))
		}
	}

	// 4. Initialize State
	version := cfg.Version
	if version == "" {
		version = domain.CurrentVersion.Number
	}
	state := domain.NewInstallationState(cfg.Mode, version, cfg.DataDir)
	state.ConfigPath = filepath.Join(cfg.DataDir, "config.json")
	if err := uc.stateRepo.SaveState(ctx, cfg.DataDir, state); err != nil {
		return nil, "", err
	}

	// 5. Materialize templates and config
	state.AdvancePhase(domain.DeploymentPhaseMaterialization)
	if err := uc.renderer.Render(cfg, cfg.DataDir); err != nil {
		state.MarkFailed("template rendering failed: " + err.Error())
		_ = uc.stateRepo.SaveState(ctx, cfg.DataDir, state)
		return nil, "", err
	}
	if err := uc.configRepo.SaveConfig(ctx, state.ConfigPath, cfg); err != nil {
		state.MarkFailed("config saving failed: " + err.Error())
		_ = uc.stateRepo.SaveState(ctx, cfg.DataDir, state)
		return nil, "", err
	}

	// 6. Deploy Stack
	state.AdvancePhase(domain.DeploymentPhaseDeployment)
	composePath := filepath.Join(cfg.DataDir, "docker-compose.yml")
	envFile := filepath.Join(cfg.DataDir, ".env")

	isDryRun := cfg.DryRun || os.Getenv("VULNEX_DRY_RUN") == "1"
	if !isDryRun {
		_ = uc.runtime.Pull(ctx, composePath, envFile) // best effort online pull
		if err := uc.runtime.Up(ctx, composePath, envFile); err != nil {
			state.MarkFailed("compose up failed: " + err.Error())
			_ = uc.stateRepo.SaveState(ctx, cfg.DataDir, state)
			return nil, "", err
		}
	}

	// 7. Verify Health & Collect Services
	state.AdvancePhase(domain.DeploymentPhaseHealthCheck)
	if !isDryRun {
		services, err := uc.runtime.Ps(ctx, composePath)
		if err != nil {
			state.MarkFailed("health query failed: " + err.Error())
			_ = uc.stateRepo.SaveState(ctx, cfg.DataDir, state)
			return nil, "", err
		}
		state.SetServices(services)
	} else {
		state.SetServices(map[string]string{"postgres": "running", "api": "running", "web": "running"})
	}

	// 8. Ephemeral Setup Token Generation & Seeding (ADR-013)
	token, err := domain.GenerateEphemeralSetupToken()
	if err != nil {
		state.MarkFailed("token generation failed: " + err.Error())
		_ = uc.stateRepo.SaveState(ctx, cfg.DataDir, state)
		return nil, "", err
	}

	tokenHashBytes := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(tokenHashBytes[:])

	seedSQL := fmt.Sprintf(
		"INSERT INTO setup_tokens (id, token_hash, expires_at) VALUES (gen_random_uuid(), '%s', NOW() + INTERVAL '1 hour');",
		tokenHash,
	)

	if !isDryRun {
		_, execErr := uc.runtime.Exec(ctx, composePath, "postgres", []string{
			"psql", "-U", "vulnex", "-d", "vulnex", "-c", seedSQL,
		})
		if execErr != nil {
			state.LastError = "warning: database setup token seeding delayed: " + execErr.Error()
		}
	}

	// 9. Mark Ready and Commit State
	state.MarkReady()
	if err := uc.stateRepo.SaveState(ctx, cfg.DataDir, state); err != nil {
		return nil, "", err
	}

	return state, token, nil
}
