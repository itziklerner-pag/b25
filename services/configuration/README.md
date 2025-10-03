# Configuration Service

Centralized configuration management with versioning, hot-reload, and audit logging for the HFT trading system.

**Language**: Go 1.21+
**Development Plan**: `../../docs/service-plans/07-configuration-service.md`

## Features

- **Configuration Management**: CRUD operations for all system configurations
- **Type Support**: Strategies, risk limits, trading pairs, and system settings
- **Versioning**: Full version history with rollback capability
- **Hot-Reload**: Real-time configuration updates via NATS pub/sub
- **Validation**: Schema validation before saving configurations
- **Audit Logging**: Complete audit trail of all configuration changes
- **Multi-Format**: Support for JSON and YAML configuration formats
- **REST & gRPC APIs**: Dual API support for flexibility
- **Prometheus Metrics**: Built-in observability

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Start all services (PostgreSQL, NATS, Configuration Service)
docker-compose up -d

# View logs
docker-compose logs -f configuration-service

# Stop services
docker-compose down
```

### Manual Setup

1. **Install Dependencies**
   ```bash
   go mod download
   ```

2. **Setup Database**
   ```bash
   # Create database
   createdb configuration_db

   # Run migrations
   make migrate-up
   ```

3. **Configure Service**
   ```bash
   cp config.example.yaml config.yaml
   # Edit config.yaml with your settings
   ```

4. **Run Service**
   ```bash
   # Using Make
   make run

   # Or directly
   go run ./cmd/server
   ```

## Configuration

Copy `config.example.yaml` to `config.yaml` and adjust settings:

```yaml
server:
  host: 0.0.0.0
  http_port: 9096
  grpc_port: 9097

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: configuration_db
  sslmode: disable

nats:
  url: nats://localhost:4222
  config_topic: config.updates
```

## API Endpoints

### Health & Metrics
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /metrics` - Prometheus metrics

### Configuration Management
- `POST /api/v1/configurations` - Create configuration
- `GET /api/v1/configurations` - List configurations (with filters)
- `GET /api/v1/configurations/:id` - Get configuration by ID
- `GET /api/v1/configurations/key/:key` - Get configuration by key
- `PUT /api/v1/configurations/:id` - Update configuration
- `DELETE /api/v1/configurations/:id` - Delete configuration

### Status Management
- `POST /api/v1/configurations/:id/activate` - Activate configuration
- `POST /api/v1/configurations/:id/deactivate` - Deactivate configuration

### Versioning
- `GET /api/v1/configurations/:id/versions` - Get all versions
- `POST /api/v1/configurations/:id/rollback` - Rollback to version

### Audit
- `GET /api/v1/configurations/:id/audit-logs` - Get audit logs

## Configuration Types

### Strategy Configuration
```json
{
  "key": "market_making_strategy",
  "type": "strategy",
  "value": {
    "name": "Market Making",
    "type": "market_making",
    "enabled": true,
    "parameters": {
      "spread": 0.002,
      "order_size": 100
    }
  },
  "format": "json",
  "description": "Market making strategy configuration"
}
```

### Risk Limit Configuration
```json
{
  "key": "default_risk_limits",
  "type": "risk_limit",
  "value": {
    "max_position_size": 10000,
    "max_loss_per_trade": 500,
    "max_daily_loss": 2000,
    "max_leverage": 10,
    "stop_loss_percent": 5
  },
  "format": "json",
  "description": "Default risk limits"
}
```

### Trading Pair Configuration
```json
{
  "key": "btc_usdt_pair",
  "type": "trading_pair",
  "value": {
    "symbol": "BTC/USDT",
    "base_currency": "BTC",
    "quote_currency": "USDT",
    "min_order_size": 0.001,
    "max_order_size": 10,
    "price_precision": 2,
    "quantity_precision": 8,
    "enabled": true
  },
  "format": "json",
  "description": "BTC/USDT trading pair"
}
```

### System Configuration
```json
{
  "key": "system_maintenance",
  "type": "system",
  "value": {
    "name": "maintenance_mode",
    "value": false,
    "type": "boolean"
  },
  "format": "json",
  "description": "System maintenance mode"
}
```

## NATS Events

Configuration updates are published to NATS for hot-reload:

**Topic Pattern**: `config.updates.{type}`

**Event Format**:
```json
{
  "id": "uuid",
  "key": "config_key",
  "type": "strategy",
  "value": {...},
  "format": "json",
  "version": 2,
  "action": "updated",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

**Actions**: `created`, `updated`, `activated`, `deactivated`, `deleted`

## Development

### Running Tests
```bash
make test
```

### Test with Coverage
```bash
make test-coverage
```

### Code Formatting
```bash
make fmt
```

### Linting
```bash
make lint
```

### Hot Reload (requires air)
```bash
make dev
```

## Database Migrations

### Apply Migrations
```bash
make migrate-up
```

### Rollback Migrations
```bash
make migrate-down
```

### Create New Migration
```bash
make migrate-create
# Enter migration name when prompted
```

## Docker

### Build Image
```bash
make docker-build
```

### Run Container
```bash
make docker-run
```

### Stop Container
```bash
make docker-stop
```

## Metrics

Prometheus metrics available at `http://localhost:9096/metrics`:

- `config_operations_total` - Total configuration operations
- `config_operation_duration_seconds` - Operation duration
- `active_configurations` - Number of active configurations by type
- `config_versions` - Current version number of configurations
- `config_update_events_total` - NATS update events published
- `config_validation_errors_total` - Validation errors by type

## Architecture

```
cmd/
  server/
    main.go           # Entry point

internal/
  api/                # HTTP handlers
    handler.go
    configuration_handler.go
    health_handler.go
    router.go

  config/             # Configuration loading
    config.go

  domain/             # Domain models
    configuration.go
    errors.go

  metrics/            # Prometheus metrics
    metrics.go

  repository/         # Database layer
    configuration_repository.go

  service/            # Business logic
    configuration_service.go

  validator/          # Configuration validation
    validator.go

migrations/           # Database migrations
  000001_init_schema.up.sql
  000001_init_schema.down.sql
```

## Security Considerations

1. **Input Validation**: All configurations are validated before saving
2. **Audit Logging**: Complete audit trail with IP and user agent tracking
3. **Access Control**: Implement authentication middleware in production
4. **Encryption**: Use TLS for NATS and database connections in production
5. **Secrets**: Never store secrets in configurations, use secret management

## Production Deployment

1. **Environment Variables**: Override config with environment variables
2. **Database**: Use managed PostgreSQL with replication
3. **NATS**: Deploy NATS cluster for high availability
4. **Monitoring**: Configure Prometheus scraping
5. **Logging**: Ship logs to centralized logging system
6. **Backups**: Regular database backups with point-in-time recovery

## Troubleshooting

### Database Connection Issues
```bash
# Test database connection
psql -h localhost -U postgres -d configuration_db

# Check migrations status
migrate -path migrations -database "postgres://..." version
```

### NATS Connection Issues
```bash
# Check NATS server
curl http://localhost:8222/varz

# Test NATS connectivity
nats-bench pub config.updates.test -m 1
```

### Service Not Starting
```bash
# Check logs
docker-compose logs configuration-service

# Verify configuration
cat config.yaml

# Test build
go build ./cmd/server
```

## Contributing

1. Follow Go best practices and conventions
2. Add tests for new features
3. Update documentation
4. Run linting before committing
5. Keep dependencies up to date

## License

Proprietary - All rights reserved
