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

func startFunc(_ *cobra.Command, _ []string) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if config.C.Telegram.Token == "" {
		log.Logger.Fatal("telegram bot token is not configured")
	}

	log.Logger.Info("starting client service",
		zap.String("tag", app.GitTag),
		zap.String("commit", app.GitCommit),
		zap.String("build_date", app.BuildDate),
	)

	db := app.WithDatabase()
	defer db.Close()

	repo := app.WithRepository(db)
	svc := app.WithServices()
	uc := app.WithUsecases(repo, svc)

	srv := httpserver.NewServer(uc)
	srv.Serve()
	defer srv.Shutdown(context.Background())

	<-ctx.Done()
	log.Logger.Info("shutting down")
}

func init() {
	rootCmd.AddCommand(startCmd)
}
