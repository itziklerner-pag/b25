.PHONY: help build run test clean docker-build docker-up docker-down lint

# Default target
help:
	@echo "Media Service - Available targets:"
	@echo "  build         - Build the service binary"
	@echo "  run           - Run the service locally"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-up     - Start all services with Docker Compose"
	@echo "  docker-down   - Stop all services"
	@echo "  lint          - Run linters"
	@echo "  deps          - Download dependencies"
	@echo "  migrate       - Run database migrations"

# Build the service
build:
	@echo "Building media service..."
	CGO_ENABLED=1 go build -o bin/media-service ./cmd/server

# Run the service
run:
	@echo "Running media service..."
	go run ./cmd/server

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t b25/media-service:latest .

# Start all services with Docker Compose
docker-up:
	@echo "Starting services..."
	docker-compose up -d

# Stop all services
docker-down:
	@echo "Stopping services..."
	docker-compose down

# Run linters
lint:
	@echo "Running linters..."
	golangci-lint run ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Run database migrations (requires service to be running)
migrate:
	@echo "Running migrations..."
	@echo "Migrations run automatically on service start"

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	cp config.example.yaml config.yaml
	cp .env.example .env
	@echo "Please update config.yaml and .env with your settings"

# Start development environment
dev:
	@echo "Starting development environment..."
	docker-compose up postgres redis minio -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Starting media service..."
	go run ./cmd/server
