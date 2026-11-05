package cli

import (
	"bytes"
	"testing"

	"github.com/vulnex/vulnex/installer/internal/domain"
)

func TestRootCommand_HasRequiredSubcommands(t *testing.T) {
	rootCmd := NewRootCommand()

	expectedSubcommands := []string{
		"version",
		"doctor",
		"install",
		"status",
		"start",
		"stop",
		"restart",
		"logs",
		"dashboard",
		"config",
		"update",
		"backup",
		"restore",
		"uninstall",
	}

	for _, name := range expectedSubcommands {
		cmd, _, err := rootCmd.Find([]string{name})
		if err != nil || cmd == nil || cmd.Name() != name {
			t.Errorf("Expected subcommand '%s' to be registered on root command", name)
		}
	}
}

func TestVersionCommand_OutputsVersion(t *testing.T) {
	rootCmd := NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Errorf("Expected version command to produce output, got empty")
	}
}

func TestInstallCommand_FailsWithoutMode(t *testing.T) {
	rootCmd := NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"install", "--non-interactive"})

	err := rootCmd.Execute()
	if err == nil {
		t.Errorf("Expected install command without --mode to fail, but it succeeded")
	}
}

func TestInstallCommand_RunsTUIWizardInteractive(t *testing.T) {
	t.Setenv("VULNEX_SKIP_PREFLIGHT", "1")
	t.Setenv("VULNEX_DRY_RUN", "1")

	rootCmd := NewRootCommand()
	buf := new(bytes.Buffer)
	in := bytes.NewBufferString("1\n\ny\n")
	rootCmd.SetIn(in)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"install"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("interactive install wizard failed: %v", err)
	}

	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("INITIAL ADMINISTRATOR SETUP TOKEN")) {
		t.Errorf("expected output to contain setup token card, got: %s", output)
	}
}

func TestResolveExitCode(t *testing.T) {
	if code := ResolveExitCode(nil); code != 0 {
		t.Errorf("expected 0 for nil error, got %d", code)
	}

	domErr := domain.NewError(domain.ExitInvalidConfig, "invalid config")
	if code := ResolveExitCode(domErr); code != int(domain.ExitInvalidConfig) {
		t.Errorf("expected %d for domain error, got %d", domain.ExitInvalidConfig, code)
	}
}
