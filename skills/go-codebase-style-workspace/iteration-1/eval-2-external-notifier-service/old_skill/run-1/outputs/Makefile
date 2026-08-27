.PHONY: build run test clean migrate-up migrate-down lint fmt

APP_NAME=client-service
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

LDFLAGS=-ldflags "\
	-X 'github.com/blkst8/client-service/internal/app.GitCommit=$(COMMIT)' \
	-X 'github.com/blkst8/client-service/internal/app.GitTag=$(VERSION)' \
	-X 'github.com/blkst8/client-service/internal/app.BuildDate=$(BUILD_TIME)'"

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
