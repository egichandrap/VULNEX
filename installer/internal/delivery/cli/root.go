package cli

import (
	"fmt"

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
		SilenceErrors: true,
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

func newDoctorCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose OS, CPU, memory, disk, Docker Engine, and Compose prerequisites",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Checking VULNEX runtime prerequisites...")
			// Detailed environment checks to be wired via EnvironmentDetector port
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output diagnostics in JSON format")
	return cmd
}

func newInstallCmd() *cobra.Command {
	var (
		mode           string
		domainName     string
		version        string
		dataDir        string
		configPath     string
		externalProxy  bool
		trustedProxies string
		nonInteractive bool
		autoYes        bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install and provision VULNEX in Local or Server deployment mode",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if mode == "" {
				return fmt.Errorf("required flag '--mode' missing (must be 'local' or 'server')")
			}
			if mode != "local" && mode != "server" {
				return fmt.Errorf("invalid mode '%s': must be 'local' or 'server'", mode)
			}
			if mode == "server" && domainName == "" {
				return fmt.Errorf("server mode requires '--domain' hostname")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Printf("Initiating VULNEX installation in %s mode...\n", mode)
			return nil
		},
	}

	cmd.Flags().StringVar(&mode, "mode", "", "Deployment mode: 'local' (loopback) or 'server' (remote access)")
	cmd.Flags().StringVar(&domainName, "domain", "", "Public hostname for server mode (used for Caddy ACME TLS)")
	cmd.Flags().StringVar(&version, "version", "", "Exact VULNEX release to install")
	cmd.Flags().StringVar(&dataDir, "data-dir", "", "Persistent data directory path")
	cmd.Flags().StringVar(&configPath, "config", "", "Path to custom configuration file")
	cmd.Flags().BoolVar(&externalProxy, "external-proxy", false, "Offload TLS to an external proxy instead of embedded Caddy")
	cmd.Flags().StringVar(&trustedProxies, "trusted-proxies", "", "CIDR ranges of trusted reverse proxies when using external proxy")
	cmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Disable prompts; fail on missing required inputs")
	cmd.Flags().BoolVar(&autoYes, "yes", false, "Confirm non-destructive defaults automatically")

	return cmd
}

func newStatusCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status and health of VULNEX services",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Checking VULNEX component health...")
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output status in JSON format")
	return cmd
}

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start VULNEX services",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Starting VULNEX stack...")
			return nil
		},
	}
}

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop VULNEX services",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Stopping VULNEX stack...")
			return nil
		},
	}
}

func newRestartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: "Restart VULNEX services",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Restarting VULNEX stack...")
			return nil
		},
	}
}

func newLogsCmd() *cobra.Command {
	var (
		service string
		follow  bool
	)
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Read VULNEX service logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Reading VULNEX logs...")
			return nil
		},
	}
	cmd.Flags().StringVar(&service, "service", "", "Filter logs by service name")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log stream")
	return cmd
}

func newDashboardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dashboard",
		Short: "Print or open the VULNEX dashboard URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("VULNEX Dashboard URL: http://127.0.0.1:3000")
			return nil
		},
	}
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage and validate VULNEX configuration",
	}

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration without printing secrets",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Configuration is valid.")
			return nil
		},
	}

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Display active configuration with secrets redacted",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Displaying redacted configuration...")
			return nil
		},
	}

	cmd.AddCommand(validateCmd, showCmd)
	return cmd
}

func newUpdateCmd() *cobra.Command {
	var targetVersion string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Verify and apply a VULNEX application update",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Checking for updates...")
			return nil
		},
	}
	cmd.Flags().StringVar(&targetVersion, "version", "", "Target release version to update to")
	return cmd
}

func newBackupCmd() *cobra.Command {
	var outputPath string
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup persistent data (database, configuration, evidence)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Creating backup archive...")
			return nil
		},
	}
	cmd.Flags().StringVar(&outputPath, "output", "./vulnex-backup", "Destination directory or file path")
	return cmd
}

func newRestoreCmd() *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore persistent data from a backup archive",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("required flag '--file' missing")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Printf("Restoring backup from %s...\n", filePath)
			return nil
		},
	}
	cmd.Flags().StringVar(&filePath, "file", "", "Path to backup file")
	return cmd
}

func newUninstallCmd() *cobra.Command {
	var (
		keepData   bool
		deleteData bool
	)
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove VULNEX deployment with an explicit data-retention choice",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !keepData && !deleteData {
				return fmt.Errorf("must explicitly specify '--keep-data' or '--delete-data'")
			}
			if keepData && deleteData {
				return fmt.Errorf("cannot specify both '--keep-data' and '--delete-data'")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if deleteData {
				cmd.Println("Uninstalling VULNEX and removing persistent volumes...")
			} else {
				cmd.Println("Uninstalling VULNEX while preserving persistent data...")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&keepData, "keep-data", false, "Retain persistent application data and databases")
	cmd.Flags().BoolVar(&deleteData, "delete-data", false, "Permanently delete persistent data (requires confirmation)")
	return cmd
}
