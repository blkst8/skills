package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"yourproject/internal/app"
	"yourproject/internal/config"
	httpserver "yourproject/internal/http"
	"yourproject/internal/log"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return startFunc()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

// startFunc loads the configuration, wires the global singleton in order
// (graceful shutdown, database, repositories, services), starts the HTTP
// server and blocks until a shutdown signal arrives.
func startFunc() error {
	if err := config.Load(configPath); err != nil {
		return err
	}

	log.Logger.Info("starting service")

	app.WithGracefulShutdown()
	app.WithDatabase()
	app.WithRepository()
	app.WithService()

	srv := httpserver.NewServer()
	srv.Serve()

	app.Wait()

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Logger.Error("failed to shutdown server", zap.Error(err))
	}

	return nil
}
