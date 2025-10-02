# Shared Code and Schemas

Shared libraries, protobuf definitions, and data schemas used across services.

## Structure

```
shared/
├── proto/       # Protobuf definitions for gRPC
├── schemas/     # Data schemas (JSON Schema, Avro, etc.)
└── lib/         # Shared libraries per language
```

## Proto (Protobuf Definitions)

Shared gRPC service definitions and message types.

### Usage

```bash
# Generate code for all languages
./scripts/generate-proto.sh

# Or per language
protoc --go_out=. --go-grpc_out=. proto/*.proto
protoc --python_out=. --grpc_python_out=. proto/*.proto
```

### Files

- `market_data.proto` - Market data messages
- `orders.proto` - Order and fill messages
- `account.proto` - Account and position messages
- `config.proto` - Configuration messages
- `common.proto` - Common types (timestamp, decimal, etc.)

## Schemas (Data Schemas)

JSON Schema and other schema definitions for validation and documentation.

### Files

- `order_request.schema.json` - Order request validation
- `strategy_config.schema.json` - Strategy configuration
- `risk_limits.schema.json` - Risk limit definitions

## Lib (Shared Libraries)

Language-specific shared libraries.

### Structure

```
lib/
├── go/
│   ├── types/       # Common Go types
│   ├── utils/       # Utility functions
│   └── metrics/     # Metrics helpers
├── rust/
│   └── common/      # Common Rust crate
└── python/
    └── b25common/   # Python package
```

### Usage

#### Go
```go
import "github.com/yourorg/b25/shared/lib/go/types"
```

#### Rust
```toml
[dependencies]
b25-common = { path = "../../shared/lib/rust/common" }
```

#### Python
```python
from b25common import types
```

## Development

When adding new shared code:

1. Add to appropriate directory (`proto/`, `schemas/`, or `lib/`)
2. Update this README
3. Regenerate code if needed (`./scripts/generate-proto.sh`)
4. Version changes appropriately
5. Update dependent services

## Versioning

Shared code follows semantic versioning:
- Major: Breaking changes to APIs/schemas
- Minor: New features, backward compatible
- Patch: Bug fixes

## Testing

Each shared library has its own tests:

```bash
# Go
cd lib/go && go test ./...

# Rust
cd lib/rust/common && cargo test

# Python
cd lib/python && pytest
```
