package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vulnex/vulnex/installer/internal/application"
	"github.com/vulnex/vulnex/installer/internal/delivery/tui"
	"github.com/vulnex/vulnex/installer/internal/domain"
	"github.com/vulnex/vulnex/installer/internal/infrastructure/docker"
	"github.com/vulnex/vulnex/installer/internal/infrastructure/storage"
	"github.com/vulnex/vulnex/installer/internal/infrastructure/sysinfo"
	"github.com/vulnex/vulnex/installer/internal/infrastructure/templates"
)

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
		skipPreflight  bool
		dryRun         bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install and provision VULNEX in Local or Server deployment mode",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if nonInteractive {
				if mode == "" {
					return fmt.Errorf("required flag '--mode' missing (must be 'local' or 'server')")
				}
				if mode != "local" && mode != "server" {
					return fmt.Errorf("invalid mode '%s': must be 'local' or 'server'", mode)
				}
				if mode == "server" && domainName == "" {
					return fmt.Errorf("server mode requires '--domain' hostname")
				}
			} else if mode != "" && mode != "local" && mode != "server" {
				return fmt.Errorf("invalid mode '%s': must be 'local' or 'server'", mode)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			store := storage.NewFileStorage()
			lock := storage.NewFileLock()
			runtime := docker.NewComposeAdapter()
			detector := sysinfo.NewDetector()
			renderer := templates.NewGenerator()

			installUC := application.NewInstallVulnexUseCase(
				runtime,
				store,
				store,
				lock,
				detector,
				renderer,
			)

			// 1. Interactive TUI Mode
			if !nonInteractive && mode == "" {
				wizard := tui.NewInstallWizard(cmd.InOrStdin(), cmd.OutOrStdout())
				cfg, _, err := wizard.Run(cmd.Context())
				if err != nil {
					return err
				}
				cfg.SkipPreflight = skipPreflight
				cfg.DryRun = dryRun

				cmd.Printf("\nInitiating VULNEX deployment for %s mode...\n", cfg.Mode)
				state, token, err := installUC.Execute(cmd.Context(), cfg)
				if err != nil {
					cmd.Printf("Deployment failed: %v\n", err)
					return err
				}

				cmd.Printf("✓ Installation successfully committed (State: %s).\n", state.Status)
				printTokenBox(cmd, token, cfg)
				return nil
			}

			// 2. Non-interactive or flag-based execution
			ingress := domain.IngressEmbeddedCaddy
			if externalProxy {
				ingress = domain.IngressExternalProxy
			}

			cfg := &domain.InstallConfig{
				Mode:           domain.DeploymentMode(mode),
				DomainName:     domainName,
				Ingress:        ingress,
				TrustedProxies: trustedProxies,
				DataDir:        dataDir,
				Version:        version,
				SkipPreflight:  skipPreflight,
				DryRun:         dryRun,
			}

			cmd.Printf("Initiating VULNEX installation in %s mode...\n", mode)
			state, token, err := installUC.Execute(cmd.Context(), cfg)
			if err != nil {
				cmd.Printf("Deployment error: %v\n", err)
				return err
			}

			cmd.Printf("✓ Installation successfully completed (Status: %s).\n", state.Status)
			printTokenBox(cmd, token, cfg)
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
	cmd.Flags().BoolVar(&skipPreflight, "skip-preflight", false, "Skip environment and runtime pre-flight checks (testing/advanced)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate installation without starting containers")

	return cmd
}

func printTokenBox(cmd *cobra.Command, token string, config *domain.InstallConfig) {
	onboardingURL := "http://localhost:3000/setup"
	if config.Mode == domain.DeploymentModeServer && config.DomainName != "" {
		onboardingURL = fmt.Sprintf("https://%s/setup", config.DomainName)
	}

	box := fmt.Sprintf(`
┌────────────────────────────────────────────────────────────────────────┐
│                   INITIAL ADMINISTRATOR SETUP TOKEN                    │
├────────────────────────────────────────────────────────────────────────┤
│  Token:  %s  │
│                                                                        │
│  SECURITY NOTICE (ADR-013):                                            │
│  - This setup token expires in 60 minutes and is single-use.           │
│  - It will NEVER be displayed again in plaintext.                      │
│                                                                        │
│  Next Step:                                                            │
│  Navigate to %s in your browser to complete           │
│  administrator account creation and platform onboarding.               │
└────────────────────────────────────────────────────────────────────────┘
`, token, onboardingURL)

	cmd.Println(box)
}
