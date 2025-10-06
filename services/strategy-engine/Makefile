.PHONY: build clean test run docker-build docker-run plugins help

# Variables
BINARY_NAME=strategy-engine
BINARY_PATH=bin/$(BINARY_NAME)
DOCKER_IMAGE=strategy-engine
DOCKER_TAG=latest
GO_FILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Default target
all: build

## help: Show this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@go build -o $(BINARY_PATH) ./cmd/server
	@echo "Build complete: $(BINARY_PATH)"

## plugins: Build Go plugin strategies
plugins:
	@echo "Building Go plugins..."
	@cd plugins/go && \
	for plugin in *.go; do \
		if [ -f "$$plugin" ] && [ "$$plugin" != "*.go" ]; then \
			name=$${plugin%.go}; \
			echo "Building plugin: $$name"; \
			go build -buildmode=plugin -o "$${name}.so" "$$plugin"; \
		fi; \
	done
	@echo "Plugins built"

## run: Build and run the service
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BINARY_PATH)

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

## test-race: Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	@go test -race -v ./...

## test-cover: Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	@go test -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

## lint: Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

## mod: Download and tidy dependencies
mod:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@find plugins/go -name "*.so" -delete
	@echo "Clean complete"

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker build complete: $(DOCKER_IMAGE):$(DOCKER_TAG)"

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -d \
		--name $(BINARY_NAME) \
		-p 9092:9092 \
		-v $(PWD)/config.yaml:/app/config.yaml \
		-v $(PWD)/plugins:/app/plugins \
		$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "Container started: $(BINARY_NAME)"

## docker-stop: Stop Docker container
docker-stop:
	@echo "Stopping Docker container..."
	@docker stop $(BINARY_NAME) || true
	@docker rm $(BINARY_NAME) || true
	@echo "Container stopped"

## docker-logs: Show Docker container logs
docker-logs:
	@docker logs -f $(BINARY_NAME)

## dev: Run in development mode with hot reload
dev:
	@echo "Running in development mode..."
	@air -c .air.toml || go run ./cmd/server

## proto: Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/order_execution.proto
	@echo "Protobuf code generated"

## install-tools: Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Tools installed"
