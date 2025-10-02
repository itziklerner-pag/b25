# Account Monitor Service

Balance tracking, P&L calculation, and position reconciliation service.

**Language**: Go 1.21+  
**Development Plan**: `../../docs/service-plans/04-account-monitor-service.md`

## Quick Start
```bash
go build -o bin/account-monitor ./cmd/server
./bin/account-monitor
```

## Configuration
Copy `config.example.yaml` to `config.yaml`

## Testing
```bash
go test ./...
```

## Metrics
http://localhost:9093/metrics
