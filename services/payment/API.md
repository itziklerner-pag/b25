# Payment Service API Documentation

Version: 1.0.0
Base URL: `http://localhost:9097/api/v1`

## Authentication

All endpoints (except webhooks and health checks) require JWT authentication.

**Header Format:**
```
Authorization: Bearer <your_jwt_token>
```

## Error Responses

All endpoints return consistent error responses:

```json
{
  "error": "Error message describing what went wrong"
}
```

**HTTP Status Codes:**
- `200 OK` - Request succeeded
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

## Endpoints

### Payments

#### Create Payment

Create a new one-time payment.

**Endpoint:** `POST /payments`

**Request Body:**
```json
{
  "user_id": "user_123",
  "amount": 1000,
  "currency": "USD",
  "payment_method": "pm_card_visa",
  "description": "Product purchase",
  "metadata": {
    "order_id": "order_789",
    "product_name": "Premium Plan"
  }
}
```

**Parameters:**
- `user_id` (string, required) - User identifier
- `amount` (integer, required) - Amount in cents (minimum 50)
- `currency` (string, required) - 3-letter ISO currency code
- `payment_method` (string, optional) - Stripe payment method ID
- `description` (string, optional) - Payment description
- `metadata` (object, optional) - Additional metadata

**Response:** `201 Created`
```json
{
  "id": "tx_abc123",
  "user_id": "user_123",
  "stripe_payment_id": "pi_xyz789",
  "amount": 1000,
  "currency": "USD",
  "status": "succeeded",
  "payment_method": "pm_card_visa",
  "description": "Product purchase",
  "metadata": "{\"order_id\":\"order_789\"}",
  "receipt_url": "https://pay.stripe.com/receipts/...",
  "created_at": "2025-10-03T10:30:00Z",
  "updated_at": "2025-10-03T10:30:00Z"
}
```

#### Get Payment

Retrieve a specific payment by ID.

**Endpoint:** `GET /payments/:id`

**Response:** `200 OK`
```json
{
  "id": "tx_abc123",
  "user_id": "user_123",
  "amount": 1000,
  "currency": "USD",
  "status": "succeeded",
  "created_at": "2025-10-03T10:30:00Z"
}
```

#### List User Payments

Get all payments for a specific user.

**Endpoint:** `GET /users/:user_id/payments`

**Query Parameters:**
- `limit` (integer, optional) - Number of results (default: 20)
- `offset` (integer, optional) - Offset for pagination (default: 0)

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "tx_abc123",
      "amount": 1000,
      "currency": "USD",
      "status": "succeeded",
      "created_at": "2025-10-03T10:30:00Z"
    }
  ],
  "limit": 20,
  "offset": 0
}
```

### Payment Methods

#### Attach Payment Method

Attach a payment method to a user.

**Endpoint:** `POST /payment-methods`

**Request Body:**
```json
{
  "user_id": "user_123",
  "payment_method": "pm_card_visa",
  "set_as_default": true
}
```

**Parameters:**
- `user_id` (string, required) - User identifier
- `payment_method` (string, required) - Stripe payment method ID
- `set_as_default` (boolean, optional) - Set as default payment method

**Response:** `201 Created`
```json
{
  "id": "pm_local_123",
  "user_id": "user_123",
  "stripe_payment_method_id": "pm_card_visa",
  "type": "card",
  "is_default": true,
  "card_brand": "visa",
  "card_last4": "4242",
  "card_exp_month": 12,
  "card_exp_year": 2025,
  "created_at": "2025-10-03T10:30:00Z"
}
```

#### List Payment Methods

Get all payment methods for a user.

**Endpoint:** `GET /users/:user_id/payment-methods`

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "pm_local_123",
      "type": "card",
      "is_default": true,
      "card_brand": "visa",
      "card_last4": "4242"
    }
  ]
}
```

#### Remove Payment Method

Detach a payment method from a user.

**Endpoint:** `DELETE /payment-methods/:id`

**Response:** `200 OK`
```json
{
  "message": "Payment method removed successfully"
}
```

### Subscriptions

#### Create Subscription

Create a new recurring subscription.

**Endpoint:** `POST /subscriptions`

**Request Body:**
```json
{
  "user_id": "user_123",
  "price_id": "price_1234567890",
  "payment_method": "pm_card_visa",
  "trial_days": 14,
  "metadata": {
    "plan_type": "premium"
  }
}
```

**Parameters:**
- `user_id` (string, required) - User identifier
- `price_id` (string, required) - Stripe price ID
- `payment_method` (string, required) - Stripe payment method ID
- `trial_days` (integer, optional) - Trial period in days
- `metadata` (object, optional) - Additional metadata

**Response:** `201 Created`
```json
{
  "id": "sub_abc123",
  "user_id": "user_123",
  "stripe_subscription_id": "sub_xyz789",
  "stripe_price_id": "price_1234567890",
  "status": "active",
  "plan_name": "Premium Plan",
  "amount": 2999,
  "currency": "USD",
  "interval": "month",
  "interval_count": 1,
  "trial_end": "2025-10-17T10:30:00Z",
  "current_period_start": "2025-10-03T10:30:00Z",
  "current_period_end": "2025-11-03T10:30:00Z",
  "created_at": "2025-10-03T10:30:00Z"
}
```

#### Get Subscription

Retrieve a specific subscription.

**Endpoint:** `GET /subscriptions/:id`

**Response:** `200 OK`
```json
{
  "id": "sub_abc123",
  "status": "active",
  "amount": 2999,
  "currency": "USD",
  "interval": "month"
}
```

#### List User Subscriptions

Get all subscriptions for a user.

**Endpoint:** `GET /users/:user_id/subscriptions`

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "sub_abc123",
      "status": "active",
      "plan_name": "Premium Plan",
      "amount": 2999,
      "interval": "month"
    }
  ]
}
```

#### Cancel Subscription

Cancel a subscription.

**Endpoint:** `DELETE /subscriptions/:id`

**Query Parameters:**
- `immediate` (boolean, optional) - Cancel immediately or at period end

**Response:** `200 OK`
```json
{
  "id": "sub_abc123",
  "status": "canceled",
  "cancel_at_period_end": false,
  "canceled_at": "2025-10-03T10:30:00Z"
}
```

### Invoices

#### Get Invoice

Retrieve a specific invoice.

**Endpoint:** `GET /invoices/:id`

**Response:** `200 OK`
```json
{
  "id": "inv_abc123",
  "user_id": "user_123",
  "stripe_invoice_id": "in_xyz789",
  "number": "INV-001",
  "status": "paid",
  "amount_due": 2999,
  "amount_paid": 2999,
  "amount_remaining": 0,
  "currency": "USD",
  "hosted_invoice_url": "https://invoice.stripe.com/...",
  "invoice_pdf": "https://pay.stripe.com/invoice/.../pdf",
  "paid_at": "2025-10-03T10:30:00Z"
}
```

#### List User Invoices

Get all invoices for a user.

**Endpoint:** `GET /users/:user_id/invoices`

**Query Parameters:**
- `limit` (integer, optional) - Number of results (default: 20)
- `offset` (integer, optional) - Offset for pagination (default: 0)

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "inv_abc123",
      "number": "INV-001",
      "status": "paid",
      "amount_due": 2999,
      "currency": "USD"
    }
  ],
  "limit": 20,
  "offset": 0
}
```

#### List Subscription Invoices

Get all invoices for a subscription.

**Endpoint:** `GET /subscriptions/:subscription_id/invoices`

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "inv_abc123",
      "number": "INV-001",
      "status": "paid",
      "amount_due": 2999
    }
  ]
}
```

### Refunds

#### Create Refund

Create a refund for a transaction.

**Endpoint:** `POST /refunds`

**Request Body:**
```json
{
  "transaction_id": "tx_abc123",
  "amount": 500,
  "reason": "requested_by_customer"
}
```

**Parameters:**
- `transaction_id` (string, required) - Transaction ID to refund
- `amount` (integer, optional) - Amount to refund in cents (omit for full refund)
- `reason` (string, required) - Refund reason (duplicate, fraudulent, requested_by_customer)

**Response:** `201 Created`
```json
{
  "id": "ref_abc123",
  "transaction_id": "tx_abc123",
  "stripe_refund_id": "re_xyz789",
  "amount": 500,
  "currency": "USD",
  "reason": "requested_by_customer",
  "status": "succeeded",
  "created_at": "2025-10-03T10:30:00Z"
}
```

### Webhooks

#### Stripe Webhook

Receive and process Stripe webhook events.

**Endpoint:** `POST /webhooks/stripe`

**Headers:**
- `Stripe-Signature` (string, required) - Webhook signature for verification

**Request Body:** Raw Stripe event payload

**Response:** `200 OK`
```json
{
  "received": true
}
```

**Supported Events:**
- `payment_intent.succeeded`
- `payment_intent.payment_failed`
- `payment_intent.canceled`
- `customer.subscription.created`
- `customer.subscription.updated`
- `customer.subscription.deleted`
- `invoice.paid`
- `invoice.payment_failed`
- `invoice.finalized`
- `charge.refunded`

### System

#### Health Check

Check service health status.

**Endpoint:** `GET /health`

**Response:** `200 OK`
```json
{
  "status": "healthy"
}
```

#### Metrics

Prometheus metrics endpoint.

**Endpoint:** `GET /metrics`

**Response:** `200 OK` (Prometheus format)
```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",status="200"} 1234
```

## Rate Limiting

- **Default Limit:** 100 requests per minute per IP
- **Header:** `X-RateLimit-Remaining` shows remaining requests
- **Response:** `429 Too Many Requests` when limit exceeded

## Pagination

List endpoints support pagination:

**Query Parameters:**
- `limit` - Number of items per page (max 100, default 20)
- `offset` - Number of items to skip (default 0)

**Example:**
```
GET /users/user_123/payments?limit=50&offset=100
```

## Webhooks

### Setting Up Webhooks

1. Configure endpoint in Stripe Dashboard
2. Add webhook URL: `https://yourdomain.com/api/v1/webhooks/stripe`
3. Select events to monitor
4. Copy webhook signing secret
5. Set `STRIPE_WEBHOOK_SECRET` environment variable

### Webhook Security

- All webhook events are verified using Stripe signature
- Events are stored for idempotency
- Duplicate events are ignored
- Failed events are logged for retry

## Examples

### Complete Payment Flow

```bash
# 1. Attach payment method
curl -X POST http://localhost:9097/api/v1/payment-methods \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_123",
    "payment_method": "pm_card_visa",
    "set_as_default": true
  }'

# 2. Create payment
curl -X POST http://localhost:9097/api/v1/payments \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_123",
    "amount": 1000,
    "currency": "USD",
    "payment_method": "pm_card_visa",
    "description": "Order #123"
  }'

# 3. Get payment status
curl http://localhost:9097/api/v1/payments/tx_abc123 \
  -H "Authorization: Bearer $JWT_TOKEN"
```

### Subscription Flow

```bash
# 1. Create subscription
curl -X POST http://localhost:9097/api/v1/subscriptions \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_123",
    "price_id": "price_monthly_premium",
    "payment_method": "pm_card_visa",
    "trial_days": 14
  }'

# 2. List user subscriptions
curl http://localhost:9097/api/v1/users/user_123/subscriptions \
  -H "Authorization: Bearer $JWT_TOKEN"

# 3. Cancel subscription (at period end)
curl -X DELETE http://localhost:9097/api/v1/subscriptions/sub_abc123 \
  -H "Authorization: Bearer $JWT_TOKEN"

# 4. Cancel subscription (immediately)
curl -X DELETE "http://localhost:9097/api/v1/subscriptions/sub_abc123?immediate=true" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

### Refund Flow

```bash
# Full refund
curl -X POST http://localhost:9097/api/v1/refunds \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_id": "tx_abc123",
    "reason": "requested_by_customer"
  }'

# Partial refund
curl -X POST http://localhost:9097/api/v1/refunds \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_id": "tx_abc123",
    "amount": 500,
    "reason": "requested_by_customer"
  }'
```

---

**API Version:** 1.0.0
**Last Updated:** 2025-10-03
