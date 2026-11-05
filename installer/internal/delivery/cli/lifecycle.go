package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/vulnex/vulnex/installer/internal/application"
	"github.com/vulnex/vulnex/installer/internal/infrastructure/docker"
	"github.com/vulnex/vulnex/installer/internal/infrastructure/storage"
)

func newStatusCmd() *cobra.Command {
	var (
		dataDir    string
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status and health of VULNEX services",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dataDir == "" {
				dataDir = "./vulnex-data"
			}

			store := storage.NewFileStorage()
			runtime := docker.NewComposeAdapter()
			uc := application.NewLifecycleUseCase(runtime, store)

			state, services, err := uc.Status(cmd.Context(), dataDir)
			if err != nil {
				return err
			}

			if jsonOutput {
				out := map[string]interface{}{
					"state":    state,
					"services": services,
				}
				data, _ := json.MarshalIndent(out, "", "  ")
				cmd.Println(string(data))
				return nil
			}

			cmd.Println("\n=== VULNEX Deployment Status ===")
			cmd.Printf("  Status:      %s\n", state.Status)
			cmd.Printf("  Mode:        %s\n", state.Mode)
			cmd.Printf("  Version:     %s\n", state.Version)
			cmd.Printf("  Phase:       %s\n", state.Phase)
			cmd.Printf("  Data Dir:    %s\n", state.DataDir)
			cmd.Println("--- Services ---")
			if len(services) == 0 {
				cmd.Println("  (no active containers detected)")
			} else {
				for svc, st := range services {
					cmd.Printf("  %-16s : %s\n", svc, st)
				}
			}
			cmd.Println("================================")
			return nil
		},
	}

	cmd.Flags().StringVar(&dataDir, "data-dir", "./vulnex-data", "Persistent data directory path")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output status in JSON format")
	return cmd
}

func newStartCmd() *cobra.Command {
	var dataDir string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start VULNEX services",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dataDir == "" {
				dataDir = "./vulnex-data"
			}
			store := storage.NewFileStorage()
			runtime := docker.NewComposeAdapter()
			uc := application.NewLifecycleUseCase(runtime, store)

			cmd.Printf("Starting VULNEX stack in %s...\n", dataDir)
			if err := uc.Start(cmd.Context(), dataDir); err != nil {
				return err
			}
			cmd.Println("✓ VULNEX stack started successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&dataDir, "data-dir", "./vulnex-data", "Persistent data directory path")
	return cmd
}

func newStopCmd() *cobra.Command {
	var dataDir string

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop VULNEX services",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dataDir == "" {
				dataDir = "./vulnex-data"
			}
			store := storage.NewFileStorage()
			runtime := docker.NewComposeAdapter()
			uc := application.NewLifecycleUseCase(runtime, store)

			cmd.Printf("Stopping VULNEX stack in %s...\n", dataDir)
			if err := uc.Stop(cmd.Context(), dataDir); err != nil {
				return err
			}
			cmd.Println("✓ VULNEX stack stopped.")
			return nil
		},
	}

	cmd.Flags().StringVar(&dataDir, "data-dir", "./vulnex-data", "Persistent data directory path")
	return cmd
}

func newRestartCmd() *cobra.Command {
	var dataDir string

	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart VULNEX services",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dataDir == "" {
				dataDir = "./vulnex-data"
			}
			store := storage.NewFileStorage()
			runtime := docker.NewComposeAdapter()
			uc := application.NewLifecycleUseCase(runtime, store)

			cmd.Printf("Restarting VULNEX stack in %s...\n", dataDir)
			if err := uc.Restart(cmd.Context(), dataDir); err != nil {
				return err
			}
			cmd.Println("✓ VULNEX stack restarted successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&dataDir, "data-dir", "./vulnex-data", "Persistent data directory path")
	return cmd
}

func newLogsCmd() *cobra.Command {
	var (
		dataDir string
		service string
		follow  bool
	)

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Read VULNEX service logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dataDir == "" {
				dataDir = "./vulnex-data"
			}
			store := storage.NewFileStorage()
			runtime := docker.NewComposeAdapter()
			uc := application.NewLifecycleUseCase(runtime, store)

			return uc.Logs(cmd.Context(), dataDir, service, follow)
		},
	}

	cmd.Flags().StringVar(&dataDir, "data-dir", "./vulnex-data", "Persistent data directory path")
	cmd.Flags().StringVar(&service, "service", "", "Filter logs by service name")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log stream")
	return cmd
}
