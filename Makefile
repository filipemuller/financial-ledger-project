.PHONY: help setup start stop clean migrate test test-unit test-integration test-coverage run build

include .env
export

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Start database and run migrations
	@echo "Starting PostgreSQL..."
	docker-compose up -d
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Running migrations..."
	go run cmd/migrate/main.go
	@echo "Setup complete!"

start: ## Start database
	docker-compose up -d

stop: ## Stop database
	docker-compose down

clean: ## Stop database and remove all data
	docker-compose down -v
	@echo "Database cleaned!"

migrate: ## Run database migrations
	go run cmd/migrate/main.go

run: ## Start the API server
	go run cmd/api/main.go

build: ## Build the API server binary
	@echo "Building API server..."
	go build -o bin/api cmd/api/main.go
	@echo "Building migration tool..."
	go build -o bin/migrate cmd/migrate/main.go
	@echo "Build complete! Binaries in ./bin/"

test: ## Run all tests
	go test -v -race ./...

test-unit: ## Run unit tests only
	go test -v -race ./internal/models/... ./tests/unit/...

test-integration: start ## Run integration tests (requires database)
	@echo "Running integration tests..."
	go test -v -race ./tests/integration/...

test-coverage: ## Run tests with coverage report
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run Go linter
	golangci-lint run ./...

fmt: ## Format Go code
	go fmt ./...
	gofmt -s -w .

tidy: ## Tidy Go modules
	go mod tidy

deps: ## Download dependencies
	go mod download

check: fmt lint test ## Run format, lint, and tests

.DEFAULT_GOAL := help
