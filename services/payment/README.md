# Payment Service

A production-ready payment processing service with Stripe integration, subscription management, and comprehensive PCI compliance features.

## Features

- **Payment Processing**: Accept one-time payments via Stripe
- **Subscription Management**: Recurring billing with trial periods
- **Invoice Generation**: Automated invoice creation and tracking
- **Refund Processing**: Full and partial refund support
- **Payment Methods**: Store and manage customer payment methods securely
- **Webhook Handling**: Real-time event processing from Stripe
- **Security**: JWT authentication, rate limiting, and encryption
- **Observability**: Prometheus metrics, structured logging, health checks
- **PCI Compliance**: Designed with security best practices

## Architecture

```
┌─────────────────┐
│   API Gateway   │
└────────┬────────┘
         │
    ┌────▼────┐
    │   API   │
    │ Handlers│
    └────┬────┘
         │
    ┌────▼────────┐
    │  Services   │
    │             │
    │ • Payment   │
    │ • Subscription│
    │ • Invoice   │
    │ • Refund    │
    │ • Webhook   │
    └────┬────────┘
         │
    ┌────▼──────────┐        ┌──────────┐
    │ Repositories  │◄──────►│PostgreSQL│
    └────┬──────────┘        └──────────┘
         │
    ┌────▼────┐              ┌────────┐
    │ Stripe  │◄────────────►│ Stripe │
    │ Client  │              │  API   │
    └─────────┘              └────────┘
         │
    ┌────▼────┐
    │  Redis  │
    │ (Cache) │
    └─────────┘
```

## Database Schema

### Tables

- **transactions**: Payment transaction records
- **subscriptions**: Recurring subscription data
- **invoices**: Invoice tracking and management
- **payment_methods**: Stored payment methods
- **refunds**: Refund transaction records
- **webhook_events**: Webhook event log for idempotency

## API Endpoints

### Payment Endpoints

```
POST   /api/v1/payments                     # Create a payment
GET    /api/v1/payments/:id                 # Get payment details
GET    /api/v1/users/:user_id/payments      # Get user's payments
```

### Payment Method Endpoints

```
POST   /api/v1/payment-methods              # Attach payment method
GET    /api/v1/users/:user_id/payment-methods  # Get user's payment methods
DELETE /api/v1/payment-methods/:id          # Remove payment method
```

### Subscription Endpoints

```
POST   /api/v1/subscriptions                # Create subscription
GET    /api/v1/subscriptions/:id            # Get subscription details
GET    /api/v1/users/:user_id/subscriptions # Get user's subscriptions
DELETE /api/v1/subscriptions/:id            # Cancel subscription
```

### Invoice Endpoints

```
GET    /api/v1/invoices/:id                           # Get invoice
GET    /api/v1/users/:user_id/invoices                # Get user's invoices
GET    /api/v1/subscriptions/:subscription_id/invoices # Get subscription invoices
```

### Refund Endpoints

```
POST   /api/v1/refunds                      # Create refund
```

### Webhook Endpoints

```
POST   /api/v1/webhooks/stripe              # Stripe webhook handler
```

### System Endpoints

```
GET    /health                              # Health check
GET    /metrics                             # Prometheus metrics
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 15+
- Redis 7+
- Stripe account (test or production)

### Installation

1. Clone the repository:
```bash
cd services/payment
```

2. Copy environment configuration:
```bash
cp .env.example .env
```

3. Update `.env` with your configuration:
```env
STRIPE_SECRET_KEY=sk_test_your_key
STRIPE_WEBHOOK_SECRET=whsec_your_secret
JWT_SECRET=your_jwt_secret
ENCRYPTION_KEY=your_32_character_encryption_key
```

4. Start dependencies:
```bash
docker-compose up postgres redis -d
```

5. Run database migrations:
```bash
make migrate-up
```

6. Start the service:
```bash
make run
```

The service will be available at `http://localhost:9097`

### Using Docker

Build and run with Docker Compose:

```bash
docker-compose up --build
```

## Configuration

Configuration can be provided via:

1. **Environment Variables** (recommended for production)
2. **YAML Configuration File** (`config.yaml`)
3. **Environment File** (`.env`)

### Key Configuration Options

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | 9097 |
| `DB_HOST` | PostgreSQL host | localhost |
| `REDIS_HOST` | Redis host | localhost |
| `STRIPE_SECRET_KEY` | Stripe secret key | - |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook secret | - |
| `JWT_SECRET` | JWT signing secret | - |
| `ENCRYPTION_KEY` | Data encryption key (32 chars) | - |
| `RATE_LIMIT_PER_MINUTE` | API rate limit | 100 |

## Usage Examples

### Create a Payment

```bash
curl -X POST http://localhost:9097/api/v1/payments \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_123",
    "amount": 1000,
    "currency": "USD",
    "description": "Product purchase",
    "payment_method": "pm_card_visa"
  }'
```

### Create a Subscription

```bash
curl -X POST http://localhost:9097/api/v1/subscriptions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_123",
    "price_id": "price_1234567890",
    "payment_method": "pm_card_visa",
    "trial_days": 14
  }'
```

### Process a Refund

```bash
curl -X POST http://localhost:9097/api/v1/refunds \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_id": "tx_123",
    "amount": 500,
    "reason": "requested_by_customer"
  }'
```

## Webhook Setup

1. Configure webhook endpoint in Stripe Dashboard:
```
https://yourdomain.com/api/v1/webhooks/stripe
```

2. Select events to monitor:
   - `payment_intent.succeeded`
   - `payment_intent.payment_failed`
   - `customer.subscription.created`
   - `customer.subscription.updated`
   - `customer.subscription.deleted`
   - `invoice.paid`
   - `invoice.payment_failed`
   - `charge.refunded`

3. Copy webhook signing secret to `.env`:
```env
STRIPE_WEBHOOK_SECRET=whsec_...
```

## Security & PCI Compliance

### Best Practices Implemented

1. **Never Store Sensitive Card Data**
   - Use Stripe Elements for card collection
   - Store only payment method IDs from Stripe

2. **Encryption at Rest**
   - Sensitive data encrypted with AES-256
   - Encryption keys managed securely

3. **Encryption in Transit**
   - TLS 1.3 for all connections
   - Certificate validation enforced

4. **Authentication & Authorization**
   - JWT-based authentication
   - Role-based access control
   - Rate limiting per IP

5. **Audit Logging**
   - All payment operations logged
   - Webhook events stored for audit trail

6. **Security Headers**
   - CORS configured properly
   - Security headers enforced

### PCI DSS Compliance Notes

This service is designed to minimize PCI DSS scope:

- **SAQ A Compliance**: Use Stripe.js/Elements for card collection
- **No Card Data Storage**: Only Stripe tokens/IDs stored
- **Encrypted Storage**: All sensitive data encrypted
- **Access Controls**: JWT authentication required
- **Logging**: Comprehensive audit trail maintained

### Production Deployment Checklist

- [ ] Use production Stripe keys
- [ ] Enable SSL/TLS certificates
- [ ] Set strong JWT secret (32+ characters)
- [ ] Set strong encryption key (exactly 32 characters)
- [ ] Configure firewall rules
- [ ] Enable database backups
- [ ] Set up monitoring and alerts
- [ ] Configure rate limiting appropriately
- [ ] Implement IP whitelisting if needed
- [ ] Review and test webhook handling
- [ ] Set up log aggregation
- [ ] Enable database connection pooling
- [ ] Configure Redis persistence

## Testing

Run all tests:
```bash
make test
```

Run with coverage:
```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Categories

- **Unit Tests**: Service and repository layer
- **Integration Tests**: Database operations
- **API Tests**: Endpoint validation
- **Webhook Tests**: Event processing

## Monitoring

### Health Check

```bash
curl http://localhost:9097/health
```

Expected response:
```json
{"status": "healthy"}
```

### Metrics

Prometheus metrics available at `/metrics`:

```bash
curl http://localhost:9097/metrics
```

Key metrics:
- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request latency
- `payment_transactions_total` - Payment transactions processed
- `payment_errors_total` - Payment processing errors
- `webhook_events_total` - Webhook events received

### Logging

Structured JSON logging with the following levels:
- `INFO`: Normal operations
- `WARN`: Warnings and retries
- `ERROR`: Error conditions
- `DEBUG`: Debug information (development only)

## Development

### Project Structure

```
services/payment/
├── cmd/
│   └── server/          # Application entry point
│       └── main.go
├── internal/
│   ├── api/            # HTTP handlers and routing
│   │   ├── handlers/   # Request handlers
│   │   ├── middleware/ # Middleware functions
│   │   └── routes.go   # Route registration
│   ├── config/         # Configuration management
│   ├── database/       # Database connections
│   ├── logger/         # Logging implementation
│   ├── models/         # Domain models
│   ├── payment/        # Stripe integration
│   ├── repository/     # Data access layer
│   └── service/        # Business logic
├── migrations/         # Database migrations
├── .env.example        # Example environment file
├── Dockerfile          # Container definition
├── docker-compose.yml  # Local development setup
├── go.mod              # Go dependencies
├── Makefile            # Build automation
└── README.md           # This file
```

### Adding New Features

1. Define models in `internal/models/`
2. Create repository in `internal/repository/`
3. Implement service in `internal/service/`
4. Add handler in `internal/api/handlers/`
5. Register routes in `internal/api/routes.go`
6. Write tests
7. Update documentation

## Troubleshooting

### Common Issues

**Database Connection Failed**
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check connection details
psql -h localhost -U postgres -d payment_db
```

**Redis Connection Failed**
```bash
# Check if Redis is running
docker-compose ps redis

# Test connection
redis-cli ping
```

**Stripe Webhook Signature Verification Failed**
- Ensure `STRIPE_WEBHOOK_SECRET` matches Stripe Dashboard
- Verify webhook endpoint URL is correct
- Check that payload hasn't been modified

**Payment Intent Creation Failed**
- Verify `STRIPE_SECRET_KEY` is correct
- Check Stripe API version compatibility
- Review Stripe Dashboard for errors

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Write/update tests
5. Update documentation
6. Submit a pull request

## License

See [LICENSE](../../LICENSE) file for details.

## Support

For issues and questions:
- Create an issue in the repository
- Check existing documentation
- Review Stripe API documentation

## Related Services

- [Order Execution Service](../order-execution/) - Order management
- [Account Monitor Service](../account-monitor/) - Account tracking
- [Configuration Service](../configuration/) - System configuration

---

**Built with Go and Stripe** | **Production-Ready** | **PCI Compliant Design**
