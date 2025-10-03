# B25 Terminal UI (TUI)

Real-time terminal-based user interface for B25 HFT trading system built with Rust and ratatui.

## Features

- **Real-time Updates**: WebSocket connection to Dashboard Server with 100ms refresh rate
- **Multi-panel Layout**:
  - Positions panel with real-time P&L tracking
  - Active orders panel with status indicators
  - Order book visualization with depth bars
  - Recent fills/trades history
  - AI signals panel (placeholder for future integration)
  - System alerts and notifications
- **Keyboard Navigation**: Vim-style keybindings for efficient navigation
- **Manual Trading Controls**: Place and cancel orders directly from the terminal
- **Color-coded Display**: Green for profit, red for loss, cyan for buy, magenta for sell
- **Auto-reconnect**: Automatic reconnection with exponential backoff on disconnect
- **Performance**: <2% CPU usage idle, <5% under load, <50MB memory

## Quick Start

```bash
# Build and run
cargo run --release

# With custom config
cargo run --release -- --config config.yaml

# With custom WebSocket URL
cargo run --release -- --url ws://localhost:8080/ws
```

## Installation

### Prerequisites

- Rust 1.75+ (stable channel)
- Terminal with 256 color support
- UTF-8 terminal encoding
- Minimum terminal size: 80x24 (recommended: 200x50)

### Build from Source

```bash
# Development build
cargo build

# Production build (optimized)
cargo build --release

# The binary will be at:
# ./target/release/b25-terminal-ui
```

### Cross-platform Builds

```bash
# Install cross-compilation tool
cargo install cross

# Linux x86_64
cross build --release --target x86_64-unknown-linux-gnu

# Linux ARM64 (AWS Graviton, Raspberry Pi)
cross build --release --target aarch64-unknown-linux-gnu

# macOS x86_64
cross build --release --target x86_64-apple-darwin

# macOS ARM64 (M1/M2)
cross build --release --target aarch64-apple-darwin
```

## Configuration

Copy the example configuration file:

```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml` to customize settings:

```yaml
connection:
  dashboard_url: "ws://localhost:8080/ws"
  reconnect_interval_ms: 1000
  max_reconnect_interval_ms: 30000

ui:
  refresh_rate_ms: 100
  color_scheme: "default"
  stale_data_threshold_s: 5

panels:
  default_symbol: "BTCUSDT"
  max_fills_display: 50
  max_signals_display: 20
  orderbook_depth_levels: 10
```

See [config.example.yaml](/home/mm/dev/b25/ui/terminal/config.example.yaml) for all available options.

## Usage

### Keyboard Shortcuts

#### Global Commands

| Key | Action |
|-----|--------|
| `q`, `Ctrl+C` | Quit application |
| `?` | Toggle help screen |
| `r` | Reload configuration |
| `Tab` | Next panel |
| `Shift+Tab` | Previous panel |
| `:` | Enter command mode |

#### Navigation

| Key | Action |
|-----|--------|
| `j` / `Down` | Scroll down |
| `k` / `Up` | Scroll up |

#### Panel-Specific (Orders)

| Key | Action |
|-----|--------|
| `c` | Cancel selected order |
| `C` | Cancel all orders |

#### Panel-Specific (Positions)

| Key | Action |
|-----|--------|
| `x` | Close selected position |
| `X` | Close all positions |

### Command Mode

Press `:` to enter command mode, then type one of the following commands:

#### Place Orders

```
:buy <symbol> <size> <price>     # Place limit buy order
:sell <symbol> <size> <price>    # Place limit sell order
:market buy <symbol> <size>      # Place market buy order
:market sell <symbol> <size>     # Place market sell order
```

Examples:
```
:buy BTCUSDT 0.5 42000
:sell ETHUSDT 2.0 2200
:market buy BTCUSDT 0.1
```

#### Cancel Orders

```
:cancel <order_id>    # Cancel specific order
```

#### Close Positions

```
:close <symbol>       # Close position for symbol
```

Press `Enter` to execute the command, or `Esc` to cancel.

## Panel Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│ B25 Trading System │ WS: ● Connected (10ms) │ 12:34:56 UTC         │ Status bar
├──────────────────────────────┬──────────────────────────────────────┤
│ ┌─ POSITIONS ──────────────┐ │ ┌─ ORDER BOOK (BTCUSDT) ──────────┐ │
│ │ Symbol  Size  Entry  P&L │ │ │  Bids   │ Price │  Asks         │ │
│ │ BTCUSDT +0.5  42000 +250 │ │ │ 2.5 ███ │42505.0│               │ │
│ └──────────────────────────┘ │ │         │       │ ███ 1.9       │ │
│                              │ │ Spread: 1.0 (0.002%)             │ │
│ ┌─ ACTIVE ORDERS ──────────┐ │ └──────────────────────────────────┘ │
│ │ ID    Symbol  Side Price │ │                                      │
│ │ 12345 BTCUSDT BUY  42400 │ │ ┌─ AI SIGNALS ───────────────────┐ │
│ └──────────────────────────┘ │ │ Strategy  Symbol  Strength      │ │
│                              │ │ MeanRev   BTCUSDT ████████      │ │
│ ┌─ RECENT FILLS ───────────┐ │ └─────────────────────────────────┘ │
│ │ Time  Symbol  Side Price │ │                                      │
│ │ 10:32 BTCUSDT BUY  42480 │ │                                      │
│ └──────────────────────────┘ │                                      │
├──────────────────────────────┴──────────────────────────────────────┤
│ ┌─ ALERTS ────────────────────────────────────────────────────────┐ │
│ │ [WARN] 10:36 Position size exceeds 50% of limit                │ │
│ │ [INFO] 10:35 Strategy 'MeanRevert' generated LONG signal       │ │
│ └─────────────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────────────┤
│ ? Help │ : Command │ q Quit │ Tab Switch │ c Cancel Order          │ Help bar
└─────────────────────────────────────────────────────────────────────┘
```

## Color Scheme

| Element | Color | Meaning |
|---------|-------|---------|
| Green | P&L Positive | Profitable position/trade |
| Red | P&L Negative | Loss position/trade |
| Cyan | Buy Orders | Buy side orders |
| Magenta | Sell Orders | Sell side orders |
| Yellow | Warnings | Warning alerts or stale data |
| White | Info | Normal text |
| Dark Gray | Borders | Unfocused panels |
| Cyan | Focused Border | Currently focused panel |

## Connection Status Indicators

| Symbol | Status | Description |
|--------|--------|-------------|
| `●` Green | Connected | Connected and receiving updates |
| `◐` Yellow | Connecting | Attempting to connect |
| `○` Gray | Disconnected | No connection |
| `✖` Red | Error | Connection error |
| `⚠` Yellow | Stale Data | No updates for >5 seconds |

## Development

### Project Structure

```
ui/terminal/
├── Cargo.toml              # Dependencies and build config
├── config.example.yaml     # Example configuration
├── src/
│   ├── main.rs            # Application entry point
│   ├── config.rs          # Configuration loading
│   ├── state.rs           # Application state management
│   ├── types.rs           # Type definitions
│   ├── websocket.rs       # WebSocket client
│   ├── keyboard.rs        # Keyboard event handling
│   └── ui/
│       ├── mod.rs         # UI module
│       ├── theme.rs       # Color scheme
│       ├── status.rs      # Status bar widget
│       ├── help.rs        # Help overlay
│       ├── positions.rs   # Positions panel
│       ├── orders.rs      # Orders panel
│       ├── fills.rs       # Fills panel
│       ├── orderbook.rs   # Order book panel
│       ├── signals.rs     # AI signals panel
│       └── alerts.rs      # Alerts panel
└── README.md
```

### Running in Development

```bash
# Run with debug logging
RUST_LOG=debug cargo run

# Run with hot reload (requires cargo-watch)
cargo install cargo-watch
cargo watch -x run

# Run tests
cargo test

# Run benchmarks
cargo bench
```

### Testing

```bash
# Unit tests
cargo test

# Integration tests
cargo test --test '*'

# Test with coverage
cargo install cargo-tarpaulin
cargo tarpaulin --out Html
```

## Performance Targets

- **CPU Usage**: <2% idle, <5% under load
- **Memory Usage**: <50 MB
- **Render Time**: <10ms per frame (100 FPS capability)
- **WebSocket Reconnection**: <1s on disconnect
- **Input Latency**: <50ms

## Troubleshooting

### Terminal Display Issues

If you see garbled characters or incorrect colors:

```bash
# Check terminal color support
echo $TERM

# Should be xterm-256color or similar
# If not, set it:
export TERM=xterm-256color
```

### WebSocket Connection Fails

1. Verify Dashboard Server is running on the configured port
2. Check firewall rules
3. Review logs with `RUST_LOG=debug`

### High CPU Usage

1. Increase `refresh_rate_ms` in config (e.g., to 200ms)
2. Enable `enable_dirty_flag_optimization` in config
3. Reduce `orderbook_depth_levels` in config

## Docker Deployment

```dockerfile
# Build in Docker
docker build -t b25-terminal-ui .

# Run with interactive terminal
docker run -it --rm \
  -e DASHBOARD_URL=ws://dashboard:8080/ws \
  b25-terminal-ui
```

## License

MIT

## Related Documentation

- [Development Plan](/home/mm/dev/b25/docs/service-plans/09-terminal-ui-service.md)
- [ratatui Documentation](https://docs.rs/ratatui)
- [crossterm Documentation](https://docs.rs/crossterm)
