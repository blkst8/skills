// Package cmd contains the Cobra commands of the service.
//
// Each file defines exactly one command.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// configPath is the value of the persistent --config flag.
var configPath string

var rootCmd = &cobra.Command{
	Use:   "yourproject",
	Short: "yourproject HTTP service",
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "config.yaml", "path to the configuration file")
}
