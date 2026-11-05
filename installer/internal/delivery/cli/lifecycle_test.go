package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vulnex/vulnex/installer/internal/delivery/cli"
)

func TestLifecycleCommands(t *testing.T) {
	tempDir := t.TempDir()
	// Create mock state
	statePath := filepath.Join(tempDir, "vulnex.json")
	_ = os.WriteFile(statePath, []byte(`{
		"id": "vulnex-local",
		"mode": "local",
		"version": "1.0.0",
		"status": "ready",
		"phase": "committed",
		"data_dir": "`+tempDir+`"
	}`), 0600)

	rootCmd := cli.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Status command
	rootCmd.SetArgs([]string{"status", "--data-dir", tempDir})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("status command failed: %v", err)
	}
	if !strings.Contains(buf.String(), "VULNEX Deployment Status") {
		t.Errorf("expected status header in output: %s", buf.String())
	}
}

func TestMaintenanceCommands_Validation(t *testing.T) {
	rootCmd := cli.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Restore requires --file
	rootCmd.SetArgs([]string{"restore"})
	if err := rootCmd.Execute(); err == nil {
		t.Error("expected restore without --file to fail, got nil")
	}

	// Uninstall requires explicit data retention
	buf.Reset()
	rootCmd.SetArgs([]string{"uninstall"})
	if err := rootCmd.Execute(); err == nil {
		t.Error("expected uninstall without retention choice to fail, got nil")
	}
}
