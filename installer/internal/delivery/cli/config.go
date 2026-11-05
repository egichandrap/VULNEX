package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vulnex/vulnex/installer/internal/application"
	"github.com/vulnex/vulnex/installer/internal/infrastructure/storage"
)

func newConfigCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage and validate VULNEX configuration",
	}

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration without printing secrets",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := storage.NewFileStorage()
			uc := application.NewValidateConfigUseCase(store)

			if err := uc.Execute(cmd.Context(), configPath); err != nil {
				return err
			}

			cmd.Println("✓ Configuration is valid and satisfies all domain invariants.")
			return nil
		},
	}
	validateCmd.Flags().StringVar(&configPath, "config", "./vulnex-config.json", "Path to VULNEX configuration file")

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Display active configuration with secrets redacted",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := storage.NewFileStorage()
			uc := application.NewShowConfigUseCase(store)

			redacted, err := uc.Execute(cmd.Context(), configPath)
			if err != nil {
				return err
			}

			data, err := json.MarshalIndent(redacted, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format configuration JSON: %w", err)
			}

			cmd.Println(string(data))
			return nil
		},
	}
	showCmd.Flags().StringVar(&configPath, "config", "./vulnex-config.json", "Path to VULNEX configuration file")

	cmd.AddCommand(validateCmd, showCmd)
	return cmd
}
