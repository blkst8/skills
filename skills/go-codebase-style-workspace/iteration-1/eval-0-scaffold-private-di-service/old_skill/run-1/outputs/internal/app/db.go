package app

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/blkst8/invoice-service/internal/config"
	"github.com/blkst8/invoice-service/internal/log"
)

// WithDatabase opens the MySQL connection pool, configures it and verifies
// connectivity before returning the handle.
func WithDatabase() *sqlx.DB {
	cfg := config.C.Database

	db, err := sqlx.Open("mysql", cfg.DSN)
	if err != nil {
		log.Logger.Fatal("failed to open database connection", zap.Error(err))
	}

	db.SetMaxOpenConns(cfg.MaxConn)
	db.SetMaxIdleConns(cfg.IdleConn)
	db.SetConnMaxLifetime(cfg.Timeout)

	if err := db.Ping(); err != nil {
		log.Logger.Fatal("failed to ping database", zap.Error(err))
	}

	return db
}
