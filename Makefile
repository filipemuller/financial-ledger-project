.PHONY: help setup start stop clean migrate test test-unit test-integration test-coverage run build

include .env
export

help:
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup:
	@echo "Starting PostgreSQL..."
	docker-compose up -d
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Running migrations..."
	go run cmd/migrate/main.go
	@echo "Setup complete!"

start:
	docker-compose up -d

stop:
	docker-compose down

clean:
	docker-compose down -v
	@echo "Database cleaned!"

migrate:
	go run cmd/migrate/main.go

run:
	go run cmd/api/main.go

build:
	@echo "Building API server..."
	go build -o bin/api cmd/api/main.go
	@echo "Building migration tool..."
	go build -o bin/migrate cmd/migrate/main.go
	@echo "Build complete! Binaries in ./bin/"

test:
	go test -v -race ./...

test-unit:
	go test -v -race ./internal/models/... ./tests/unit/...

test-integration: start
	@echo "Running integration tests..."
	go test -v -race ./tests/integration/...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...
	gofmt -s -w .

tidy:
	go mod tidy

deps:
	go mod download

check: fmt lint test

.DEFAULT_GOAL := help
