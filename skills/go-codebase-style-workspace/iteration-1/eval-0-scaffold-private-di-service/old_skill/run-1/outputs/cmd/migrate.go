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
	Use:   "migrate",
	Short: "Apply all pending database migrations",
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := loadConfig(); err != nil {
			return err
		}

		m, err := migrate.New("file://migrations", config.C.Database.MigrateDSN())
		if err != nil {
			return fmt.Errorf("failed to initialize migrations: %w", err)
		}
		defer m.Close()

		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations: %w", err)
		}

		log.Logger.Info("database migrations applied")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
