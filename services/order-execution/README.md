# Order Execution Service

Order lifecycle management and exchange communication service.

**Language**: Go 1.21+  
**Development Plan**: `../../docs/service-plans/02-order-execution-service.md`

## Quick Start
```bash
go build -o bin/order-execution ./cmd/server
./bin/order-execution
```

## Configuration
Copy `config.example.yaml` to `config.yaml`

## Testing
```bash
go test ./...
```

## Metrics
http://localhost:9091/metrics
