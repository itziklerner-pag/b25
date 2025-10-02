# B25 Monorepo Structure

Complete directory structure of the B25 HFT trading system monorepo.

## 📁 Directory Tree

```
b25/
├── .github/
│   └── workflows/
│       ├── ci.yml                    # Selective service CI/CD
│       └── docker-build.yml          # Docker image builds
│
├── services/                         # Backend microservices
│   ├── market-data/                 # Rust - Market data ingestion
│   │   ├── src/
│   │   ├── tests/
│   │   ├── Cargo.toml
│   │   ├── Dockerfile
│   │   ├── config.example.yaml
│   │   └── README.md
│   │
│   ├── order-execution/             # Go - Order management
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   ├── pkg/
│   │   ├── go.mod
│   │   ├── Dockerfile
│   │   ├── config.example.yaml
│   │   └── README.md
│   │
│   ├── strategy-engine/             # Go - Strategy execution
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   ├── plugins/                # Strategy plugins
│   │   │   ├── go/                # Go plugins
│   │   │   └── python/            # Python plugins
│   │   ├── go.mod
│   │   ├── Dockerfile
│   │   └── README.md
│   │
│   ├── account-monitor/             # Go - Balance & P&L
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   ├── go.mod
│   │   ├── Dockerfile
│   │   └── README.md
│   │
│   ├── dashboard-server/            # Go - WebSocket server
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   ├── go.mod
│   │   ├── Dockerfile
│   │   └── README.md
│   │
│   ├── risk-manager/                # Go - Risk management
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   ├── go.mod
│   │   ├── Dockerfile
│   │   └── README.md
│   │
│   ├── configuration/               # Go - Config management
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   ├── migrations/             # DB migrations
│   │   ├── go.mod
│   │   ├── Dockerfile
│   │   └── README.md
│   │
│   ├── metrics/                     # Observability configs
│   │   ├── prometheus/
│   │   │   ├── prometheus.yml
│   │   │   └── alerts/
│   │   ├── alertmanager/
│   │   │   └── alertmanager.yml
│   │   ├── grafana/
│   │   │   ├── provisioning/
│   │   │   │   ├── datasources/
│   │   │   │   └── dashboards/
│   │   │   └── dashboards/
│   │   └── README.md
│   │
│   └── README.md                    # Services overview
│
├── ui/                              # User interfaces
│   ├── terminal/                    # Rust - Terminal UI
│   │   ├── src/
│   │   ├── tests/
│   │   ├── Cargo.toml
│   │   ├── config.example.yaml
│   │   └── README.md
│   │
│   └── web/                         # React/JavaScript - Web dashboard
│       ├── src/
│       │   ├── components/
│       │   ├── pages/
│       │   ├── hooks/
│       │   ├── store/
│       │   └── utils/
│       ├── public/
│       ├── tests/
│       ├── package.json
│       ├── vite.config.js
│       ├── Dockerfile
│       └── README.md
│
├── shared/                          # Shared code and schemas
│   ├── proto/                       # Protobuf definitions
│   │   ├── market_data.proto
│   │   ├── orders.proto
│   │   ├── account.proto
│   │   ├── config.proto
│   │   └── common.proto
│   │
│   ├── schemas/                     # Data schemas
│   │   ├── order_request.schema.json
│   │   ├── strategy_config.schema.json
│   │   └── risk_limits.schema.json
│   │
│   ├── lib/                         # Shared libraries
│   │   ├── go/
│   │   │   ├── types/
│   │   │   ├── utils/
│   │   │   └── metrics/
│   │   ├── rust/
│   │   │   └── common/
│   │   └── python/
│   │       └── b25common/
│   │
│   └── README.md
│
├── docker/                          # Docker configurations
│   ├── docker-compose.dev.yml       # Development compose
│   ├── docker-compose.prod.yml      # Production compose
│   └── docker-compose.test.yml      # Testing compose
│
├── k8s/                             # Kubernetes manifests
│   ├── namespace.yaml
│   ├── configmaps/
│   ├── secrets/
│   ├── deployments/
│   ├── services/
│   └── ingress/
│
├── scripts/                         # Utility scripts
│   ├── test-all.sh                 # Run all tests
│   ├── test-integration.sh         # Integration tests
│   ├── build-all.sh                # Build all services
│   ├── generate-proto.sh           # Generate protobuf code
│   └── README.md
│
├── tests/                           # Cross-service tests
│   ├── integration/                # Integration tests
│   │   ├── market_data_test.go
│   │   ├── order_flow_test.go
│   │   └── README.md
│   │
│   └── e2e/                        # End-to-end tests
│       ├── trading_flow_test.go
│       └── README.md
│
├── docs/                            # Documentation
│   ├── README.md                   # Documentation overview
│   ├── INDEX.md                    # Documentation index
│   ├── SYSTEM_ARCHITECTURE.md      # Architecture blueprint
│   ├── COMPONENT_SPECIFICATIONS.md # Component specs
│   ├── IMPLEMENTATION_GUIDE.md     # Implementation guide
│   ├── sub-systems.md              # Microservices design
│   │
│   └── service-plans/              # Development plans
│       ├── 01-market-data-service.md
│       ├── 02-order-execution-service.md
│       ├── 03-strategy-engine-service.md
│       ├── 04-account-monitor-service.md
│       ├── 05-dashboard-server-service.md
│       ├── 06-risk-manager-service.md
│       ├── 07-configuration-service.md
│       ├── 08-metrics-observability-service.md
│       ├── 09-terminal-ui-service.md
│       └── 10-web-dashboard-service.md
│
├── .gitignore                       # Git ignore rules
├── README.md                        # Project README
├── CONTRIBUTING.md                  # Contribution guide
├── MONOREPO_STRUCTURE.md           # This file
└── LICENSE                          # License file

```

## 🎯 Key Principles

### Service Independence
- Each service has its own build/test/deploy cycle
- Services can use different tech stacks
- Clear API boundaries via protobuf/gRPC

### Shared Code Management
- Common code in `shared/` directory
- Protobuf definitions for inter-service contracts
- Language-specific shared libraries

### CI/CD Optimization
- Selective builds: Only changed services are tested/built
- Path-based filtering in GitHub Actions
- Fast feedback loops

### Development Workflow
- Clone once, access all services
- Unified Docker Compose for local development
- Consistent tooling across services

## 📊 Service Matrix

| Service | Language | Port | Dependencies |
|---------|----------|------|--------------|
| market-data | Rust | 9090 | Redis, QuestDB |
| order-execution | Go | 9091 | Redis, NATS |
| strategy-engine | Go+Python | 9092 | NATS |
| account-monitor | Go | 9093 | TimescaleDB, Redis |
| dashboard-server | Go | 8080 | Redis, all services |
| risk-manager | Go | 9095 | PostgreSQL |
| configuration | Go | 9096 | PostgreSQL |
| terminal-ui | Rust | - | dashboard-server |
| web-dashboard | React | 3000 | dashboard-server |

## 🔄 Data Flow

```
Exchange WebSocket
    ↓
market-data (port 9090)
    ↓ (Redis Pub/Sub)
strategy-engine (port 9092)
    ↓ (gRPC)
order-execution (port 9091)
    ↓ (REST API)
Exchange REST API
    ↓ (WebSocket user stream)
account-monitor (port 9093)
    ↓
All → dashboard-server (port 8080)
    ↓ (WebSocket)
terminal-ui / web-dashboard (port 3000)
```

## 🚀 Quick Commands

```bash
# Start infrastructure
docker-compose -f docker/docker-compose.dev.yml up -d redis postgres timescaledb nats

# Build all services
./scripts/build-all.sh

# Test all services
./scripts/test-all.sh

# Start a specific service
cd services/market-data && cargo run

# Start all services
docker-compose -f docker/docker-compose.dev.yml up
```

## 📈 Metrics & Observability

All services expose:
- **Metrics**: `http://localhost:<port>/metrics` (Prometheus format)
- **Health**: `http://localhost:<port>/health`

Observability stack:
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001
- **Alertmanager**: http://localhost:9093

## 🔐 Configuration

Each service can be configured via:
1. Environment variables (`.env` files)
2. YAML config files (`config.yaml`)
3. Configuration service (runtime)

See individual service READMEs for specific options.

## 📝 Development Status

- [x] Monorepo structure
- [x] Documentation
- [x] Service development plans
- [x] CI/CD pipeline
- [ ] Service implementations
- [ ] Integration tests
- [ ] Production deployment

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

---

**Monorepo Benefits:**
✅ Single source of truth
✅ Atomic cross-service changes
✅ Simplified dependency management
✅ Unified CI/CD
✅ Better developer experience
