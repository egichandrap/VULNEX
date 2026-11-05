package application_test

import (
	"context"
	"testing"

	"github.com/vulnex/vulnex/installer/internal/application"
	"github.com/vulnex/vulnex/installer/internal/domain"
)

type mockRuntime struct {
	pulled   bool
	upped    bool
	downed   bool
	execCmds [][]string
}

func (m *mockRuntime) Pull(ctx context.Context, composePath string, envFile string) error {
	m.pulled = true
	return nil
}
func (m *mockRuntime) Up(ctx context.Context, composePath string, envFile string) error {
	m.upped = true
	return nil
}
func (m *mockRuntime) Down(ctx context.Context, composePath string, removeVolumes bool) error {
	m.downed = true
	return nil
}
func (m *mockRuntime) Ps(ctx context.Context, composePath string) (map[string]string, error) {
	return map[string]string{"postgres": "running", "api": "running", "web": "running"}, nil
}
func (m *mockRuntime) Logs(ctx context.Context, composePath string, service string, follow bool) error {
	return nil
}
func (m *mockRuntime) Exec(ctx context.Context, composePath string, service string, cmd []string) (string, error) {
	m.execCmds = append(m.execCmds, cmd)
	return "INSERT 0 1", nil
}

type mockStateRepo struct {
	state *domain.InstallationState
}

func (m *mockStateRepo) LoadState(ctx context.Context, dataDir string) (*domain.InstallationState, error) {
	return m.state, nil
}
func (m *mockStateRepo) SaveState(ctx context.Context, dataDir string, state *domain.InstallationState) error {
	m.state = state
	return nil
}

type mockFileLock struct{}

func (m *mockFileLock) AcquireLock(ctx context.Context, dataDir string) (func() error, error) {
	return func() error { return nil }, nil
}

type mockTemplateRenderer struct {
	renderedCfg *domain.InstallConfig
}

func (m *mockTemplateRenderer) Render(cfg *domain.InstallConfig, targetDir string) error {
	m.renderedCfg = cfg
	return nil
}

func TestInstallVulnexUseCase_LocalSuccess(t *testing.T) {
	tempDir := t.TempDir()
	configRepo := newMockConfigRepo()
	stateRepo := &mockStateRepo{}
	fileLock := &mockFileLock{}
	runtime := &mockRuntime{}
	renderer := &mockTemplateRenderer{}
	detector := &mockEnvDetector{
		os:        "linux",
		arch:      "amd64",
		dockerOk:  true,
		composeOk: true,
		memBytes:  8 * 1024 * 1024 * 1024,
		diskBytes: 50 * 1024 * 1024 * 1024,
		portOk:    true,
		regOk:     true,
	}

	uc := application.NewInstallVulnexUseCase(
		runtime,
		configRepo,
		stateRepo,
		fileLock,
		detector,
		renderer,
	)

	cfg := &domain.InstallConfig{
		Mode:    domain.DeploymentModeLocal,
		DataDir: tempDir,
	}

	state, token, err := uc.Execute(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if state.Status != domain.InstallationStatusReady {
		t.Errorf("expected ready status, got %s", state.Status)
	}
	if state.Phase != domain.DeploymentPhaseCommitted {
		t.Errorf("expected committed phase, got %s", state.Phase)
	}
	if len(token) != 64 {
		t.Errorf("expected 64-char hex setup token, got length %d", len(token))
	}
	if !runtime.upped {
		t.Error("expected runtime.Up to be called")
	}
	if len(runtime.execCmds) == 0 {
		t.Error("expected psql exec command to inject setup token into database")
	}
}
