.PHONY: all build test clean run docker-build docker-run lint help

# Variables
BINARY_NAME=search-service
DOCKER_IMAGE=b25/search-service
VERSION?=1.0.0
GO=go
GOTEST=$(GO) test
GOVET=$(GO) vet
GOLINT=golangci-lint

all: clean lint test build

## build: Build the binary
build:
	@echo "Building..."
	@$(GO) build -o $(BINARY_NAME) -v ./cmd/server

## test: Run tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v -race -coverprofile=coverage.out ./...

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@$(GO) clean

## run: Run the application
run: build
	@echo "Running..."
	@./$(BINARY_NAME) -config config.yaml

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run --rm -p 9097:9097 -p 9098:9098 \
		-e SEARCH_ELASTICSEARCH_ADDRESSES=http://host.docker.internal:9200 \
		-e SEARCH_REDIS_ADDRESS=host.docker.internal:6379 \
		-e SEARCH_NATS_URL=nats://host.docker.internal:4222 \
		$(DOCKER_IMAGE):latest

## lint: Run linter
lint:
	@echo "Running linter..."
	@$(GOVET) ./...
	@if command -v $(GOLINT) > /dev/null; then \
		$(GOLINT) run ./...; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...

## mod: Tidy go modules
mod:
	@echo "Tidying go modules..."
	@$(GO) mod tidy
	@$(GO) mod verify

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GO) mod download

## integration-test: Run integration tests
integration-test:
	@echo "Running integration tests..."
	@$(GOTEST) -v -tags=integration ./tests/integration/...

## benchmark: Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	@$(GOTEST) -bench=. -benchmem ./...

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
