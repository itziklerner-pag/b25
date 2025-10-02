# Web Dashboard

Modern web-based dashboard for HFT trading system.

**Stack**: React 18 + TypeScript + Vite  
**Development Plan**: `../../docs/service-plans/10-web-dashboard-service.md`

## Quick Start
```bash
npm install
npm run dev
```

## Access
http://localhost:3000

## Features
- Real-time WebSocket updates (250ms)
- TradingView charts
- Mobile-responsive design
- Manual trading interface
- P&L analytics

## Building
```bash
# Development
npm run dev

# Production build
npm run build

# Preview production build
npm run preview
```

## Testing
```bash
# Unit tests
npm run test

# E2E tests
npm run test:e2e

# Coverage
npm run test:coverage
```

## Deployment
```bash
# Docker
docker build -t b25/web-dashboard .
docker run -p 3000:80 b25/web-dashboard
```
