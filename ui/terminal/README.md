# Terminal UI (TUI)

Real-time terminal-based user interface for HFT trading.

**Language**: Rust 1.75+ with ratatui  
**Development Plan**: `../../docs/service-plans/09-terminal-ui-service.md`

## Quick Start
```bash
cargo run --release
```

## Features
- 6-panel layout (positions, orders, fills, orderbook, signals, alerts)
- 100ms refresh rate
- <2% CPU usage
- Keyboard-driven navigation
- Manual trading controls

## Building
```bash
# Development
cargo build

# Production (optimized)
cargo build --release

# Cross-platform
cargo build --release --target x86_64-unknown-linux-gnu
```

## Configuration
Copy `config.example.yaml` to `config.yaml`

## Testing
```bash
cargo test
```

## Usage
```bash
./target/release/terminal-ui

# With config file
./target/release/terminal-ui --config config.yaml
```
