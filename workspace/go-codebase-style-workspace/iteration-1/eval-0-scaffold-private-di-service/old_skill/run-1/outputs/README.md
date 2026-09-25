# invoice-service

REST API for managing invoices, plus a background job that reconciles
overdue invoices every 5 minutes.

## Features

- Invoice REST API (create, get, list) built with Echo
- MySQL persistence via sqlx with interface-driven repositories
- Background invoice reconciliation worker (5-minute interval)
- Zap structured logging, Viper YAML configuration, graceful shutdown
- Private dependency injection wiring in `cmd/start.go`

## Prerequisites

- Go 1.24+
- MySQL 8.0+

## Installation

```bash
go mod download
```

## Configuration

Copy the example config and edit it:

```bash
cp config.example.yaml config.yaml
```

Never commit the real `config.yaml`; it contains secrets.

## Usage

Apply the database migrations:

```bash
go run . migrate --config config.yaml
```

Start the server and workers:

```bash
go run . start --config config.yaml
```

## API

| Method | Path                  | Description                |
|--------|-----------------------|----------------------------|
| GET    | `/healthz`            | Liveness probe             |
| GET    | `/metrics`            | Prometheus metrics         |
| POST   | `/api/v1/invoices`    | Create an invoice          |
| GET    | `/api/v1/invoices`    | List invoices (limit/offset) |
| GET    | `/api/v1/invoices/:id`| Get a single invoice       |

## Development

Run tests:

```bash
make test
```

Build the binary (build info injected via ldflags):

```bash
make build
```
