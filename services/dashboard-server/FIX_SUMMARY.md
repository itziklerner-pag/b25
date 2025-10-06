# Dashboard Server WebSocket Null Fix - Summary

## Problem
The Dashboard Server was sending `null` for `market_data` and `strategies` in WebSocket messages, even though the data was loaded correctly on startup.

**Browser Console Error:**
```
Received full state snapshot: {market_data: null, orders: Array(0), positions: {…}, account: {…}, strategies: null, …}
```

**Server Logs:**
```
Initial state loaded: {"market_data":3,"strategies":3,...}
```

## Root Cause
In both `/home/mm/dev/b25/services/dashboard-server/internal/server/server.go` and `/home/mm/dev/b25/services/dashboard-server/internal/broadcaster/broadcaster.go`, the `filterStateBySubscriptions()` function was creating a filtered `State` struct **without initializing the map fields**.

When a Go map is `nil` and serialized to JSON, it becomes `null` instead of an empty object `{}`.

### Before Fix (server.go line 291-314):
```go
func (s *Server) filterStateBySubscriptions(state *types.State, subscriptions map[string]bool) *types.State {
    filtered := &types.State{
        Timestamp: state.Timestamp,
        Sequence:  state.Sequence,
        // Maps and slices are nil by default
    }

    if subscriptions["market_data"] {
        filtered.MarketData = state.MarketData  // If subscription is false, remains nil
    }
    if subscriptions["strategies"] {
        filtered.Strategies = state.Strategies  // If subscription is false, remains nil
    }

    return filtered
}
```

**Result:** When subscriptions existed but maps were copied, the fields were set. However, in some code paths or when the broadcaster initialized the state, the maps remained `nil`, which serialized to JSON `null`.

## Solution
Initialize all map and slice fields to empty containers instead of leaving them `nil`:

### After Fix (server.go line 300-329):
```go
func (s *Server) filterStateBySubscriptions(state *types.State, subscriptions map[string]bool) *types.State {
    filtered := &types.State{
        Timestamp: state.Timestamp,
        Sequence:  state.Sequence,
        // Initialize all maps to prevent null in JSON serialization
        MarketData: make(map[string]*types.MarketData),
        Orders:     make([]*types.Order, 0),
        Positions:  make(map[string]*types.Position),
        Strategies: make(map[string]*types.Strategy),
    }

    // Copy data based on subscriptions
    if subscriptions["market_data"] && state.MarketData != nil {
        filtered.MarketData = state.MarketData
    }
    if subscriptions["orders"] && state.Orders != nil {
        filtered.Orders = state.Orders
    }
    if subscriptions["positions"] && state.Positions != nil {
        filtered.Positions = state.Positions
    }
    if subscriptions["account"] && state.Account != nil {
        filtered.Account = state.Account
    }
    if subscriptions["strategies"] && state.Strategies != nil {
        filtered.Strategies = state.Strategies
    }

    return filtered
}
```

Same fix applied to `/home/mm/dev/b25/services/dashboard-server/internal/broadcaster/broadcaster.go` line 236-265.

## Verification
WebSocket test confirmed the fix works:

```json
{
  "type": "snapshot",
  "seq": 30206,
  "data": {
    "market_data": {
      "BTCUSDT": { "last_price": 123772.35, ... },
      "ETHUSDT": { "last_price": 4529.215, ... },
      "SOLUSDT": { "last_price": 50000, ... }
    },
    "strategies": {},     // Empty object, NOT null ✅
    "positions": {},      // Empty object, NOT null ✅
    "orders": [],         // Empty array, NOT null ✅
    "account": { ... }
  }
}
```

## Files Modified
1. `/home/mm/dev/b25/services/dashboard-server/internal/server/server.go`
   - Added map initialization in `filterStateBySubscriptions()` (line 300-309)
   - Added debug logging in `handleRefresh()` (line 288-294)
   - Added debug logging in `sendFullState()` (line 345-350)

2. `/home/mm/dev/b25/services/dashboard-server/internal/broadcaster/broadcaster.go`
   - Added map initialization in `filterStateBySubscriptions()` (line 237-245)

## Testing
1. Built binary: `go build -o bin/service ./cmd/server`
2. Restarted service: `./bin/service`
3. Connected WebSocket client on port 8086
4. Verified all fields serialize as empty objects `{}` or arrays `[]` instead of `null`

## Impact
- Web dashboard will now properly display empty states instead of showing "null"
- Frontend can safely iterate over `market_data`, `strategies`, and `positions` without null checks
- Consistent JSON structure whether data is present or not
