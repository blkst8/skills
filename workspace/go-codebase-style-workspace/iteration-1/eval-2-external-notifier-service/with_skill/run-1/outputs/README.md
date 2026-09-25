# client-service

Go HTTP service that manages client sign-ups and sends a Telegram welcome
message to every new client with a Telegram ID.

## Features

- Create and fetch clients (`POST /api/v1/clients`, `GET /api/v1/clients/:id`)
- Telegram Bot API notifier client (`internal/service/notifier`)
- Welcome notification is best-effort: if Telegram is down, client creation
  still succeeds and the failure is logged
- Layered architecture (handlers → usecases → repositories/services),
  interface-driven with private dependency injection
- zap structured logging, Viper YAML config, sqlx/MySQL, Echo HTTP server
  with graceful shutdown

## Prerequisites

- Go 1.24+
- MySQL 8.0+
- A Telegram bot token (create a bot via @BotFather)
- golang-migrate CLI for migrations

## Installation

```
git clone https://github.com/blkst8/client-service
cd client-service
go mod download
```

## Configuration

Copy the example config and edit:

```
cp config.example.yaml config.yaml
```

Set `database.dsn` and `telegram.token` for your environment. Never commit
the real `config.yaml` with secrets.

## Database

Apply migrations:

```
make migrate-up DATABASE_URL="mysql://user:pass@tcp(localhost:3306)/clients"
```

## Usage

Start the server:

```
./client-service start --config config.yaml
```

## API Documentation

- `POST /api/v1/clients` — create a client. Body:
  `{"name": "Alice", "email": "alice@example.com", "telegram_id": "12345"}`.
  On success the client is persisted and a Telegram welcome message is sent
  when `telegram_id` is present (failures are logged, not fatal).
- `GET /api/v1/clients/:id` — fetch a client (404 when unknown).
- `GET /healthz` — liveness probe.

## Development

Run tests:

```
make test
```

## License

MIT License
