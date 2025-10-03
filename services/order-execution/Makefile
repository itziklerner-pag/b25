.PHONY: build test clean run docker proto help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=order-execution
BINARY_PATH=bin/$(BINARY_NAME)

# Docker parameters
DOCKER_IMAGE=order-execution
DOCKER_TAG=latest

# Protobuf parameters
PROTO_DIR=proto
PROTO_FILES=$(wildcard $(PROTO_DIR)/*.proto)

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	$(GOBUILD) -o $(BINARY_PATH) -v ./cmd/server
	@echo "Build complete: $(BINARY_PATH)"

run: ## Run the service locally
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run ./cmd/server

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -race -cover ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./...

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

proto: ## Generate protobuf code
	@echo "Generating protobuf code..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)
	@echo "Protobuf generation complete"

proto-install: ## Install protobuf tools
	@echo "Installing protobuf tools..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Protobuf tools installed"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -d \
		--name $(BINARY_NAME) \
		-p 50051:50051 \
		-p 9091:9091 \
		-e BINANCE_API_KEY=$(BINANCE_API_KEY) \
		-e BINANCE_SECRET_KEY=$(BINANCE_SECRET_KEY) \
		-e REDIS_ADDRESS=redis:6379 \
		-e NATS_ADDRESS=nats://nats:4222 \
		$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "Container started: $(BINARY_NAME)"

docker-stop: ## Stop Docker container
	@echo "Stopping Docker container..."
	docker stop $(BINARY_NAME)
	docker rm $(BINARY_NAME)
	@echo "Container stopped"

docker-logs: ## View Docker container logs
	docker logs -f $(BINARY_NAME)

lint: ## Run linters
	@echo "Running linters..."
	@golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	@gofmt -s -w .
	@go mod tidy

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

# Development commands
dev: ## Run in development mode with hot reload (requires air)
	@echo "Starting development server..."
	@air

install-tools: proto-install ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/cosmtrek/air@latest
	@echo "Development tools installed"

# Load testing
load-test: ## Run load test (requires ghz)
	@echo "Running load test..."
	@ghz --insecure \
		--proto proto/order.proto \
		--call order.OrderService/CreateOrder \
		-d '{"symbol":"BTCUSDT","side":"BUY","type":"LIMIT","quantity":0.001,"price":45000}' \
		-c 10 -n 1000 \
		localhost:50051

# Deployment
deploy-k8s: ## Deploy to Kubernetes
	@echo "Deploying to Kubernetes..."
	@kubectl apply -f k8s/

# Database migrations (if needed)
migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	# Add migration commands here

migrate-down: ## Run database migrations down
	@echo "Running migrations down..."
	# Add migration commands here

.DEFAULT_GOAL := help
