// Package cmd provides the invoice-service CLI commands.
//
// Each file in this package is a Cobra command: root.go (this file),
// start.go (server + workers) and migrate.go (database migrations).
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/blkst8/invoice-service/internal/config"
	"github.com/blkst8/invoice-service/internal/log"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "invoice-service",
	Short: "Invoice service: REST API and background reconciliation jobs",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		if err := config.Load(cfgFile); err != nil {
			log.Logger.Fatal("failed to load config", zap.String("path", cfgFile), zap.Error(err))
		}
		log.SetLevel(config.C.Logger.Level)
	},
}

// Execute runs the root command and exits with status 1 on failure.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.yaml", "path to the configuration file")
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(migrateCmd)
}
