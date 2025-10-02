.PHONY: build run test clean docker-build docker-run lint fmt help

# Variables
BINARY_NAME=api-gateway
DOCKER_IMAGE=b25/api-gateway
VERSION?=1.0.0

# Default target
.DEFAULT_GOAL := help

## build: Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) -ldflags="-X main.Version=$(VERSION)" ./cmd/server

## run: Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	@go run ./cmd/server

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

## test-coverage: Run tests with coverage report
test-coverage: test
	@go tool cover -html=coverage.out

## lint: Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out
	@go clean

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):latest

## docker-push: Push Docker image
docker-push:
	@echo "Pushing Docker image..."
	@docker push $(DOCKER_IMAGE):$(VERSION)
	@docker push $(DOCKER_IMAGE):latest

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

## help: Show this help
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
