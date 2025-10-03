.PHONY: build run test clean docker-build docker-run help

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the service in debug mode
	cargo build

release: ## Build the service in release mode
	cargo build --release

run: ## Run the service locally
	cargo run --release

test: ## Run tests
	cargo test

test-verbose: ## Run tests with output
	cargo test -- --nocapture

check: ## Check code without building
	cargo check

fmt: ## Format code
	cargo fmt

lint: ## Run clippy linter
	cargo clippy -- -D warnings

clean: ## Clean build artifacts
	cargo clean

docker-build: ## Build Docker image
	docker build -t b25/market-data:latest .

docker-run: ## Run Docker container
	docker run -p 9090:9090 -v $(PWD)/config.example.yaml:/app/config.yaml b25/market-data:latest

compose-up: ## Start services with docker-compose
	cp config.example.yaml config.yaml
	docker-compose up -d

compose-down: ## Stop services with docker-compose
	docker-compose down

compose-logs: ## View docker-compose logs
	docker-compose logs -f market-data

compose-monitoring: ## Start with monitoring stack
	cp config.example.yaml config.yaml
	docker-compose --profile monitoring up -d

health: ## Check service health
	curl -s http://localhost:9090/health | jq

metrics: ## View service metrics
	curl -s http://localhost:9090/metrics

watch-metrics: ## Watch metrics continuously
	watch -n 2 'curl -s http://localhost:9090/metrics | grep -E "(websocket_connected|processing_latency|orderbook_updates)"'

subscribe: ## Subscribe to Redis orderbook updates (requires redis-cli)
	redis-cli SUBSCRIBE 'orderbook:BTCUSDT'

install-deps: ## Install system dependencies (Ubuntu/Debian)
	sudo apt-get update && sudo apt-get install -y pkg-config libssl-dev

install-rust: ## Install Rust toolchain
	curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

benchmark: ## Run benchmarks
	cargo bench

profile: ## Profile the application
	cargo build --release
	perf record --call-graph=dwarf ./target/release/market-data-service
	perf report
