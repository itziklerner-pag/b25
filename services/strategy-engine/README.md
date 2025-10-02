# Strategy Engine Service

Trading strategy execution and signal generation service.

**Language**: Go 1.21+ (core) + Python 3.11+ (plugins)  
**Development Plan**: `../../docs/service-plans/03-strategy-engine-service.md`

## Quick Start
```bash
go build -o bin/strategy-engine ./cmd/server
./bin/strategy-engine
```

## Configuration
Copy `config.example.yaml` to `config.yaml`

## Testing
```bash
go test ./...
```

## Metrics
http://localhost:9092/metrics
