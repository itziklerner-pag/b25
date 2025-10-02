# Scripts

Utility scripts for building, testing, and deploying the trading system.

## Available Scripts

### Testing

- **`test-all.sh`** - Run all unit tests across all services
- **`test-integration.sh`** - Run integration tests with Docker Compose
- **`test-e2e.sh`** - Run end-to-end tests (TODO)

### Building

- **`build-all.sh`** - Build all services locally
- **`docker-build-all.sh`** - Build all Docker images (TODO)

### Development

- **`dev-start.sh`** - Start development environment (TODO)
- **`dev-stop.sh`** - Stop development environment (TODO)
- **`generate-proto.sh`** - Generate code from protobuf definitions (TODO)

### Deployment

- **`deploy-staging.sh`** - Deploy to staging environment (TODO)
- **`deploy-prod.sh`** - Deploy to production (TODO)

## Usage

```bash
# Make scripts executable
chmod +x scripts/*.sh

# Run all tests
./scripts/test-all.sh

# Build all services
./scripts/build-all.sh

# Run integration tests
./scripts/test-integration.sh
```

## CI/CD Integration

These scripts are used by GitHub Actions workflows in `.github/workflows/`
