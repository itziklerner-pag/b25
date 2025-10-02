.PHONY: help build run test clean docker-build docker-run migrate-up migrate-down

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the server binary
	go build -o messaging-server ./cmd/server

build-migrate: ## Build the migrate binary
	go build -o migrate ./cmd/migrate

run: ## Run the server
	go run ./cmd/server/main.go

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -f messaging-server migrate
	rm -f coverage.out coverage.html

docker-build: ## Build Docker image
	docker build -t messaging-service:latest .

docker-run: ## Run with Docker Compose
	docker-compose up -d

docker-down: ## Stop Docker Compose
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f messaging

migrate-up: build-migrate ## Run database migrations
	./migrate up

migrate-down: build-migrate ## Rollback last migration
	./migrate down

lint: ## Run linter
	golangci-lint run

fmt: ## Format code
	go fmt ./...

mod-tidy: ## Tidy go modules
	go mod tidy

mod-download: ## Download dependencies
	go mod download

dev: ## Run in development mode
	air

install-tools: ## Install development tools
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
