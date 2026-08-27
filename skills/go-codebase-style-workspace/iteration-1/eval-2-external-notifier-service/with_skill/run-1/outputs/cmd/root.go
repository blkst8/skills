// Package cmd provides the Cobra commands of the client service.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/blkst8/client-service/internal/app"
	"github.com/blkst8/client-service/internal/config"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   app.Name,
	Short: "Client service with Telegram welcome notifications",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		if err := config.Load(configPath); err != nil {
			fmt.Fprintln(os.Stderr, "failed to load config:", err)
			os.Exit(1)
		}
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "config.yaml", "path to the configuration file")
}
