# client-service

Go service that stores clients and sends a Telegram welcome message to every
new sign-up. Built with Echo, zap, Viper, Cobra, and sqlx/MySQL, following the
blkst8 Go codebase style guide (private dependency injection pattern).

## Features

- Client CRUD flow over an Echo HTTP API (`POST /api/v1/clients`, `GET /api/v1/clients/:id`)
- Telegram Bot API notifier client (`internal/services/notifier`)
- Welcome message is best effort: if Telegram is down the error is logged and
  client creation still succeeds
- Structured logging with zap, Viper-based configuration, graceful shutdown

## Prerequisites

- Go 1.24+
- MySQL 8.0+
- golang-migrate (for database migrations)

## Installation

```bash
git clone https://github.com/blkst8/client-service
cd client-service
go mod download
```

## Configuration

Copy the example config and edit it:

```bash
cp config.example.yaml config.yaml
```

Set `telegram.bot_token` to the token issued by @BotFather. Never commit the
real value. With `telegram.enabled: false` a no-op notifier is wired and no
messages are sent. `telegram.timeout` bounds every Bot API call so a Telegram
outage can never hang client creation.

Apply migrations:

```bash
make migrate-up DATABASE_URL="mysql://user:pass@tcp(localhost:3306)/clients"
```

## Usage

Start the server:

```bash
./client-service start --config config.yaml
```

Create a client (sends the Telegram welcome message when `telegram_id` is set):

```bash
curl -X POST http://localhost:8080/api/v1/clients \
  -H 'Content-Type: application/json' \
  -d '{"name": "Alice", "email": "alice@example.com", "telegram_id": "42"}'
```

## API Documentation

- `POST /api/v1/clients` — create a client, triggers the Telegram welcome message (best effort)
- `GET /api/v1/clients/:id` — fetch a client by ID (404 when unknown)
- `GET /healthz` — liveness
- `GET /metrics` — Prometheus metrics

## Development

Run tests:

```bash
make test
```
