# Web Dashboard

Modern web-based dashboard for B25 high-frequency trading system.

**Stack**: React 18 + TypeScript + Vite
**Development Plan**: `../../docs/service-plans/10-web-dashboard-service.md`

## Features

- **Real-time Updates**: WebSocket connection with 250ms refresh rate
- **Trading Interface**: Manual order placement with limit/market orders
- **Live Order Book**: Real-time order book visualization with depth charts
- **Position Management**: Track and close positions with P&L breakdown
- **Analytics Dashboard**: P&L charts and performance metrics
- **System Monitoring**: Service health and latency tracking
- **Responsive Design**: Mobile-first design with dark mode support
- **Auto-Reconnect**: Automatic WebSocket reconnection with exponential backoff

## Tech Stack

- **Framework**: React 18 with TypeScript
- **Build Tool**: Vite 5
- **State Management**: Zustand (WebSocket state) + TanStack Query (REST API)
- **Styling**: Tailwind CSS with shadcn/ui components
- **Charts**: Apache ECharts for advanced visualizations
- **Routing**: React Router v6
- **Testing**: Vitest (unit) + Playwright (E2E)

## Quick Start

### Prerequisites
- Node.js 20+
- npm or yarn

### Installation

```bash
# Install dependencies
npm install

# Copy environment file
cp .env.example .env

# Start development server
npm run dev
```

Access the dashboard at: http://localhost:3000

### Environment Variables

Create a `.env` file with the following variables:

```env
VITE_WS_URL=ws://localhost:8080/ws
VITE_API_URL=http://localhost:8080/api
VITE_AUTH_URL=http://localhost:3001
```

## Development

### Available Scripts

```bash
# Development server with hot reload
npm run dev

# Type checking
npm run type-check

# Linting
npm run lint

# Format code
npm run format

# Build for production
npm run build

# Preview production build
npm run preview
```

## Testing

### Unit Tests

```bash
# Run unit tests
npm run test

# Run with coverage
npm run test:coverage

# Watch mode
npm run test -- --watch
```

### E2E Tests

```bash
# Run E2E tests
npm run test:e2e

# Run in UI mode
npx playwright test --ui

# Run specific browser
npx playwright test --project=chromium
```

## Building

### Production Build

```bash
# Build optimized production bundle
npm run build

# Output will be in ./dist directory
```

### Bundle Analysis

```bash
# Build and view bundle stats
npm run build

# Open dist/stats.html to analyze bundle size
```

## Deployment

### Docker

```bash
# Build Docker image
docker build -t b25/web-dashboard .

# Run container
docker run -p 3000:80 \
  -e VITE_WS_URL=ws://your-server:8080/ws \
  -e VITE_API_URL=http://your-server:8080/api \
  b25/web-dashboard

# Access at http://localhost:3000
```

### Docker Compose

The web dashboard is included in the main docker-compose setup:

```bash
# From repository root
docker-compose -f docker/docker-compose.dev.yml up web-dashboard
```

### Static Hosting

The built application is a static SPA that can be hosted on:
- Nginx (included in Dockerfile)
- Vercel
- Netlify
- AWS S3 + CloudFront
- Any static file server

## Project Structure

```
ui/web/
├── src/
│   ├── components/          # Reusable UI components
│   │   ├── ui/             # Base UI components (shadcn)
│   │   ├── Layout.tsx      # Main layout with sidebar
│   │   ├── ConnectionStatus.tsx
│   │   ├── ErrorBoundary.tsx
│   │   └── ThemeProvider.tsx
│   ├── pages/              # Page components
│   │   ├── DashboardPage.tsx
│   │   ├── PositionsPage.tsx
│   │   ├── OrdersPage.tsx
│   │   ├── OrderBookPage.tsx
│   │   ├── AnalyticsPage.tsx
│   │   ├── TradingPage.tsx
│   │   ├── SystemPage.tsx
│   │   └── LoginPage.tsx
│   ├── hooks/              # Custom React hooks
│   │   └── useWebSocket.ts # WebSocket connection hook
│   ├── store/              # Zustand state management
│   │   └── trading.ts      # Trading state store
│   ├── lib/                # Utility functions
│   │   └── utils.ts        # Helper functions
│   ├── types/              # TypeScript type definitions
│   │   └── index.ts
│   ├── config/             # Configuration
│   │   └── env.ts          # Environment config
│   ├── test/               # Test utilities
│   │   └── setup.ts
│   ├── App.tsx             # Root component
│   ├── main.tsx            # Entry point
│   └── index.css           # Global styles
├── tests/
│   └── e2e/                # E2E tests
│       └── dashboard.spec.ts
├── public/                 # Static assets
├── Dockerfile              # Production Docker image
├── nginx.conf              # Nginx configuration
├── vite.config.ts          # Vite configuration
├── tailwind.config.ts      # Tailwind configuration
├── tsconfig.json           # TypeScript configuration
├── vitest.config.ts        # Vitest configuration
├── playwright.config.ts    # Playwright configuration
└── package.json
```

## WebSocket Protocol

The dashboard connects to the Dashboard Server via WebSocket on `/ws`.

### Message Format

```typescript
// Outgoing messages
{
  "type": "subscribe" | "unsubscribe" | "order" | "ping",
  "channel"?: string,
  "channels"?: string[],
  "action"?: "create" | "cancel" | "close",
  "data"?: any
}

// Incoming messages
{
  "type": "pong" | "account" | "position" | "order" | "trade" | "orderbook" | "system_health",
  "data": any,
  "timestamp"?: number
}
```

### Subscriptions

- `account` - Account balance and equity updates
- `positions` - Position updates
- `orders` - Order status updates
- `trades` - Recent trades
- `system_health` - Service health metrics

## Performance Targets

- First Contentful Paint: < 1.8s
- Time to Interactive: < 3.9s
- Bundle Size: < 500KB gzipped
- WebSocket Latency: < 100ms
- UI Update Rate: 250ms (configurable)

## Browser Support

- Chrome/Edge (latest 2 versions)
- Firefox (latest 2 versions)
- Safari (latest 2 versions)
- Mobile browsers (iOS Safari, Chrome Android)

## Accessibility

- WCAG 2.1 AA compliant
- Keyboard navigation support
- Screen reader compatible
- High contrast mode support

## Security

- Content Security Policy (CSP) enabled
- XSS protection headers
- Secure WebSocket connections (wss://) in production
- Authentication token management
- No sensitive data in localStorage

## Troubleshooting

### WebSocket Connection Issues

1. Check that Dashboard Server is running on port 8080
2. Verify VITE_WS_URL in `.env`
3. Check browser console for connection errors
4. Ensure firewall allows WebSocket connections

### Build Failures

```bash
# Clear cache and reinstall
rm -rf node_modules package-lock.json
npm install

# Clear Vite cache
rm -rf .vite
```

### Development Server Issues

```bash
# Check port availability
lsof -i :3000

# Try different port
npm run dev -- --port 3001
```

## Contributing

1. Follow the existing code style
2. Write tests for new features
3. Update documentation
4. Run linter before committing: `npm run lint`
5. Ensure all tests pass: `npm test && npm run test:e2e`

## License

See repository root for license information.

## Support

- Documentation: `../../docs/service-plans/10-web-dashboard-service.md`
- Issues: GitHub Issues
- Main README: `../../README.md`

---

**Built with React 18 + TypeScript + Vite for ultra-low latency trading.**
