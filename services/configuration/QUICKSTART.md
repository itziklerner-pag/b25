# Configuration Service - Quick Start Guide

Get the Configuration Service up and running in 5 minutes.

## Prerequisites

- Go 1.21 or later
- Docker and Docker Compose (recommended)
- PostgreSQL 15+ (if running locally)
- NATS server (if running locally)

## Option 1: Docker Compose (Recommended)

The fastest way to get started:

```bash
# 1. Navigate to the service directory
cd services/configuration

# 2. Start all services (PostgreSQL, NATS, Configuration Service)
docker-compose up -d

# 3. Check service health
curl http://localhost:9096/health

# 4. View logs
docker-compose logs -f configuration-service

# 5. Try the API
curl http://localhost:9096/api/v1/configurations
```

That's it! The service is running on:
- HTTP API: http://localhost:9096
- gRPC API: localhost:9097
- Metrics: http://localhost:9096/metrics

## Option 2: Local Development

### Step 1: Install Dependencies

```bash
cd services/configuration
go mod download
```

### Step 2: Start Dependencies

**PostgreSQL:**
```bash
docker run -d --name config-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=configuration_db \
  -p 5432:5432 \
  postgres:15-alpine
```

**NATS:**
```bash
docker run -d --name config-nats \
  -p 4222:4222 \
  -p 8222:8222 \
  nats:latest --http_port 8222
```

### Step 3: Configure the Service

```bash
cp config.example.yaml config.yaml
# Edit config.yaml if needed
```

### Step 4: Run Database Migrations

```bash
# Install golang-migrate if not already installed
# macOS: brew install golang-migrate
# Linux: Download from https://github.com/golang-migrate/migrate

make migrate-up
```

### Step 5: Run the Service

```bash
# Using Make
make run

# Or directly
go run ./cmd/server
```

## Quick API Test

### 1. Create a Configuration

```bash
curl -X POST http://localhost:9096/api/v1/configurations \
  -H "Content-Type: application/json" \
  -d '{
    "key": "my_strategy",
    "type": "strategy",
    "value": {
      "name": "Test Strategy",
      "type": "market_making",
      "enabled": true,
      "parameters": {"spread": 0.002}
    },
    "format": "json",
    "description": "My test strategy",
    "created_by": "admin"
  }'
```

### 2. List Configurations

```bash
curl http://localhost:9096/api/v1/configurations
```

### 3. Get Configuration by Key

```bash
curl http://localhost:9096/api/v1/configurations/key/my_strategy
```

### 4. Update Configuration

```bash
# Get the ID from the previous response
CONFIG_ID="paste-id-here"

curl -X PUT http://localhost:9096/api/v1/configurations/${CONFIG_ID} \
  -H "Content-Type: application/json" \
  -d '{
    "value": {
      "name": "Test Strategy",
      "type": "market_making",
      "enabled": true,
      "parameters": {"spread": 0.003}
    },
    "format": "json",
    "description": "Updated test strategy",
    "updated_by": "admin",
    "change_reason": "Adjusted spread"
  }'
```

### 5. View Version History

```bash
curl http://localhost:9096/api/v1/configurations/${CONFIG_ID}/versions
```

## Test Hot-Reload with NATS

### Terminal 1: Run the NATS Subscriber

```bash
# Build the subscriber
go build -o bin/nats-subscriber ./examples/nats_subscriber.go

# Run it
./bin/nats-subscriber
```

### Terminal 2: Update a Configuration

```bash
# Use the API examples script
./examples/api_examples.sh
```

You should see real-time updates in Terminal 1!

## Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test -v ./internal/validator/...
```

## Useful Commands

```bash
# Format code
make fmt

# Lint code
make lint

# Build binary
make build

# Clean build artifacts
make clean

# View logs (Docker)
docker-compose logs -f configuration-service

# Stop services (Docker)
docker-compose down

# Restart service (Docker)
docker-compose restart configuration-service
```

## Development Workflow

1. **Make changes** to the code
2. **Run tests**: `make test`
3. **Format code**: `make fmt`
4. **Build**: `make build`
5. **Run**: `make run` or `docker-compose up --build`

## Example Use Cases

### Use Case 1: Strategy Configuration Management

```bash
# Create strategy
curl -X POST http://localhost:9096/api/v1/configurations \
  -H "Content-Type: application/json" \
  -d '{
    "key": "arbitrage_btc",
    "type": "strategy",
    "value": {
      "name": "BTC Arbitrage",
      "type": "arbitrage",
      "enabled": true,
      "parameters": {
        "exchanges": ["binance", "coinbase"],
        "min_profit": 0.005
      }
    },
    "format": "json",
    "created_by": "trader1"
  }'

# Services listening to NATS will automatically receive the update
# and reload the strategy configuration
```

### Use Case 2: Risk Limit Updates

```bash
# Update risk limits
curl -X PUT http://localhost:9096/api/v1/configurations/${CONFIG_ID} \
  -H "Content-Type: application/json" \
  -d '{
    "value": {
      "max_position_size": 20000,
      "max_loss_per_trade": 1000,
      "max_daily_loss": 5000
    },
    "format": "json",
    "updated_by": "risk_manager",
    "change_reason": "Market volatility adjustment"
  }'

# Risk manager service receives update via NATS and adjusts limits
```

### Use Case 3: Emergency Configuration Rollback

```bash
# Something went wrong? Rollback!
curl -X POST http://localhost:9096/api/v1/configurations/${CONFIG_ID}/rollback \
  -H "Content-Type: application/json" \
  -d '{
    "version": 1,
    "rolled_back_by": "admin",
    "reason": "Emergency rollback - strategy causing losses"
  }'

# All services receive the rollback event and reload previous config
```

## Monitoring

### Prometheus Metrics

```bash
# View all metrics
curl http://localhost:9096/metrics

# Query specific metrics
curl http://localhost:9096/metrics | grep config_operations

# Set up Prometheus scraping (add to prometheus.yml):
# - job_name: 'configuration-service'
#   static_configs:
#     - targets: ['localhost:9096']
```

### Health Checks

```bash
# Service health
curl http://localhost:9096/health

# Readiness check
curl http://localhost:9096/ready
```

## Troubleshooting

### Problem: Service won't start

**Solution:**
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check if NATS is running
curl http://localhost:8222/varz

# Check configuration
cat config.yaml

# Check logs
docker-compose logs configuration-service
```

### Problem: Database connection error

**Solution:**
```bash
# Test PostgreSQL connection
psql -h localhost -U postgres -d configuration_db

# Run migrations
make migrate-up

# Check database config in config.yaml
```

### Problem: NATS not receiving events

**Solution:**
```bash
# Check NATS connection
curl http://localhost:8222/varz

# Test NATS subscription
nats sub "config.updates.*"

# Check NATS configuration in config.yaml
```

## Next Steps

1. **Explore the API**: Use the `examples/api_examples.sh` script
2. **Implement Hot-Reload**: Use the NATS subscriber pattern in your services
3. **Add Custom Validators**: Extend `internal/validator/validator.go`
4. **Set up Monitoring**: Configure Prometheus and Grafana
5. **Read the full README**: Check `README.md` for detailed documentation

## Support

- Full Documentation: [README.md](./README.md)
- API Examples: [examples/api_examples.sh](./examples/api_examples.sh)
- NATS Subscriber: [examples/nats_subscriber.go](./examples/nats_subscriber.go)
- Tests: [internal/validator/validator_test.go](./internal/validator/validator_test.go)

Happy configuring! 🚀
