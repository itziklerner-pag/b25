# B25 Trading Platform - Deployment Guide

This guide covers deployment scenarios for the B25 trading platform.

## Quick Start

### Development Environment

```bash
# 1. Copy environment file
cp .env.example .env

# 2. Start all services
./scripts/dev-start.sh

# 3. Access services
# - Web Dashboard: http://localhost:3000
# - API Gateway: http://localhost:8000
# - Grafana: http://localhost:3001 (admin/admin)
```

### Stop Development Environment

```bash
# Stop services (keep data)
./scripts/dev-stop.sh

# Stop services and remove all data
REMOVE_VOLUMES=true ./scripts/dev-stop.sh
```

## Service Architecture

### Infrastructure Services
- **Redis** (6379): In-memory cache and pub/sub
- **PostgreSQL** (5432): Configuration and relational data
- **TimescaleDB** (5433): Time-series data storage
- **NATS** (4222): Message broker with JetStream

### Core Trading Services
- **Market Data** (8080, 50051): Real-time market data ingestion
- **Order Execution** (8081, 50052): Order routing and execution
- **Strategy Engine** (8082, 50053): Trading strategy execution
- **Risk Manager** (8083, 50054): Risk controls and limits
- **Account Monitor** (8084, 50055): Account tracking and analytics
- **Configuration** (8085, 50056): Centralized configuration management
- **Dashboard Server** (8086): WebSocket server for real-time updates

### API & UI Services
- **API Gateway** (8000): Main API entry point
- **Auth Service** (8001): Authentication and authorization
- **Web Dashboard** (3000): React-based trading dashboard

### Observability Stack
- **Prometheus** (9090): Metrics collection
- **Grafana** (3001): Metrics visualization
- **Alertmanager** (9093): Alert routing

## Port Mapping

| Service | HTTP | gRPC | Metrics |
|---------|------|------|---------|
| Market Data | 8080 | 50051 | 9100 |
| Order Execution | 8081 | 50052 | 9101 |
| Strategy Engine | 8082 | 50053 | 9102 |
| Risk Manager | 8083 | 50054 | 9103 |
| Account Monitor | 8084 | 50055 | 9104 |
| Configuration | 8085 | 50056 | 9105 |
| Dashboard Server | 8086 | - | 9106 |
| API Gateway | 8000 | - | 9000 |
| Auth Service | 8001 | - | 9001 |

## Network Architecture

All services communicate via the `b25-network` Docker bridge network:

```
┌─────────────────────────────────────────────────────────┐
│                    b25-network (bridge)                  │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  ┌───────────────────────────────────────────────┐      │
│  │         Infrastructure Layer                   │      │
│  │  • Redis  • PostgreSQL  • TimescaleDB  • NATS │      │
│  └───────────────────────────────────────────────┘      │
│                         ↑                                │
│  ┌──────────────────────────────────────────────────┐   │
│  │         Core Trading Services                     │   │
│  │  • Market Data  • Order Execution                │   │
│  │  • Strategy Engine  • Risk Manager               │   │
│  │  • Account Monitor  • Configuration              │   │
│  └──────────────────────────────────────────────────┘   │
│                         ↑                                │
│  ┌──────────────────────────────────────────────────┐   │
│  │         API & UI Layer                            │   │
│  │  • API Gateway  • Auth  • Dashboard Server       │   │
│  │  • Web Dashboard                                  │   │
│  └──────────────────────────────────────────────────┘   │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

## Production Deployment

### Prerequisites

1. **Docker and Docker Compose** installed
2. **Production environment file** configured
3. **SSL certificates** for HTTPS
4. **Domain names** configured
5. **Container registry** access (GitHub Container Registry)

### Deployment Steps

```bash
# 1. Create production environment file
cp .env.production.example .env.production

# 2. Update .env.production with:
#    - Strong passwords (min 32 characters)
#    - JWT secrets (min 64 characters)
#    - Exchange API credentials
#    - Domain names
#    - SSL certificate paths

# 3. Build and push Docker images
DOCKER_REGISTRY=ghcr.io/yourorg VERSION=1.0.0 PUSH=true ./scripts/docker-build-all.sh

# 4. Deploy to production
ENV_FILE=.env.production VERSION=1.0.0 ./scripts/deploy-prod.sh
```

### Health Checks

All services implement health check endpoints:

```bash
# Check individual service
curl http://localhost:8080/health

# Check all services
for port in 8080 8081 8082 8083 8084 8085 8086 8000 8001; do
  echo "Checking port $port..."
  curl -f "http://localhost:$port/health" && echo " ✓" || echo " ✗"
done
```

## Service Dependencies

Services start in the following order to ensure dependencies are ready:

1. **Infrastructure**: redis, postgres, timescaledb, nats
2. **Observability**: prometheus, grafana, alertmanager
3. **Core Services**: configuration → market-data → order-execution → strategy-engine → risk-manager → account-monitor → dashboard-server
4. **API/UI**: auth → api-gateway → web-dashboard

## Environment Variables

### Required for All Environments

```bash
# Database
POSTGRES_PASSWORD=<secure-password>
TIMESCALEDB_PASSWORD=<secure-password>
REDIS_PASSWORD=<secure-password>

# Security
JWT_SECRET=<64-char-random-string>

# Exchange
EXCHANGE_API_KEY=<your-api-key>
EXCHANGE_SECRET_KEY=<your-secret>
```

### Production-Specific

```bash
# Docker
DOCKER_REGISTRY=ghcr.io/yourorg
VERSION=1.0.0

# URLs
API_URL=https://api.yourdomain.com
WS_URL=wss://ws.yourdomain.com
CORS_ORIGINS=https://yourdomain.com

# Monitoring
GRAFANA_ROOT_URL=https://grafana.yourdomain.com
SLACK_WEBHOOK_URL=https://hooks.slack.com/...

# SSL
SSL_CERT_PATH=/etc/nginx/ssl/cert.pem
SSL_KEY_PATH=/etc/nginx/ssl/key.pem
```

## Scaling

### Horizontal Scaling

Scale specific services based on load:

```bash
# Scale order execution service
docker-compose -f docker/docker-compose.prod.yml up -d --scale order-execution=3

# Scale strategy engine
docker-compose -f docker/docker-compose.prod.yml up -d --scale strategy-engine=2
```

### Resource Limits

Services have resource limits defined in docker-compose.prod.yml:

- **Market Data**: 2 CPU, 2GB RAM
- **Order Execution**: 4 CPU, 4GB RAM
- **Strategy Engine**: 4 CPU, 8GB RAM
- **TimescaleDB**: 4 CPU, 8GB RAM

## Monitoring

### Prometheus Metrics

All services expose metrics at `/metrics` endpoint on their metrics port:

- Market Data: http://localhost:9100/metrics
- Order Execution: http://localhost:9101/metrics
- Strategy Engine: http://localhost:9102/metrics

### Grafana Dashboards

Access Grafana at http://localhost:3001:

- **Trading Overview**: Real-time P&L, positions, orders
- **System Performance**: CPU, memory, network
- **Service Health**: Request rates, error rates, latency

### Alerts

Alertmanager routes alerts to:

- **Critical**: PagerDuty + SMS
- **Warning**: Slack
- **Info**: Email

## Backup and Recovery

### Manual Backup

```bash
# Backup all volumes
docker run --rm -v b25-postgres-data-prod:/data -v $(pwd)/backups:/backup \
  alpine tar czf /backup/postgres-$(date +%Y%m%d).tar.gz /data

# Backup TimescaleDB
docker run --rm -v b25-timescaledb-data-prod:/data -v $(pwd)/backups:/backup \
  alpine tar czf /backup/timescaledb-$(date +%Y%m%d).tar.gz /data
```

### Automated Backup

Backups run daily at 2 AM (configured via BACKUP_SCHEDULE):

```bash
# View backup status
docker-compose -f docker/docker-compose.prod.yml logs backup-service
```

### Restore

```bash
# Stop services
docker-compose -f docker/docker-compose.prod.yml down

# Restore volume
docker run --rm -v b25-postgres-data-prod:/data -v $(pwd)/backups:/backup \
  alpine tar xzf /backup/postgres-20250101.tar.gz -C /

# Start services
./scripts/deploy-prod.sh
```

## Troubleshooting

### Service Won't Start

```bash
# Check logs
docker-compose -f docker/docker-compose.dev.yml logs -f [service-name]

# Check service status
docker-compose -f docker/docker-compose.dev.yml ps

# Restart specific service
docker-compose -f docker/docker-compose.dev.yml restart [service-name]
```

### Connection Issues

```bash
# Test Redis connection
docker exec -it b25-redis redis-cli ping

# Test PostgreSQL connection
docker exec -it b25-postgres psql -U b25 -d b25_config -c "SELECT 1"

# Test NATS connection
curl http://localhost:8222/healthz
```

### Performance Issues

```bash
# Check resource usage
docker stats

# Check service metrics
curl http://localhost:9100/metrics | grep -E "cpu|memory"

# View Prometheus targets
open http://localhost:9090/targets
```

## Security Considerations

### Network Security

- All services communicate on private Docker network
- Only necessary ports exposed to host
- HTTPS/TLS termination at nginx reverse proxy

### Authentication

- JWT-based authentication with rotating secrets
- API keys for service-to-service communication
- Role-based access control (RBAC)

### Data Security

- Encrypted database connections (SSL mode required)
- Secrets stored in environment variables (not in code)
- Sensitive data encrypted at rest

### Audit Logging

All critical operations logged:

- Order creation/modification/cancellation
- Configuration changes
- Authentication attempts
- Risk limit breaches

## CI/CD Integration

### GitHub Actions Workflow

Automated pipeline on push to main:

1. **Test**: Run unit and integration tests
2. **Build**: Build Docker images for all services
3. **Scan**: Security scanning with Trivy
4. **Push**: Push images to GitHub Container Registry
5. **Deploy**: Automatic deployment to staging

### Manual Production Deployment

```bash
# Tag release
git tag -a v1.0.0 -m "Release 1.0.0"
git push origin v1.0.0

# GitHub Actions will build and tag images
# Then deploy manually:
VERSION=v1.0.0 ./scripts/deploy-prod.sh
```

## Support

For deployment issues:

1. Check logs: `docker-compose -f docker/docker-compose.dev.yml logs`
2. Review this guide
3. Check service health endpoints
4. Contact DevOps team
