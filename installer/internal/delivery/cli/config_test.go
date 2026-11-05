package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vulnex/vulnex/installer/internal/delivery/cli"
)

func TestConfigValidateCommand(t *testing.T) {
	tempDir := t.TempDir()
	validConfig := filepath.Join(tempDir, "valid.json")
	_ = os.WriteFile(validConfig, []byte(`{"mode":"local","data_dir":"./data"}`), 0600)

	invalidConfig := filepath.Join(tempDir, "invalid.json")
	_ = os.WriteFile(invalidConfig, []byte(`{"mode":"server"}`), 0600) // missing domain

	rootCmd := cli.NewRootCommand()
	outBuf := &bytes.Buffer{}
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(outBuf)

	// Valid config execution
	rootCmd.SetArgs([]string{"config", "validate", "--config", validConfig})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected valid config validation to succeed, got %v", err)
	}
	if !strings.Contains(outBuf.String(), "Configuration is valid") {
		t.Errorf("expected success message in output, got: %s", outBuf.String())
	}

	// Invalid config execution
	outBuf.Reset()
	rootCmd.SetArgs([]string{"config", "validate", "--config", invalidConfig})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected invalid config validation to fail, got nil")
	}
}

func TestConfigShowCommand_RedactsSecrets(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")
	_ = os.WriteFile(configFile, []byte(`{
		"mode": "server",
		"domain_name": "sec.vulnex.dev",
		"data_dir": "/var/lib/vulnex",
		"database_password": "super-secret-cleartext-password"
	}`), 0600)

	rootCmd := cli.NewRootCommand()
	outBuf := &bytes.Buffer{}
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(outBuf)

	rootCmd.SetArgs([]string{"config", "show", "--config", configFile})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config show failed: %v", err)
	}

	output := outBuf.String()
	if strings.Contains(output, "super-secret-cleartext-password") {
		t.Errorf("SECURITY LEAK: output contains cleartext password: %s", output)
	}
	if !strings.Contains(output, "********") {
		t.Errorf("expected redacted password '********' in output, got: %s", output)
	}
}
