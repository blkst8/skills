package app

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/blkst8/client-service/internal/config"
	"github.com/blkst8/client-service/internal/log"
)

// WithDatabase opens the MySQL connection pool, configures pooling from the
// configuration, verifies the connection with a ping, and returns it.
func WithDatabase() *sqlx.DB {
	cfg := config.C.Database

	db, err := sqlx.Open("mysql", cfg.DSN)
	if err != nil {
		log.Logger.Fatal("failed to connect to database", zap.Error(err))
	}

	db.SetMaxOpenConns(cfg.MaxConn)
	db.SetMaxIdleConns(cfg.IdleConn)
	db.SetConnMaxLifetime(cfg.Timeout)

	if err := db.Ping(); err != nil {
		log.Logger.Fatal("failed to ping database", zap.Error(err))
	}

	return db
}
