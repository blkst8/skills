# Configuration, Logging & Database

Covers: Viper config structs and YAML, zap logging setup/usage, DB connection
setup, migrations, and sqlx query patterns.

---

## Configuration Management (Viper)

**File**: `internal/config/config.go`

```go
package config

import (
    "github.com/spf13/viper"
    "time"
)

var C *Config

type Config struct {
    Logger     Logger     `yaml:"logger"`
    HTTPServer HTTPServer `yaml:"http_server"`
    Database   Database   `yaml:"database"`
    Worker     Worker     `yaml:"worker"`
    Workers    Workers    `yaml:"workers"`
}

type Logger struct {
    Level string `yaml:"level"`
}

type HTTPServer struct {
    Listen            string        `yaml:"listen"`
    ReadTimeout       time.Duration `yaml:"read_timeout"`
    WriteTimeout      time.Duration `yaml:"write_timeout"`
    ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
    IdleTimeout       time.Duration `yaml:"idle_timeout"`
}

type Database struct {
    DSN      string        `yaml:"dsn"`
    MaxConn  int           `yaml:"max_conn"`
    IdleConn int           `yaml:"idle_conn"`
    Timeout  time.Duration `yaml:"timeout"`
}

type Worker struct {
    Enabled       bool          `yaml:"enabled"`
    JobsIntervals JobsIntervals `yaml:"jobs_intervals"`
}

type JobsIntervals struct {
    SyncDatabases time.Duration `yaml:"sync_databases"`
    CleanupJobs   time.Duration `yaml:"cleanup_jobs"`
}

// Workers configures the worker pool (internal/workers).
type Workers struct {
    Enabled   bool          `yaml:"enabled"`
    Count     int           `yaml:"count"`
    QueueSize int           `yaml:"queue_size"`
    Timeout   time.Duration `yaml:"timeout"`
    Retries   int           `yaml:"retries"`
}

func Load(configPath string) error {
    viper.SetConfigFile(configPath)
    viper.SetConfigType("yaml")

    if err := viper.ReadInConfig(); err != nil {
        return err
    }

    if err := viper.Unmarshal(&C); err != nil {
        return err
    }

    return nil
}
```

### Configuration File (YAML)

**File**: `config.example.yaml`

```yaml
logger:
  level: info

http_server:
  listen: :8080
  read_timeout: 30s
  write_timeout: 30s
  read_header_timeout: 10s
  idle_timeout: 120s

database:
  dsn: "user:pass@tcp(localhost:3306)/dbname?parseTime=true"
  max_conn: 25
  idle_conn: 5
  timeout: 30s

worker:
  enabled: true
  jobs_intervals:
    sync_databases: 5m
    cleanup_jobs: 1h

workers:
  enabled: true
  count: 4
  queue_size: 128
  timeout: 2m
  retries: 3
```

### Best Practices

- Provide `config.example.yaml` in the repo
- Use environment-specific configs (dev, staging, prod)
- Never commit actual `config.yaml` with secrets
- Support environment variable overrides
- Validate configuration on startup

---

## Logging (zap)

### Setup

**File**: `internal/log/log.go`

```go
package log

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func init() {
    var err error
    config := zap.NewProductionConfig()
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

    Logger, err = config.Build()
    if err != nil {
        panic(err)
    }
}
```

### Usage Patterns

```go
import (
    "go.uber.org/zap"
    "yourproject/internal/log"
)

// Info level
log.Logger.Info("server started", zap.String("address", ":8080"))

// Error level with error
log.Logger.Error("failed to connect", zap.Error(err))

// With multiple fields
log.Logger.Info("request processed",
    zap.String("method", "POST"),
    zap.String("path", "/api/clients"),
    zap.Int("status", 200),
    zap.Duration("latency", duration),
)

// Structured logging
log.Logger.Debug("processing item",
    zap.String("id", itemID),
    zap.Any("details", item),
)
```

Never use `fmt.Println` or the standard `log` package in application code.

### HTTP Request Logging

Use middleware for automatic request/response logging:

```go
// internal/http/middlewares/logger.go
func ZapLogger(log *zap.Logger, skipURLs ...string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(ctx echo.Context) error {
            start := time.Now()
            err := next(ctx)

            // Log with structured fields
            log.Info("request",
                zap.String("method", ctx.Request().Method),
                zap.String("uri", ctx.Request().RequestURI),
                zap.Int("status", ctx.Response().Status),
                zap.Duration("latency", time.Since(start)),
            )

            return err
        }
    }
}
```

---

## Database Layer

### Connection Setup

**Global Singleton pattern** — `internal/app/db.go`:

```go
package app

import (
    "time"
    _ "github.com/go-sql-driver/mysql"
    "github.com/jmoiron/sqlx"
    "go.uber.org/zap"

    "yourproject/internal/config"
    "yourproject/internal/log"
)

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
```

**Private Dependencies pattern** — same file returns `*sqlx.DB` instead of
mutating `A`; see `references/wiring-patterns.md`.

### Migrations

Use `golang-migrate` for database migrations:

```bash
# Create migration
migrate create -ext sql -dir migrations -seq init

# Apply migrations
migrate -path migrations -database "mysql://user:pass@tcp(localhost:3306)/db" up

# Rollback
migrate -path migrations -database "mysql://user:pass@tcp(localhost:3306)/db" down 1
```

**Migration files**:

```sql
-- migrations/000001_init.up.sql
CREATE TABLE clients (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- migrations/000001_init.down.sql
DROP TABLE IF EXISTS clients;
```

### Query Patterns

```go
// Named query (INSERT/UPDATE)
query := `INSERT INTO clients (name, email) VALUES (:name, :email)`
_, err := db.NamedExecContext(ctx, query, &client)

// Get single row
query := `SELECT * FROM clients WHERE id = ?`
var client models.Client
err := db.GetContext(ctx, &client, query, id)

// Get multiple rows
query := `SELECT * FROM clients WHERE active = ?`
var clients []models.Client
err := db.SelectContext(ctx, &clients, query, true)

// IN clause with sqlx
query := `DELETE FROM clients WHERE id IN (?)`
query, args, err := sqlx.In(query, ids)
if err != nil {
    return err
}
query = db.Rebind(query)
_, err = db.ExecContext(ctx, query, args...)
```
