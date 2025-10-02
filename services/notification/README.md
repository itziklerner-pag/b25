# Notification Service

Multi-channel notification service for the B25 HFT trading system with support for email, SMS, push notifications, and webhooks.

## Features

- **Multi-Channel Delivery**: Email (SendGrid), SMS (Twilio), Push (FCM), Webhooks
- **Template System**: Reusable notification templates with variable substitution
- **Delivery Queue**: Asynchronous delivery with Asynq and Redis
- **Retry Logic**: Automatic retry with exponential backoff
- **User Preferences**: Per-channel notification preferences with quiet hours
- **Rate Limiting**: Prevent notification spam with configurable limits
- **Notification History**: Complete delivery tracking and status
- **Alert Rules**: Automated notifications based on system events
- **RESTful API**: Complete API for managing notifications
- **Event-Driven**: Subscribe to NATS/Redis for real-time notifications

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                 Notification Service                     │
│                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐             │
│  │   API    │  │  Worker  │  │ Scheduler│             │
│  │  (Gin)   │  │ (Asynq)  │  │ (Cron)   │             │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘             │
│       │             │              │                    │
│       └─────────────┼──────────────┘                    │
│                     │                                   │
│            ┌────────▼────────┐                          │
│            │ Business Logic   │                          │
│            │  - Validation    │                          │
│            │  - Rate Limiting │                          │
│            │  - Preferences   │                          │
│            └────────┬────────┘                          │
│                     │                                   │
│       ┌─────────────┼─────────────┐                    │
│       │             │             │                    │
│  ┌────▼────┐  ┌────▼────┐  ┌────▼────┐               │
│  │ Providers│  │  Queue  │  │   DB    │               │
│  │ SendGrid │  │  Redis  │  │Postgres │               │
│  │ Twilio   │  │         │  │         │               │
│  │   FCM    │  │         │  │         │               │
│  └──────────┘  └─────────┘  └─────────┘               │
└─────────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- SendGrid API key (for email)
- Twilio credentials (for SMS)
- Firebase credentials (for push)

### Installation

1. Clone the repository
2. Copy configuration:
   ```bash
   cp config.example.yaml config.yaml
   cp .env.example .env
   ```

3. Update configuration with your credentials

4. Run database migrations:
   ```bash
   psql -U notification_user -d notification_db -f migrations/001_init_schema.up.sql
   ```

5. Install dependencies:
   ```bash
   go mod download
   ```

6. Run the service:
   ```bash
   go run cmd/server/main.go
   ```

### Docker

Build and run with Docker:

```bash
docker build -t notification-service .
docker run -p 9097:9097 -p 9098:9098 notification-service
```

With Docker Compose:

```bash
docker-compose -f ../../docker/docker-compose.dev.yml up notification-service
```

## Configuration

Configuration can be provided via YAML file or environment variables.

### Key Configuration Options

```yaml
server:
  port: 9097
  mode: development

database:
  host: localhost
  port: 5432
  dbname: notification_db

redis:
  host: localhost
  port: 6379

email:
  provider: sendgrid
  sendgrid:
    api_key: your_key_here

sms:
  provider: twilio
  twilio:
    account_sid: your_sid
    auth_token: your_token

push:
  provider: fcm
  fcm:
    credentials_file: /path/to/firebase.json
```

See `config.example.yaml` for all options.

## API Documentation

### Create Notification

```bash
POST /api/v1/notifications
Content-Type: application/json

{
  "user_id": "user123",
  "channel": "email",
  "template_name": "order_fill",
  "template_data": {
    "order_id": "ORD123",
    "symbol": "BTCUSDT",
    "quantity": 0.5
  },
  "priority": "high"
}
```

### List Notifications

```bash
GET /api/v1/notifications?user_id=uuid&channel=email&status=sent&limit=20&offset=0
```

### Get Notification

```bash
GET /api/v1/notifications/{id}
```

### Get User Notifications

```bash
GET /api/v1/notifications/user/{user_id}?limit=20&offset=0
```

## Notification Channels

### Email (SendGrid)

- HTML and plain text support
- Template rendering
- Delivery tracking
- Bounce handling

### SMS (Twilio)

- International support
- Delivery receipts
- Character limit handling

### Push (Firebase FCM)

- iOS and Android support
- Rich notifications
- Device management
- Badge counts

### Webhooks

- Configurable endpoints
- Signature verification
- Retry on failure

## Template System

Create reusable templates:

```sql
INSERT INTO notification_templates (name, type, channel, subject, body_template)
VALUES (
  'order_fill',
  'order_fill',
  'email',
  'Order Filled: {{.Symbol}}',
  '<h1>Order Filled</h1><p>Your order for {{.Quantity}} {{.Symbol}} has been filled.</p>'
);
```

Use in API:

```json
{
  "template_name": "order_fill",
  "template_data": {
    "symbol": "BTCUSDT",
    "quantity": 0.5
  }
}
```

## User Preferences

Users can configure notification preferences per channel and category:

```sql
INSERT INTO notification_preferences (user_id, channel, category, is_enabled, quiet_hours_enabled, quiet_hours_start, quiet_hours_end)
VALUES (
  'user-uuid',
  'email',
  'trading_alerts',
  true,
  true,
  '22:00',
  '08:00'
);
```

## Rate Limiting

Prevent notification spam:

- Email: 50 per hour per user
- SMS: 20 per hour per user
- Push: 100 per hour per user

Configurable in `config.yaml`.

## Monitoring

### Health Checks

- Liveness: `GET /health`
- Readiness: `GET /ready`

### Metrics

Prometheus metrics at `GET /metrics` (port 9098):

- `notification_sent_total{channel, status}`
- `notification_delivery_duration_seconds`
- `notification_queue_size`
- `rate_limit_hits_total`

### Logging

Structured JSON logs with:
- Request IDs
- Correlation IDs
- Performance metrics
- Error details

## Event-Driven Notifications

Subscribe to system events:

### NATS Subscriptions

```yaml
subscriptions:
  nats:
    enabled: true
    subjects:
      - "trading.alerts.*"
      - "risk.violations.*"
      - "orders.critical.*"
```

### Alert Rules

Automated notifications based on events:

```sql
INSERT INTO alert_rules (name, event_source, event_type, channels, priority)
VALUES (
  'high_risk_alert',
  'risk',
  'drawdown_exceeded',
  ARRAY['email', 'sms'],
  'critical'
);
```

## Database Schema

Key tables:

- `notifications`: Core notification records
- `notification_templates`: Reusable templates
- `notification_preferences`: User preferences
- `notification_events`: Delivery tracking
- `alert_rules`: Automated notifications
- `users`: Notification recipients
- `user_devices`: Push notification devices

See `migrations/001_init_schema.up.sql` for full schema.

## Development

### Running Tests

```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./tests/integration/...

# With coverage
go test -cover ./...
```

### Code Structure

```
services/notification/
├── cmd/server/          # Main application
├── internal/
│   ├── api/            # HTTP handlers
│   ├── config/         # Configuration
│   ├── middleware/     # HTTP middleware
│   ├── models/         # Data models
│   ├── providers/      # Channel providers
│   ├── queue/          # Queue implementation
│   ├── repository/     # Database layer
│   ├── service/        # Business logic
│   └── templates/      # Template engine
├── migrations/         # Database migrations
└── tests/             # Test files
```

## Deployment

### Environment Variables

See `.env.example` for required variables.

### Database Migrations

Run migrations before deployment:

```bash
psql -U notification_user -d notification_db -f migrations/001_init_schema.up.sql
```

### Scaling

- Horizontal scaling: Multiple service instances
- Queue workers: Scale concurrency in config
- Database: Read replicas for queries
- Redis: Cluster for high availability

## Troubleshooting

### Common Issues

**Notifications not sending:**
- Check provider credentials in config
- Verify queue is processing (check Redis)
- Check logs for errors

**Rate limit errors:**
- Adjust limits in config
- Check Redis for rate limit keys
- Reset limits if needed

**Template errors:**
- Verify template syntax
- Check variable names match data
- Review template registration logs

## Contributing

See CONTRIBUTING.md for development guidelines.

## License

Part of the B25 HFT trading system.

## Support

For issues and questions, contact the development team.
