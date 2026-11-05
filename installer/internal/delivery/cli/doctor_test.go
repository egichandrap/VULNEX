package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/vulnex/vulnex/installer/internal/delivery/cli"
)

func TestDoctorCommand_Execution(t *testing.T) {
	rootCmd := cli.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"doctor", "--json"})

	// Execution should complete (may pass or fail exit depending on host Docker, but produces valid JSON output)
	_ = rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, `"status":`) {
		t.Errorf("expected doctor --json to output status field, got: %s", output)
	}
	if !strings.Contains(output, `"checks":`) {
		t.Errorf("expected doctor --json to output checks field, got: %s", output)
	}
}
