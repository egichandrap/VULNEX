package main

import (
	"os"

	"github.com/vulnex/vulnex/installer/internal/delivery/cli"
)

func main() {
	rootCmd := cli.NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
