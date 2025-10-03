# Quick Start Guide

Get the B25 Web Dashboard running in 3 minutes.

## Prerequisites

- Node.js 20+ installed
- Dashboard Server running on port 8080 (or update `.env`)

## Step-by-Step Setup

### 1. Install Dependencies

```bash
cd ui/web
npm install
```

This will install all required packages (~500MB, takes 2-3 minutes).

### 2. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` if your Dashboard Server is on a different host:

```env
VITE_WS_URL=ws://localhost:8080/ws
VITE_API_URL=http://localhost:8080/api
VITE_AUTH_URL=http://localhost:3001
```

### 3. Start Development Server

```bash
npm run dev
```

Output should show:
```
VITE v5.2.0  ready in 500 ms

➜  Local:   http://localhost:3000/
➜  Network: use --host to expose
```

### 4. Open in Browser

Navigate to: **http://localhost:3000**

You should see:
- Login page (if auth is enabled)
- Dashboard with connection status indicator
- Real-time updates if WebSocket connects successfully

## Troubleshooting

### WebSocket Not Connecting

**Symptom**: Connection status shows "Disconnected" or "Connection Error"

**Solutions**:
1. Check Dashboard Server is running: `curl http://localhost:8080/health`
2. Verify WebSocket URL in `.env`
3. Check browser console for error messages
4. Ensure no firewall blocking WebSocket connections

### Build Errors

**Symptom**: `npm install` fails or `npm run dev` won't start

**Solutions**:
```bash
# Clear everything and start fresh
rm -rf node_modules package-lock.json
npm cache clean --force
npm install
```

### Port 3000 Already in Use

**Solutions**:
```bash
# Use different port
npm run dev -- --port 3001

# Or kill process on port 3000
lsof -ti:3000 | xargs kill -9
```

### TypeScript Errors

**Symptom**: Red underlines in IDE or build errors

**Solutions**:
```bash
# Restart TypeScript server in VS Code
# Command Palette > TypeScript: Restart TS Server

# Or check types
npm run type-check
```

## Development Workflow

### Making Changes

1. Edit files in `src/`
2. Vite will hot-reload automatically
3. Check browser console for errors
4. Use React DevTools for debugging

### Adding New Pages

1. Create component in `src/pages/NewPage.tsx`
2. Add route in `src/App.tsx`
3. Add navigation link in `src/components/Layout.tsx`

### Adding New Components

1. Create in `src/components/`
2. Import and use in pages
3. Add tests in `src/components/*.test.tsx`

## Testing

### Run Unit Tests
```bash
npm run test
```

### Run E2E Tests
```bash
# Install Playwright browsers (first time only)
npx playwright install

# Run tests
npm run test:e2e
```

### Check Code Quality
```bash
# Linting
npm run lint

# Formatting
npm run format

# Type checking
npm run type-check
```

## Building for Production

### Local Build
```bash
npm run build
npm run preview
```

### Docker Build
```bash
docker build -t b25/web-dashboard .
docker run -p 3000:80 b25/web-dashboard
```

## Common Tasks

### Update Dependencies
```bash
npm update
```

### Check Bundle Size
```bash
npm run build
# Open dist/stats.html
```

### Clean Build
```bash
rm -rf dist .vite
npm run build
```

## Key Files to Know

- `src/App.tsx` - Main application with routing
- `src/store/trading.ts` - Trading state management
- `src/hooks/useWebSocket.ts` - WebSocket connection
- `src/components/Layout.tsx` - Main layout
- `src/pages/DashboardPage.tsx` - Main dashboard

## Useful Commands

```bash
# Development
npm run dev              # Start dev server
npm run build            # Production build
npm run preview          # Preview production build

# Testing
npm run test             # Unit tests
npm run test:e2e         # E2E tests
npm run test:coverage    # Coverage report

# Code Quality
npm run lint             # ESLint
npm run format           # Prettier
npm run type-check       # TypeScript

# Docker
docker build -t b25/web-dashboard .
docker run -p 3000:80 b25/web-dashboard
```

## Getting Help

1. Check `README.md` for full documentation
2. Check `IMPLEMENTATION.md` for technical details
3. Check browser console for errors
4. Check `../../docs/service-plans/10-web-dashboard-service.md` for design docs

## Success Checklist

- [ ] Dependencies installed (`node_modules/` exists)
- [ ] Environment configured (`.env` file exists)
- [ ] Dev server starts without errors
- [ ] Browser opens to http://localhost:3000
- [ ] Connection status shows (Connecting or Connected)
- [ ] Dashboard displays account stats
- [ ] Navigation between pages works
- [ ] Dark mode toggle works

If all checks pass, you're ready to trade! 🎉

---

**Time to get running**: ~3 minutes
**Time to first contribution**: ~15 minutes
