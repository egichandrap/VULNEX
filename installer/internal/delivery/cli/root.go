package cli

import (
	"github.com/vulnex/vulnex/installer/internal/domain"

	"github.com/spf13/cobra"
)

// NewRootCommand creates the top-level 'vulnex' Cobra command.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vulnex",
		Short: "VULNEX — Defensive Vulnerability Assessment Platform CLI",
		Long: `VULNEX CLI is the unified control plane for diagnosing, installing,
and managing VULNEX deployments in Local and Server modes.`,
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	cmd.AddCommand(
		newVersionCmd(),
		newDoctorCmd(),
		newInstallCmd(),
		newStatusCmd(),
		newStartCmd(),
		newStopCmd(),
		newRestartCmd(),
		newLogsCmd(),
		newDashboardCmd(),
		newConfigCmd(),
		newUpdateCmd(),
		newBackupCmd(),
		newRestoreCmd(),
		newUninstallCmd(),
	)

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print VULNEX CLI version and release information",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(domain.CurrentVersion.String())
		},
	}
}
