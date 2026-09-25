// Package cmd contains the Cobra commands of the application.
//
// Each file in this package defines exactly one command:
//   - root.go: the root command and shared flags
//   - start.go: the start command, which boots the HTTP server
package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/blkst8/client-service/internal/log"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "client-service",
	Short: "Client service with Telegram welcome notifications",
	Long:  "client-service stores clients and sends a Telegram welcome message to every new sign-up.",
}

// Execute runs the root command and exits the process on failure.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Logger.Fatal("command execution failed", zap.Error(err))
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "config.yaml", "path to the configuration file")
}
