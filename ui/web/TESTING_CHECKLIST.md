# Testing Checklist - Market Data Display Fix

## Quick Verification Steps

### 1. Start the Development Server
```bash
cd /home/mm/dev/b25/ui/web
npm run dev
```

### 2. Open Browser DevTools

**Console Tab** - You should see:
```
[WebSocket] Connecting to: ws://localhost:8080/ws
[WebSocket] Connected
[WebSocket] Received market_data: {BTCUSDT: {...}, ETHUSDT: {...}}
[Store] Updated market data for BTCUSDT: {last_price: 123360.05}
[Store] Updated market data for ETHUSDT: {last_price: 4499.18}
[MarketPrices] Rendering with data: [...]
```

**Network Tab** - WebSocket connection:
- Status: 101 Switching Protocols
- Messages tab should show incoming JSON with `market_data` field

### 3. Visual Verification

#### Dashboard Page (/)
- [ ] "Market Prices" card displays
- [ ] BTC price shows $123,360.05 (or current live price)
- [ ] ETH price shows $4,499.18 (or current live price)
- [ ] Prices update every ~1 second
- [ ] Green/red arrows show for price changes
- [ ] "Last update" timestamp updates

#### Trading Page (/trading)
- [ ] "Current Market Price" banner displays at top
- [ ] Price matches Dashboard display
- [ ] "Use Market Price" button appears for limit orders
- [ ] Clicking button fills price field with current market price
- [ ] Price updates in real-time

#### Account Stats (all pages)
- [ ] Balance shows correctly (e.g., $10,000)
- [ ] Equity displays
- [ ] Unrealized P&L updates

### 4. Redux DevTools Check

Open Redux DevTools (if installed) or use Zustand DevTools:

```javascript
// In browser console:
window.__ZUSTAND_STORE__ = useTradingStore.getState();
console.log(window.__ZUSTAND_STORE__.marketData);
```

Should show:
```javascript
Map(2) {
  'BTCUSDT' => { last_price: 123360.05, ... },
  'ETHUSDT' => { last_price: 4499.18, ... }
}
```

### 5. Manual Store Inspection

In browser console:
```javascript
// Check if market data is in store
const state = useTradingStore.getState();
console.log('Market Data:', Array.from(state.marketData.entries()));
console.log('Last Update:', new Date(state.lastUpdate).toLocaleString());
```

### 6. Component Re-render Test

The MarketPrices component should re-render when data updates:

```javascript
// In browser console, watch for this log every ~1 second:
// [MarketPrices] Rendering with data: [...]
```

## Common Issues & Solutions

### Issue: "Waiting for market data..." message persists

**Possible Causes:**
1. WebSocket not connected
2. Backend not sending market_data
3. Data structure mismatch

**Debug:**
```javascript
// Check WebSocket status
useTradingStore.getState().status // should be 'connected'

// Check if data is being received
// Look for "[WebSocket] Received market_data" in console
```

### Issue: Prices show $0.00

**Possible Causes:**
1. formatCurrency() receiving undefined
2. Data not in store

**Debug:**
```javascript
// Check raw data
const marketData = useTradingStore.getState().marketData;
console.log(marketData.get('BTCUSDT')); // Should show {last_price: ...}
```

### Issue: Component not updating

**Possible Causes:**
1. Zustand selector not detecting changes
2. Map reference not updating

**Debug:**
```javascript
// Check lastUpdate is changing
setInterval(() => {
  console.log('Last Update:', useTradingStore.getState().lastUpdate);
}, 1000);
```

## Performance Check

### Expected Metrics:
- **WebSocket Latency**: < 100ms
- **Component Re-render Time**: < 16ms (60fps)
- **Bundle Size**: ~1MB (acceptable for charts)
- **Memory Usage**: Stable (no leaks)

### Monitor in DevTools:
```javascript
// Performance tab
// Record during 30 seconds of live updates
// Check for:
// - Smooth FPS (should be 60fps)
// - No memory leaks (heap should stabilize)
// - No layout thrashing
```

## Success Criteria

✅ **All of these should be true:**

1. Dashboard displays live BTC/ETH prices
2. Prices update automatically every 1-2 seconds
3. No console errors
4. WebSocket status shows "connected"
5. Trading page shows current market price
6. "Use Market Price" button works
7. Prices match between Dashboard and Trading pages
8. Console shows `[Store] Updated market data for...` logs
9. Redux/Zustand DevTools shows populated marketData Map
10. No TypeScript errors (`npm run build` succeeds)

## Before/After Comparison

### BEFORE (Broken):
- Market prices: $0.00
- Console: `{market_data: {BTCUSDT: {last_price: 123360.05}}}`
- Store: orderbooks Map has invalid data
- Components: Reading orderbook.bids (undefined)

### AFTER (Fixed):
- Market prices: $123,360.05 (live)
- Console: `[Store] Updated market data for BTCUSDT`
- Store: marketData Map has proper MarketData objects
- Components: Reading marketData.get('BTCUSDT').last_price
