# Configuration Service

Centralized configuration management with versioning and hot-reload.

**Language**: Go 1.21+  
**Development Plan**: `../../docs/service-plans/07-configuration-service.md`

## Quick Start
```bash
go build -o bin/configuration ./cmd/server
./bin/configuration
```

## Configuration
Copy `config.example.yaml` to `config.yaml`

## Testing
```bash
go test ./...
```

## Metrics
http://localhost:9096/metrics
