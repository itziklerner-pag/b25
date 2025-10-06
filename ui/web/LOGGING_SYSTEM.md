# B25 Web Dashboard - Professional Logging System

## Overview

A comprehensive, production-ready logging system with configurable log levels, color-coded output, and runtime controls via the Debug Panel.

## Features

### Core Logger (`/src/utils/logger.ts`)

- **5 Log Levels**: ERROR, WARN, INFO, DEBUG, TRACE
- **Color-Coded Output**: Different colors for each log level in development
- **Timestamps**: Precise timestamps on all logs (HH:MM:SS.mmm format)
- **Context/Module Support**: Organize logs by module (e.g., `[WebSocket]`, `[Store]`)
- **Environment-Based Defaults**:
  - Production: WARN level (only errors and warnings)
  - Development: DEBUG level (verbose logging)
- **Runtime Configuration**: Change log level via Debug Panel or localStorage
- **Performance Optimized**: Zero overhead when log level is disabled
- **Singleton Pattern**: Single logger instance across the app

### Log Levels

| Level | Value | Description | When to Use |
|-------|-------|-------------|-------------|
| ERROR | 0 | Critical errors that need immediate attention | Exceptions, failed operations, fatal errors |
| WARN | 1 | Warning conditions that should be reviewed | Deprecations, fallbacks, potential issues |
| INFO | 2 | General informational messages | App initialization, major state changes |
| DEBUG | 3 | Detailed debugging information | Function calls, data updates, flow tracking |
| TRACE | 4 | Very detailed trace information | Component renders, minor state changes |

### Usage Examples

```typescript
import { logger } from '@/utils/logger';

// With context and message
logger.error('WebSocket', 'Connection failed', error);
logger.warn('Store', 'Update slow', { duration: 100 });
logger.info('WebSocket', 'Connected to server');
logger.debug('Store', 'Market data updated', { symbol, price });
logger.trace('Component', 'Render cycle', { props });

// Without context (simple logging)
logger.error('Failed to fetch data', error);
logger.info('Application started');
```

## Configuration

### Environment Variables (`.env`)

```bash
# Production default (only errors and warnings)
VITE_LOG_LEVEL=WARN

# Development default (verbose debugging)
VITE_LOG_LEVEL=DEBUG
```

### Runtime Configuration

**Via Debug Panel:**
1. Press `Ctrl+D` (or `Cmd+D` on Mac) to open Debug Panel
2. Click desired log level button (ERROR, WARN, INFO, DEBUG, TRACE)
3. Changes apply immediately
4. Settings persist in localStorage

**Via Code:**
```typescript
import { logger } from '@/utils/logger';

// Set log level programmatically
logger.setLevel('DEBUG');
logger.setLevel('INFO');

// Get current level
const currentLevel = logger.getLevel(); // Returns LogLevel enum
const levelString = logger.getLevelString(); // Returns 'DEBUG', 'INFO', etc.
```

**Via Browser Console:**
```javascript
// Change log level from browser console
localStorage.setItem('b25_log_level', 'TRACE');
```

## Files Updated

### New Files
- `/src/utils/logger.ts` - Core logging system
- `/src/.env.development` - Development environment config
- `LOGGING_SYSTEM.md` - This documentation

### Updated Files
1. **Core Modules:**
   - `/src/hooks/useWebSocket.ts` - WebSocket logging
   - `/src/store/trading.ts` - Store action logging

2. **Components:**
   - `/src/components/MarketPrices.tsx` - Market data logging
   - `/src/components/DebugPanel.tsx` - Log level controls
   - `/src/components/ErrorBoundary.tsx` - Error logging
   - `/src/components/ServiceMonitor.tsx` - Service health logging

3. **Pages:**
   - `/src/App.tsx` - App initialization logging
   - `/src/pages/DashboardPage.tsx` - Dashboard logging
   - `/src/pages/TradingPage.tsx` - Trading operations logging

4. **Config:**
   - `/.env` - Added VITE_LOG_LEVEL

## Debug Panel Controls

### New Features

1. **Log Level Selection:**
   - 5 buttons for each log level
   - Active level is highlighted in color
   - Shows current level and what it logs

2. **Clear Console:**
   - Button to clear browser console
   - Useful for starting fresh debugging session

3. **Visual Feedback:**
   - ERROR: Red background
   - WARN: Amber background
   - INFO: Blue background
   - DEBUG: Violet background
   - TRACE: Gray background

### Keyboard Shortcut
- `Ctrl+D` or `Cmd+D` - Toggle Debug Panel

## Default Log Levels

### Production (`.env`)
```bash
VITE_LOG_LEVEL=WARN
```
- **Logs**: Errors and warnings only
- **Purpose**: Minimize console noise in production
- **Performance**: Maximum performance (most logs disabled)

### Development (`.env.development`)
```bash
VITE_LOG_LEVEL=DEBUG
```
- **Logs**: Everything except TRACE
- **Purpose**: Comprehensive debugging without excessive detail
- **Performance**: Acceptable for development

### Override Priority
1. **localStorage** (`b25_log_level`) - Highest priority (user override)
2. **Environment variable** (`VITE_LOG_LEVEL`) - Medium priority
3. **Default** (WARN for prod, DEBUG for dev) - Lowest priority

## How to Change Log Levels

### Method 1: Debug Panel (Recommended)
1. Press `Ctrl+D` to open Debug Panel
2. Click desired log level button
3. Changes apply immediately and persist across page refreshes

### Method 2: Environment Variable
Edit `.env` or `.env.development`:
```bash
VITE_LOG_LEVEL=INFO  # Change to desired level
```
Restart dev server for changes to take effect.

### Method 3: Browser Console
```javascript
// Temporary change (until page refresh)
localStorage.setItem('b25_log_level', 'TRACE');
location.reload();
```

### Method 4: Code
```typescript
import { logger } from '@/utils/logger';
logger.setLevel('DEBUG');
```

## Color Coding (Development Only)

| Level | Text Color | Background Color |
|-------|-----------|------------------|
| ERROR | #ef4444 (red-500) | #fee2e2 (red-100) |
| WARN | #f59e0b (amber-500) | #fef3c7 (amber-100) |
| INFO | #3b82f6 (blue-500) | #dbeafe (blue-100) |
| DEBUG | #8b5cf6 (violet-500) | #ede9fe (violet-100) |
| TRACE | #6b7280 (gray-500) | #f3f4f6 (gray-100) |

*Note: Colors are disabled in production for performance*

## Performance Considerations

1. **Conditional Logging**: Logs only execute if level is enabled
2. **No Overhead**: Disabled logs have zero performance impact
3. **Production Mode**: Simple console output without styling overhead
4. **Development Mode**: Rich colored output with context

## Best Practices

### When to Use Each Level

**ERROR** - Only for actual errors:
```typescript
logger.error('WebSocket', 'Connection failed', error);
logger.error('API', 'Failed to fetch data', { status: 500 });
```

**WARN** - Potential issues or deprecations:
```typescript
logger.warn('Store', 'Slow update detected', { duration: 200 });
logger.warn('Config', 'Using fallback configuration');
```

**INFO** - Important application events:
```typescript
logger.info('App', 'Dashboard initialized');
logger.info('WebSocket', 'Connected to server');
logger.info('Store', 'Selected symbol changed', { symbol: 'BTCUSDT' });
```

**DEBUG** - Detailed debugging information:
```typescript
logger.debug('Store', 'Market data updated', { symbol, price });
logger.debug('WebSocket', 'Received message', { type: message.type });
```

**TRACE** - Very detailed trace (usually disabled):
```typescript
logger.trace('Component', 'Render cycle', { props });
logger.trace('WebSocket', 'Heartbeat sent');
```

### Context Naming Conventions

- Use PascalCase for contexts: `WebSocket`, `Store`, `MarketPrices`
- Keep contexts short and descriptive
- Be consistent across related files

## Troubleshooting

### Logs Not Appearing

1. Check current log level: Open Debug Panel and see active level
2. Ensure your log is at or above current level
3. Check browser console settings (filter might be active)

### Too Many Logs

1. Lower the log level to WARN or ERROR
2. Use Debug Panel to change level without code changes
3. In production, ensure `VITE_LOG_LEVEL=WARN`

### Lost Log Settings

- Log level is stored in localStorage (`b25_log_level`)
- Clearing localStorage will reset to environment default
- Set environment variable for persistent default

## Future Enhancements

Potential improvements for the logging system:

1. **Remote Logging**: Send logs to external service (e.g., Sentry, LogRocket)
2. **Log History**: Store recent logs for viewing in Debug Panel
3. **Log Export**: Download logs as JSON/CSV
4. **Log Filtering**: Filter by context/module in Debug Panel
5. **Log Search**: Search through log history
6. **Performance Metrics**: Track log volume and performance impact

## Migration from console.log

Old code:
```typescript
console.log('[WebSocket] Connected');
console.error('[WebSocket] Error:', error);
console.warn('[Store] Slow update');
```

New code:
```typescript
logger.info('WebSocket', 'Connected');
logger.error('WebSocket', 'Connection error', error);
logger.warn('Store', 'Slow update');
```

## Summary

The B25 Web Dashboard now has a professional, production-ready logging system with:

✅ 5 configurable log levels (ERROR, WARN, INFO, DEBUG, TRACE)
✅ Color-coded console output in development
✅ Runtime log level changes via Debug Panel
✅ Environment-based defaults (WARN for prod, DEBUG for dev)
✅ localStorage persistence for user preferences
✅ Zero performance overhead when disabled
✅ Context/module organization
✅ Timestamps on all logs

**Default Levels:**
- **Production**: `WARN` (errors and warnings only)
- **Development**: `DEBUG` (all logs except trace)

**How to Change:**
- Press `Ctrl+D` to open Debug Panel
- Click desired log level
- Settings persist automatically
