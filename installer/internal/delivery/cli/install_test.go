package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/vulnex/vulnex/installer/internal/delivery/cli"
)

func TestInstallCommand_NonInteractiveLocal(t *testing.T) {
	tempDir := t.TempDir()
	rootCmd := cli.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	rootCmd.SetArgs([]string{
		"install",
		"--mode", "local",
		"--data-dir", tempDir,
		"--non-interactive",
		"--skip-preflight",
		"--dry-run",
	})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("install non-interactive failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "INITIAL ADMINISTRATOR SETUP TOKEN") {
		t.Errorf("expected setup token in output, got: %s", out)
	}
	if !strings.Contains(out, "Installation successfully completed") {
		t.Errorf("expected success message, got: %s", out)
	}
}

func TestDashboardCommand_Output(t *testing.T) {
	tempDir := t.TempDir()
	rootCmd := cli.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	rootCmd.SetArgs([]string{"dashboard", "--data-dir", tempDir})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("dashboard command failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "http://localhost:3000") {
		t.Errorf("expected dashboard URL in output, got: %s", out)
	}
}
