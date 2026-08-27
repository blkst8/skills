package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/blkst8/invoice-service/internal/app"
	"github.com/blkst8/invoice-service/internal/config"
	httpserver "github.com/blkst8/invoice-service/internal/http"
	"github.com/blkst8/invoice-service/internal/log"
	"github.com/blkst8/invoice-service/internal/worker"
	workerhandlers "github.com/blkst8/invoice-service/internal/worker/handlers"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the HTTP server and the background workers",
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := loadConfig(); err != nil {
			return err
		}

		startFunc()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

// startFunc wires the whole application together using the private
// dependency injection pattern: every layer receives its dependencies
// explicitly from the layer above it, with no global state.
func startFunc() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Logger.Info("starting invoice-service",
		zap.String("git_tag", app.GitTag),
		zap.String("git_commit", app.GitCommit),
		zap.String("build_date", app.BuildDate),
	)

	db := app.WithDatabase()
	defer db.Close()

	repo := app.WithRepository(db)
	svc := app.WithServices(db, repo)

	if config.C.Worker.Enabled {
		reconcileJob := worker.NewWorker(config.C.Worker.JobsIntervals.ReconcileInvoices)
		reconcileJob.RunAsync(workerhandlers.NewReconcileInvoices(svc).Handle)
		defer reconcileJob.Close()
	}

	srv := httpserver.NewServer(svc)
	srv.Serve()
	defer srv.Shutdown(context.Background())

	log.Logger.Info("invoice-service started", zap.String("address", config.C.HTTPServer.Listen))

	<-ctx.Done()
}
