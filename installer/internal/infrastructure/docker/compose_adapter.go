package docker

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vulnex/vulnex/installer/internal/domain"
)

// CommandRunner abstracts process execution for testing.
type CommandRunner interface {
	Run(ctx context.Context, cmd string, args ...string) ([]byte, error)
}

type osCommandRunner struct{}

func (r *osCommandRunner) Run(ctx context.Context, cmd string, args ...string) ([]byte, error) {
	c := exec.CommandContext(ctx, cmd, args...)
	return c.CombinedOutput()
}

// ComposeAdapter executes Docker Compose commands.
type ComposeAdapter struct {
	runner CommandRunner
}

func NewComposeAdapter() *ComposeAdapter {
	return &ComposeAdapter{runner: &osCommandRunner{}}
}

func NewComposeAdapterWithRunner(runner CommandRunner) *ComposeAdapter {
	return &ComposeAdapter{runner: runner}
}

func (a *ComposeAdapter) Pull(ctx context.Context, composePath string, envFile string) error {
	args := []string{"compose", "-f", filepath.Clean(composePath)}
	if envFile != "" {
		args = append(args, "--env-file", filepath.Clean(envFile))
	}
	args = append(args, "pull")

	out, err := a.runner.Run(ctx, "docker", args...)
	if err != nil {
		return domain.WrapError(domain.ExitDeploymentFailed, "failed to pull docker compose images: "+string(out), err)
	}
	return nil
}

func (a *ComposeAdapter) Up(ctx context.Context, composePath string, envFile string) error {
	args := []string{"compose", "-f", filepath.Clean(composePath)}
	if envFile != "" {
		args = append(args, "--env-file", filepath.Clean(envFile))
	}
	args = append(args, "up", "-d")

	out, err := a.runner.Run(ctx, "docker", args...)
	if err != nil {
		return domain.WrapError(domain.ExitDeploymentFailed, "failed to start docker compose stack: "+string(out), err)
	}
	return nil
}

func (a *ComposeAdapter) Down(ctx context.Context, composePath string, removeVolumes bool) error {
	args := []string{"compose", "-f", filepath.Clean(composePath), "down"}
	if removeVolumes {
		args = append(args, "-v")
	}

	out, err := a.runner.Run(ctx, "docker", args...)
	if err != nil {
		return domain.WrapError(domain.ExitGeneralFailure, "failed to stop docker compose stack: "+string(out), err)
	}
	return nil
}

func (a *ComposeAdapter) Ps(ctx context.Context, composePath string) (map[string]string, error) {
	args := []string{"compose", "-f", filepath.Clean(composePath), "ps", "--format", "json"}
	out, err := a.runner.Run(ctx, "docker", args...)
	if err != nil {
		return nil, domain.WrapError(domain.ExitGeneralFailure, "failed to query docker compose ps: "+string(out), err)
	}

	services := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var item struct {
			Service string `json:"Service"`
			State   string `json:"State"`
			Status  string `json:"Status"`
		}
		if err := json.Unmarshal([]byte(line), &item); err == nil && item.Service != "" {
			state := item.State
			if state == "" {
				state = item.Status
			}
			services[item.Service] = state
		}
	}

	return services, nil
}

func (a *ComposeAdapter) Logs(ctx context.Context, composePath string, service string, follow bool) error {
	args := []string{"compose", "-f", filepath.Clean(composePath), "logs"}
	if follow {
		args = append(args, "-f")
	}
	if service != "" {
		args = append(args, service)
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (a *ComposeAdapter) Exec(ctx context.Context, composePath string, service string, command []string) (string, error) {
	args := []string{"compose", "-f", filepath.Clean(composePath), "exec", "-T", service}
	args = append(args, command...)

	out, err := a.runner.Run(ctx, "docker", args...)
	if err != nil {
		return "", domain.WrapError(domain.ExitGeneralFailure, "container exec failed: "+string(out), err)
	}
	return string(out), nil
}
