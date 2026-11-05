package docker_test

import (
	"context"
	"testing"

	"github.com/vulnex/vulnex/installer/internal/infrastructure/docker"
)

type mockCommandRunner struct {
	lastCmd  string
	lastArgs []string
	output   string
	err      error
}

func (m *mockCommandRunner) Run(ctx context.Context, cmd string, args ...string) ([]byte, error) {
	m.lastCmd = cmd
	m.lastArgs = args
	return []byte(m.output), m.err
}

func TestComposeAdapter_Up(t *testing.T) {
	runner := &mockCommandRunner{}
	adapter := docker.NewComposeAdapterWithRunner(runner)

	err := adapter.Up(context.Background(), "/data/docker-compose.yml", "/data/.env")
	if err != nil {
		t.Fatalf("Up failed: %v", err)
	}

	if runner.lastCmd != "docker" {
		t.Errorf("expected command docker, got %s", runner.lastCmd)
	}
	expectedArgs := []string{"compose", "-f", "/data/docker-compose.yml", "--env-file", "/data/.env", "up", "-d"}
	for i, arg := range expectedArgs {
		if runner.lastArgs[i] != arg {
			t.Errorf("arg %d: expected %s, got %s", i, arg, runner.lastArgs[i])
		}
	}
}

func TestComposeAdapter_Down(t *testing.T) {
	runner := &mockCommandRunner{}
	adapter := docker.NewComposeAdapterWithRunner(runner)

	// Down without volume removal
	_ = adapter.Down(context.Background(), "/data/docker-compose.yml", false)
	if runner.lastArgs[len(runner.lastArgs)-1] == "-v" {
		t.Errorf("unexpected -v flag in Down")
	}

	// Down with volume removal
	_ = adapter.Down(context.Background(), "/data/docker-compose.yml", true)
	if runner.lastArgs[len(runner.lastArgs)-1] != "-v" {
		t.Errorf("expected -v flag when removeVolumes is true")
	}
}

func TestComposeAdapter_Exec(t *testing.T) {
	runner := &mockCommandRunner{output: "INSERT 0 1\n"}
	adapter := docker.NewComposeAdapterWithRunner(runner)

	out, err := adapter.Exec(context.Background(), "/data/docker-compose.yml", "postgres", []string{"psql", "-c", "SELECT 1"})
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}
	if out != "INSERT 0 1\n" {
		t.Errorf("unexpected output: %s", out)
	}
}
