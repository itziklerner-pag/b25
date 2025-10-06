# Quick Start: B25 Logging System

## TL;DR

```typescript
import { logger } from '@/utils/logger';

// Basic usage
logger.error('WebSocket', 'Connection failed', error);
logger.warn('Store', 'Update slow', { duration: 100 });
logger.info('WebSocket', 'Connected');
logger.debug('Store', 'Data updated', { symbol, price });
logger.trace('Component', 'Render', { props });
```

## Change Log Level

**Fastest Way:**
1. Press `Ctrl+D` (or `Cmd+D`)
2. Click log level button (ERROR, WARN, INFO, DEBUG, TRACE)
3. Done! Changes apply immediately

## Current Defaults

- **Production**: `WARN` - Only errors and warnings
- **Development**: `DEBUG` - Everything except trace

## Examples from Codebase

### WebSocket Logging
```typescript
// /src/hooks/useWebSocket.ts
logger.info('WebSocket', 'Connecting to', { url });
logger.info('WebSocket', 'Connected successfully');
logger.warn('WebSocket', 'Disconnected', { code, reason });
logger.error('WebSocket', 'Connection error', error);
logger.debug('WebSocket', 'Received message', { type, channel });
logger.trace('WebSocket', 'Heartbeat sent');
```

### Store Logging
```typescript
// /src/store/trading.ts
logger.info('Store', 'Selected symbol changed', { symbol });
logger.debug('Store', 'Updated market data', { symbol, price });
logger.warn('Store', 'Cannot send order - not connected');
```

### Component Logging
```typescript
// /src/components/MarketPrices.tsx
logger.trace('MarketPrices', 'Component rendering', { count });
logger.debug('MarketPrices', 'Data updated', { pairs });
```

## Log Levels Explained

| Level | What Appears | Use For |
|-------|--------------|---------|
| ERROR | Errors only | Production troubleshooting |
| WARN | Errors + Warnings | Production default |
| INFO | Errors + Warnings + Info | General debugging |
| DEBUG | All except trace | Development default |
| TRACE | Everything | Deep debugging |

## Tips

1. **Production**: Set to WARN or ERROR to reduce noise
2. **Development**: Set to DEBUG for good visibility
3. **Deep Debugging**: Use TRACE temporarily for specific issues
4. **Performance**: Logs only execute when level is enabled (zero overhead when disabled)

## Debug Panel Features

Press `Ctrl+D` to access:
- **Log Level Control** - Change level with one click
- **Clear Console** - Start fresh debugging
- **WebSocket Status** - Connection and latency
- **Store Updates** - Data update tracking
- **Market Data** - Real-time price info

## Configuration Files

### Production (`.env`)
```bash
VITE_LOG_LEVEL=WARN
```

### Development (`.env.development`)
```bash
VITE_LOG_LEVEL=DEBUG
```

## Common Tasks

**See all WebSocket activity:**
1. Press `Ctrl+D`
2. Click `DEBUG`
3. Watch console for `[WebSocket]` logs

**Debug render issues:**
1. Press `Ctrl+D`
2. Click `TRACE`
3. Watch for component render logs

**Production debugging:**
1. Keep at `WARN` normally
2. Temporarily switch to `INFO` or `DEBUG` if needed
3. User settings persist in localStorage

That's it! Happy logging! 🚀
