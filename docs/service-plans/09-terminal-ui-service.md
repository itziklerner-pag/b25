# Terminal UI (TUI) Service - Development Plan

**Service:** Terminal UI (TUI) Service
**Version:** 1.0
**Last Updated:** 2025-10-02
**Status:** Planning Phase

---

## 1. Technology Stack Recommendation

### Core TUI Framework

**Recommended: Rust with ratatui**

**Rationale:**
- Ultra-low CPU usage (<2% idle, <5% under load)
- Excellent WebSocket client libraries (tokio-tungstenite)
- Zero-cost abstractions for efficient rendering
- Compiled binary with no runtime dependencies
- Strong type safety for complex UI state management

**Alternative Options:**
- **Go + Bubbletea:** Easier development, slightly higher memory footprint
- **Python + Textual:** Rapid prototyping, requires Python runtime
- **Go + tview:** Simple but less flexible for complex layouts

### Technology Stack Components

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **TUI Framework** | ratatui 0.26+ | Terminal rendering and layout |
| **Async Runtime** | tokio 1.35+ | Async WebSocket and event loop |
| **WebSocket Client** | tokio-tungstenite 0.21+ | Dashboard server connection |
| **Serialization** | serde + serde_json | Message deserialization |
| **Event Handling** | crossterm 0.27+ | Keyboard input and terminal control |
| **State Management** | Custom with Arc<RwLock> | Thread-safe state updates |
| **Testing** | insta (snapshots) + cargo test | Rendering and logic tests |
| **Benchmarking** | criterion | Performance regression detection |

### Key Libraries

```toml
[dependencies]
ratatui = "0.26"
tokio = { version = "1.35", features = ["full"] }
tokio-tungstenite = "0.21"
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
crossterm = { version = "0.27", features = ["event-stream"] }
chrono = "0.4"
humantime = "2.1"
anyhow = "1.0"
tracing = "0.1"
tracing-subscriber = "0.3"

[dev-dependencies]
insta = "1.34"
criterion = "0.5"
mockito = "1.2"
```

---

## 2. Architecture Design

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Terminal UI Process                   │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  ┌──────────────┐    ┌──────────────┐   ┌────────────┐ │
│  │   WebSocket  │───▶│    State     │◀──│  Keyboard  │ │
│  │    Client    │    │   Manager    │   │   Handler  │ │
│  └──────────────┘    └──────────────┘   └────────────┐ │
│         │                    │                        │ │
│         │                    ▼                        │ │
│         │            ┌──────────────┐                 │ │
│         │            │   UI State   │                 │ │
│         │            │  (Arc<RwLock>)│                │ │
│         │            └──────────────┘                 │ │
│         │                    │                        │ │
│         │                    ▼                        │ │
│         │            ┌──────────────┐                 │ │
│         └───────────▶│   Renderer   │◀────────────────┘ │
│                      │  (100ms tick) │                  │
│                      └──────────────┘                  │
│                             │                          │
│                             ▼                          │
│                      ┌──────────────┐                  │
│                      │   Terminal   │                  │
│                      │    Output    │                  │
│                      └──────────────┘                  │
└─────────────────────────────────────────────────────────┘
```

### Component Responsibilities

#### WebSocket Client
- Connect to Dashboard Server (configurable endpoint)
- Auto-reconnection with exponential backoff (1s, 2s, 4s, 8s, max 30s)
- Subscribe to data streams: positions, orders, orderbook, signals, alerts
- Deserialize incoming messages into typed structs
- Handle connection state: Disconnected → Connecting → Connected → Error

#### State Manager
- Thread-safe state updates using Arc<RwLock<AppState>>
- Maintain state for all UI panels
- Handle state transitions and validation
- Timestamp tracking for stale data detection (>5s = stale warning)

#### Keyboard Handler
- Non-blocking input capture using crossterm event stream
- Global keyboard shortcuts
- Context-sensitive commands (depends on focused panel)
- Input mode switching (normal/command/insert)

#### Renderer
- 100ms tick rate (10 FPS) for smooth updates
- Selective rendering: only redraw changed panels
- Layout calculation with responsive sizing
- Color theme application
- Efficient terminal buffer management

---

## 3. Development Phases

### Phase 1: Basic TUI Framework and Layout (Week 1)
**Goal:** Running TUI with static layout and mock data

**Tasks:**
- [ ] Initialize Rust project with dependencies
- [ ] Setup crossterm terminal backend
- [ ] Implement basic ratatui layout with 6 panels
- [ ] Create panel components (empty containers)
- [ ] Add keyboard handler for quit (q/Ctrl+C)
- [ ] Implement 100ms render loop
- [ ] Add connection status bar

**Deliverables:**
- Runnable binary showing empty panel layout
- Basic keyboard navigation between panels (Tab/Shift+Tab)

**Testing:**
- Snapshot tests for layout rendering
- Terminal resize handling tests

---

### Phase 2: WebSocket Client Integration (Week 1-2)
**Goal:** Live connection to Dashboard Server

**Tasks:**
- [ ] Implement WebSocket client with tokio-tungstenite
- [ ] Add reconnection logic with exponential backoff
- [ ] Define message types and deserialization
- [ ] Implement connection state machine
- [ ] Add connection status indicators
- [ ] Handle subscription management
- [ ] Implement graceful shutdown

**Deliverables:**
- WebSocket client connecting to Dashboard Server
- Connection status display in UI
- Reconnection on disconnect

**Testing:**
- Mock WebSocket server for testing
- Reconnection scenario tests
- Message deserialization tests

---

### Phase 3: Position Panel Implementation (Week 2)
**Goal:** Display real-time positions

**Tasks:**
- [ ] Define Position state structure
- [ ] Implement position data updates from WebSocket
- [ ] Create position table widget
- [ ] Add color coding (green/red for P&L)
- [ ] Implement sorting (by symbol, size, P&L)
- [ ] Add entry price and current price display
- [ ] Calculate and display unrealized P&L

**Panel Data:**
```
┌─ POSITIONS ─────────────────────────────────────────┐
│ Symbol   Size    Entry     Current   Unreal P&L  %  │
│ BTCUSDT  +0.5    42000.0   42500.0   +250.0    +1.2%│
│ ETHUSDT  -2.0    2200.0    2180.0    +40.0     +0.9%│
└─────────────────────────────────────────────────────┘
```

**Testing:**
- Position calculation unit tests
- P&L calculation verification
- Color coding logic tests

---

### Phase 4: Order Book Panel Implementation (Week 2-3)
**Goal:** Real-time order book visualization

**Tasks:**
- [ ] Define OrderBook state structure
- [ ] Implement order book updates (full snapshot + deltas)
- [ ] Create bid/ask table widgets
- [ ] Add depth bar visualization (ASCII bars)
- [ ] Implement price level aggregation
- [ ] Add spread calculation and display
- [ ] Implement mid-price calculation
- [ ] Add order book imbalance indicator

**Panel Data:**
```
┌─ ORDER BOOK (BTCUSDT) ──────────────────────────────┐
│        Bids          │ Price   │        Asks         │
│ 2.5 ████████████     │ 42505.0 │                     │
│ 3.2 ████████████████ │ 42504.5 │                     │
│ 1.8 ███████          │ 42504.0 │                     │
│ ──────────────────── │ ─────── │ ──────────────────  │
│                      │ 42506.0 │ ████████      1.9   │
│                      │ 42506.5 │ ██████████    2.3   │
│                      │ 42507.0 │ ████████████  2.8   │
│ Spread: 1.0 (0.002%) │ Mid: 42505.5            │
└─────────────────────────────────────────────────────┘
```

**Testing:**
- Order book update logic tests
- Depth calculation tests
- Spread and mid-price calculation tests

---

### Phase 5: Orders and Fills Panels (Week 3)
**Goal:** Display active orders and recent fills

**Tasks:**
- [ ] Define Order and Fill state structures
- [ ] Implement order list widget
- [ ] Add order status color coding
- [ ] Implement fills list widget (last 50 fills)
- [ ] Add time formatting (relative time)
- [ ] Implement order filtering (active/all)
- [ ] Add order cancellation keyboard shortcut

**Orders Panel:**
```
┌─ ACTIVE ORDERS ─────────────────────────────────────┐
│ ID     Symbol   Side Type  Price    Size   Status   │
│ 12345  BTCUSDT  BUY  LIMIT 42400.0  0.5    NEW      │
│ 12346  ETHUSDT  SELL LIMIT 2210.0   1.0    PARTIALLY│
└─────────────────────────────────────────────────────┘
```

**Fills Panel:**
```
┌─ RECENT FILLS ──────────────────────────────────────┐
│ Time    Symbol   Side Price    Size   Fee    P&L    │
│ 10:32   BTCUSDT  BUY  42480.0  0.3    0.12   +45.0  │
│ 10:28   ETHUSDT  SELL 2205.0   0.5    0.22   +12.5  │
│ 10:15   BTCUSDT  SELL 42450.0  0.2    0.08   -8.0   │
└─────────────────────────────────────────────────────┘
```

**Testing:**
- Order state transition tests
- Fill recording tests
- Time formatting tests

---

### Phase 6: AI Signals and Alerts Panels (Week 3-4)
**Goal:** Display AI-generated signals and system alerts

**Tasks:**
- [ ] Define Signal and Alert state structures
- [ ] Implement signals list widget
- [ ] Add signal strength visualization (bars/colors)
- [ ] Implement alerts panel with priority coloring
- [ ] Add alert dismissal functionality
- [ ] Implement signal filtering by strategy
- [ ] Add alert history (last 100 alerts)

**AI Signals Panel:**
```
┌─ AI SIGNALS ────────────────────────────────────────┐
│ Time  Strategy    Symbol   Signal  Strength  Price  │
│ 10:35 MeanRevert  BTCUSDT  LONG    ████████  42500  │
│ 10:33 Momentum    ETHUSDT  SHORT   ████      2200   │
│ 10:30 Arbitrage   BTCUSDT  NEUTRAL ██        42480  │
└─────────────────────────────────────────────────────┘
```

**Alerts Panel:**
```
┌─ ALERTS ────────────────────────────────────────────┐
│ [WARN]  10:36  Position size exceeds 50% of limit   │
│ [INFO]  10:35  Strategy 'MeanRevert' generated LONG │
│ [ERROR] 10:32  Order rejected: Insufficient balance │
└─────────────────────────────────────────────────────┘
```

**Testing:**
- Signal strength calculation tests
- Alert priority sorting tests
- Alert dismissal logic tests

---

### Phase 7: Keyboard Controls and Manual Trading (Week 4)
**Goal:** Interactive trading functionality

**Tasks:**
- [ ] Implement command mode (press ':' like vim)
- [ ] Add order placement commands
- [ ] Implement order cancellation (select + 'c')
- [ ] Add position close commands
- [ ] Implement symbol switching ('/' for search)
- [ ] Add panel focus switching (Tab/Shift+Tab)
- [ ] Implement help overlay ('?' key)
- [ ] Add configuration reload ('r' key)

**Keyboard Shortcuts:**
```
Global:
  q, Ctrl+C    : Quit application
  ?            : Show help overlay
  r            : Reload configuration
  Tab          : Next panel
  Shift+Tab    : Previous panel
  /            : Symbol search
  :            : Command mode

Command Mode:
  :buy <symbol> <size> <price>   : Place limit buy order
  :sell <symbol> <size> <price>  : Place limit sell order
  :market <side> <symbol> <size> : Place market order
  :cancel <order_id>             : Cancel order
  :close <symbol>                : Close position

Panel-Specific:
  Orders Panel:
    c : Cancel selected order
    C : Cancel all orders

  Positions Panel:
    x : Close selected position
    X : Close all positions
```

**Testing:**
- Keyboard input parsing tests
- Command execution tests
- Panel focus management tests

---

### Phase 8: Performance Optimization and Testing (Week 4-5)
**Goal:** Production-ready TUI with comprehensive testing

**Tasks:**
- [ ] Implement selective rendering (dirty flag system)
- [ ] Optimize data structure updates
- [ ] Add rendering performance benchmarks
- [ ] Implement CPU usage monitoring
- [ ] Add memory usage tracking
- [ ] Create comprehensive test suite
- [ ] Add snapshot tests for all panels
- [ ] Implement load testing (high-frequency updates)
- [ ] Add error recovery mechanisms
- [ ] Create user documentation

**Performance Targets:**
- CPU usage: <2% idle, <5% under load
- Memory usage: <50 MB
- Render time: <10ms per frame (100 FPS capability)
- WebSocket reconnection: <1s on disconnect
- Input latency: <50ms

**Testing:**
- Performance regression tests with criterion
- Memory leak detection
- WebSocket stress testing (1000 msgs/sec)
- Terminal resize edge cases
- Long-running stability tests (24h+)

---

## 4. Implementation Details

### Panel Layout Design

```
Terminal Size: 200x50 (typical full-screen terminal)

┌─────────────────────────────────────────────────────────────────────────┐
│ B25 Trading System  │  WS: Connected (10ms)  │  12:34:56 UTC           │ ← Status bar (1 row)
├───────────────────────────────────┬─────────────────────────────────────┤
│                                   │                                     │
│  ┌─ POSITIONS ─────────────────┐ │  ┌─ ORDER BOOK (BTCUSDT) ────────┐ │
│  │ Symbol   Size   Entry   P&L  │ │  │   Bids    │ Price │   Asks    │ │
│  │ BTCUSDT  +0.5   42000  +250  │ │  │ 2.5 ████  │42505.0│           │ │
│  │ ETHUSDT  -2.0   2200   +40   │ │  │ 3.2 █████ │42504.5│           │ │
│  │                              │ │  │ 1.8 ███   │42504.0│           │ │
│  └──────────────────────────────┘ │  │ ────────  │───────│  ──────── │ │
│                                   │  │           │42506.0│ ████  1.9 │ │
│  ┌─ ACTIVE ORDERS ──────────────┐ │  │           │42506.5│ █████ 2.3 │ │
│  │ ID    Symbol  Side  Price     │ │  │ Spread: 1.0 (0.002%)        │ │
│  │ 12345 BTCUSDT BUY   42400.0   │ │  └─────────────────────────────┘ │
│  │ 12346 ETHUSDT SELL  2210.0    │ │                                   │
│  └──────────────────────────────┘ │  ┌─ AI SIGNALS ─────────────────┐ │
│                                   │  │ Strategy    Symbol  Strength  │ │
│  ┌─ RECENT FILLS ───────────────┐ │  │ MeanRevert  BTCUSDT ████████  │ │
│  │ Time  Symbol  Side  Price     │ │  │ Momentum    ETHUSDT ████      │ │
│  │ 10:32 BTCUSDT BUY   42480.0   │ │  └─────────────────────────────┘ │
│  │ 10:28 ETHUSDT SELL  2205.0    │ │                                   │
│  └──────────────────────────────┘ │                                   │
│                                   │                                     │
├───────────────────────────────────┴─────────────────────────────────────┤
│ ┌─ ALERTS ──────────────────────────────────────────────────────────┐  │
│ │ [WARN]  10:36  Position size exceeds 50% of limit                 │  │
│ │ [INFO]  10:35  Strategy 'MeanRevert' generated LONG signal        │  │
│ └───────────────────────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────────────────────┤
│ ? Help │ : Command │ q Quit │ Tab Switch Panel │ / Search             │ ← Help bar (1 row)
└─────────────────────────────────────────────────────────────────────────┘

Left Column:  40% width (positions, orders, fills)
Right Column: 60% width (orderbook, signals)
Alerts:       Full width, 3 rows
Status:       Full width, 1 row
Help:         Full width, 1 row
```

### Responsive Layout Breakpoints

```rust
// Layout adapts to terminal size
fn calculate_layout(terminal_size: (u16, u16)) -> Layout {
    match terminal_size {
        // Small terminal: vertical stack
        (_, height) if height < 30 => Layout::Vertical,

        // Narrow terminal: single column
        (width, _) if width < 100 => Layout::SingleColumn,

        // Standard layout
        (width, height) if width >= 100 && height >= 30 => Layout::TwoColumn,

        // Wide terminal: three columns
        (width, _) if width >= 250 => Layout::ThreeColumn,

        _ => Layout::TwoColumn,
    }
}
```

---

### WebSocket Message Handling

```rust
use serde::{Deserialize, Serialize};
use tokio_tungstenite::{connect_async, tungstenite::Message};
use anyhow::Result;

// Message types from Dashboard Server
#[derive(Debug, Deserialize)]
#[serde(tag = "type")]
enum DashboardMessage {
    #[serde(rename = "positions")]
    Positions { data: Vec<Position> },

    #[serde(rename = "orders")]
    Orders { data: Vec<Order> },

    #[serde(rename = "orderbook")]
    OrderBook { symbol: String, data: OrderBookData },

    #[serde(rename = "fills")]
    Fills { data: Vec<Fill> },

    #[serde(rename = "signals")]
    Signals { data: Vec<Signal> },

    #[serde(rename = "alerts")]
    Alerts { data: Vec<Alert> },
}

// WebSocket client with reconnection
pub struct WsClient {
    url: String,
    state_tx: mpsc::Sender<StateUpdate>,
    reconnect_interval: Duration,
}

impl WsClient {
    pub async fn connect_with_retry(&self) -> Result<()> {
        let mut backoff = 1;

        loop {
            match self.connect().await {
                Ok(_) => {
                    tracing::info!("WebSocket connected");
                    backoff = 1; // Reset on success
                }
                Err(e) => {
                    tracing::error!("Connection failed: {}, retrying in {}s", e, backoff);
                    tokio::time::sleep(Duration::from_secs(backoff)).await;
                    backoff = (backoff * 2).min(30); // Exponential backoff, max 30s
                }
            }
        }
    }

    async fn connect(&self) -> Result<()> {
        let (ws_stream, _) = connect_async(&self.url).await?;
        let (write, mut read) = ws_stream.split();

        // Send subscription message
        self.send_subscription(&write).await?;

        // Message processing loop
        while let Some(msg) = read.next().await {
            match msg? {
                Message::Text(text) => {
                    let dashboard_msg: DashboardMessage = serde_json::from_str(&text)?;
                    self.handle_message(dashboard_msg).await?;
                }
                Message::Close(_) => {
                    tracing::warn!("WebSocket closed by server");
                    return Err(anyhow::anyhow!("Connection closed"));
                }
                _ => {}
            }
        }

        Ok(())
    }

    async fn handle_message(&self, msg: DashboardMessage) -> Result<()> {
        let update = match msg {
            DashboardMessage::Positions { data } => StateUpdate::Positions(data),
            DashboardMessage::Orders { data } => StateUpdate::Orders(data),
            DashboardMessage::OrderBook { symbol, data } =>
                StateUpdate::OrderBook(symbol, data),
            DashboardMessage::Fills { data } => StateUpdate::Fills(data),
            DashboardMessage::Signals { data } => StateUpdate::Signals(data),
            DashboardMessage::Alerts { data } => StateUpdate::Alerts(data),
        };

        self.state_tx.send(update).await?;
        Ok(())
    }
}
```

---

### Efficient Rendering Strategy

```rust
use ratatui::{Frame, layout::Rect};
use std::sync::Arc;
use parking_lot::RwLock;

// Dirty flag system for selective rendering
#[derive(Default)]
struct DirtyFlags {
    positions: bool,
    orders: bool,
    orderbook: bool,
    fills: bool,
    signals: bool,
    alerts: bool,
    status: bool,
}

pub struct AppState {
    positions: Vec<Position>,
    orders: Vec<Order>,
    orderbook: OrderBookData,
    fills: Vec<Fill>,
    signals: Vec<Signal>,
    alerts: Vec<Alert>,
    connection_status: ConnectionStatus,
    dirty: DirtyFlags,
}

impl AppState {
    pub fn update_positions(&mut self, positions: Vec<Position>) {
        if self.positions != positions {
            self.positions = positions;
            self.dirty.positions = true;
        }
    }

    pub fn clear_dirty(&mut self) {
        self.dirty = DirtyFlags::default();
    }
}

// Renderer with selective updates
pub fn render(f: &mut Frame, state: &Arc<RwLock<AppState>>) {
    let state = state.read();

    // Only redraw panels with dirty flag set
    if state.dirty.positions {
        render_positions_panel(f, area, &state.positions);
    }

    if state.dirty.orderbook {
        render_orderbook_panel(f, area, &state.orderbook);
    }

    // ... render other dirty panels
}

// Render loop with 100ms tick
pub async fn render_loop(
    terminal: &mut Terminal<CrosstermBackend<io::Stdout>>,
    state: Arc<RwLock<AppState>>,
) -> Result<()> {
    let mut interval = tokio::time::interval(Duration::from_millis(100));

    loop {
        interval.tick().await;

        terminal.draw(|f| {
            render(f, &state);
        })?;

        state.write().clear_dirty();
    }
}
```

---

### Keyboard Shortcuts Implementation

```rust
use crossterm::event::{Event, KeyCode, KeyModifiers};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
enum InputMode {
    Normal,
    Command,
}

pub struct KeyboardHandler {
    mode: InputMode,
    command_buffer: String,
    focused_panel: Panel,
}

impl KeyboardHandler {
    pub fn handle_key(&mut self, event: Event) -> Option<Action> {
        match event {
            Event::Key(key) => match self.mode {
                InputMode::Normal => match key.code {
                    KeyCode::Char('q') => Some(Action::Quit),
                    KeyCode::Char('c') if key.modifiers.contains(KeyModifiers::CONTROL) =>
                        Some(Action::Quit),
                    KeyCode::Char('?') => Some(Action::ShowHelp),
                    KeyCode::Char('r') => Some(Action::ReloadConfig),
                    KeyCode::Tab => Some(Action::NextPanel),
                    KeyCode::BackTab => Some(Action::PrevPanel),
                    KeyCode::Char('/') => Some(Action::StartSearch),
                    KeyCode::Char(':') => {
                        self.mode = InputMode::Command;
                        self.command_buffer.clear();
                        Some(Action::EnterCommandMode)
                    }
                    KeyCode::Char('c') => self.handle_panel_shortcut('c'),
                    KeyCode::Char('x') => self.handle_panel_shortcut('x'),
                    _ => None,
                },
                InputMode::Command => match key.code {
                    KeyCode::Enter => {
                        let cmd = self.command_buffer.clone();
                        self.command_buffer.clear();
                        self.mode = InputMode::Normal;
                        Some(Action::ExecuteCommand(cmd))
                    }
                    KeyCode::Esc => {
                        self.command_buffer.clear();
                        self.mode = InputMode::Normal;
                        Some(Action::CancelCommand)
                    }
                    KeyCode::Char(c) => {
                        self.command_buffer.push(c);
                        None
                    }
                    KeyCode::Backspace => {
                        self.command_buffer.pop();
                        None
                    }
                    _ => None,
                },
            },
            _ => None,
        }
    }

    fn handle_panel_shortcut(&self, key: char) -> Option<Action> {
        match (self.focused_panel, key) {
            (Panel::Orders, 'c') => Some(Action::CancelSelectedOrder),
            (Panel::Orders, 'C') => Some(Action::CancelAllOrders),
            (Panel::Positions, 'x') => Some(Action::CloseSelectedPosition),
            (Panel::Positions, 'X') => Some(Action::CloseAllPositions),
            _ => None,
        }
    }
}
```

---

### Color Scheme and Styling

```rust
use ratatui::style::{Color, Modifier, Style};

pub struct ColorScheme {
    // Status colors
    pub connected: Color,
    pub disconnected: Color,
    pub warning: Color,
    pub error: Color,

    // P&L colors
    pub profit: Color,
    pub loss: Color,
    pub neutral: Color,

    // Order colors
    pub buy: Color,
    pub sell: Color,

    // UI elements
    pub border: Color,
    pub border_focused: Color,
    pub text: Color,
    pub text_dim: Color,
    pub highlight: Color,
}

impl Default for ColorScheme {
    fn default() -> Self {
        Self {
            connected: Color::Green,
            disconnected: Color::Red,
            warning: Color::Yellow,
            error: Color::Red,

            profit: Color::Green,
            loss: Color::Red,
            neutral: Color::Gray,

            buy: Color::Cyan,
            sell: Color::Magenta,

            border: Color::DarkGray,
            border_focused: Color::Cyan,
            text: Color::White,
            text_dim: Color::DarkGray,
            highlight: Color::Yellow,
        }
    }
}

// Style helpers
pub fn profit_style(value: f64) -> Style {
    let colors = ColorScheme::default();
    if value > 0.0 {
        Style::default().fg(colors.profit)
    } else if value < 0.0 {
        Style::default().fg(colors.loss)
    } else {
        Style::default().fg(colors.neutral)
    }
}

pub fn order_side_style(side: &str) -> Style {
    let colors = ColorScheme::default();
    match side {
        "BUY" => Style::default().fg(colors.buy).add_modifier(Modifier::BOLD),
        "SELL" => Style::default().fg(colors.sell).add_modifier(Modifier::BOLD),
        _ => Style::default(),
    }
}
```

---

## 5. UI Layout Specification

### Screen Layout Details

**Status Bar (Top - 1 row):**
- Left: Application name and version
- Center: WebSocket connection status and latency
- Right: Current time (UTC)

**Main Area (Dynamic height):**

**Left Column (40% width):**
1. **Positions Panel (30% height)**
   - Columns: Symbol, Size, Entry Price, Current Price, Unrealized P&L, P&L %
   - Sort by: P&L (default), Symbol, Size
   - Color: Green for profit, Red for loss

2. **Active Orders Panel (35% height)**
   - Columns: Order ID, Symbol, Side, Type, Price, Size, Status
   - Color: Cyan for BUY, Magenta for SELL
   - Status indicators: NEW, PARTIALLY_FILLED, FILLED, CANCELED

3. **Recent Fills Panel (35% height)**
   - Columns: Time, Symbol, Side, Price, Size, Fee, P&L
   - Shows last 50 fills
   - Time format: HH:MM or relative (e.g., "2m ago")

**Right Column (60% width):**
1. **Order Book Panel (60% height)**
   - Split view: Bids (left) | Price (center) | Asks (right)
   - Depth visualization: ASCII bars proportional to size
   - Highlight: Best bid/ask
   - Footer: Spread (absolute and %), Mid price
   - Symbol selector: Default from focused position

2. **AI Signals Panel (40% height)**
   - Columns: Time, Strategy Name, Symbol, Signal Type, Strength, Target Price
   - Strength visualization: ASCII bars (8 levels)
   - Color: Green for LONG, Red for SHORT, Gray for NEUTRAL
   - Shows last 20 signals

**Alerts Panel (Bottom - 3 rows):**
- Full width, scrollable
- Format: [LEVEL] Timestamp Message
- Levels: ERROR (red), WARN (yellow), INFO (white)
- Auto-dismiss after 30s for INFO, manual dismiss for WARN/ERROR

**Help Bar (Bottom - 1 row):**
- Shows context-sensitive shortcuts
- Changes based on focused panel and input mode

---

### Data Display Formats

**Numbers:**
- Prices: 2-8 decimal places (depends on symbol tick size)
- Sizes: 1-6 decimal places (depends on symbol lot size)
- P&L: 2 decimal places with currency symbol
- Percentages: 2 decimal places with % sign

**Time:**
- Absolute: HH:MM:SS UTC
- Relative: "Xs ago", "Xm ago", "Xh ago" for recent events (<24h)

**Status Indicators:**
- Connection: ● Connected (green), ● Connecting (yellow), ● Disconnected (red)
- Stale data: ⚠ Warning if no update for >5s

---

### Color Coding Scheme

| Element | Color | Meaning |
|---------|-------|---------|
| **P&L Positive** | Green | Profitable position/trade |
| **P&L Negative** | Red | Loss position/trade |
| **Buy Orders** | Cyan | Buy side orders |
| **Sell Orders** | Magenta | Sell side orders |
| **Long Signals** | Green | AI suggests long position |
| **Short Signals** | Red | AI suggests short position |
| **Error Alerts** | Red | Critical errors |
| **Warning Alerts** | Yellow | Warnings |
| **Info Alerts** | White | Informational |
| **Focused Border** | Cyan | Currently focused panel |
| **Normal Border** | Dark Gray | Unfocused panels |
| **Stale Data** | Yellow | Data older than 5s |

---

## 6. Testing Strategy

### Unit Tests

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_position_pnl_calculation() {
        let position = Position {
            symbol: "BTCUSDT".to_string(),
            size: 0.5,
            entry_price: 42000.0,
            current_price: 42500.0,
            side: PositionSide::Long,
        };

        assert_eq!(position.unrealized_pnl(), 250.0);
        assert_eq!(position.pnl_percent(), 1.19); // (500 / 42000) * 100
    }

    #[test]
    fn test_orderbook_spread_calculation() {
        let orderbook = OrderBookData {
            bids: vec![(42500.0, 2.5), (42499.5, 3.0)],
            asks: vec![(42501.0, 1.8), (42501.5, 2.2)],
        };

        assert_eq!(orderbook.spread(), 1.0);
        assert_eq!(orderbook.mid_price(), 42500.5);
        assert_eq!(orderbook.spread_percent(), 0.00235); // (1.0 / 42500.5) * 100
    }
}
```

---

### Snapshot Tests (Rendering)

```rust
use insta::assert_snapshot;

#[test]
fn test_positions_panel_rendering() {
    let positions = vec![
        Position {
            symbol: "BTCUSDT".to_string(),
            size: 0.5,
            entry_price: 42000.0,
            current_price: 42500.0,
            side: PositionSide::Long,
        },
    ];

    let rendered = render_positions_to_string(&positions, (50, 10));
    assert_snapshot!(rendered);
}

#[test]
fn test_orderbook_panel_rendering() {
    let orderbook = OrderBookData {
        bids: vec![(42500.0, 2.5), (42499.5, 3.0), (42499.0, 1.8)],
        asks: vec![(42501.0, 1.8), (42501.5, 2.2), (42502.0, 2.8)],
    };

    let rendered = render_orderbook_to_string(&orderbook, (60, 15));
    assert_snapshot!(rendered);
}
```

---

### WebSocket Reconnection Tests

```rust
use mockito::{mock, Server};
use tokio::time::{sleep, Duration};

#[tokio::test]
async fn test_websocket_reconnection() {
    let mut server = Server::new_async().await;
    let url = server.url();

    // First connection succeeds
    let _m1 = server.mock("GET", "/ws")
        .with_status(101)
        .create_async()
        .await;

    let client = WsClient::new(&url);
    client.connect().await.expect("First connection failed");

    // Simulate disconnect
    drop(_m1);
    sleep(Duration::from_secs(1)).await;

    // Second connection after reconnect
    let _m2 = server.mock("GET", "/ws")
        .with_status(101)
        .create_async()
        .await;

    // Verify reconnection happened
    assert!(client.is_connected());
}

#[tokio::test]
async fn test_exponential_backoff() {
    let server = Server::new_async().await;
    let url = server.url();

    let client = WsClient::new(&url);

    // First retry: 1s
    let start = Instant::now();
    client.connect_with_retry().await;
    assert!(start.elapsed() >= Duration::from_secs(1));
    assert!(start.elapsed() < Duration::from_secs(2));

    // Second retry: 2s
    let start = Instant::now();
    client.connect_with_retry().await;
    assert!(start.elapsed() >= Duration::from_secs(2));
    assert!(start.elapsed() < Duration::from_secs(3));
}
```

---

### Performance Tests

```rust
use criterion::{black_box, criterion_group, criterion_main, Criterion};

fn benchmark_rendering(c: &mut Criterion) {
    let state = create_mock_app_state();

    c.bench_function("render_full_ui", |b| {
        b.iter(|| {
            render_to_buffer(black_box(&state))
        });
    });

    // Target: <10ms per frame
}

fn benchmark_orderbook_update(c: &mut Criterion) {
    let mut orderbook = OrderBookData::default();
    let update = create_mock_orderbook_update();

    c.bench_function("apply_orderbook_update", |b| {
        b.iter(|| {
            orderbook.apply_update(black_box(&update))
        });
    });

    // Target: <1ms per update
}

criterion_group!(benches, benchmark_rendering, benchmark_orderbook_update);
criterion_main!(benches);
```

---

### Input Handling Tests

```rust
#[test]
fn test_keyboard_command_parsing() {
    let handler = KeyboardHandler::new();

    // Test buy command
    let action = handler.parse_command("buy BTCUSDT 0.5 42000");
    assert_eq!(action, Some(Action::PlaceOrder {
        side: OrderSide::Buy,
        symbol: "BTCUSDT",
        size: 0.5,
        price: Some(42000.0),
    }));

    // Test cancel command
    let action = handler.parse_command("cancel 12345");
    assert_eq!(action, Some(Action::CancelOrder { order_id: 12345 }));

    // Test invalid command
    let action = handler.parse_command("invalid");
    assert_eq!(action, None);
}

#[test]
fn test_panel_focus_navigation() {
    let mut ui_state = UiState::default();

    // Start at first panel
    assert_eq!(ui_state.focused_panel, Panel::Positions);

    // Tab to next panel
    ui_state.next_panel();
    assert_eq!(ui_state.focused_panel, Panel::Orders);

    // Shift+Tab to previous panel
    ui_state.prev_panel();
    assert_eq!(ui_state.focused_panel, Panel::Positions);
}
```

---

## 7. Deployment

### Standalone Binary

```bash
# Build for current platform
cargo build --release

# Binary location
./target/release/b25-tui

# Cross-platform builds
cargo install cross

# Linux x86_64
cross build --release --target x86_64-unknown-linux-gnu

# Linux ARM64 (e.g., AWS Graviton)
cross build --release --target aarch64-unknown-linux-gnu

# macOS x86_64
cross build --release --target x86_64-apple-darwin

# macOS ARM64 (M1/M2)
cross build --release --target aarch64-apple-darwin
```

---

### Configuration File

```toml
# config.toml
[connection]
dashboard_url = "ws://localhost:8080/ws"
reconnect_interval_ms = 1000
max_reconnect_interval_ms = 30000

[ui]
refresh_rate_ms = 100
color_scheme = "default"  # or "solarized", "gruvbox"
show_milliseconds = false

[panels]
default_symbol = "BTCUSDT"
max_fills_display = 50
max_signals_display = 20
max_alerts_display = 100
orderbook_depth_levels = 10

[keyboard]
quit_keys = ["q", "Ctrl+C"]
help_key = "?"
reload_key = "r"
command_key = ":"
search_key = "/"

[performance]
enable_dirty_flag_optimization = true
max_cpu_percent = 5.0
max_memory_mb = 50

[logging]
level = "info"  # debug, info, warn, error
file = "/var/log/b25-tui.log"
```

---

### Running the TUI

```bash
# With default config
./b25-tui

# With custom config
./b25-tui --config /path/to/config.toml

# With environment variable
DASHBOARD_URL=ws://prod.example.com:8080/ws ./b25-tui

# With debug logging
RUST_LOG=debug ./b25-tui

# Docker container
docker run -it --rm \
  -e DASHBOARD_URL=ws://dashboard:8080/ws \
  b25-tui:latest
```

---

### Docker Deployment

```dockerfile
# Dockerfile
FROM rust:1.75 as builder

WORKDIR /app
COPY Cargo.toml Cargo.lock ./
COPY src ./src

RUN cargo build --release

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/target/release/b25-tui /usr/local/bin/

ENV DASHBOARD_URL=ws://dashboard:8080/ws
ENV RUST_LOG=info

ENTRYPOINT ["/usr/local/bin/b25-tui"]
```

```yaml
# docker-compose.yml
services:
  tui:
    build:
      context: ./tui
      dockerfile: Dockerfile
    environment:
      DASHBOARD_URL: ws://dashboard:8080/ws
      RUST_LOG: info
    stdin_open: true
    tty: true
    networks:
      - trading-net

networks:
  trading-net:
    external: true
```

---

## 8. Observability

### Self-Monitoring Display

**Connection Status (Status Bar):**
```
WS: ● Connected (10ms) | Last Update: 0.5s ago
```

- Connection state indicator (color-coded dot)
- Round-trip latency
- Time since last message received
- Warning if no message for >5s

**Performance Metrics (Hidden Panel - Press 'p'):**
```
┌─ PERFORMANCE ────────────────────────────────┐
│ CPU Usage:       2.3%                        │
│ Memory Usage:    32.5 MB                     │
│ Render Time:     6.2ms (avg), 12.1ms (p99)  │
│ Messages/sec:    250                         │
│ Frames/sec:      10.0                        │
│ Dropped Frames:  0                           │
└──────────────────────────────────────────────┘
```

**Connection State Indicators:**
- ● Green: Connected and receiving updates
- ● Yellow: Connected but no updates for >5s (stale)
- ● Orange: Reconnecting (with retry count)
- ● Red: Disconnected

**Panel State Indicators:**
- Normal border: Up-to-date data
- Yellow border: Stale data (>5s old)
- Red border: Error state (failed to update)

---

### Logging

```rust
use tracing::{info, warn, error, debug};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

pub fn init_logging(config: &LoggingConfig) {
    let file_appender = tracing_appender::rolling::daily(&config.log_dir, "tui.log");
    let (non_blocking, _guard) = tracing_appender::non_blocking(file_appender);

    tracing_subscriber::registry()
        .with(tracing_subscriber::EnvFilter::new(&config.level))
        .with(tracing_subscriber::fmt::layer()
            .with_writer(non_blocking)
            .with_ansi(false)
            .json())
        .init();

    info!("TUI started with config: {:?}", config);
}

// Throughout the code
debug!("WebSocket message received: {:?}", msg);
info!("Connected to dashboard server at {}", url);
warn!("No update received for 5s, marking data as stale");
error!("Failed to parse message: {}", err);
```

---

### Health Metrics Export

```rust
// Optional: Export metrics for external monitoring
use prometheus::{Counter, Gauge, Histogram, Registry};

pub struct TuiMetrics {
    ws_messages_received: Counter,
    ws_connection_status: Gauge,
    render_duration: Histogram,
    cpu_usage: Gauge,
    memory_usage: Gauge,
}

impl TuiMetrics {
    pub fn new(registry: &Registry) -> Self {
        // Register metrics
        let ws_messages_received = Counter::new(
            "tui_ws_messages_received_total",
            "Total WebSocket messages received"
        ).unwrap();

        let ws_connection_status = Gauge::new(
            "tui_ws_connection_status",
            "WebSocket connection status (1=connected, 0=disconnected)"
        ).unwrap();

        let render_duration = Histogram::with_opts(
            HistogramOpts::new(
                "tui_render_duration_seconds",
                "Time spent rendering frames"
            ).buckets(vec![0.001, 0.005, 0.010, 0.025, 0.050, 0.100])
        ).unwrap();

        registry.register(Box::new(ws_messages_received.clone())).unwrap();
        registry.register(Box::new(ws_connection_status.clone())).unwrap();
        registry.register(Box::new(render_duration.clone())).unwrap();

        Self {
            ws_messages_received,
            ws_connection_status,
            render_duration,
            cpu_usage: Gauge::new("tui_cpu_usage", "CPU usage percentage").unwrap(),
            memory_usage: Gauge::new("tui_memory_mb", "Memory usage in MB").unwrap(),
        }
    }

    pub fn record_ws_message(&self) {
        self.ws_messages_received.inc();
    }

    pub fn update_connection_status(&self, connected: bool) {
        self.ws_connection_status.set(if connected { 1.0 } else { 0.0 });
    }

    pub fn observe_render_duration(&self, duration: Duration) {
        self.render_duration.observe(duration.as_secs_f64());
    }
}
```

---

## 9. Key Component Code Examples

### Main Application Structure

```rust
use anyhow::Result;
use crossterm::{
    event::{self, DisableMouseCapture, EnableMouseCapture, Event},
    execute,
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};
use ratatui::{backend::CrosstermBackend, Terminal};
use std::sync::Arc;
use parking_lot::RwLock;
use tokio::sync::mpsc;

mod config;
mod ui;
mod ws;
mod state;
mod keyboard;

use config::Config;
use state::{AppState, StateUpdate};
use ws::WsClient;
use keyboard::KeyboardHandler;

#[tokio::main]
async fn main() -> Result<()> {
    // Load configuration
    let config = Config::load()?;

    // Initialize logging
    init_logging(&config.logging)?;

    // Setup terminal
    enable_raw_mode()?;
    let mut stdout = std::io::stdout();
    execute!(stdout, EnterAlternateScreen, EnableMouseCapture)?;
    let backend = CrosstermBackend::new(stdout);
    let mut terminal = Terminal::new(backend)?;

    // Initialize shared state
    let state = Arc::new(RwLock::new(AppState::default()));

    // Create channels
    let (state_tx, mut state_rx) = mpsc::channel::<StateUpdate>(1000);
    let (action_tx, mut action_rx) = mpsc::channel::<Action>(100);

    // Spawn WebSocket client
    let ws_client = WsClient::new(config.connection.dashboard_url.clone(), state_tx);
    let ws_handle = tokio::spawn(async move {
        ws_client.connect_with_retry().await
    });

    // Spawn keyboard handler
    let keyboard_handler = KeyboardHandler::new(action_tx.clone());
    let keyboard_handle = tokio::spawn(async move {
        keyboard_handler.run().await
    });

    // Spawn state updater
    let state_clone = state.clone();
    let state_update_handle = tokio::spawn(async move {
        while let Some(update) = state_rx.recv().await {
            state_clone.write().apply_update(update);
        }
    });

    // Main render loop
    let render_handle = tokio::spawn({
        let state = state.clone();
        async move {
            let mut interval = tokio::time::interval(
                std::time::Duration::from_millis(100)
            );

            loop {
                interval.tick().await;

                if let Err(e) = terminal.draw(|f| {
                    ui::render(f, &state);
                }) {
                    tracing::error!("Render error: {}", e);
                }
            }
        }
    });

    // Action handler loop
    loop {
        if let Some(action) = action_rx.recv().await {
            match action {
                Action::Quit => break,
                action => handle_action(action, &state).await?,
            }
        }
    }

    // Cleanup
    ws_handle.abort();
    keyboard_handle.abort();
    state_update_handle.abort();
    render_handle.abort();

    disable_raw_mode()?;
    execute!(
        terminal.backend_mut(),
        LeaveAlternateScreen,
        DisableMouseCapture
    )?;
    terminal.show_cursor()?;

    Ok(())
}

async fn handle_action(action: Action, state: &Arc<RwLock<AppState>>) -> Result<()> {
    match action {
        Action::NextPanel => state.write().next_panel(),
        Action::PrevPanel => state.write().prev_panel(),
        Action::CancelOrder { order_id } => {
            // Send cancel request via RPC/WebSocket
            tracing::info!("Canceling order {}", order_id);
        }
        Action::ClosePosition { symbol } => {
            // Send close position request
            tracing::info!("Closing position for {}", symbol);
        }
        _ => {}
    }
    Ok(())
}
```

---

## 10. Project Structure

```
tui/
├── Cargo.toml
├── Cargo.lock
├── config.example.toml
├── README.md
├── Dockerfile
├── src/
│   ├── main.rs                 # Application entry point
│   ├── config.rs               # Configuration loading
│   ├── state.rs                # Application state management
│   ├── ws/
│   │   ├── mod.rs              # WebSocket client module
│   │   ├── client.rs           # WS connection and reconnection
│   │   ├── messages.rs         # Message type definitions
│   │   └── subscription.rs     # Subscription management
│   ├── ui/
│   │   ├── mod.rs              # UI module
│   │   ├── layout.rs           # Layout calculation
│   │   ├── render.rs           # Main render function
│   │   ├── panels/
│   │   │   ├── positions.rs    # Position panel widget
│   │   │   ├── orders.rs       # Orders panel widget
│   │   │   ├── orderbook.rs    # Order book panel widget
│   │   │   ├── fills.rs        # Fills panel widget
│   │   │   ├── signals.rs      # AI signals panel widget
│   │   │   └── alerts.rs       # Alerts panel widget
│   │   ├── status.rs           # Status bar widget
│   │   ├── help.rs             # Help overlay widget
│   │   └── theme.rs            # Color scheme definitions
│   ├── keyboard/
│   │   ├── mod.rs              # Keyboard handling module
│   │   ├── handler.rs          # Key event handler
│   │   ├── commands.rs         # Command parsing
│   │   └── shortcuts.rs        # Shortcut definitions
│   └── metrics.rs              # Performance metrics
├── tests/
│   ├── integration/
│   │   ├── websocket_test.rs   # WS integration tests
│   │   └── ui_test.rs          # UI integration tests
│   └── snapshots/              # Insta snapshot files
├── benches/
│   └── render_bench.rs         # Performance benchmarks
└── examples/
    ├── mock_dashboard.rs       # Mock dashboard server for testing
    └── demo_data.rs            # Demo with mock data
```

---

## 11. Development Timeline

| Phase | Duration | Deliverable |
|-------|----------|-------------|
| Phase 1 | Week 1 | Basic TUI framework with empty panels |
| Phase 2 | Week 1-2 | WebSocket client with live connection |
| Phase 3 | Week 2 | Position panel with real-time updates |
| Phase 4 | Week 2-3 | Order book panel with depth visualization |
| Phase 5 | Week 3 | Orders and fills panels |
| Phase 6 | Week 3-4 | AI signals and alerts panels |
| Phase 7 | Week 4 | Keyboard controls and manual trading |
| Phase 8 | Week 4-5 | Performance optimization and testing |
| **Total** | **4-5 weeks** | Production-ready TUI |

---

## 12. Success Metrics

**Performance:**
- ✅ CPU usage < 2% idle, < 5% under load
- ✅ Memory usage < 50 MB
- ✅ Render time < 10ms per frame
- ✅ WebSocket reconnection < 1s
- ✅ Input latency < 50ms

**Reliability:**
- ✅ 24+ hour stability test without crashes
- ✅ Automatic reconnection on disconnect
- ✅ Graceful degradation on data loss
- ✅ No memory leaks (verified with valgrind)

**Usability:**
- ✅ Intuitive keyboard navigation
- ✅ Clear visual indicators for all states
- ✅ Responsive to terminal resize
- ✅ Comprehensive help documentation
- ✅ Low learning curve for basic operations

**Testing:**
- ✅ 80%+ code coverage
- ✅ All integration tests passing
- ✅ Performance benchmarks within targets
- ✅ Snapshot tests for all panels

---

## 13. Future Enhancements (Post-MVP)

**Phase 2 Features:**
1. Chart panel with candlestick visualization (using plotters-rs)
2. Mouse support for panel interaction
3. Multiple symbol tabs
4. Custom panel layouts (user-configurable)
5. Trade history search and filtering
6. Strategy performance comparison panel
7. Risk metrics visualization
8. Export data to CSV

**Phase 3 Features:**
1. TUI recording/replay for debugging
2. Plugin system for custom panels
3. Multi-exchange support with selector
4. Alert rules editor
5. Position calculator widget
6. Hotkey macros
7. Split-screen mode (multiple orderbooks)
8. Integration with external tools (vim-style external editing)

---

## 14. Dependencies and Prerequisites

**Build Requirements:**
- Rust 1.75+ (stable channel)
- cargo, rustc
- Linux/macOS/Windows with terminal emulator supporting 256 colors

**Runtime Requirements:**
- Dashboard Server running and accessible
- WebSocket endpoint exposed
- Terminal size: minimum 80x24, recommended 200x50
- UTF-8 terminal support

**Development Tools:**
- rust-analyzer (IDE support)
- cargo-watch (hot reload during development)
- cargo-criterion (benchmarking)
- cargo-insta (snapshot testing)

---

## 15. References

**Documentation:**
- [ratatui Documentation](https://docs.rs/ratatui)
- [crossterm Documentation](https://docs.rs/crossterm)
- [tokio-tungstenite Documentation](https://docs.rs/tokio-tungstenite)

**Examples:**
- [ratatui Examples](https://github.com/ratatui-org/ratatui/tree/main/examples)
- [gitui Source Code](https://github.com/extrawurst/gitui) - Production TUI app
- [bottom Source Code](https://github.com/ClementTsang/bottom) - System monitor TUI

**Design Inspiration:**
- htop (process monitor)
- k9s (Kubernetes TUI)
- lazygit (git TUI)
- ticker (stock market TUI)

---

**End of Development Plan**

This comprehensive plan provides the roadmap for building a production-ready Terminal UI service. The architecture prioritizes performance, reliability, and usability while maintaining clean separation of concerns and testability.
