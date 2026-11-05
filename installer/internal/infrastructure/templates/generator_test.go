package templates_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vulnex/vulnex/installer/internal/domain"
	"github.com/vulnex/vulnex/installer/internal/infrastructure/templates"
)

func TestGenerator_RenderLocal(t *testing.T) {
	tempDir := t.TempDir()
	gen := templates.NewGenerator()

	cfg := &domain.InstallConfig{
		Mode:             domain.DeploymentModeLocal,
		DatabasePassword: "random-secure-password-123",
		Version:          "1.0.0",
	}

	if err := gen.Render(cfg, tempDir); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	composeFile := filepath.Join(tempDir, "docker-compose.yml")
	composeData, err := os.ReadFile(composeFile)
	if err != nil {
		t.Fatalf("missing compose file: %v", err)
	}
	if !strings.Contains(string(composeData), "127.0.0.1:3000:3000") {
		t.Errorf("expected local loopback port binding in compose file")
	}

	envFile := filepath.Join(tempDir, ".env")
	envData, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("missing .env file: %v", err)
	}
	if !strings.Contains(string(envData), "random-secure-password-123") {
		t.Errorf("expected generated password in .env")
	}
}

func TestGenerator_RenderServer(t *testing.T) {
	tempDir := t.TempDir()
	gen := templates.NewGenerator()

	cfg := &domain.InstallConfig{
		Mode:             domain.DeploymentModeServer,
		DomainName:       "sec.example.com",
		Ingress:          domain.IngressEmbeddedCaddy,
		DatabasePassword: "db-pass-server-456",
		Version:          "1.0.0",
	}

	if err := gen.Render(cfg, tempDir); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	composeFile := filepath.Join(tempDir, "docker-compose.yml")
	composeData, err := os.ReadFile(composeFile)
	if err != nil {
		t.Fatalf("missing compose file: %v", err)
	}
	if !strings.Contains(string(composeData), "caddy:") {
		t.Errorf("expected caddy service in server compose file")
	}

	caddyFile := filepath.Join(tempDir, "caddy", "Caddyfile")
	caddyData, err := os.ReadFile(caddyFile)
	if err != nil {
		t.Fatalf("missing Caddyfile: %v", err)
	}
	if !strings.Contains(string(caddyData), "reverse_proxy api:8080") {
		t.Errorf("expected api reverse proxy in Caddyfile")
	}
}
