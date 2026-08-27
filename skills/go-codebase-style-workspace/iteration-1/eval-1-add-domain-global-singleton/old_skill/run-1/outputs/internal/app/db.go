package app

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"yourproject/internal/config"
	"yourproject/internal/log"
)

// WithDatabase opens the database connection and stores it on the
// application singleton.
func WithDatabase() {
	var err error
	cfg := config.C.Database

	A.Database, err = sqlx.Open("mysql", cfg.DSN)
	if err != nil {
		log.Logger.Fatal("failed to connect to database", zap.Error(err))
	}

	A.Database.SetMaxOpenConns(cfg.MaxConn)
	A.Database.SetMaxIdleConns(cfg.IdleConn)
	A.Database.SetConnMaxLifetime(cfg.Timeout)

	// Test connection
	if err := A.Database.Ping(); err != nil {
		log.Logger.Fatal("failed to ping database", zap.Error(err))
	}
}
