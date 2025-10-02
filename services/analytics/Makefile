.PHONY: build run test clean docker-build docker-run lint fmt help

# Variables
SERVICE_NAME := analytics
BINARY_NAME := analytics-server
DOCKER_IMAGE := b25/$(SERVICE_NAME)
VERSION := 1.0.0

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the service binary
	@echo "Building $(SERVICE_NAME)..."
	@go build -o bin/$(BINARY_NAME) -ldflags="-X main.version=$(VERSION)" ./cmd/server
	@echo "Built bin/$(BINARY_NAME)"

run: ## Run the service locally
	@echo "Running $(SERVICE_NAME)..."
	@go run ./cmd/server -config config.yaml

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linters
	@echo "Running linters..."
	@go vet ./...
	@gofmt -l -s .
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -s -w .

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.txt coverage.html
	@go clean

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .
	@echo "Built $(DOCKER_IMAGE):$(VERSION)"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 9097:9097 -p 9098:9098 -p 9099:9099 \
		-v $(PWD)/config.yaml:/app/config.yaml \
		$(DOCKER_IMAGE):latest

docker-push: ## Push Docker image
	@echo "Pushing Docker image..."
	@docker push $(DOCKER_IMAGE):$(VERSION)
	@docker push $(DOCKER_IMAGE):latest

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

migrate-up: ## Run database migrations
	@echo "Running database migrations..."
	@psql "$(DB_URL)" -f migrations/001_initial_schema.sql

migrate-down: ## Rollback database migrations
	@echo "Rolling back database migrations..."
	@echo "Manual rollback required - review migrations/001_initial_schema.sql"

dev: ## Run in development mode with hot reload
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		make run; \
	fi

.DEFAULT_GOAL := help
