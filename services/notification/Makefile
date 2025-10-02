.PHONY: help build run test clean docker-build docker-run migrate-up migrate-down lint fmt

# Variables
SERVICE_NAME := notification-service
DOCKER_IMAGE := b25/$(SERVICE_NAME)
GO_FILES := $(shell find . -name '*.go' -not -path './vendor/*')

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the service binary
	@echo "Building $(SERVICE_NAME)..."
	@go build -o bin/$(SERVICE_NAME) ./cmd/server

run: ## Run the service locally
	@echo "Running $(SERVICE_NAME)..."
	@go run ./cmd/server/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -cover ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -tags=integration ./tests/integration/...

lint: ## Run linters
	@echo "Running linters..."
	@golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	@gofmt -s -w $(GO_FILES)
	@go mod tidy

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod verify

migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	@psql -U notification_user -d notification_db -f migrations/001_init_schema.up.sql

migrate-down: ## Run database migrations down
	@echo "Running migrations down..."
	@psql -U notification_user -d notification_db -f migrations/001_init_schema.down.sql

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 9097:9097 -p 9098:9098 --env-file .env $(DOCKER_IMAGE):latest

docker-compose-up: ## Start services with docker-compose
	@echo "Starting services..."
	@docker-compose -f ../../docker/docker-compose.dev.yml up -d notification-service

docker-compose-down: ## Stop services with docker-compose
	@echo "Stopping services..."
	@docker-compose -f ../../docker/docker-compose.dev.yml down

dev: ## Run in development mode with hot reload
	@echo "Starting development mode..."
	@air

proto-gen: ## Generate protobuf code
	@echo "Generating protobuf code..."
	@protoc --go_out=. --go-grpc_out=. ../../shared/proto/*.proto

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

seed-db: ## Seed database with sample data
	@echo "Seeding database..."
	@psql -U notification_user -d notification_db -f tests/seed.sql

.DEFAULT_GOAL := help
