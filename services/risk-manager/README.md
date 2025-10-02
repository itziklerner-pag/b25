# Risk Manager Service

Risk management, limit enforcement, and emergency controls service.

**Language**: Go 1.21+  
**Development Plan**: `../../docs/service-plans/06-risk-manager-service.md`

## Quick Start
```bash
go build -o bin/risk-manager ./cmd/server
./bin/risk-manager
```

## Configuration
Copy `config.example.yaml` to `config.yaml`

## Testing
```bash
go test ./...
```

## Metrics
http://localhost:9095/metrics
