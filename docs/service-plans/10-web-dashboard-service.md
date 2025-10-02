# Web Dashboard Service - Development Plan

**Service:** Web Dashboard Service (Sub-system 10)
**Purpose:** Web-based user interface for real-time trading system monitoring and control
**Last Updated:** 2025-10-02
**Version:** 1.0

---

## 1. Technology Stack Recommendation

### Core Framework: **React 18+ with JavaScript (ES6+)**

**IMPORTANT: JavaScript-Only Policy**

This project enforces a **strict JavaScript-only policy**:
- NO TypeScript allowed
- All code must be written in pure JavaScript (ES6+)
- Use JSDoc comments for type documentation
- No `.ts` or `.tsx` files - use `.js` and `.jsx` only

**Rationale:**
- Industry-standard with extensive ecosystem
- Excellent WebSocket integration libraries
- JSDoc provides documentation without build complexity
- Server Components support for future optimization
- Large talent pool for maintenance
- Simpler debugging and runtime behavior

**Alternative Consideration:** SolidJS (best performance) or Svelte (simplest DX)

### Build Tool: **Vite 5+**

**Rationale:**
- Lightning-fast HMR for development experience
- Optimized production builds with Rollup
- Native ESM support
- Excellent JavaScript (ES6+) support
- Minimal configuration
- No TypeScript compilation needed

### UI Framework: **Tailwind CSS + shadcn/ui**

**Rationale:**
- Tailwind: Utility-first, minimal bundle size with purging
- shadcn/ui: Headless, accessible components (copy-paste approach)
- Full customization control
- Excellent dark mode support
- No runtime CSS-in-JS overhead

**Alternative:** Ant Design (more opinionated, faster prototyping)

### Charting Library: **Lightweight TradingView + Apache ECharts**

**Rationale:**
- **TradingView Lightweight Charts:**
  - Built for financial data
  - Real-time performance
  - Small bundle size (~200KB)
  - Candlestick, line, area charts
- **Apache ECharts:**
  - Complex visualizations (P&L, risk metrics)
  - Excellent performance with large datasets
  - WebGL acceleration support

**Alternative:** Recharts (simpler, React-native) for non-critical charts

### WebSocket Client: **Native WebSocket API + Custom Hook**

**Rationale:**
- No dependency overhead
- Full control over reconnection logic
- Simple integration with state management
- SSE fallback capability if needed

### State Management: **Zustand + TanStack Query**

**Rationale:**
- **Zustand:**
  - Minimal boilerplate
  - No provider hell
  - Excellent DevTools
  - Perfect for global WebSocket state
- **TanStack Query:**
  - Server state management
  - Caching and invalidation
  - REST API queries for historical data
  - Optimistic updates

**Alternative:** Jotai (atomic state) or Valtio (proxy-based)

### Testing Stack

| Layer | Tool | Purpose |
|-------|------|---------|
| Unit | **Vitest** | Component and hook testing (Vite-native) |
| Integration | **Testing Library** | User-centric component testing |
| E2E | **Playwright** | Full browser automation |
| Visual | **Chromatic/Percy** | Visual regression detection |
| Performance | **Lighthouse CI** | Web Vitals monitoring |

### Additional Libraries

```json
{
  "dependencies": {
    "react": "^18.3.0",
    "react-dom": "^18.3.0",
    "zustand": "^4.5.0",
    "@tanstack/react-query": "^5.28.0",
    "lightweight-charts": "^4.1.0",
    "echarts": "^5.5.0",
    "echarts-for-react": "^3.0.0",
    "react-router-dom": "^6.22.0",
    "date-fns": "^3.3.0",
    "clsx": "^2.1.0",
    "tailwind-merge": "^2.2.0",
    "lucide-react": "^0.344.0",
    "sonner": "^1.4.0"
  },
  "devDependencies": {
    "@vitejs/plugin-react": "^4.2.0",
    "vite": "^5.1.0",
    "vitest": "^1.3.0",
    "@testing-library/react": "^14.2.0",
    "@testing-library/user-event": "^14.5.0",
    "@playwright/test": "^1.42.0",
    "tailwindcss": "^3.4.0",
    "autoprefixer": "^10.4.0",
    "postcss": "^8.4.0",
    "eslint": "^8.57.0",
    "prettier": "^3.2.0"
  }

  NOTE: NO TypeScript - JavaScript (ES6+) only!
}
```

---

## 2. Architecture Design

### Component Hierarchy

```
App (Root)
├── WebSocketProvider (Context)
├── ThemeProvider (Dark/Light)
└── Router
    ├── Layout (Shell)
    │   ├── Navbar
    │   │   ├── Logo
    │   │   ├── Navigation
    │   │   └── ConnectionStatus
    │   ├── Sidebar (Mobile drawer)
    │   └── ErrorBoundary
    │       └── Outlet (Page content)
    └── Routes
        ├── DashboardPage (/)
        │   ├── AccountSummary
        │   ├── PositionsGrid
        │   ├── RecentOrdersList
        │   └── QuickStats
        ├── PositionsPage (/positions)
        │   ├── PositionsTable
        │   │   └── PositionRow (actions)
        │   └── PositionChart
        ├── OrdersPage (/orders)
        │   ├── ActiveOrdersTable
        │   ├── OrderHistoryTable
        │   └── OrderForm
        ├── OrderBookPage (/orderbook)
        │   ├── SymbolSelector
        │   ├── OrderBookVisualization
        │   │   ├── BidDepthChart
        │   │   ├── AskDepthChart
        │   │   └── SpreadIndicator
        │   └── RecentTrades
        ├── ChartsPage (/analytics)
        │   ├── PnLChart (Cumulative)
        │   ├── DailyPnLChart
        │   ├── WinRateChart
        │   ├── DrawdownChart
        │   └── TradeDistribution
        ├── TradingPage (/trade)
        │   ├── TradingForm
        │   │   ├── SymbolInput
        │   │   ├── OrderTypeSelector
        │   │   ├── PriceInput
        │   │   ├── QuantityInput
        │   │   └── SubmitButton
        │   ├── LiveOrderBook (Mini)
        │   └── PositionPreview
        └── SystemPage (/system)
            ├── ServiceHealthGrid
            ├── LatencyMetrics
            ├── ErrorLog
            └── ConfigViewer
```

### WebSocket Integration Pattern

```javascript
// Custom hook pattern with auto-reconnect
/**
 * @typedef {Object} WebSocketHook
 * @property {TradingData} data
 * @property {'connecting'|'connected'|'disconnected'|'error'} status
 * @property {number} latency
 * @property {Function} reconnect
 * @property {Function} subscribe - Subscribe to a channel
 * @property {Function} unsubscribe - Unsubscribe from a channel
 */

// Store pattern
interface WebSocketStore {
  // Connection state
  connection: WebSocket | null;
  status: ConnectionStatus;
  latency: number;
  lastUpdate: Date;

  // Data streams
  orderbook: OrderBookState;
  positions: Position[];
  orders: Order[];
  account: AccountState;
  trades: Trade[];
  systemHealth: HealthMetrics;

  // Actions
  connect: () => void;
  disconnect: () => void;
  subscribe: (channel: string) => void;
  sendOrder: (order: OrderRequest) => void;
}
```

**Implementation Strategy:**
1. Single WebSocket connection to Dashboard Server
2. Message routing by topic/type
3. Automatic reconnection with exponential backoff
4. Heartbeat/ping-pong for connection validation
5. Message queue during disconnection
6. Optimistic updates with rollback

### State Management Approach

**Three-Layer Architecture:**

1. **WebSocket State (Zustand)**
   - Real-time data from server
   - Connection management
   - Subscription handling

2. **Server State (TanStack Query)**
   - Historical data queries
   - REST API integration
   - Cache management

3. **UI State (React State)**
   - Form inputs
   - Modal visibility
   - Filters and sorting

```typescript
// Example Zustand store structure
interface TradingStore {
  // Real-time data
  orderbook: Map<string, OrderBook>;
  positions: Map<string, Position>;
  orders: Map<string, Order>;
  account: AccountState | null;

  // Connection state
  ws: WebSocket | null;
  status: ConnectionStatus;
  latency: number;

  // Actions
  updateOrderBook: (symbol: string, data: OrderBook) => void;
  updatePosition: (position: Position) => void;
  updateOrder: (order: Order) => void;
  updateAccount: (account: AccountState) => void;
}
```

### Routing Structure

```typescript
// Route configuration
const routes = [
  {
    path: '/',
    element: <Layout />,
    errorElement: <ErrorPage />,
    children: [
      { index: true, element: <DashboardPage /> },
      { path: 'positions', element: <PositionsPage /> },
      { path: 'orders', element: <OrdersPage /> },
      { path: 'orderbook', element: <OrderBookPage /> },
      { path: 'analytics', element: <ChartsPage /> },
      { path: 'trade', element: <TradingPage /> },
      { path: 'system', element: <SystemPage /> },
    ],
  },
];
```

### Responsive Design Strategy

**Breakpoint System (Tailwind):**
- `sm`: 640px (Mobile landscape)
- `md`: 768px (Tablet)
- `lg`: 1024px (Desktop)
- `xl`: 1280px (Large desktop)
- `2xl`: 1536px (Ultra-wide)

**Mobile-First Approach:**
1. **Mobile (<768px)**
   - Single column layout
   - Bottom navigation
   - Swipeable tabs
   - Condensed tables (horizontal scroll)
   - Essential metrics only

2. **Tablet (768-1024px)**
   - Two column grid
   - Side navigation
   - Expandable panels
   - Full tables with pagination

3. **Desktop (>1024px)**
   - Multi-column dashboard
   - Sidebar navigation
   - Multi-panel layouts
   - Dense information display

**Performance Targets:**
- Mobile: FCP < 1.8s, TTI < 3.9s
- Desktop: FCP < 1.2s, TTI < 2.5s
- Bundle: Main < 150KB, Total < 500KB (gzipped)
- 60fps scrolling and animations

---

## 3. Development Phases

### Phase 1: Project Setup and Layout (Week 1, Days 1-2)

**Goals:**
- Initialize Vite + React + JavaScript project
- Configure Tailwind CSS and shadcn/ui
- Implement layout shell and routing
- Setup development tooling

**Tasks:**
1. Project initialization
   ```bash
   npm create vite@latest web-dashboard -- --template react
   cd web-dashboard
   npm install
   ```

   **IMPORTANT**: Use `react` template, NOT `react-ts`. JavaScript only!

2. Install dependencies
   ```bash
   npm install zustand @tanstack/react-query react-router-dom
   npm install clsx tailwind-merge date-fns lucide-react sonner
   npm install -D tailwindcss postcss autoprefixer
   npx tailwindcss init -p
   ```

3. Configure Tailwind and theming
4. Implement Layout component with navbar and sidebar
5. Setup routing structure
6. Configure ESLint and Prettier
7. Setup Vitest configuration

**Deliverables:**
- Working dev server
- Navigation structure
- Dark/light theme toggle
- Basic error boundary

### Phase 2: WebSocket Client and State Management (Week 1, Days 3-4)

**Goals:**
- Implement WebSocket connection manager
- Create Zustand store for real-time data
- Build connection status indicator
- Implement auto-reconnection logic

**Tasks:**
1. Create `useWebSocket` hook
   ```javascript
   // src/hooks/useWebSocket.js
   /**
    * Custom WebSocket hook with auto-reconnection
    * @param {string} url - WebSocket URL
    * @returns {Object} WebSocket connection state and methods
    */
   export function useWebSocket(url) {
     const [status, setStatus] = useState<ConnectionStatus>('disconnected');
     const wsRef = useRef<WebSocket | null>(null);

     const connect = useCallback(() => {
       const ws = new WebSocket(url);

       ws.onopen = () => setStatus('connected');
       ws.onclose = () => {
         setStatus('disconnected');
         // Reconnect after 3s
         setTimeout(connect, 3000);
       };
       ws.onerror = () => setStatus('error');
       ws.onmessage = (event) => {
         const message = JSON.parse(event.data);
         handleMessage(message);
       };

       wsRef.current = ws;
     }, [url]);

     useEffect(() => {
       connect();
       return () => wsRef.current?.close();
     }, [connect]);

     return { status, send: wsRef.current?.send };
   }
   ```

2. Create Zustand store
   ```javascript
   // src/store/trading.js
   export const useTradingStore = create<TradingStore>((set, get) => ({
     orderbook: new Map(),
     positions: new Map(),
     orders: new Map(),
     account: null,

     updateOrderBook: (symbol, data) => {
       set(state => ({
         orderbook: new Map(state.orderbook).set(symbol, data)
       }));
     },

     // ... other actions
   }));
   ```

3. Implement message router
4. Build connection status component
5. Add latency measurement (ping-pong)
6. Implement subscription management

**Deliverables:**
- Working WebSocket connection
- Real-time data flow to store
- Connection status indicator
- Auto-reconnection

### Phase 3: Account Overview Page (Week 1, Days 5-6)

**Goals:**
- Display account balance and equity
- Show active positions summary
- Display recent orders
- Implement quick stats cards

**Tasks:**
1. Create `AccountSummary` component
   ```javascript
   // src/components/AccountSummary.jsx
   export function AccountSummary() {
     const account = useTradingStore(state => state.account);

     return (
       <Card>
         <CardHeader>Account Overview</CardHeader>
         <CardContent className="grid grid-cols-2 md:grid-cols-4 gap-4">
           <Stat label="Balance" value={account?.balance} />
           <Stat label="Equity" value={account?.equity} />
           <Stat label="P&L (24h)" value={account?.pnl24h} trend />
           <Stat label="Margin Ratio" value={account?.marginRatio} />
         </CardContent>
       </Card>
     );
   }
   ```

2. Create `PositionsGrid` component (compact view)
3. Create `RecentOrdersList` component
4. Create `QuickStats` cards
5. Implement responsive grid layout
6. Add skeleton loading states

**Deliverables:**
- Functional dashboard page
- Real-time data updates
- Mobile-responsive layout
- Loading states

### Phase 4: Positions and Orders Pages (Week 2, Days 1-3)

**Goals:**
- Build detailed positions table
- Build active orders table with cancel action
- Implement order history view
- Add filtering and sorting

**Tasks:**
1. Create `PositionsTable` component
   ```javascript
   // src/components/PositionsTable.jsx
   export function PositionsTable() {
     const positions = useTradingStore(state =>
       Array.from(state.positions.values())
     );

     return (
       <Table>
         <TableHeader>
           <TableRow>
             <TableHead>Symbol</TableHead>
             <TableHead>Side</TableHead>
             <TableHead>Size</TableHead>
             <TableHead>Entry Price</TableHead>
             <TableHead>Mark Price</TableHead>
             <TableHead>P&L</TableHead>
             <TableHead>Actions</TableHead>
           </TableRow>
         </TableHeader>
         <TableBody>
           {positions.map(position => (
             <PositionRow key={position.symbol} position={position} />
           ))}
         </TableBody>
       </Table>
     );
   }
   ```

2. Create `ActiveOrdersTable` with cancel functionality
3. Create `OrderHistoryTable` with pagination
4. Implement table sorting and filtering
5. Add position close action
6. Add order modification modal
7. Create mobile card view (alternative to table)

**Deliverables:**
- Positions page with real-time updates
- Orders page with management actions
- Filtering and sorting
- Mobile-responsive tables

### Phase 5: Order Book Visualization (Week 2, Days 4-5)

**Goals:**
- Real-time order book depth chart
- Bid/ask spread visualization
- Recent trades list
- Symbol selection

**Tasks:**
1. Create `OrderBookVisualization` component
   ```typescript
   // src/components/OrderBookVisualization.tsx
   export function OrderBookVisualization({ symbol }: Props) {
     const orderbook = useTradingStore(state =>
       state.orderbook.get(symbol)
     );

     const chartData = useMemo(() => ({
       bids: orderbook?.bids.map((level, i) => ({
         price: level.price,
         volume: level.volume,
         cumulative: orderbook.bids.slice(0, i+1)
           .reduce((sum, l) => sum + l.volume, 0)
       })) || [],
       asks: orderbook?.asks.map((level, i) => ({
         price: level.price,
         volume: level.volume,
         cumulative: orderbook.asks.slice(0, i+1)
           .reduce((sum, l) => sum + l.volume, 0)
       })) || []
     }), [orderbook]);

     return <DepthChart data={chartData} />;
   }
   ```

2. Implement depth chart with ECharts
3. Create bid/ask book table (traditional view)
4. Add spread indicator
5. Create recent trades list
6. Implement symbol selector dropdown
7. Optimize rendering (virtualization if needed)

**Deliverables:**
- Real-time order book visualization
- Depth chart with cumulative volume
- Symbol switching
- Performant updates (250ms)

### Phase 6: Charts and Analytics Page (Week 3, Days 1-3)

**Goals:**
- Cumulative P&L chart
- Daily P&L bar chart
- Win rate and statistics
- Drawdown visualization

**Tasks:**
1. Create `PnLChart` component
   ```typescript
   // src/components/PnLChart.tsx
   export function PnLChart() {
     const { data } = useQuery({
       queryKey: ['pnl-history'],
       queryFn: () => fetch('/api/pnl/history').then(r => r.json())
     });

     const chartConfig = {
       xAxis: { type: 'time' },
       yAxis: { type: 'value' },
       series: [{
         data: data?.map(d => [d.timestamp, d.cumulative_pnl]),
         type: 'line',
         smooth: true,
         areaStyle: {}
       }]
     };

     return <ReactECharts option={chartConfig} />;
   }
   ```

2. Create `DailyPnLChart` (bar chart)
3. Create `WinRateChart` (pie/donut chart)
4. Create `DrawdownChart`
5. Create `TradeDistribution` chart
6. Implement time range selector (1D, 1W, 1M, ALL)
7. Add chart export functionality

**Deliverables:**
- Analytics dashboard with multiple charts
- Historical data integration
- Interactive time range selection
- Export charts as images

### Phase 7: Manual Trading Interface (Week 3, Days 4-5)

**Goals:**
- Order submission form
- Real-time validation
- Order preview
- Quick order templates

**Tasks:**
1. Create `TradingForm` component
   ```typescript
   // src/components/TradingForm.tsx
   export function TradingForm() {
     const [order, setOrder] = useState<OrderDraft>({
       symbol: 'BTCUSDT',
       side: 'BUY',
       type: 'LIMIT',
       price: 0,
       quantity: 0
     });

     const account = useTradingStore(state => state.account);

     const validation = useMemo(() => ({
       sufficientBalance: account?.balance >= order.quantity * order.price,
       validPrice: order.price > 0,
       validQuantity: order.quantity > 0
     }), [order, account]);

     const handleSubmit = () => {
       useTradingStore.getState().sendOrder(order);
     };

     return (
       <form onSubmit={handleSubmit}>
         {/* Form fields */}
       </form>
     );
   }
   ```

2. Implement order type selector (LIMIT, MARKET, STOP)
3. Add price and quantity inputs with validation
4. Create order preview calculation
5. Implement quick order buttons (25%, 50%, 75%, 100%)
6. Add order confirmation modal
7. Show order submission feedback (toast)

**Deliverables:**
- Functional trading form
- Real-time validation
- Order submission via WebSocket
- User feedback on success/error

### Phase 8: Mobile Responsiveness (Week 3, Day 6)

**Goals:**
- Optimize all pages for mobile
- Implement bottom navigation
- Add swipe gestures
- Optimize touch targets

**Tasks:**
1. Implement bottom navigation for mobile
2. Convert tables to card views on mobile
3. Add swipe gestures for tab navigation
4. Optimize touch target sizes (min 44x44px)
5. Test on multiple device sizes
6. Implement pull-to-refresh
7. Add haptic feedback for actions

**Deliverables:**
- Fully mobile-responsive app
- Touch-optimized interactions
- Native-like experience

### Phase 9: Testing and Optimization (Week 4, Days 1-6)

**Goals:**
- Comprehensive test coverage
- Performance optimization
- Accessibility audit
- Production build

**Tasks:**
1. Write component unit tests (Vitest)
   ```typescript
   // src/components/AccountSummary.test.tsx
   import { render, screen } from '@testing-library/react';
   import { AccountSummary } from './AccountSummary';

   test('renders account balance', () => {
     render(<AccountSummary />);
     expect(screen.getByText(/balance/i)).toBeInTheDocument();
   });
   ```

2. Write WebSocket integration tests
3. Write E2E tests (Playwright)
   ```typescript
   // tests/e2e/dashboard.spec.ts
   import { test, expect } from '@playwright/test';

   test('dashboard loads and displays data', async ({ page }) => {
     await page.goto('/');
     await expect(page.locator('h1')).toContainText('Dashboard');
     // Wait for WebSocket connection
     await expect(page.locator('[data-status="connected"]'))
       .toBeVisible({ timeout: 5000 });
   });
   ```

4. Implement visual regression tests
5. Run Lighthouse audit
6. Optimize bundle size
   - Code splitting by route
   - Lazy load charts
   - Tree shake unused dependencies
7. Implement error boundaries
8. Add loading skeletons
9. Accessibility audit (WCAG 2.1 AA)
10. Production build configuration

**Deliverables:**
- 80%+ test coverage
- All E2E tests passing
- Lighthouse score > 90
- Bundle size < 500KB gzipped
- WCAG 2.1 AA compliant

---

## 4. Implementation Details

### WebSocket Hook Implementation

```javascript
// src/hooks/useWebSocket.js
import { useEffect, useRef, useState, useCallback } from 'react';
import { useTradingStore } from '../store/trading';

interface UseWebSocketOptions {
  url: string;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
  heartbeatInterval?: number;
}

export function useWebSocket({
  url,
  reconnectInterval = 3000,
  maxReconnectAttempts = 10,
  heartbeatInterval = 30000
}: UseWebSocketOptions) {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const heartbeatRef = useRef<NodeJS.Timeout>();
  const [status, setStatus] = useState<ConnectionStatus>('disconnected');
  const [latency, setLatency] = useState(0);

  const handleMessage = useCallback((event: MessageEvent) => {
    const message = JSON.parse(event.data);

    // Handle pong for latency measurement
    if (message.type === 'pong') {
      const now = Date.now();
      setLatency(now - message.timestamp);
      return;
    }

    // Route message to appropriate store action
    switch (message.type) {
      case 'orderbook':
        useTradingStore.getState().updateOrderBook(message.symbol, message.data);
        break;
      case 'position':
        useTradingStore.getState().updatePosition(message.data);
        break;
      case 'order':
        useTradingStore.getState().updateOrder(message.data);
        break;
      case 'account':
        useTradingStore.getState().updateAccount(message.data);
        break;
      default:
        console.warn('Unknown message type:', message.type);
    }
  }, []);

  const sendHeartbeat = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: 'ping',
        timestamp: Date.now()
      }));
    }
  }, []);

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    setStatus('connecting');
    const ws = new WebSocket(url);

    ws.onopen = () => {
      console.log('WebSocket connected');
      setStatus('connected');
      reconnectAttemptsRef.current = 0;

      // Start heartbeat
      heartbeatRef.current = setInterval(sendHeartbeat, heartbeatInterval);

      // Subscribe to initial channels
      ws.send(JSON.stringify({
        type: 'subscribe',
        channels: ['account', 'positions', 'orders']
      }));
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected');
      setStatus('disconnected');

      // Clear heartbeat
      if (heartbeatRef.current) {
        clearInterval(heartbeatRef.current);
      }

      // Attempt reconnection
      if (reconnectAttemptsRef.current < maxReconnectAttempts) {
        reconnectAttemptsRef.current++;
        console.log(`Reconnecting... (attempt ${reconnectAttemptsRef.current})`);
        setTimeout(connect, reconnectInterval * Math.pow(2, reconnectAttemptsRef.current - 1));
      } else {
        setStatus('error');
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      setStatus('error');
    };

    ws.onmessage = handleMessage;

    wsRef.current = ws;
  }, [url, reconnectInterval, maxReconnectAttempts, handleMessage, sendHeartbeat, heartbeatInterval]);

  const disconnect = useCallback(() => {
    if (heartbeatRef.current) {
      clearInterval(heartbeatRef.current);
    }
    wsRef.current?.close();
    wsRef.current = null;
  }, []);

  const subscribe = useCallback((channel: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: 'subscribe',
        channel
      }));
    }
  }, []);

  const unsubscribe = useCallback((channel: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: 'unsubscribe',
        channel
      }));
    }
  }, []);

  useEffect(() => {
    connect();
    return disconnect;
  }, [connect, disconnect]);

  return {
    status,
    latency,
    connect,
    disconnect,
    subscribe,
    unsubscribe,
    send: wsRef.current?.send.bind(wsRef.current)
  };
}
```

### State Management Structure

```javascript
// src/store/trading.js
import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

export interface OrderBook {
  symbol: string;
  bids: PriceLevel[];
  asks: PriceLevel[];
  timestamp: number;
}

export interface PriceLevel {
  price: number;
  volume: number;
}

export interface Position {
  symbol: string;
  side: 'LONG' | 'SHORT';
  size: number;
  entryPrice: number;
  markPrice: number;
  unrealizedPnl: number;
  leverage: number;
  liquidationPrice: number;
}

export interface Order {
  id: string;
  symbol: string;
  side: 'BUY' | 'SELL';
  type: 'LIMIT' | 'MARKET' | 'STOP_LIMIT' | 'STOP_MARKET';
  status: 'NEW' | 'PARTIALLY_FILLED' | 'FILLED' | 'CANCELED' | 'REJECTED';
  price?: number;
  quantity: number;
  filledQuantity: number;
  avgFillPrice?: number;
  timestamp: number;
}

export interface AccountState {
  balance: number;
  equity: number;
  availableBalance: number;
  unrealizedPnl: number;
  marginRatio: number;
  totalPnl24h: number;
}

interface TradingStore {
  // State
  orderbook: Map<string, OrderBook>;
  positions: Map<string, Position>;
  orders: Map<string, Order>;
  account: AccountState | null;
  selectedSymbol: string;

  // Actions
  updateOrderBook: (symbol: string, data: OrderBook) => void;
  updatePosition: (position: Position) => void;
  updateOrder: (order: Order) => void;
  removeOrder: (orderId: string) => void;
  updateAccount: (account: AccountState) => void;
  setSelectedSymbol: (symbol: string) => void;

  // WebSocket actions
  sendOrder: (order: OrderRequest) => void;
  cancelOrder: (orderId: string) => void;
  closePosition: (symbol: string) => void;
}

export const useTradingStore = create<TradingStore>()(
  devtools(
    (set, get) => ({
      // Initial state
      orderbook: new Map(),
      positions: new Map(),
      orders: new Map(),
      account: null,
      selectedSymbol: 'BTCUSDT',

      // Update actions
      updateOrderBook: (symbol, data) => {
        set(state => {
          const newOrderbook = new Map(state.orderbook);
          newOrderbook.set(symbol, data);
          return { orderbook: newOrderbook };
        });
      },

      updatePosition: (position) => {
        set(state => {
          const newPositions = new Map(state.positions);
          newPositions.set(position.symbol, position);
          return { positions: newPositions };
        });
      },

      updateOrder: (order) => {
        set(state => {
          const newOrders = new Map(state.orders);
          newOrders.set(order.id, order);
          return { orders: newOrders };
        });
      },

      removeOrder: (orderId) => {
        set(state => {
          const newOrders = new Map(state.orders);
          newOrders.delete(orderId);
          return { orders: newOrders };
        });
      },

      updateAccount: (account) => {
        set({ account });
      },

      setSelectedSymbol: (symbol) => {
        set({ selectedSymbol: symbol });
      },

      // WebSocket actions (these need ws reference)
      sendOrder: (order) => {
        // This will be called from component with ws reference
        console.log('Send order:', order);
      },

      cancelOrder: (orderId) => {
        console.log('Cancel order:', orderId);
      },

      closePosition: (symbol) => {
        console.log('Close position:', symbol);
      }
    }),
    { name: 'TradingStore' }
  )
);
```

### Key Component Implementations

#### Connection Status Indicator

```javascript
// src/components/ConnectionStatus.jsx
import { useMemo } from 'react';
import { Wifi, WifiOff, AlertCircle } from 'lucide-react';
import { cn } from '../lib/utils';

interface ConnectionStatusProps {
  status: ConnectionStatus;
  latency: number;
}

export function ConnectionStatus({ status, latency }: ConnectionStatusProps) {
  const { icon: Icon, color, text } = useMemo(() => {
    switch (status) {
      case 'connected':
        return {
          icon: Wifi,
          color: 'text-green-500',
          text: `Connected (${latency}ms)`
        };
      case 'connecting':
        return {
          icon: Wifi,
          color: 'text-yellow-500 animate-pulse',
          text: 'Connecting...'
        };
      case 'disconnected':
        return {
          icon: WifiOff,
          color: 'text-gray-500',
          text: 'Disconnected'
        };
      case 'error':
        return {
          icon: AlertCircle,
          color: 'text-red-500',
          text: 'Connection Error'
        };
    }
  }, [status, latency]);

  return (
    <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-muted/50">
      <Icon className={cn('h-4 w-4', color)} />
      <span className="text-sm font-medium">{text}</span>
    </div>
  );
}
```

#### Order Book Depth Chart

```typescript
// src/components/OrderBookDepthChart.tsx
import { useMemo } from 'react';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';
import { useTradingStore } from '../store/trading';

interface OrderBookDepthChartProps {
  symbol: string;
  height?: number;
}

export function OrderBookDepthChart({ symbol, height = 400 }: OrderBookDepthChartProps) {
  const orderbook = useTradingStore(state => state.orderbook.get(symbol));

  const chartOption: EChartsOption = useMemo(() => {
    if (!orderbook) return {};

    // Calculate cumulative volumes
    const bids = orderbook.bids.map((level, i) => ({
      price: level.price,
      cumulative: orderbook.bids
        .slice(0, i + 1)
        .reduce((sum, l) => sum + l.volume, 0)
    })).reverse();

    const asks = orderbook.asks.map((level, i) => ({
      price: level.price,
      cumulative: orderbook.asks
        .slice(0, i + 1)
        .reduce((sum, l) => sum + l.volume, 0)
    }));

    return {
      grid: {
        left: '3%',
        right: '3%',
        bottom: '10%',
        top: '10%',
        containLabel: true
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'cross'
        },
        formatter: (params: any) => {
          const data = params[0];
          return `Price: ${data.name}<br/>Volume: ${data.value.toFixed(4)}`;
        }
      },
      xAxis: {
        type: 'category',
        data: [...bids.map(b => b.price), ...asks.map(a => a.price)],
        axisLabel: {
          formatter: (value: number) => value.toFixed(2)
        }
      },
      yAxis: {
        type: 'value',
        axisLabel: {
          formatter: (value: number) => value.toFixed(2)
        }
      },
      series: [
        {
          name: 'Bids',
          type: 'line',
          data: bids.map(b => b.cumulative),
          itemStyle: { color: '#10b981' },
          areaStyle: {
            color: 'rgba(16, 185, 129, 0.2)'
          },
          smooth: false,
          step: 'end'
        },
        {
          name: 'Asks',
          type: 'line',
          data: [...Array(bids.length).fill(0), ...asks.map(a => a.cumulative)],
          itemStyle: { color: '#ef4444' },
          areaStyle: {
            color: 'rgba(239, 68, 68, 0.2)'
          },
          smooth: false,
          step: 'start'
        }
      ]
    };
  }, [orderbook]);

  if (!orderbook) {
    return <div className="flex items-center justify-center h-96">Loading order book...</div>;
  }

  return (
    <ReactECharts
      option={chartOption}
      style={{ height }}
      opts={{ renderer: 'canvas' }}
    />
  );
}
```

#### Trading Form

```typescript
// src/components/TradingForm.tsx
import { useState, useMemo } from 'react';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { useTradingStore } from '../store/trading';
import { toast } from 'sonner';

type OrderSide = 'BUY' | 'SELL';
type OrderType = 'LIMIT' | 'MARKET';

export function TradingForm() {
  const [side, setSide] = useState<OrderSide>('BUY');
  const [type, setType] = useState<OrderType>('LIMIT');
  const [price, setPrice] = useState('');
  const [quantity, setQuantity] = useState('');

  const account = useTradingStore(state => state.account);
  const selectedSymbol = useTradingStore(state => state.selectedSymbol);

  const orderValue = useMemo(() => {
    const p = parseFloat(price) || 0;
    const q = parseFloat(quantity) || 0;
    return p * q;
  }, [price, quantity]);

  const validation = useMemo(() => {
    const p = parseFloat(price) || 0;
    const q = parseFloat(quantity) || 0;

    return {
      validPrice: type === 'MARKET' || p > 0,
      validQuantity: q > 0,
      sufficientBalance: account ? account.availableBalance >= orderValue : false
    };
  }, [price, quantity, type, account, orderValue]);

  const isValid = validation.validPrice && validation.validQuantity && validation.sufficientBalance;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!isValid) {
      toast.error('Invalid order parameters');
      return;
    }

    const order = {
      symbol: selectedSymbol,
      side,
      type,
      price: type === 'LIMIT' ? parseFloat(price) : undefined,
      quantity: parseFloat(quantity)
    };

    // Send via WebSocket
    useTradingStore.getState().sendOrder(order);

    toast.success('Order submitted');

    // Reset form
    setPrice('');
    setQuantity('');
  };

  const setPercentage = (percentage: number) => {
    if (!account) return;

    const available = account.availableBalance;
    const p = parseFloat(price) || 0;

    if (p > 0) {
      const q = (available * percentage / 100) / p;
      setQuantity(q.toFixed(8));
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4 p-4 border rounded-lg">
      <div className="flex gap-2">
        <Button
          type="button"
          variant={side === 'BUY' ? 'default' : 'outline'}
          className={side === 'BUY' ? 'bg-green-600 hover:bg-green-700' : ''}
          onClick={() => setSide('BUY')}
        >
          Buy
        </Button>
        <Button
          type="button"
          variant={side === 'SELL' ? 'default' : 'outline'}
          className={side === 'SELL' ? 'bg-red-600 hover:bg-red-700' : ''}
          onClick={() => setSide('SELL')}
        >
          Sell
        </Button>
      </div>

      <div>
        <Label>Order Type</Label>
        <Select value={type} onValueChange={(v) => setType(v as OrderType)}>
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="LIMIT">Limit</SelectItem>
            <SelectItem value="MARKET">Market</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {type === 'LIMIT' && (
        <div>
          <Label>Price</Label>
          <Input
            type="number"
            step="0.01"
            value={price}
            onChange={(e) => setPrice(e.target.value)}
            placeholder="0.00"
          />
        </div>
      )}

      <div>
        <Label>Quantity</Label>
        <Input
          type="number"
          step="0.00000001"
          value={quantity}
          onChange={(e) => setQuantity(e.target.value)}
          placeholder="0.00"
        />
        <div className="flex gap-2 mt-2">
          <Button type="button" size="sm" variant="outline" onClick={() => setPercentage(25)}>
            25%
          </Button>
          <Button type="button" size="sm" variant="outline" onClick={() => setPercentage(50)}>
            50%
          </Button>
          <Button type="button" size="sm" variant="outline" onClick={() => setPercentage(75)}>
            75%
          </Button>
          <Button type="button" size="sm" variant="outline" onClick={() => setPercentage(100)}>
            100%
          </Button>
        </div>
      </div>

      <div className="p-3 bg-muted rounded">
        <div className="flex justify-between text-sm">
          <span>Order Value:</span>
          <span className="font-medium">${orderValue.toFixed(2)}</span>
        </div>
        <div className="flex justify-between text-sm mt-1">
          <span>Available:</span>
          <span className="font-medium">${account?.availableBalance.toFixed(2) ?? '0.00'}</span>
        </div>
      </div>

      <Button
        type="submit"
        disabled={!isValid}
        className="w-full"
        variant={side === 'BUY' ? 'default' : 'destructive'}
      >
        {side === 'BUY' ? 'Buy' : 'Sell'} {selectedSymbol}
      </Button>

      {!validation.sufficientBalance && (
        <p className="text-sm text-red-500">Insufficient balance</p>
      )}
    </form>
  );
}
```

### Chart Configurations

#### Cumulative P&L Chart

```typescript
// src/components/PnLChart.tsx
import { useQuery } from '@tanstack/react-query';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';

export function PnLChart() {
  const { data: pnlHistory, isLoading } = useQuery({
    queryKey: ['pnl-history'],
    queryFn: async () => {
      const response = await fetch('/api/analytics/pnl');
      return response.json();
    },
    refetchInterval: 60000 // Refresh every minute
  });

  const chartOption: EChartsOption = {
    grid: {
      left: '3%',
      right: '3%',
      bottom: '10%',
      top: '10%',
      containLabel: true
    },
    tooltip: {
      trigger: 'axis',
      formatter: (params: any) => {
        const data = params[0];
        const date = new Date(data.name).toLocaleString();
        const value = data.value.toFixed(2);
        const color = data.value >= 0 ? 'green' : 'red';
        return `${date}<br/>P&L: <span style="color: ${color}">$${value}</span>`;
      }
    },
    xAxis: {
      type: 'time',
      axisLabel: {
        formatter: (value: number) => {
          return new Date(value).toLocaleDateString();
        }
      }
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        formatter: (value: number) => `$${value}`
      },
      splitLine: {
        lineStyle: {
          type: 'dashed'
        }
      }
    },
    series: [
      {
        name: 'Cumulative P&L',
        type: 'line',
        data: pnlHistory?.map((d: any) => [d.timestamp, d.cumulative_pnl]) || [],
        smooth: true,
        lineStyle: {
          width: 2
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(16, 185, 129, 0.3)' },
              { offset: 1, color: 'rgba(16, 185, 129, 0.05)' }
            ]
          }
        },
        itemStyle: {
          color: '#10b981'
        }
      }
    ],
    dataZoom: [
      {
        type: 'inside',
        start: 0,
        end: 100
      },
      {
        start: 0,
        end: 100
      }
    ]
  };

  if (isLoading) {
    return <div className="flex items-center justify-center h-96">Loading chart...</div>;
  }

  return <ReactECharts option={chartOption} style={{ height: 400 }} />;
}
```

### Theme and Styling

#### Tailwind Configuration

```typescript
// tailwind.config.ts
import type { Config } from 'tailwindcss';

const config: Config = {
  darkMode: ['class'],
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        border: 'hsl(var(--border))',
        input: 'hsl(var(--input))',
        ring: 'hsl(var(--ring))',
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        primary: {
          DEFAULT: 'hsl(var(--primary))',
          foreground: 'hsl(var(--primary-foreground))',
        },
        success: {
          DEFAULT: '#10b981',
          foreground: '#ffffff',
        },
        danger: {
          DEFAULT: '#ef4444',
          foreground: '#ffffff',
        },
        // Trading-specific colors
        buy: '#10b981',
        sell: '#ef4444',
        profit: '#10b981',
        loss: '#ef4444',
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
      },
    },
  },
  plugins: [require('tailwindcss-animate')],
};

export default config;
```

#### Dark Theme CSS Variables

```css
/* src/index.css */
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    --background: 0 0% 100%;
    --foreground: 222.2 84% 4.9%;
    --primary: 222.2 47.4% 11.2%;
    --primary-foreground: 210 40% 98%;
    --border: 214.3 31.8% 91.4%;
    --input: 214.3 31.8% 91.4%;
    --ring: 222.2 84% 4.9%;
  }

  .dark {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    --primary: 210 40% 98%;
    --primary-foreground: 222.2 47.4% 11.2%;
    --border: 217.2 32.6% 17.5%;
    --input: 217.2 32.6% 17.5%;
    --ring: 212.7 26.8% 83.9%;
  }
}

@layer utilities {
  .text-balance {
    text-wrap: balance;
  }

  .bg-grid-pattern {
    background-image: url("data:image/svg+xml,%3Csvg width='40' height='40' viewBox='0 0 40 40' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M0 0h40v40H0z' fill='none'/%3E%3Cpath d='M0 0h1v40H0zM0 0h40v1H0z' fill='%23e5e7eb' fill-opacity='0.1'/%3E%3C/svg%3E");
  }
}
```

---

## 5. Page Specifications

### Dashboard Home (Account Overview)

**Route:** `/`

**Layout:**
```
┌─────────────────────────────────────────────┐
│  Account Summary                            │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐      │
│  │ Bal  │ │ Eq   │ │ P&L  │ │ Margin│      │
│  └──────┘ └──────┘ └──────┘ └──────┘      │
└─────────────────────────────────────────────┘
┌─────────────────────────┬───────────────────┐
│  Open Positions         │  Recent Orders    │
│  ┌──────────────────┐   │  ┌─────────────┐ │
│  │ Position 1       │   │  │ Order 1     │ │
│  │ Position 2       │   │  │ Order 2     │ │
│  └──────────────────┘   │  └─────────────┘ │
└─────────────────────────┴───────────────────┘
```

**Components:**
- `AccountSummary`: 4-card grid with balance, equity, P&L, margin
- `PositionsGrid`: Compact positions table (max 5 rows, link to full page)
- `RecentOrdersList`: Last 10 orders with status
- `QuickStats`: Win rate, total trades, avg P&L

**Real-time Updates:**
- Account metrics: Every message
- Positions: On position updates
- Orders: On order updates

### Positions Page

**Route:** `/positions`

**Features:**
- Full positions table with sorting
- Close position action
- Position details modal
- P&L breakdown per position
- Risk metrics (margin, liquidation price)

**Columns:**
- Symbol
- Side (LONG/SHORT)
- Size
- Entry Price
- Mark Price
- Unrealized P&L ($ and %)
- Leverage
- Liquidation Price
- Actions (Close, Modify)

### Active Orders Page

**Route:** `/orders`

**Features:**
- Active orders table
- Order history table (separate tabs)
- Cancel order action
- Modify order modal
- Filter by symbol, status
- Bulk cancel option

**Columns:**
- Order ID
- Symbol
- Side
- Type
- Price
- Quantity
- Filled
- Status
- Time
- Actions

### Order Book Visualization

**Route:** `/orderbook`

**Components:**
- Symbol selector dropdown
- Depth chart (cumulative volume)
- Traditional book table (bids/asks)
- Recent trades list
- Spread indicator
- Book imbalance metric

**Real-time Updates:**
- Order book: Every 250ms (throttled)
- Trades: Every update

### P&L Charts and Analytics

**Route:** `/analytics`

**Charts:**
1. **Cumulative P&L**: Line chart with area fill
2. **Daily P&L**: Bar chart (green/red)
3. **Win Rate**: Donut chart (wins vs losses)
4. **Drawdown**: Area chart showing peak-to-trough
5. **Trade Distribution**: Histogram of P&L per trade
6. **Performance by Symbol**: Bar chart comparison

**Time Range Selector:**
- 1 Day
- 1 Week
- 1 Month
- 3 Months
- All Time

**Metrics Cards:**
- Total Trades
- Win Rate
- Avg Win / Avg Loss
- Sharpe Ratio
- Max Drawdown
- Profit Factor

### Manual Trading Interface

**Route:** `/trade`

**Layout:**
```
┌──────────────────┬───────────────────┐
│  Trading Form    │  Order Book       │
│  ┌────────────┐  │  ┌─────────────┐  │
│  │ Symbol     │  │  │ Asks        │  │
│  │ Side       │  │  │ ----------- │  │
│  │ Type       │  │  │ Bids        │  │
│  │ Price      │  │  └─────────────┘  │
│  │ Quantity   │  │                   │
│  │ [Submit]   │  │  Recent Trades    │
│  └────────────┘  │  ┌─────────────┐  │
│                  │  │ Trade 1     │  │
│  Position Preview│  │ Trade 2     │  │
│  ┌────────────┐  │  └─────────────┘  │
│  │ Est. P&L   │  │                   │
│  └────────────┘  │                   │
└──────────────────┴───────────────────┘
```

**Features:**
- Order form with validation
- Quick percentage buttons (25%, 50%, 75%, 100%)
- Order preview calculation
- Live order book integration
- Recent trades feed
- Position impact preview
- Order confirmation modal

### System Health Page

**Route:** `/system`

**Components:**
- Service health grid (all microservices)
- Latency metrics per service
- WebSocket connection status
- Error log viewer
- System metrics (CPU, memory, network)
- Circuit breaker status

**Real-time Monitoring:**
- Health checks every 30s
- Latency measurements
- Error log streaming

---

## 6. Testing Strategy

### Component Unit Tests (Vitest + Testing Library)

```typescript
// src/components/AccountSummary.test.tsx
import { render, screen } from '@testing-library/react';
import { describe, it, expect, beforeEach } from 'vitest';
import { AccountSummary } from './AccountSummary';
import { useTradingStore } from '../store/trading';

describe('AccountSummary', () => {
  beforeEach(() => {
    useTradingStore.setState({
      account: {
        balance: 10000,
        equity: 10500,
        availableBalance: 8000,
        unrealizedPnl: 500,
        marginRatio: 0.2,
        totalPnl24h: 250
      }
    });
  });

  it('renders account metrics', () => {
    render(<AccountSummary />);

    expect(screen.getByText(/10,000/)).toBeInTheDocument();
    expect(screen.getByText(/10,500/)).toBeInTheDocument();
  });

  it('displays P&L with correct color', () => {
    render(<AccountSummary />);

    const pnlElement = screen.getByText(/500/);
    expect(pnlElement).toHaveClass('text-green-500');
  });
});
```

### WebSocket Integration Tests

```typescript
// src/hooks/useWebSocket.test.ts
import { renderHook, waitFor } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { useWebSocket } from './useWebSocket';
import { WS } from 'vitest-websocket-mock';

describe('useWebSocket', () => {
  it('connects to WebSocket server', async () => {
    const server = new WS('ws://localhost:8080');
    const { result } = renderHook(() => useWebSocket({ url: 'ws://localhost:8080' }));

    await waitFor(() => {
      expect(result.current.status).toBe('connected');
    });

    server.close();
  });

  it('handles incoming messages', async () => {
    const server = new WS('ws://localhost:8080');
    const { result } = renderHook(() => useWebSocket({ url: 'ws://localhost:8080' }));

    await waitFor(() => {
      expect(result.current.status).toBe('connected');
    });

    server.send(JSON.stringify({
      type: 'account',
      data: { balance: 10000 }
    }));

    // Verify state update
    await waitFor(() => {
      expect(useTradingStore.getState().account?.balance).toBe(10000);
    });

    server.close();
  });

  it('reconnects on disconnection', async () => {
    const server = new WS('ws://localhost:8080');
    const { result } = renderHook(() => useWebSocket({
      url: 'ws://localhost:8080',
      reconnectInterval: 100
    }));

    await waitFor(() => {
      expect(result.current.status).toBe('connected');
    });

    server.close();

    await waitFor(() => {
      expect(result.current.status).toBe('disconnected');
    });

    // Should reconnect
    const server2 = new WS('ws://localhost:8080');

    await waitFor(() => {
      expect(result.current.status).toBe('connected');
    }, { timeout: 5000 });

    server2.close();
  });
});
```

### E2E Tests (Playwright)

```typescript
// tests/e2e/dashboard.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('loads and displays account summary', async ({ page }) => {
    // Wait for WebSocket connection
    await expect(page.locator('[data-status="connected"]')).toBeVisible({ timeout: 10000 });

    // Verify account summary is visible
    await expect(page.locator('text=Account Overview')).toBeVisible();

    // Verify metrics cards
    await expect(page.locator('text=Balance')).toBeVisible();
    await expect(page.locator('text=Equity')).toBeVisible();
  });

  test('navigates to positions page', async ({ page }) => {
    await page.click('text=Positions');

    await expect(page).toHaveURL('/positions');
    await expect(page.locator('h1:has-text("Positions")')).toBeVisible();
  });

  test('real-time order book updates', async ({ page }) => {
    await page.goto('/orderbook');

    // Select symbol
    await page.selectOption('[data-testid="symbol-selector"]', 'BTCUSDT');

    // Wait for order book to load
    await expect(page.locator('[data-testid="orderbook-chart"]')).toBeVisible();

    // Verify updates (check timestamp changes)
    const initialTimestamp = await page.locator('[data-testid="orderbook-timestamp"]').textContent();

    await page.waitForTimeout(1000);

    const updatedTimestamp = await page.locator('[data-testid="orderbook-timestamp"]').textContent();

    expect(updatedTimestamp).not.toBe(initialTimestamp);
  });
});

// tests/e2e/trading.spec.ts
test.describe('Trading', () => {
  test('submits limit order', async ({ page }) => {
    await page.goto('/trade');

    // Fill order form
    await page.fill('[data-testid="price-input"]', '50000');
    await page.fill('[data-testid="quantity-input"]', '0.01');

    // Submit order
    await page.click('[data-testid="submit-order"]');

    // Verify success toast
    await expect(page.locator('text=Order submitted')).toBeVisible();

    // Verify order appears in orders list
    await page.goto('/orders');
    await expect(page.locator('text=BTCUSDT')).toBeVisible();
  });
});
```

### Visual Regression Tests

```typescript
// tests/visual/pages.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Visual Regression', () => {
  test('dashboard page snapshot', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('[data-status="connected"]');

    await expect(page).toHaveScreenshot('dashboard.png', {
      fullPage: true,
      animations: 'disabled'
    });
  });

  test('order book page snapshot', async ({ page }) => {
    await page.goto('/orderbook');
    await page.waitForSelector('[data-testid="orderbook-chart"]');

    await expect(page).toHaveScreenshot('orderbook.png', {
      fullPage: true,
      animations: 'disabled'
    });
  });
});
```

### Performance Testing

```typescript
// tests/performance/lighthouse.spec.ts
import { test } from '@playwright/test';
import { playAudit } from 'playwright-lighthouse';

test.describe('Performance', () => {
  test('lighthouse audit - dashboard', async ({ page }) => {
    await page.goto('/');

    await playAudit({
      page,
      thresholds: {
        performance: 90,
        accessibility: 95,
        'best-practices': 90,
        seo: 80,
        pwa: 50
      },
      port: 9222
    });
  });
});
```

**Test Coverage Goals:**
- Unit tests: 80%+ coverage
- Integration tests: Key user flows
- E2E tests: Critical paths (login, trading, monitoring)
- Visual regression: All major pages
- Performance: Lighthouse scores > 90

---

## 7. Deployment

### Build Process

```json
// package.json scripts
{
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview",
    "test": "vitest",
    "test:e2e": "playwright test",
    "test:coverage": "vitest --coverage",
    "lint": "eslint . --ext js,jsx --report-unused-disable-directives --max-warnings 0",
    "format": "prettier --write \"src/**/*.{js,jsx,css}\""
  }

  NOTE: No TypeScript compilation needed!
}
```

### Production Build Configuration

```javascript
// vite.config.js
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { visualizer } from 'rollup-plugin-visualizer';

export default defineConfig({
  plugins: [
    react(),
    visualizer({
      filename: './dist/stats.html',
      open: false,
      gzipSize: true,
      brotliSize: true
    })
  ],
  build: {
    target: 'esnext',
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true
      }
    },
    rollupOptions: {
      output: {
        manualChunks: {
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          'charts': ['echarts', 'echarts-for-react', 'lightweight-charts'],
          'ui': ['lucide-react', 'sonner']
        }
      }
    },
    chunkSizeWarningLimit: 1000
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true
      }
    }
  }
});
```

### Dockerfile

```dockerfile
# Multi-stage build for optimized production image
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json package-lock.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Build application
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy built assets from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost/health || exit 1

CMD ["nginx", "-g", "daemon off;"]
```

### Nginx Configuration

```nginx
# nginx.conf
server {
    listen 80;
    server_name _;

    root /usr/share/nginx/html;
    index index.html;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1000;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' ws: wss:;" always;

    # Cache static assets
    location ~* \.(js|css|png|jpg|jpeg|gif|svg|ico|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # WebSocket proxy
    location /ws {
        proxy_pass http://dashboard-server:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 86400;
    }

    # API proxy
    location /api {
        proxy_pass http://dashboard-server:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Health check endpoint
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }

    # SPA routing - serve index.html for all routes
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

### Environment Configuration

```javascript
// src/config/env.js
/**
 * @typedef {Object} Config
 * @property {string} websocketUrl
 * @property {string} apiUrl
 * @property {'development'|'production'} environment
 * @property {boolean} enableDevTools
 */

/** @type {Config} */
export const config = {
  websocketUrl: import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws',
  apiUrl: import.meta.env.VITE_API_URL || 'http://localhost:8080/api',
  environment: import.meta.env.MODE as 'development' | 'production',
  enableDevTools: import.meta.env.DEV
};
```

```bash
# .env.production
VITE_WS_URL=wss://dashboard.example.com/ws
VITE_API_URL=https://api.example.com
```

### Docker Compose Integration

```yaml
# docker-compose.yml (excerpt)
services:
  web-dashboard:
    build:
      context: ./web-dashboard
      dockerfile: Dockerfile
    container_name: web-dashboard
    ports:
      - "3000:80"
    environment:
      - VITE_WS_URL=ws://dashboard-server:8080/ws
      - VITE_API_URL=http://dashboard-server:8080/api
    networks:
      - trading-net
    depends_on:
      - dashboard-server
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost/health"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 10s
    restart: unless-stopped

networks:
  trading-net:
    driver: bridge
```

### Static Hosting Alternatives

**Option 1: Nginx (Recommended)**
- Full control over caching and headers
- WebSocket proxy configuration
- Included in Dockerfile above

**Option 2: Caddy**
```dockerfile
FROM caddy:alpine
COPY --from=builder /app/dist /usr/share/caddy
COPY Caddyfile /etc/caddy/Caddyfile
```

```caddyfile
# Caddyfile
:80 {
    root * /usr/share/caddy
    file_server

    @api {
        path /api/*
    }
    reverse_proxy @api dashboard-server:8080

    @ws {
        path /ws
    }
    reverse_proxy @ws dashboard-server:8080

    try_files {path} /index.html
}
```

**Option 3: Vercel/Netlify (Cloud)**
- Automatic CDN distribution
- Easy deployment from Git
- Limited backend proxy capabilities
- Use `vercel.json` or `netlify.toml` for SPA routing

---

## 8. Observability

### Built-in Connection Status

```typescript
// src/components/ConnectionMonitor.tsx
import { useEffect, useState } from 'react';
import { Activity, AlertTriangle } from 'lucide-react';
import { Alert, AlertDescription, AlertTitle } from './ui/alert';

export function ConnectionMonitor() {
  const [connectionQuality, setConnectionQuality] = useState<'good' | 'poor' | 'bad'>('good');
  const latency = useTradingStore(state => state.latency);
  const status = useTradingStore(state => state.status);

  useEffect(() => {
    if (status !== 'connected') {
      setConnectionQuality('bad');
      return;
    }

    if (latency < 100) {
      setConnectionQuality('good');
    } else if (latency < 500) {
      setConnectionQuality('poor');
    } else {
      setConnectionQuality('bad');
    }
  }, [latency, status]);

  if (connectionQuality === 'good') return null;

  return (
    <Alert variant={connectionQuality === 'bad' ? 'destructive' : 'default'}>
      <AlertTriangle className="h-4 w-4" />
      <AlertTitle>Connection Issue</AlertTitle>
      <AlertDescription>
        {status !== 'connected'
          ? 'Disconnected from server. Attempting to reconnect...'
          : `High latency detected (${latency}ms). Data updates may be delayed.`
        }
      </AlertDescription>
    </Alert>
  );
}
```

### Latency Indicators

```typescript
// src/components/LatencyIndicator.tsx
import { useMemo } from 'react';
import { Activity } from 'lucide-react';
import { cn } from '../lib/utils';

interface LatencyIndicatorProps {
  latency: number;
}

export function LatencyIndicator({ latency }: LatencyIndicatorProps) {
  const { color, label } = useMemo(() => {
    if (latency < 50) return { color: 'text-green-500', label: 'Excellent' };
    if (latency < 100) return { color: 'text-green-400', label: 'Good' };
    if (latency < 200) return { color: 'text-yellow-500', label: 'Fair' };
    if (latency < 500) return { color: 'text-orange-500', label: 'Poor' };
    return { color: 'text-red-500', label: 'Bad' };
  }, [latency]);

  return (
    <div className="flex items-center gap-2">
      <Activity className={cn('h-4 w-4', color)} />
      <span className="text-sm">
        {latency}ms <span className="text-muted-foreground">({label})</span>
      </span>
    </div>
  );
}
```

### Error Boundary and Monitoring

```typescript
// src/components/ErrorBoundary.tsx
import React, { Component, ErrorInfo, ReactNode } from 'react';
import { AlertTriangle } from 'lucide-react';
import { Button } from './ui/button';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);

    // Send to error monitoring service (e.g., Sentry)
    if (window.Sentry) {
      window.Sentry.captureException(error, {
        contexts: {
          react: {
            componentStack: errorInfo.componentStack
          }
        }
      });
    }
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null });
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex items-center justify-center min-h-screen p-4">
          <div className="max-w-md w-full space-y-4 text-center">
            <AlertTriangle className="h-16 w-16 text-red-500 mx-auto" />
            <h1 className="text-2xl font-bold">Something went wrong</h1>
            <p className="text-muted-foreground">
              {this.state.error?.message || 'An unexpected error occurred'}
            </p>
            <Button onClick={this.handleReset}>
              Try Again
            </Button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
```

### Performance Monitoring

```typescript
// src/utils/performance.ts
export function measureWebVitals() {
  if (typeof window === 'undefined') return;

  // First Contentful Paint
  const observer = new PerformanceObserver((list) => {
    for (const entry of list.getEntries()) {
      if (entry.name === 'first-contentful-paint') {
        console.log('FCP:', entry.startTime);
        sendMetric('fcp', entry.startTime);
      }
    }
  });

  observer.observe({ entryTypes: ['paint'] });

  // Cumulative Layout Shift
  let clsValue = 0;
  const clsObserver = new PerformanceObserver((list) => {
    for (const entry of list.getEntries() as any[]) {
      if (!entry.hadRecentInput) {
        clsValue += entry.value;
      }
    }
  });

  clsObserver.observe({ entryTypes: ['layout-shift'] });

  // Send CLS on page hide
  window.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'hidden') {
      console.log('CLS:', clsValue);
      sendMetric('cls', clsValue);
    }
  });
}

function sendMetric(name: string, value: number) {
  // Send to analytics service
  if (navigator.sendBeacon) {
    const data = JSON.stringify({ metric: name, value, timestamp: Date.now() });
    navigator.sendBeacon('/api/metrics', data);
  }
}

// Usage in App.tsx
useEffect(() => {
  measureWebVitals();
}, []);
```

### Client-Side Error Logging

```typescript
// src/utils/errorLogger.ts
export class ErrorLogger {
  private static instance: ErrorLogger;
  private errors: Array<{ timestamp: number; message: string; stack?: string }> = [];

  private constructor() {
    this.setupGlobalHandlers();
  }

  static getInstance(): ErrorLogger {
    if (!ErrorLogger.instance) {
      ErrorLogger.instance = new ErrorLogger();
    }
    return ErrorLogger.instance;
  }

  private setupGlobalHandlers() {
    window.addEventListener('error', (event) => {
      this.log({
        timestamp: Date.now(),
        message: event.message,
        stack: event.error?.stack
      });
    });

    window.addEventListener('unhandledrejection', (event) => {
      this.log({
        timestamp: Date.now(),
        message: `Unhandled Promise Rejection: ${event.reason}`,
        stack: event.reason?.stack
      });
    });
  }

  log(error: { timestamp: number; message: string; stack?: string }) {
    this.errors.push(error);

    // Keep only last 100 errors
    if (this.errors.length > 100) {
      this.errors.shift();
    }

    // Send to backend
    fetch('/api/errors', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(error)
    }).catch(() => {
      // Silently fail if error reporting fails
    });
  }

  getErrors() {
    return this.errors;
  }
}

// Initialize in App.tsx
useEffect(() => {
  ErrorLogger.getInstance();
}, []);
```

---

## Summary

This comprehensive development plan provides:

1. **Modern Technology Stack**: React 18, Vite, JavaScript (ES6+), Tailwind CSS, Zustand, TanStack Query (NO TypeScript)
2. **Scalable Architecture**: Component hierarchy, WebSocket integration, state management
3. **Phased Development**: 9 phases over 4 weeks with clear deliverables
4. **Production-Ready Implementation**: WebSocket hooks, chart components, trading forms
5. **Comprehensive Testing**: Unit, integration, E2E, visual regression, performance
6. **Optimized Deployment**: Multi-stage Docker build, Nginx configuration, static hosting
7. **Built-in Observability**: Connection monitoring, latency tracking, error boundaries

**Key Deliverables:**
- Real-time trading dashboard with sub-second updates
- Mobile-responsive design (mobile-first approach)
- Order management and manual trading interface
- Advanced charting and analytics
- Production-ready containerized deployment
- 80%+ test coverage with E2E automation
- Lighthouse score > 90
- Bundle size < 500KB gzipped
- WCAG 2.1 AA accessibility compliance

**Development Timeline:** 4 weeks (24 working days)

**Next Steps:**
1. Initialize project with Vite + React + JavaScript (NO TypeScript!)
2. Setup development environment and tooling
3. Begin Phase 1: Project setup and layout
4. Follow phased development plan sequentially
5. Remember: All code must be in pure JavaScript (ES6+) with JSDoc for documentation

---

**End of Web Dashboard Service Development Plan**
