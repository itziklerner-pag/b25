# B25 - High-Frequency Trading System

A production-grade, microservices-based high-frequency trading system for cryptocurrency futures markets.

## 🏗️ Monorepo Structure

```
b25/
├── services/           # Backend microservices
│   ├── market-data/           # Rust - Real-time data ingestion
│   ├── order-execution/       # Go - Order lifecycle management
│   ├── strategy-engine/       # Go - Trading strategies
│   ├── account-monitor/       # Go - Balance & P&L tracking
│   ├── dashboard-server/      # Go - WebSocket state aggregation
│   ├── risk-manager/          # Go - Risk management
│   ├── configuration/         # Go - Config management
│   └── metrics/               # Prometheus/Grafana setup
├── ui/                 # User interfaces
│   ├── terminal/              # Rust - Terminal UI (TUI)
│   └── web/                   # React/JavaScript - Web dashboard
├── shared/             # Shared code and schemas
│   ├── proto/                 # Protobuf definitions
│   ├── schemas/               # Data schemas
│   └── lib/                   # Shared libraries
├── docker/             # Docker configurations
├── k8s/                # Kubernetes manifests (optional)
├── scripts/            # Build, deploy, test scripts
├── tests/              # Integration and E2E tests
└── docs/               # Documentation
```

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Git
- Language-specific tools (installed per service):
  - Rust 1.75+ (for market-data, terminal)
  - Go 1.21+ (for most backend services)
  - Node.js 20+ (for web dashboard)
  - Python 3.11+ (for strategy plugins)

### Local Development

```bash
# Clone repository
git clone <repo-url>
cd b25

# Start all services with Docker Compose
docker-compose -f docker/docker-compose.dev.yml up

# Or start specific services
docker-compose -f docker/docker-compose.dev.yml up market-data order-execution

# Access dashboards
# - Web Dashboard: http://localhost:3000
# - Grafana: http://localhost:3001
# - Prometheus: http://localhost:9090
```

### Build Individual Services

Each service has its own README with build instructions:

```bash
# Market Data Service (Rust)
cd services/market-data
cargo build --release

# Order Execution Service (Go)
cd services/order-execution
go build -o bin/order-execution ./cmd/server

# Web Dashboard
cd ui/web
npm install
npm run dev
```

## 📚 Documentation

- **[System Architecture](docs/SYSTEM_ARCHITECTURE.md)** - Complete architectural blueprint
- **[Component Specifications](docs/COMPONENT_SPECIFICATIONS.md)** - Detailed component specs
- **[Implementation Guide](docs/IMPLEMENTATION_GUIDE.md)** - Step-by-step implementation
- **[Sub-Systems](docs/sub-systems.md)** - Microservices architecture
- **[Service Plans](docs/service-plans/)** - Development plans for each service

## 🏢 Architecture Overview

### Four-Layer Architecture

1. **Persistence Layer**: Redis, TimescaleDB, PostgreSQL
2. **Data Plane**: Market data, order execution (ultra-low latency)
3. **Control Plane**: Strategy engine, risk manager
4. **Observability Layer**: Metrics, logging, dashboards

### Communication

- **Synchronous**: gRPC for inter-service RPC
- **Asynchronous**: NATS/Redis Pub/Sub for events
- **Real-time**: WebSocket for UI updates

### Performance Targets

| Metric | Target | Critical |
|--------|--------|----------|
| Market Data Latency | <100μs | <500μs |
| Order Execution | <10ms | <50ms |
| Strategy Decision | <500μs | <2ms |
| System Uptime | 99.99% | 99.9% |

## 🧪 Testing

```bash
# Run all unit tests
./scripts/test-all.sh

# Run integration tests
./scripts/test-integration.sh

# Run E2E tests
./scripts/test-e2e.sh

# Service-specific tests
cd services/order-execution
go test ./...
```

## 🚢 Deployment

### Development
```bash
docker-compose -f docker/docker-compose.dev.yml up
```

### Production
```bash
docker-compose -f docker/docker-compose.prod.yml up -d
```

### Kubernetes (Optional)
```bash
kubectl apply -f k8s/
```

## 🔧 Configuration

Each service is configured via:
- Environment variables (`.env` files)
- Configuration files (`config.yaml`)
- Configuration Service (runtime updates)

See individual service READMEs for specific configuration options.

## 📊 Monitoring

- **Grafana**: http://localhost:3001 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Alertmanager**: http://localhost:9093

Pre-configured dashboards:
- System Health Overview
- Performance Metrics
- Business Metrics (P&L, Trades)
- Per-Service Dashboards

## 🔐 Security

⚠️ **CRITICAL**: Never commit secrets to the repository!

- API keys: Use environment variables or secrets management
- Credentials: Store in `.env` files (gitignored)
- Production: Use Docker secrets or Kubernetes secrets
- TLS: Required for all external connections

## 📜 JavaScript-Only Policy

**IMPORTANT:** This project follows a strict JavaScript-only policy:

- All code must be written in pure JavaScript (ES6+)
- No TypeScript syntax or type annotations allowed
- Use JSDoc comments for type documentation when needed
- No `.ts` or `.tsx` files in source code
- Focus on clean, well-documented JavaScript

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## 🤝 Contributing

1. Each service can be developed independently
2. Update shared protobuf definitions in `shared/proto/`
3. Run tests before committing
4. Follow service-specific style guides
5. Update documentation

## 📝 Development Workflow

### Adding a New Feature

1. Create feature branch: `git checkout -b feature/my-feature`
2. Develop in relevant service: `cd services/my-service`
3. Write tests: Follow service testing guide
4. Update documentation: Update service README
5. Submit PR: CI will run tests automatically

### CI/CD Pipeline

- **On Push**: Run tests for changed services only
- **On PR**: Full integration test suite
- **On Merge to Main**: Deploy to staging
- **On Tag**: Deploy to production

## 🐛 Troubleshooting

### Common Issues

**Docker build failures**
```bash
# Clean Docker cache
docker system prune -a
docker-compose build --no-cache
```

**Service won't start**
```bash
# Check logs
docker-compose logs <service-name>

# Check health
docker-compose ps
```

**WebSocket connection issues**
```bash
# Verify dashboard-server is running
curl http://localhost:8080/health
```

See individual service READMEs for service-specific troubleshooting.

## 📈 Performance Tuning

- Market Data: Adjust buffer sizes in config
- Order Execution: Tune connection pool sizes
- Strategy Engine: Optimize strategy algorithms
- Dashboard: Adjust update rates (100ms TUI, 250ms Web)

## 🗺️ Roadmap

- [x] Architecture documentation
- [x] Service development plans
- [ ] Core services implementation
- [ ] Integration testing
- [ ] Production deployment
- [ ] Multi-exchange support
- [ ] Advanced strategies
- [ ] Machine learning integration

## 📄 License

[Specify your license]

## 💬 Support

- Issues: GitHub Issues
- Documentation: `docs/` directory
- Service-specific: See individual READMEs

---

**Built for high-frequency trading. Optimized for ultra-low latency. Designed for reliability.**
