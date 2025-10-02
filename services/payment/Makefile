.PHONY: help build run test clean docker-build docker-up docker-down migrate-up migrate-down

help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application locally"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-up     - Start Docker containers"
	@echo "  docker-down   - Stop Docker containers"
	@echo "  migrate-up    - Run database migrations"
	@echo "  migrate-down  - Rollback database migrations"

build:
	@echo "Building payment service..."
	go build -o bin/payment-service ./cmd/server

run:
	@echo "Running payment service..."
	go run ./cmd/server/main.go

test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

docker-build:
	@echo "Building Docker image..."
	docker build -t payment-service:latest .

docker-up:
	@echo "Starting Docker containers..."
	docker-compose up -d

docker-down:
	@echo "Stopping Docker containers..."
	docker-compose down

migrate-up:
	@echo "Running migrations..."
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/payment_db?sslmode=disable" up

migrate-down:
	@echo "Rolling back migrations..."
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/payment_db?sslmode=disable" down

deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

lint:
	@echo "Running linter..."
	golangci-lint run

dev:
	@echo "Starting development environment..."
	docker-compose up postgres redis -d
	sleep 3
	make migrate-up
	make run
