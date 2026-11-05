package cli

import (
	"github.com/spf13/cobra"
	"github.com/vulnex/vulnex/installer/internal/application"
	"github.com/vulnex/vulnex/installer/internal/infrastructure/storage"
)

func newDashboardCmd() *cobra.Command {
	var dataDir string

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Print or open the VULNEX dashboard URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dataDir == "" {
				dataDir = "./vulnex-data"
			}

			store := storage.NewFileStorage()
			uc := application.NewDashboardUseCase(store, store)

			url, err := uc.ResolveURL(cmd.Context(), dataDir)
			if err != nil {
				return err
			}

			cmd.Printf("VULNEX Dashboard URL: %s\n", url)
			if err := uc.OpenBrowser(url); err != nil {
				cmd.Printf("(Browser launch unavailable: %v. Please open URL manually)\n", err)
			} else {
				cmd.Println("Opening dashboard in default web browser...")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dataDir, "data-dir", "./vulnex-data", "Persistent data directory path")
	return cmd
}
