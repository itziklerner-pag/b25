.PHONY: build run test clean docker-build docker-run migrate-up migrate-down help

# Variables
BINARY_NAME=content-service
DOCKER_IMAGE=content-service:latest
DB_DSN=postgres://postgres:postgres@localhost:5432/content_db?sslmode=disable

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) cmd/server/main.go

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	@go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@go clean

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -d --name content-service \
		-p 8080:8080 \
		-e DATABASE_HOST=host.docker.internal \
		$(DOCKER_IMAGE)

# Stop and remove Docker container
docker-stop:
	@echo "Stopping Docker container..."
	@docker stop content-service || true
	@docker rm content-service || true

# Run migrations up
migrate-up:
	@echo "Running database migrations..."
	@migrate -path migrations -database "$(DB_DSN)" up

# Run migrations down
migrate-down:
	@echo "Rolling back database migrations..."
	@migrate -path migrations -database "$(DB_DSN)" down

# Create new migration
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Start development environment
dev:
	@echo "Starting development environment..."
	@docker-compose up -d postgres
	@sleep 2
	@$(MAKE) migrate-up
	@$(MAKE) run

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  docker-stop   - Stop Docker container"
	@echo "  migrate-up    - Run database migrations"
	@echo "  migrate-down  - Rollback database migrations"
	@echo "  migrate-create- Create new migration"
	@echo "  deps          - Install dependencies"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  dev           - Start development environment"
	@echo "  help          - Show this help message"
