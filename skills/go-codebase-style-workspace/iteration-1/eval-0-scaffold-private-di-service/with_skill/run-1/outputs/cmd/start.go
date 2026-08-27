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
	"github.com/blkst8/invoice-service/internal/workers"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the HTTP API and background workers",
	Run:   startFunc,
}

// startFunc wires the whole application using private dependency injection:
// db → repositories → external services → usecases → server/workers.
// Handlers and job handlers receive only the usecase bundle.
func startFunc(_ *cobra.Command, _ []string) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Logger.Info("starting invoice-service",
		zap.String("version", app.GitTag),
		zap.String("commit", app.GitCommit),
	)

	db := app.WithDatabase()
	defer db.Close()

	repo := app.WithRepository(db)
	svc := app.WithServices()
	uc := app.WithUsecases(repo, svc)

	// Interval-based background jobs (ticker workers).
	if config.C.Worker.Enabled {
		reconcileJob := worker.NewWorker(config.C.Worker.JobsIntervals.ReconcileInvoices)
		reconcileJob.RunAsync(workerhandlers.NewReconcileInvoices(uc).Handle)
		defer reconcileJob.Close()
	}

	// Worker pool for concurrent, on-demand tasks.
	if config.C.Workers.Enabled {
		pool := workers.NewPool(config.C.Workers)
		pool.Start()
		defer pool.Stop()
		if err := pool.Submit(workers.NewReconcileInvoicesTask(uc)); err != nil {
			log.Logger.Error("failed to submit task", zap.Error(err))
		}
	}

	srv := httpserver.NewServer(uc)
	srv.Serve()
	defer srv.Shutdown(context.Background())

	<-ctx.Done()
}
