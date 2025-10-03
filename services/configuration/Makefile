.PHONY: build run test clean docker-build docker-run migrate-up migrate-down

# Variables
SERVICE_NAME=configuration-service
BINARY_NAME=bin/$(SERVICE_NAME)
DOCKER_IMAGE=$(SERVICE_NAME):latest
POSTGRES_DSN=postgres://postgres:postgres@localhost:5432/configuration_db?sslmode=disable

# Build the application
build:
	@echo "Building $(SERVICE_NAME)..."
	@mkdir -p bin
	@go build -o $(BINARY_NAME) ./cmd/server

# Run the application
run: build
	@echo "Running $(SERVICE_NAME)..."
	@./$(BINARY_NAME)

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	@go tool cover -html=coverage.out

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -p 9096:9096 -p 9097:9097 --name $(SERVICE_NAME) $(DOCKER_IMAGE)

# Stop and remove Docker container
docker-stop:
	@docker stop $(SERVICE_NAME) || true
	@docker rm $(SERVICE_NAME) || true

# Run migrations up
migrate-up:
	@echo "Running migrations up..."
	@migrate -path migrations -database "$(POSTGRES_DSN)" up

# Run migrations down
migrate-down:
	@echo "Running migrations down..."
	@migrate -path migrations -database "$(POSTGRES_DSN)" down

# Create a new migration
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run

# Run the service with hot reload (requires air)
dev:
	@air

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  run           - Build and run the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  docker-stop   - Stop and remove Docker container"
	@echo "  migrate-up    - Run database migrations up"
	@echo "  migrate-down  - Run database migrations down"
	@echo "  migrate-create- Create a new migration"
	@echo "  deps          - Install dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  dev           - Run with hot reload"
	@echo "  help          - Show this help message"
