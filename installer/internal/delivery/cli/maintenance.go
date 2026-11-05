package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vulnex/vulnex/installer/internal/application"
	"github.com/vulnex/vulnex/installer/internal/infrastructure/docker"
	"github.com/vulnex/vulnex/installer/internal/infrastructure/storage"
)

func newBackupCmd() *cobra.Command {
	var (
		dataDir    string
		outputPath string
	)

	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup persistent data (database, configuration, evidence)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dataDir == "" {
				dataDir = "./vulnex-data"
			}
			if outputPath == "" {
				outputPath = "./vulnex-backup"
			}

			store := storage.NewFileStorage()
			runtime := docker.NewComposeAdapter()
			uc := application.NewMaintenanceUseCase(runtime, store)

			cmd.Printf("Creating VULNEX backup from %s to %s...\n", dataDir, outputPath)
			archivePath, err := uc.Backup(cmd.Context(), dataDir, outputPath)
			if err != nil {
				return err
			}

			cmd.Printf("✓ Backup successfully created at: %s\n", archivePath)
			return nil
		},
	}

	cmd.Flags().StringVar(&dataDir, "data-dir", "./vulnex-data", "Persistent data directory path")
	cmd.Flags().StringVar(&outputPath, "output", "./vulnex-backup", "Destination directory or file path")
	return cmd
}

func newRestoreCmd() *cobra.Command {
	var (
		dataDir  string
		filePath string
	)

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
			if dataDir == "" {
				dataDir = "./vulnex-data"
			}

			store := storage.NewFileStorage()
			runtime := docker.NewComposeAdapter()
			uc := application.NewMaintenanceUseCase(runtime, store)

			cmd.Printf("Restoring backup from %s into %s...\n", filePath, dataDir)
			if err := uc.Restore(cmd.Context(), dataDir, filePath); err != nil {
				return err
			}

			cmd.Println("✓ Backup successfully restored.")
			return nil
		},
	}

	cmd.Flags().StringVar(&dataDir, "data-dir", "./vulnex-data", "Persistent data directory path")
	cmd.Flags().StringVar(&filePath, "file", "", "Path to backup file")
	return cmd
}

func newUpdateCmd() *cobra.Command {
	var (
		dataDir       string
		targetVersion string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Verify and apply a VULNEX application update",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dataDir == "" {
				dataDir = "./vulnex-data"
			}
			if targetVersion == "" {
				targetVersion = "latest"
			}

			store := storage.NewFileStorage()
			runtime := docker.NewComposeAdapter()
			uc := application.NewMaintenanceUseCase(runtime, store)

			cmd.Printf("Applying VULNEX update in %s (version: %s)...\n", dataDir, targetVersion)
			if err := uc.Update(cmd.Context(), dataDir, targetVersion); err != nil {
				return err
			}

			cmd.Println("✓ Update applied successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&dataDir, "data-dir", "./vulnex-data", "Persistent data directory path")
	cmd.Flags().StringVar(&targetVersion, "version", "", "Target release version to update to")
	return cmd
}

func newUninstallCmd() *cobra.Command {
	var (
		dataDir    string
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
			if dataDir == "" {
				dataDir = "./vulnex-data"
			}

			store := storage.NewFileStorage()
			runtime := docker.NewComposeAdapter()
			uc := application.NewMaintenanceUseCase(runtime, store)

			if deleteData {
				cmd.Printf("Uninstalling VULNEX in %s and permanently deleting all persistent data...\n", dataDir)
			} else {
				cmd.Printf("Uninstalling VULNEX in %s while preserving persistent data...\n", dataDir)
			}

			if err := uc.Uninstall(cmd.Context(), dataDir, deleteData); err != nil {
				return err
			}

			cmd.Println("✓ Uninstallation completed.")
			return nil
		},
	}

	cmd.Flags().StringVar(&dataDir, "data-dir", "./vulnex-data", "Persistent data directory path")
	cmd.Flags().BoolVar(&keepData, "keep-data", false, "Retain persistent application data and databases")
	cmd.Flags().BoolVar(&deleteData, "delete-data", false, "Permanently delete persistent data (requires confirmation)")
	return cmd
}
