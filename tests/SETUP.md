# Test Environment Setup Guide

Complete setup guide for the B25 Trading System test suite.

## Prerequisites

### Required Software

1. **Go 1.21 or higher**
   ```bash
   # Check Go version
   go version

   # Install if needed (Ubuntu/Debian)
   wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
   sudo rm -rf /usr/local/go
   sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
   export PATH=$PATH:/usr/local/go/bin
   ```

2. **Docker & Docker Compose**
   ```bash
   # Check Docker
   docker --version
   docker-compose --version

   # Install if needed (Ubuntu/Debian)
   curl -fsSL https://get.docker.com -o get-docker.sh
   sudo sh get-docker.sh
   sudo usermod -aG docker $USER
   ```

3. **Git** (for cloning the repository)
   ```bash
   git --version
   ```

## Quick Setup

### 1. Clone Repository
```bash
git clone <repository-url>
cd b25/tests
```

### 2. Download Dependencies
```bash
# Option A: Using Makefile
make deps

# Option B: Manual
cd testutil && go mod download
cd ../integration && go mod download
cd ../e2e && go mod download
```

### 3. Start Test Infrastructure
```bash
# Using Makefile
make setup

# Or using Docker Compose directly
docker-compose -f testutil/docker/docker-compose.test.yml up -d
```

### 4. Verify Setup
```bash
# Check all services are healthy
docker-compose -f testutil/docker/docker-compose.test.yml ps

# Expected output:
# - redis-test      (healthy)
# - nats-test       (healthy)
# - postgres-test   (healthy)
# - timescale-test  (healthy)
# - mock-exchange   (healthy)
```

### 5. Run Tests
```bash
# All tests
make test

# Or specific test suites
make integration
make e2e
make benchmark
```

## Detailed Setup

### Infrastructure Components

#### 1. Redis (Cache & Pub/Sub)
- **Port:** 6380
- **Purpose:** Order caching, market data cache
- **Test Connection:**
  ```bash
  redis-cli -p 6380 ping
  # Should return: PONG
  ```

#### 2. NATS (Message Queue)
- **Port:** 4223 (client), 8223 (monitoring)
- **Purpose:** Event streaming, order updates
- **Test Connection:**
  ```bash
  curl http://localhost:8223/varz
  # Should return: JSON with server info
  ```

#### 3. PostgreSQL (Database)
- **Port:** 5433
- **Database:** b25_test
- **User/Pass:** testuser/testpass
- **Purpose:** Order history, account data
- **Test Connection:**
  ```bash
  psql -h localhost -p 5433 -U testuser -d b25_test -c "SELECT 1;"
  # Should return: 1
  ```

#### 4. TimescaleDB (Time Series)
- **Port:** 5434
- **Database:** timescale_test
- **User/Pass:** testuser/testpass
- **Purpose:** Market data history, metrics
- **Test Connection:**
  ```bash
  psql -h localhost -p 5434 -U testuser -d timescale_test -c "SELECT 1;"
  ```

#### 5. Mock Exchange Server
- **HTTP Port:** 8545
- **WebSocket Port:** 8546
- **Purpose:** Simulated cryptocurrency exchange
- **Test Connection:**
  ```bash
  curl http://localhost:8545/api/v3/exchangeInfo
  # Should return: JSON with exchange info
  ```

### Environment Variables

Create a `.env` file in the `tests` directory:

```bash
# Infrastructure
REDIS_ADDR=localhost:6380
NATS_ADDR=nats://localhost:4223
POSTGRES_ADDR=localhost:5433
POSTGRES_USER=testuser
POSTGRES_PASSWORD=testpass
POSTGRES_DB=b25_test

# Mock Exchange
MOCK_EXCHANGE_HTTP=http://localhost:8545
MOCK_EXCHANGE_WS=ws://localhost:8546

# Test Configuration
TEST_TIMEOUT=10m
TEST_PARALLEL=4
```

Load environment variables:
```bash
export $(cat .env | xargs)
```

## Running Tests

### Integration Tests

Test individual service integrations:

```bash
# All integration tests
cd integration
go test -v ./...

# Specific test
go test -v -run TestMarketDataPipeline

# With coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### E2E Tests

Test complete system workflows:

```bash
# All E2E tests
cd e2e
go test -v ./...

# Trading flow only
go test -v -run TestTradingFlow

# Failover tests only
go test -v -run TestFailover
```

### Performance Benchmarks

```bash
# All benchmarks
cd e2e
go test -v -run TestLatencyBenchmark

# Specific benchmark
go test -v -run TestMarketDataIngestionLatency

# With profiling
go test -v -run TestLatencyBenchmark -cpuprofile=cpu.prof -memprofile=mem.prof
go tool pprof cpu.prof
```

### Using Makefile

```bash
# Run everything
make test

# Individual suites
make integration
make e2e
make benchmark

# Quick test (skips long tests)
make quick-test

# Parallel execution
make parallel

# With coverage report
make coverage
```

## Troubleshooting

### Services Won't Start

**Problem:** Docker containers fail to start

**Solutions:**
```bash
# Check Docker daemon
sudo systemctl status docker

# Check port conflicts
sudo lsof -i :6380  # Redis
sudo lsof -i :4223  # NATS
sudo lsof -i :5433  # PostgreSQL

# Restart Docker
sudo systemctl restart docker

# Clean and rebuild
make clean
make setup
```

### Tests Are Skipping

**Problem:** Tests skip with "service not running" messages

**Solutions:**
```bash
# Verify all services are healthy
docker-compose -f testutil/docker/docker-compose.test.yml ps

# Check service logs
docker-compose -f testutil/docker/docker-compose.test.yml logs redis-test
docker-compose -f testutil/docker/docker-compose.test.yml logs nats-test

# Restart specific service
docker-compose -f testutil/docker/docker-compose.test.yml restart redis-test
```

### Connection Timeouts

**Problem:** Tests fail with connection timeouts

**Solutions:**
```bash
# Increase timeout in test code
go test -v -timeout 30m ./...

# Check firewall rules
sudo ufw status

# Verify network connectivity
docker network inspect b25-test-network
```

### Database Schema Issues

**Problem:** Database schema not initialized

**Solutions:**
```bash
# Manually run initialization script
docker exec -i b25-postgres-test psql -U testuser -d b25_test < testutil/docker/init-db.sql

# Or recreate database
docker-compose -f testutil/docker/docker-compose.test.yml down -v
docker-compose -f testutil/docker/docker-compose.test.yml up -d
```

### Go Module Issues

**Problem:** Import errors or missing dependencies

**Solutions:**
```bash
# Clean and download
cd testutil && go clean -modcache && go mod download
cd ../integration && go clean -modcache && go mod download
cd ../e2e && go clean -modcache && go mod download

# Update dependencies
cd testutil && go get -u ./... && go mod tidy
cd ../integration && go get -u ./... && go mod tidy
cd ../e2e && go get -u ./... && go mod tidy
```

## Performance Tuning

### For Development

```yaml
# docker-compose.test.yml modifications
services:
  redis-test:
    environment:
      - REDIS_SAVE=""  # Disable persistence
```

### For CI/CD

```bash
# Use smaller timeouts
export TEST_TIMEOUT=5m

# Reduce parallelism
export TEST_PARALLEL=2

# Skip slow tests
go test -short ./...
```

### For Benchmarking

```bash
# Disable other processes
docker-compose -f testutil/docker/docker-compose.test.yml stop order-execution-test strategy-engine-test

# Run only mock exchange
docker-compose -f testutil/docker/docker-compose.test.yml up -d redis-test nats-test mock-exchange
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Start test infrastructure
        run: |
          cd tests
          docker-compose -f testutil/docker/docker-compose.test.yml up -d
          sleep 15

      - name: Run integration tests
        run: |
          cd tests/integration
          go test -v ./...

      - name: Run E2E tests
        run: |
          cd tests/e2e
          go test -v ./...

      - name: Cleanup
        if: always()
        run: |
          cd tests
          docker-compose -f testutil/docker/docker-compose.test.yml down -v
```

### GitLab CI

```yaml
test:
  image: golang:1.21
  services:
    - docker:dind
  before_script:
    - cd tests
    - docker-compose -f testutil/docker/docker-compose.test.yml up -d
    - sleep 15
  script:
    - cd integration && go test -v ./...
    - cd ../e2e && go test -v ./...
  after_script:
    - cd tests
    - docker-compose -f testutil/docker/docker-compose.test.yml down -v
```

## Development Workflow

### Before Committing

```bash
# Run all tests
make test

# Check test coverage
make coverage

# Run linter
make lint
```

### Adding New Tests

1. Create test file: `*_test.go`
2. Use testify suite structure
3. Add to appropriate directory (integration/e2e)
4. Update README.md
5. Run tests: `go test -v -run YourNewTest`

### Debugging Tests

```bash
# Run single test with verbose output
go test -v -run TestSpecificTest

# Add logging in test
s.T().Logf("Debug info: %v", variable)

# Use delve debugger
dlv test -- -test.run TestSpecificTest
```

## Maintenance

### Regular Cleanup

```bash
# Weekly cleanup
make clean

# Remove old Docker images
docker image prune -a

# Clear Go build cache
go clean -cache
```

### Update Dependencies

```bash
# Update all modules
cd testutil && go get -u ./... && go mod tidy
cd ../integration && go get -u ./... && go mod tidy
cd ../e2e && go get -u ./... && go mod tidy

# Test after update
make test
```

### Monitor Resource Usage

```bash
# Check Docker resource usage
docker stats

# Check disk space
docker system df

# Clean unused resources
docker system prune -a
```

## Support

For issues:
1. Check service logs: `make logs`
2. Verify service status: `make ps`
3. Review troubleshooting section
4. Open GitHub issue with logs

For questions:
- See README.md
- Check test examples
- Review inline documentation
