// Package cmd contains the cobra commands for invoice-service.
//
// Each file in this package defines exactly one command.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/blkst8/invoice-service/internal/config"
	"github.com/blkst8/invoice-service/internal/log"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "invoice-service",
	Short: "Invoice service: REST API and background jobs",
	Long: `invoice-service exposes a REST API for managing invoices and runs
background jobs such as the periodic invoice reconciliation.

Use the --config flag to point every command at a configuration file.`,
}

// Execute runs the root command and terminates the process with a non-zero
// exit code when a command fails.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config.yaml", "path to the configuration file")
}

// loadConfig reads the configuration file and applies the configured log level.
func loadConfig() error {
	if err := config.Load(configPath); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := log.SetLevel(config.C.Logger.Level); err != nil {
		return fmt.Errorf("failed to configure logger: %w", err)
	}

	return nil
}
