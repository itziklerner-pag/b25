# Dashboard Server Service

WebSocket state aggregation and broadcasting service.

**Language**: Go 1.21+  
**Development Plan**: `../../docs/service-plans/05-dashboard-server-service.md`

## Quick Start
```bash
go build -o bin/dashboard-server ./cmd/server
./bin/dashboard-server
```

## Configuration
Copy `config.example.yaml` to `config.yaml`

## Testing
```bash
go test ./...
```

## Metrics
http://localhost:9094/metrics
