# Dashboard Server Admin Page - Visual Preview

## Page Layout

```
┌──────────────────────────────────────────────────────────────────┐
│                                                                  │
│                    ⚡ Dashboard Server                          │
│         Real-time WebSocket Aggregation & Broadcasting          │
│                        ● Healthy                                 │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘

┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│ ⚡ SERVICE      │ │ 👥 CONNECTED    │ │ 📊 STATE        │ │ 📈 GOROUTINES   │
│    UPTIME       │ │    CLIENTS      │ │    SEQUENCE     │ │                 │
│                 │ │                 │ │                 │ │                 │
│   2h 30m 15s    │ │       3         │ │    12,456       │ │      45         │
│                 │ │                 │ │                 │ │                 │
│ Since 10:30 AM  │ │ Active WebSocket│ │ Total updates   │ │ Runtime         │
│                 │ │  connections    │ │                 │ │ concurrency     │
└─────────────────┘ └─────────────────┘ └─────────────────┘ └─────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│ ℹ️  SERVICE INFORMATION                                    ●     │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐            │
│  │ Service Name │ │ Version      │ │ Go Version   │            │
│  │ dashboard-   │ │ 1.0.0        │ │ go1.21.0     │            │
│  │ server       │ │              │ │              │            │
│  └──────────────┘ └──────────────┘ └──────────────┘            │
│                                                                  │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐            │
│  │ CPU Cores    │ │ WS Format    │ │ Last Update  │            │
│  │ 8            │ │ JSON/        │ │ 15s ago      │            │
│  │              │ │ MessagePack  │ │              │            │
│  └──────────────┘ └──────────────┘ └──────────────┘            │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│ 💾 AGGREGATED STATE                                              │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐            │
│  │ Market Data  │ │ Active Orders│ │ Open Positions│           │
│  │ 5            │ │ 12           │ │ 3            │            │
│  └──────────────┘ └──────────────┘ └──────────────┘            │
│                                                                  │
│  ┌──────────────┐                                               │
│  │ Active       │                                               │
│  │ Strategies   │                                               │
│  │ 2            │                                               │
│  └──────────────┘                                               │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│ 🖥️  BACKEND SERVICES                                             │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────┐ ┌──────────────────┐                     │
│  │ ● Order          │ │ ● Strategy       │                     │
│  │   Execution      │ │   Engine         │                     │
│  │   Connected      │ │   Connected      │                     │
│  └──────────────────┘ └──────────────────┘                     │
│                                                                  │
│  ┌──────────────────┐ ┌──────────────────┐                     │
│  │ ● Account        │ │ ● Redis          │                     │
│  │   Monitor        │ │                  │                     │
│  │   Connected      │ │   Connected      │                     │
│  └──────────────────┘ └──────────────────┘                     │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│ 📋 API ENDPOINT TESTING                                          │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  [Test Health] [Test Debug] [Test History API] [Test WebSocket] │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ Testing /health...                                         │ │
│  │                                                            │ │
│  │ Status: 200                                                │ │
│  │                                                            │ │
│  │ {                                                          │ │
│  │   "status": "ok",                                          │ │
│  │   "service": "dashboard-server"                            │ │
│  │ }                                                          │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## Color Scheme

### Backgrounds
- **Page Background**: Dark blue gradient (#0f172a → #1e293b)
- **Card Background**: Semi-transparent slate (#1e293b60) with blur
- **Input/Code Blocks**: Very dark slate (#0f172a80)

### Text Colors
- **Primary Text**: Light gray (#f1f5f9)
- **Secondary Text**: Medium gray (#94a3b8)
- **Muted Text**: Dark gray (#64748b)
- **Labels**: Blue-gray (#94a3b8)

### Status Colors
- **Healthy/Connected**: Green (#10b981) ●
- **Warning**: Yellow (#fbbf24) ●
- **Error/Disconnected**: Red (#ef4444) ●

### Accent Colors
- **Primary Button**: Blue gradient (#3b82f6 → #2563eb)
- **Secondary Button**: Purple gradient (#8b5cf6 → #7c3aed)
- **Title Gradient**: Blue to purple (#60a5fa → #a78bfa)

## Interactive Elements

### Buttons
```
┌─────────────────────┐
│   Test Health       │  ← Hover: Lifts up slightly
└─────────────────────┘    Click: Brief press down
     Blue gradient          Shadow: Glowing blue
```

### Cards
```
┌─────────────────────┐
│ 📊 STATE SEQUENCE   │  ← Hover: Lifts up
│                     │    Shadow: Increases
│    12,456           │    Border: Subtle glow
│                     │
│ Total state updates │
└─────────────────────┘
```

### Service Indicators
```
● Connected    ← Pulsing animation (fades in/out)
  Green glow      Updates every 2 seconds
```

### Status Badge
```
┌──────────────┐
│ ● Healthy    │  ← Green background with glow
└──────────────┘    Rounded pill shape
                    Changes color with status
```

## Responsive Breakpoints

### Desktop (1400px+)
- 4 metric cards in a row
- Service info: 3 columns
- Backend services: 4 in a row

### Tablet (768px - 1400px)
- 2 metric cards in a row
- Service info: 2 columns
- Backend services: 2 in a row

### Mobile (<768px)
- 1 metric card per row
- Service info: 1 column
- Backend services: 1 per row
- Scrollable test results

## Animations

### On Page Load
1. Header fades in from top
2. Metric cards slide in from bottom (staggered)
3. Sections fade in sequentially

### Continuous Animations
- **Service indicator**: Pulse (2s cycle)
- **Auto-refresh dot**: Pulse (2s cycle)
- **Loading spinner**: Rotation (0.8s)

### On Interaction
- **Button hover**: Lift up 2px, shadow increase
- **Card hover**: Lift up 2px, shadow increase
- **Button press**: Scale down 98%

## Typography Hierarchy

```
⚡ Dashboard Server                    ← 2.5rem, Bold, Gradient
Real-time WebSocket...                 ← 1.1rem, Regular

📊 STATE SEQUENCE                      ← 0.875rem, Uppercase, Bold
12,456                                 ← 2rem, Bold
Total state updates                    ← 0.875rem, Regular

ℹ️  SERVICE INFORMATION                ← 1.5rem, Bold
Service Name                           ← 0.75rem, Uppercase
dashboard-server                       ← 1rem, Semibold
```

## Visual Hierarchy

### Primary Focus
1. **Header with title** - Largest, center, gradient
2. **Status badge** - Color-coded, immediately visible
3. **Metric cards** - Large numbers, high contrast

### Secondary Focus
4. **Section titles** - Clear labels with icons
5. **Backend service status** - Visual indicators
6. **Test buttons** - Call-to-action styling

### Tertiary Focus
7. **Service info grid** - Detailed metadata
8. **Labels and captions** - Supporting text
9. **Test results** - Expandable content

## Icon Set

Used inline SVGs for:
- ⚡ Lightning (Service/Speed)
- 👥 Users (Clients)
- 📊 Chart (Metrics)
- 📈 Trending (Performance)
- ℹ️  Info (Information)
- 💾 Database (State)
- 🖥️  Server (Backend)
- 📋 Clipboard (Testing)

## Accessibility Features

### Keyboard Navigation
- All buttons are keyboard accessible
- Tab order follows visual flow
- Focus indicators visible

### Screen Reader Support
- Semantic HTML (header, section, nav)
- ARIA labels on interactive elements
- Alt text on icons (via aria-label)

### Color Contrast
- All text meets WCAG AA standards
- Status indicators have text labels
- Not reliant on color alone

### Motion
- Animations are subtle
- No auto-playing video
- Respects prefers-reduced-motion (optional enhancement)

## Browser DevTools View

### Network Tab
```
GET /                           200  25KB  50ms  (HTML)
GET /api/service-info          200   2KB  30ms  (JSON)
GET /api/service-info          200   2KB  28ms  (Auto-refresh)
GET /api/service-info          200   2KB  29ms  (Auto-refresh)
```

### Console
```
No errors
No warnings
Clean console
```

### Performance
- First Contentful Paint: <100ms
- Time to Interactive: <200ms
- Lighthouse Score: 90+

## Example Test Output

### Test Health
```javascript
Testing /health...

Status: 200

{
  "status": "ok",
  "service": "dashboard-server"
}
```

### Test WebSocket
```javascript
Connecting to WebSocket...
Connected!
Subscribing to all channels...

Received: snapshot (seq: 12456)
Market Data: 5 symbols
Orders: 12
Positions: 3
Strategies: 2

Connection closed
Test completed
```

## Mobile View

```
┌─────────────────────┐
│  ⚡ Dashboard Server│
│  Real-time WebSocket│
│     ● Healthy       │
└─────────────────────┘

┌─────────────────────┐
│ ⚡ SERVICE UPTIME   │
│    2h 30m 15s       │
│ Since 10:30 AM      │
└─────────────────────┘

┌─────────────────────┐
│ 👥 CONNECTED CLIENTS│
│         3           │
│ Active WebSocket    │
└─────────────────────┘

[Scrollable content]

┌─────────────────────┐
│ [Test Health]       │
│ [Test Debug]        │
│ [Test History]      │
│ [Test WebSocket]    │
└─────────────────────┘
```

---

This preview shows the visual design and layout of the admin page. The actual implementation uses modern CSS with glass-morphism effects, smooth animations, and a professional dark theme optimized for developer experience.
