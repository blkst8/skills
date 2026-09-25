# invoice-service

Go service that manages invoices: a REST API (Echo), MySQL persistence
(sqlx) and a background job that reconciles invoices every 5 minutes.
Dependencies are wired with the **private dependency injection** pattern —
no global state, everything is chained in `cmd/start.go`.

## Features

- Invoice REST API (create / get / list) behind JWT authentication
- MySQL storage via sqlx with golang-migrate migrations
- Ticker worker that reconciles invoices on a fixed interval (default 5m)
- Optional concurrent worker pool (`internal/workers`, disabled by default)
- Prometheus metrics on `/metrics`, liveness on `/healthz`
- zap structured logging, viper configuration with built-in defaults

## Prerequisites

- Go 1.24+
- MySQL 8.0+
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI (for `make migrate-up`)

## Installation

```bash
git clone https://github.com/blkst8/invoice-service
cd invoice-service
go mod download
```

## Configuration

Copy example config and edit:

```bash
cp config.example.yaml config.yaml
```

All keys have built-in defaults (`internal/config/builtin.go`); never commit
the real `config.yaml` with secrets.

## Migrations

```bash
export DATABASE_URL="mysql://invoice:invoice@tcp(localhost:3306)/invoice_service?parseTime=true"
make migrate-up      # apply
make migrate-down    # roll back one step
# or via the built-in command:
./invoice-service migrate up --config config.yaml
```

## Usage

Start the server and background workers:

```bash
./invoice-service start --config config.yaml
```

## API Documentation

| Method | Path                  | Description                       |
|--------|-----------------------|-----------------------------------|
| GET    | `/healthz`            | Liveness probe                    |
| GET    | `/metrics`            | Prometheus metrics                |
| POST   | `/api/v1/invoices`    | Create an invoice (JWT required)  |
| GET    | `/api/v1/invoices`    | List invoices (JWT required)      |
| GET    | `/api/v1/invoices/:id`| Get one invoice (JWT required)    |

Reconciliation job: every `worker.jobs_intervals.reconcile_invoices` (5m)
pending invoices past their due date are marked overdue and ops is notified
via the configured webhook.

## Development

Run tests:

```bash
make test
```

## License

MIT License
