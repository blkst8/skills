# Build, Deployment, Dependencies & Documentation

Covers: go.mod workflow, essential dependencies, Makefile, build-info
injection, systemd service, godoc conventions, README template, and the
quality checklist.

---

## Dependency Management (Go Modules)

```bash
# Initialize module
go mod init github.com/yourusername/projectname

# Add dependency
go get github.com/labstack/echo/v4

# Update dependencies
go get -u ./...

# Tidy dependencies
go mod tidy

# Vendor dependencies (optional)
go mod vendor
```

### Essential Dependencies

```
github.com/labstack/echo/v4              # Web framework
go.uber.org/zap                          # Logging
github.com/spf13/viper                   # Configuration
github.com/spf13/cobra                   # CLI commands
github.com/jmoiron/sqlx                  # Database
github.com/go-sql-driver/mysql           # MySQL driver
github.com/golang-migrate/migrate        # Migrations
github.com/prometheus/client_golang      # Metrics
github.com/golang-jwt/jwt/v5             # JWT
go.uber.org/automaxprocs                 # Auto maxprocs (containers)
```

---

## Makefile

```makefile
.PHONY: build run test clean migrate-up migrate-down

APP_NAME=yourapp
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

LDFLAGS=-ldflags "\
    -X 'github.com/yourusername/$(APP_NAME)/internal/app.GitCommit=$(COMMIT)' \
    -X 'github.com/yourusername/$(APP_NAME)/internal/app.GitTag=$(VERSION)' \
    -X 'github.com/yourusername/$(APP_NAME)/internal/app.BuildDate=$(BUILD_TIME)'"

build:
	go build $(LDFLAGS) -o $(APP_NAME) .

run:
	go run . start --config config.yaml

test:
	go test -v -race -coverprofile=coverage.out ./...

test-coverage:
	go tool cover -html=coverage.out

clean:
	rm -f $(APP_NAME)
	rm -f coverage.out

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

lint:
	golangci-lint run

fmt:
	go fmt ./...
	goimports -w .

docker-build:
	docker build -t $(APP_NAME):$(VERSION) .
```

---

## Build Information

Inject build info into the application:

```go
// internal/app/app.go
var (
    GitCommit       string
    GitRef          string
    GitTag          string
    BuildDate       string
    CompilerVersion string
)

const (
    Name    = "yourapp"
    Version = 1
)

func Banner() string {
    return fmt.Sprintf(
        "App: %s v%d\nTag: %s\nCommit: %s\nBuild: %s\n",
        Name, Version, GitTag, GitCommit, BuildDate,
    )
}
```

---

## Systemd Service

```ini
# deployments/systemd/yourapp.service
[Unit]
Description=Your Application Service
After=network.target mysql.service

[Service]
Type=simple
User=yourapp
Group=yourapp
WorkingDirectory=/opt/yourapp
ExecStart=/opt/yourapp/yourapp start --config /etc/yourapp/config.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

---

## Documentation

### Package Documentation

```go
// Package handlers provides HTTP request handlers for the application.
//
// Each handler is responsible for:
//   - Validating request input
//   - Calling appropriate service methods
//   - Formatting response output
//   - Logging errors
package handlers
```

### Function Documentation

```go
// CreateClient handles the creation of a new client.
//
// It validates the request, generates a password if not provided,
// stores the client in the database, and returns the generated links.
//
// Request body:
//   - telegram_id (string, required): Telegram user ID
//   - quota (int, required): Quota in bytes
//   - expire_days (int, required): Number of days until expiration
//   - raw_password (string, optional): Custom password
//
// Response:
//   - links ([]string): Generated connection links
//
// Returns HTTP 200 on success, 400 on validation error, 500 on server error.
func CreateClient(ctx echo.Context) error {
    // ...
}
```

### README Structure

```markdown
# Project Name

Brief description of what the project does.

## Features

- Feature 1
- Feature 2
- Feature 3

## Prerequisites

- Go 1.24+
- MySQL 8.0+
- Optional dependencies

## Installation

git clone https://github.com/yourusername/project
cd project
go mod download

## Configuration

Copy example config and edit:
cp config.example.yaml config.yaml

## Usage

Start the server:
./yourapp start --config config.yaml

## API Documentation

Available endpoints...

## Development

Run tests:
make test

## License

MIT License
```

---

## Best Practices Summary

### Code Quality

1. **Always use context.Context** as the first parameter
2. **Check errors immediately** after function calls
3. **Use interfaces** for dependencies (testability)
4. **Structured logging** with zap
5. **Named returns** only when they improve readability
6. **Table-driven tests** for comprehensive coverage

### Performance

1. **Connection pooling** for databases
2. **Graceful shutdown** for all resources
3. **Context timeouts** for operations
4. **Proper HTTP timeout** configuration
5. **Use automaxprocs** for container environments

### Security

1. **Never commit secrets** to version control
2. **Use environment variables** or secure config management
3. **Validate all inputs** at API boundaries
4. **Use prepared statements** for SQL queries
5. **Implement rate limiting** and authentication

### Maintainability

1. **Small, focused functions** (< 50 lines ideal)
2. **Clear separation of concerns** between layers
3. **Consistent naming conventions** throughout
4. **Comprehensive error messages** for debugging
5. **Keep dependencies minimal** and up-to-date

---

## Checklist for New Projects

- [ ] Initialize Go module
- [ ] Set up project structure (cmd/, internal/, pkg/)
- [ ] Configure logging with zap
- [ ] Set up configuration with viper
- [ ] Wire application dependencies (singleton or DI — pick one, stay consistent)
- [ ] Set up database connection and migrations
- [ ] Implement repository layer with interfaces
- [ ] Create service packages for external dependency clients
- [ ] Create usecase package for big business logic (handlers → usecase → repository/service)
- [ ] Create HTTP server with middleware
- [ ] Add graceful shutdown handling
- [ ] Implement worker pool (`internal/workers`) with config-driven settings
- [ ] Set up metrics and monitoring
- [ ] Create Makefile for common tasks
- [ ] Write README with setup instructions
- [ ] Add example configuration file
- [ ] Set up CI/CD pipeline
- [ ] Write unit tests for critical paths
- [ ] Document API endpoints
