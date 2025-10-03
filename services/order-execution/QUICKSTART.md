# Order Execution Service - Quick Start Guide

## Prerequisites

```bash
# Install Go 1.21+
go version

# Install protoc
# macOS: brew install protobuf
# Ubuntu: apt-get install protobuf-compiler

# Start dependencies (Redis + NATS)
docker-compose up -d redis nats
```

## Setup (One-time)

```bash
# Run setup script
./scripts/setup.sh

# Or manually:
go mod download
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/order.proto
```

## Configuration

```bash
# Copy example config
cp config.example.yaml config.yaml

# Edit with your API keys
nano config.yaml
```

**For testing, use Binance Testnet:**
- Get testnet API keys: https://testnet.binancefuture.com/
- Set `testnet: true` in config.yaml

## Run

### Option 1: Direct Go
```bash
go run ./cmd/server
```

### Option 2: Build and Run
```bash
make build
./bin/order-execution
```

### Option 3: Docker
```bash
docker build -t order-execution .
docker run -p 50051:50051 -p 9091:9091 \
  -e BINANCE_API_KEY=your_key \
  -e BINANCE_SECRET_KEY=your_secret \
  order-execution
```

## Test the Service

### Check Health
```bash
curl http://localhost:9091/health
```

### View Metrics
```bash
curl http://localhost:9091/metrics
```

### Test with grpcurl

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List services
grpcurl -plaintext localhost:50051 list

# Create order
grpcurl -plaintext -d '{
  "symbol": "BTCUSDT",
  "side": "BUY",
  "type": "LIMIT",
  "quantity": 0.001,
  "price": 45000,
  "time_in_force": "GTC"
}' localhost:50051 order.OrderService/CreateOrder
```

## Common Commands

```bash
# Run tests
make test

# Run with coverage
make test-coverage

# Format code
make fmt

# Build binary
make build

# Clean build artifacts
make clean

# View all commands
make help
```

## Troubleshooting

### "Connection refused" on port 6379
**Issue**: Redis not running
**Fix**: `docker-compose up -d redis`

### "Connection refused" on port 4222
**Issue**: NATS not running
**Fix**: `docker-compose up -d nats`

### "Binance API error 401"
**Issue**: Invalid API credentials
**Fix**: Check API key/secret in config.yaml

### "Symbol not registered"
**Issue**: Service couldn't load exchange info
**Fix**: Check internet connection and Binance API access

## Development Workflow

1. **Make changes** to code
2. **Format**: `make fmt`
3. **Test**: `make test`
4. **Build**: `make build`
5. **Run**: `./bin/order-execution`

## Next Steps

- Read full [README.md](README.md) for detailed documentation
- Check [proto/order.proto](proto/order.proto) for API definition
- Explore [internal/](internal/) packages for implementation details
- Review [config.example.yaml](config.example.yaml) for all options

## Quick Reference

| Port  | Service              |
|-------|---------------------|
| 50051 | gRPC API            |
| 9091  | HTTP (health + metrics) |

| Endpoint              | Purpose           |
|----------------------|-------------------|
| `/health`            | Health check (JSON) |
| `/health/ready`      | Readiness probe    |
| `/health/live`       | Liveness probe     |
| `/metrics`           | Prometheus metrics |

## Example API Calls

### Create Limit Order (Maker)
```bash
grpcurl -plaintext -d '{
  "symbol": "BTCUSDT",
  "side": "BUY",
  "type": "POST_ONLY",
  "quantity": 0.001,
  "price": 44000,
  "time_in_force": "GTX",
  "post_only": true
}' localhost:50051 order.OrderService/CreateOrder
```

### Create Market Order
```bash
grpcurl -plaintext -d '{
  "symbol": "BTCUSDT",
  "side": "BUY",
  "type": "MARKET",
  "quantity": 0.001,
  "time_in_force": "IOC"
}' localhost:50051 order.OrderService/CreateOrder
```

### Cancel Order
```bash
grpcurl -plaintext -d '{
  "order_id": "abc-123-def",
  "symbol": "BTCUSDT"
}' localhost:50051 order.OrderService/CancelOrder
```

### Get Order
```bash
grpcurl -plaintext -d '{
  "order_id": "abc-123-def"
}' localhost:50051 order.OrderService/GetOrder
```

## Environment Variables

```bash
# Required
export BINANCE_API_KEY="your_api_key"
export BINANCE_SECRET_KEY="your_secret_key"

# Optional (with defaults)
export REDIS_ADDRESS="localhost:6379"
export NATS_ADDRESS="nats://localhost:4222"
export LOG_LEVEL="info"
export BINANCE_TESTNET="true"
```

## Performance Tips

1. **Use POST_ONLY orders** for maker rebates (negative fees)
2. **Monitor rate limits** in metrics (`rate_limit_hits_total`)
3. **Watch circuit breaker** state for exchange health
4. **Cache hit ratio** should be >80% for good performance

## Support

- Documentation: [README.md](README.md)
- API Spec: [proto/order.proto](proto/order.proto)
- Issues: GitHub Issues
