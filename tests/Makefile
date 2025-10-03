# Makefile for B25 Trading System Tests

.PHONY: help test integration e2e benchmark setup teardown clean

# Default target
help:
	@echo "B25 Trading System - Test Targets"
	@echo "=================================="
	@echo ""
	@echo "Setup:"
	@echo "  make setup         - Start test infrastructure"
	@echo "  make teardown      - Stop test infrastructure"
	@echo ""
	@echo "Tests:"
	@echo "  make test          - Run all tests"
	@echo "  make integration   - Run integration tests only"
	@echo "  make e2e          - Run E2E tests only"
	@echo "  make benchmark     - Run performance benchmarks"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean         - Clean test data and containers"
	@echo "  make deps          - Download Go dependencies"
	@echo "  make lint          - Run linter on tests"
	@echo ""

# Start test infrastructure
setup:
	@echo "Starting test infrastructure..."
	@docker-compose -f testutil/docker/docker-compose.test.yml up -d
	@echo "Waiting for services to be healthy..."
	@sleep 10
	@echo "Test infrastructure ready!"

# Stop test infrastructure
teardown:
	@echo "Stopping test infrastructure..."
	@docker-compose -f testutil/docker/docker-compose.test.yml down -v
	@echo "Test infrastructure stopped"

# Run all tests
test: setup
	@./run_all_tests.sh

# Run integration tests
integration: setup
	@KEEP_RUNNING=1 ./run_integration_tests.sh

# Run E2E tests
e2e: setup
	@KEEP_RUNNING=1 ./run_e2e_tests.sh

# Run benchmarks
benchmark: setup
	@KEEP_RUNNING=1 ./run_e2e_tests.sh --benchmark

# Download dependencies
deps:
	@echo "Downloading test dependencies..."
	@cd testutil && go mod download
	@cd integration && go mod download
	@cd e2e && go mod download
	@echo "Dependencies downloaded"

# Clean everything
clean:
	@echo "Cleaning test environment..."
	@docker-compose -f testutil/docker/docker-compose.test.yml down -v
	@docker system prune -f
	@rm -f integration/coverage.out
	@rm -f e2e/coverage.out
	@echo "Clean complete"

# Run linter
lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./integration/...
	@golangci-lint run ./e2e/...
	@golangci-lint run ./testutil/...

# Quick test (no cleanup)
quick-test:
	@cd integration && go test -v -short ./...
	@cd e2e && go test -v -short -run="^Test[^B]" ./...

# Generate test coverage report
coverage: setup
	@echo "Generating coverage report..."
	@cd integration && go test -v -coverprofile=coverage.out ./...
	@cd e2e && go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=integration/coverage.out -o integration/coverage.html
	@go tool cover -html=e2e/coverage.out -o e2e/coverage.html
	@echo "Coverage reports generated:"
	@echo "  - integration/coverage.html"
	@echo "  - e2e/coverage.html"

# Run tests in parallel
parallel: setup
	@echo "Running tests in parallel..."
	@cd integration && go test -v -parallel 4 ./...
	@cd e2e && go test -v -parallel 4 ./...

# Performance profiling
profile: setup
	@echo "Running performance profiling..."
	@cd e2e && go test -v -run=TestLatencyBenchmark -cpuprofile=cpu.prof -memprofile=mem.prof
	@echo "Profiles generated:"
	@echo "  - e2e/cpu.prof"
	@echo "  - e2e/mem.prof"
	@echo "Analyze with: go tool pprof e2e/cpu.prof"

# Continuous testing (watch mode - requires entr)
watch:
	@echo "Starting watch mode (requires 'entr')..."
	@find . -name "*.go" | entr -c make quick-test

# Docker logs
logs:
	@docker-compose -f testutil/docker/docker-compose.test.yml logs -f

# Show running services
ps:
	@docker-compose -f testutil/docker/docker-compose.test.yml ps
