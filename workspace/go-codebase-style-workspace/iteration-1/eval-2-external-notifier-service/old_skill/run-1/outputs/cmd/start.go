package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/blkst8/client-service/internal/app"
	"github.com/blkst8/client-service/internal/config"
	httpserver "github.com/blkst8/client-service/internal/http"
	"github.com/blkst8/client-service/internal/log"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the HTTP server",
	Run:   startFunc,
}

// startFunc wires every dependency of the application in order (database,
// repositories, notifier client, services) and starts the HTTP server.
//
// The service bundle produced here is threaded into the HTTP server, which is
// why no global state is needed anywhere in the application.
func startFunc(_ *cobra.Command, _ []string) {
	if err := config.Load(configPath); err != nil {
		log.Logger.Fatal("failed to load configuration", zap.String("path", configPath), zap.Error(err))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	db := app.WithDatabase()
	defer db.Close()

	repo := app.WithRepository(db)
	svc := app.WithServices(repo)

	srv := httpserver.NewServer(svc)
	srv.Serve()
	defer srv.Shutdown(context.Background())

	log.Logger.Info("server started", zap.String("address", config.C.HTTPServer.Listen))

	<-ctx.Done()
}

func init() {
	rootCmd.AddCommand(startCmd)
}
