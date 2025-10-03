.PHONY: build run test clean docker proto help

# Variables
SERVICE_NAME = account-monitor
BINARY_NAME = $(SERVICE_NAME)
DOCKER_IMAGE = $(SERVICE_NAME):latest
GO_FILES = $(shell find . -name '*.go' -type f)
PROTO_FILES = $(shell find pkg/proto -name '*.proto' -type f)

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the service binary
	@echo "Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) ./cmd/server
	@echo "Build complete: bin/$(BINARY_NAME)"

run: ## Run the service locally
	@echo "Running $(SERVICE_NAME)..."
	@go run ./cmd/server/main.go

test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	@go test -race ./...

test-cover: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -cover ./...
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

docker-build: ## Build Docker image
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run --rm -p 50051:50051 -p 8080:8080 -p 9093:9093 \
		--env-file .env \
		$(DOCKER_IMAGE)

proto: ## Generate protobuf code
	@echo "Generating protobuf code..."
	@protoc --go_out=. --go-grpc_out=. $(PROTO_FILES)
	@echo "Protobuf generation complete"

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -s -w .

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

dev: ## Run in development mode with hot reload
	@echo "Starting development mode..."
	@air -c .air.toml || go run ./cmd/server/main.go

migrate: ## Run database migrations (manual)
	@echo "Running migrations..."
	@psql -h localhost -U trading -d trading -f internal/storage/migrations/001_initial.sql

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

.DEFAULT_GOAL := help
