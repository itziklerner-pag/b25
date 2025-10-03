.PHONY: help build run test clean docker-build docker-run fmt lint deps install

# Variables
APP_NAME := dashboard-server
BINARY_PATH := bin/$(APP_NAME)
DOCKER_IMAGE := $(APP_NAME):latest

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

deps: ## Download dependencies
	go mod download
	go mod verify

build: ## Build the binary
	@echo "Building $(APP_NAME)..."
	@mkdir -p bin
	go build -o $(BINARY_PATH) ./cmd/server

run: ## Run the server
	go run cmd/server/main.go

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

fmt: ## Format code
	go fmt ./...
	gofmt -s -w .

lint: ## Run linter
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1; }
	golangci-lint run ./...

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

install: build ## Install binary to GOPATH
	@echo "Installing $(APP_NAME)..."
	cp $(BINARY_PATH) $(GOPATH)/bin/

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container
	docker run -d --name $(APP_NAME) -p 8080:8080 $(DOCKER_IMAGE)

docker-stop: ## Stop Docker container
	docker stop $(APP_NAME) || true
	docker rm $(APP_NAME) || true

dev: ## Run in development mode with hot reload
	@command -v air >/dev/null 2>&1 || { echo "air not installed. Run: go install github.com/cosmtrek/air@latest"; exit 1; }
	air

.DEFAULT_GOAL := help
