package cmd

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"

	"github.com/blkst8/invoice-service/internal/config"
	"github.com/blkst8/invoice-service/internal/log"
)

var migrateCmd = &cobra.Command{
	Use:       "migrate [up|down]",
	Short:     "Apply or roll back database migrations",
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: []string{"up", "down"},
	RunE: func(_ *cobra.Command, args []string) error {
		m, err := migrate.New("file://migrations", "mysql://"+config.C.Database.DSN)
		if err != nil {
			return fmt.Errorf("failed to create migrator: %w", err)
		}
		defer m.Close()

		switch args[0] {
		case "up":
			if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return fmt.Errorf("failed to apply migrations: %w", err)
			}
			log.Logger.Info("migrations applied")
		case "down":
			if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return fmt.Errorf("failed to roll back migration: %w", err)
			}
			log.Logger.Info("migration rolled back")
		}

		return nil
	},
}
