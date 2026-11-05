package tui

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/vulnex/vulnex/installer/internal/domain"
)

func TestInstallWizard_LocalModeInteractive(t *testing.T) {
	// Simulated user input:
	// 1: Local Mode
	// (empty): Accept default data dir
	// y: Confirm
	input := "1\n\ny\n"
	in := strings.NewReader(input)
	out := &bytes.Buffer{}

	wizard := NewInstallWizard(in, out)
	config, token, err := wizard.Run(context.Background())
	if err != nil {
		t.Fatalf("wizard execution failed: %v", err)
	}

	if config.Mode != domain.DeploymentModeLocal {
		t.Errorf("expected local mode, got: %s", config.Mode)
	}

	if len(token) != 64 {
		t.Errorf("expected 64-char hex setup token, got length %d", len(token))
	}

	outputStr := out.String()
	if !strings.Contains(outputStr, "VULNEX") {
		t.Errorf("expected output to contain VULNEX banner")
	}
	if !strings.Contains(outputStr, "INITIAL ADMINISTRATOR SETUP TOKEN") {
		t.Errorf("expected output to contain token setup box")
	}
	if !strings.Contains(outputStr, token) {
		t.Errorf("expected output to display generated token")
	}
}

func TestInstallWizard_ServerModeInteractive(t *testing.T) {
	// Simulated user input:
	// 2: Server Mode
	// vulnex.corp.net: Domain name
	// 1: Embedded Caddy
	// /opt/vulnex-data: Custom data dir
	// y: Confirm
	input := "2\nvulnex.corp.net\n1\n/opt/vulnex-data\ny\n"
	in := strings.NewReader(input)
	out := &bytes.Buffer{}

	wizard := NewInstallWizard(in, out)
	config, token, err := wizard.Run(context.Background())
	if err != nil {
		t.Fatalf("wizard execution failed: %v", err)
	}

	if config.Mode != domain.DeploymentModeServer {
		t.Errorf("expected server mode, got: %s", config.Mode)
	}
	if config.DomainName != "vulnex.corp.net" {
		t.Errorf("expected domain vulnex.corp.net, got: %s", config.DomainName)
	}
	if config.Ingress != domain.IngressEmbeddedCaddy {
		t.Errorf("expected embedded caddy ingress, got: %s", config.Ingress)
	}
	if config.DataDir != "/opt/vulnex-data" {
		t.Errorf("expected data dir /opt/vulnex-data, got: %s", config.DataDir)
	}

	outputStr := out.String()
	if !strings.Contains(outputStr, "vulnex.corp.net") {
		t.Errorf("expected output to contain domain name")
	}
	if !strings.Contains(outputStr, token) {
		t.Errorf("expected output to display generated token")
	}
}

func TestInstallWizard_AbortOnConfirmation(t *testing.T) {
	// User selects local mode but answers 'n' to confirmation
	input := "1\n\nn\n"
	in := strings.NewReader(input)
	out := &bytes.Buffer{}

	wizard := NewInstallWizard(in, out)
	_, _, err := wizard.Run(context.Background())
	if err == nil {
		t.Fatal("expected error on installation abort, got nil")
	}
	if !strings.Contains(err.Error(), "aborted") {
		t.Errorf("expected error message to mention 'aborted', got: %v", err)
	}
}
