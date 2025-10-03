.PHONY: help build run test clean proto docker migrate

help: ## Display this help message
	@echo "Risk Manager Service - Available Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

proto: ## Generate protobuf code
	@echo "Generating protobuf code..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/risk_manager.proto

build: ## Build the service
	@echo "Building Risk Manager Service..."
	go build -o bin/risk-manager cmd/server/main.go

run: ## Run the service locally
	@echo "Running Risk Manager Service..."
	go run cmd/server/main.go

test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/ coverage.out coverage.html

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t risk-manager:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 50051:50051 -p 9095:9095 --env-file .env risk-manager:latest

migrate-up: ## Run database migrations up
	@echo "Running migrations..."
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/risk_manager?sslmode=disable" up

migrate-down: ## Run database migrations down
	@echo "Reverting migrations..."
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/risk_manager?sslmode=disable" down

migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@echo "Creating migration: $(NAME)"
	migrate create -ext sql -dir migrations -seq $(NAME)

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

dev: ## Run in development mode with hot reload (requires air)
	@echo "Running in development mode..."
	air

.DEFAULT_GOAL := help
