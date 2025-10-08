# Configuration Service - Admin Dashboard

## Overview
Modern, responsive admin dashboard for the Configuration Service. Provides real-time monitoring, health checks, and configuration management interface.

## Features

### Service Monitoring
- Real-time health status
- Database connectivity status
- NATS messaging status
- Service uptime tracking
- Configuration count display

### API Explorer
- Complete endpoint documentation
- HTTP method indicators (GET, POST, PUT, DELETE)
- Authentication requirements
- Endpoint descriptions

### Configuration Management
- View all configurations
- Filter by service
- Status indicators (Active/Inactive)
- Version tracking
- Last updated timestamps

### Design Features
- Gradient background with card-based layout
- Smooth hover animations
- Color-coded HTTP methods
- Responsive mobile-first design
- Loading states and error handling
- Empty state handling

## Access

### Local Development
- Direct: `http://localhost:8085/`
- Admin route: `http://localhost:8085/admin`

### Production (via Nginx)
- Main route: `https://yourdomain.com/services/configuration/`
- Admin route: `https://yourdomain.com/services/configuration/admin`

## API Integration

The dashboard uses the following API base URL:
```javascript
const API_BASE = window.location.origin + '/services/configuration';
```

This ensures the admin page works both in development and production environments.

## Endpoints Used

### Public Endpoints
- `GET /health` - Service health check
- `GET /ready` - Service readiness with component status
- `GET /api/service-info` - Service metadata and endpoint list

### Protected Endpoints
- `GET /api/v1/configurations` - List all configurations (requires X-API-Key header)

## Browser Compatibility
- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## Technology Stack
- Vanilla JavaScript (no framework dependencies)
- CSS3 with custom properties
- Fetch API for HTTP requests
- Modern CSS Grid and Flexbox

## Customization

### Colors
The admin page uses CSS custom properties for easy theming:
```css
--primary: #3b82f6
--success: #10b981
--warning: #f59e0b
--danger: #ef4444
```

### Refresh Intervals
- Health check: 30 seconds
- Configuration count: 30 seconds
- Uptime: 1 second

## Development

To modify the admin page:
1. Edit `/home/mm/dev/b25/services/configuration/web/index.html`
2. Rebuild service (if needed): `make build`
3. Restart service or reload in browser

No build step required - it's a static HTML file with inline CSS and JavaScript.

## Security Notes

- Configuration API endpoints require authentication via `X-API-Key` header
- Health and readiness endpoints are public
- CORS is enabled for cross-origin requests
- No sensitive data is displayed without authentication

## Performance

- Lightweight: ~25KB total size
- No external dependencies
- Fast initial load
- Efficient polling intervals
- Lazy loading of configuration data

## Future Enhancements
- [ ] Real-time updates via WebSocket
- [ ] Configuration editing interface
- [ ] Audit log viewer
- [ ] Advanced filtering and search
- [ ] Dark mode toggle
- [ ] Export configurations to JSON/YAML
- [ ] Diff viewer for configuration versions
