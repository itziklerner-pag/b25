# Messaging Service - Quick Start Guide

Get the messaging service up and running in minutes!

## Prerequisites

- Docker and Docker Compose (recommended)
- OR Go 1.21+, PostgreSQL 15+, Redis 7+, NATS

## Option 1: Docker (Recommended)

### 1. Start the service

```bash
cd services/messaging
docker-compose up -d
```

This will start:
- PostgreSQL database
- Redis cache
- NATS message queue
- Messaging service

### 2. Check service health

```bash
curl http://localhost:9097/health
```

Expected response:
```json
{"status":"ok"}
```

### 3. View logs

```bash
docker-compose logs -f messaging
```

## Option 2: Local Development

### 1. Install dependencies

```bash
go mod download
```

### 2. Setup PostgreSQL database

```sql
CREATE DATABASE messaging;
```

### 3. Copy and configure

```bash
cp config.example.yaml config.yaml
# Edit config.yaml with your database credentials
```

### 4. Run migrations

```bash
make migrate-up
```

### 5. Start the server

```bash
make run
```

## Testing the Service

### 1. Create a test user

First, you need to generate a JWT token. For testing, you can use this simple script:

```go
// generate_token.go
package main

import (
    "fmt"
    "time"
    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
)

func main() {
    secret := "your-secret-key-change-in-production"
    userID := uuid.New()

    claims := jwt.MapClaims{
        "user_id": userID.String(),
        "username": "testuser",
        "exp": time.Now().Add(24 * time.Hour).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString([]byte(secret))

    fmt.Println("Token:", tokenString)
    fmt.Println("User ID:", userID)
}
```

### 2. Create a conversation

```bash
curl -X POST http://localhost:9097/api/v1/conversations \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "direct",
    "member_ids": ["USER_ID_1", "USER_ID_2"]
  }'
```

### 3. Send a message

```bash
curl -X POST http://localhost:9097/api/v1/conversations/CONVERSATION_ID/messages \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Hello, world!",
    "type": "text"
  }'
```

### 4. Get messages

```bash
curl http://localhost:9097/api/v1/conversations/CONVERSATION_ID/messages?limit=20 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 5. Connect via WebSocket

```javascript
// JavaScript example
const token = "YOUR_TOKEN";
const ws = new WebSocket(`ws://localhost:9097/ws?token=${token}`);

ws.onopen = () => {
    console.log('Connected');

    // Send a message
    ws.send(JSON.stringify({
        type: 'message.send',
        data: {
            conversation_id: 'CONVERSATION_ID',
            content: 'Hello via WebSocket!'
        }
    }));
};

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    console.log('Received:', message);
};
```

### 6. Search messages

```bash
curl "http://localhost:9097/api/v1/search/messages?q=hello&limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Common Commands

```bash
# Build the service
make build

# Run tests
make test

# Run with coverage
make test-coverage

# Format code
make fmt

# Run linter
make lint

# View Docker logs
make docker-logs

# Stop Docker services
make docker-down

# Clean build artifacts
make clean
```

## API Endpoints Summary

### Conversations
- `POST /api/v1/conversations` - Create
- `GET /api/v1/conversations` - List
- `GET /api/v1/conversations/:id` - Get
- `PUT /api/v1/conversations/:id` - Update
- `POST /api/v1/conversations/:id/members` - Add member
- `DELETE /api/v1/conversations/:id/members/:userId` - Remove member

### Messages
- `POST /api/v1/conversations/:id/messages` - Send
- `GET /api/v1/conversations/:id/messages` - List
- `PUT /api/v1/messages/:id` - Edit
- `DELETE /api/v1/messages/:id` - Delete

### Reactions & Read Receipts
- `POST /api/v1/messages/:id/reactions` - Add reaction
- `DELETE /api/v1/messages/:id/reactions/:emoji` - Remove reaction
- `POST /api/v1/messages/:id/read` - Mark as read

### Search & Presence
- `GET /api/v1/search/messages?q=query` - Search
- `GET /api/v1/users/online` - Online users

### Health & Metrics
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics

## WebSocket Messages

### Client → Server
```json
{"type": "message.send", "data": {"conversation_id": "...", "content": "..."}}
{"type": "typing.start", "data": {"conversation_id": "..."}}
{"type": "typing.stop", "data": {"conversation_id": "..."}}
{"type": "presence.update", "data": {"status": "online"}}
{"type": "message.read", "data": {"message_id": "..."}}
```

### Server → Client
```json
{"type": "message.new", "data": {...}}
{"type": "typing.indicator", "data": {"user_id": "...", "is_typing": true}}
{"type": "presence.changed", "data": {"user_id": "...", "status": "online"}}
{"type": "message.read_receipt", "data": {"message_id": "...", "user_id": "..."}}
```

## Environment Variables

Key environment variables (prefix with `MESSAGING_`):

```bash
MESSAGING_DATABASE_HOST=localhost
MESSAGING_DATABASE_PORT=5432
MESSAGING_DATABASE_NAME=messaging
MESSAGING_DATABASE_USER=postgres
MESSAGING_DATABASE_PASSWORD=postgres
MESSAGING_REDIS_HOST=localhost
MESSAGING_REDIS_PORT=6379
MESSAGING_NATS_URL=nats://localhost:4222
MESSAGING_AUTH_JWT_SECRET=your-secret-key
MESSAGING_LOGGING_LEVEL=info
```

## Troubleshooting

### Database connection failed
- Ensure PostgreSQL is running
- Check credentials in config.yaml
- Verify database exists

### WebSocket connection failed
- Check JWT token is valid
- Ensure token is in query parameter: `?token=...`
- Check CORS settings

### Messages not appearing in real-time
- Verify WebSocket connection is established
- Check both users are connected
- Review server logs for errors

### Migration errors
- Ensure database is accessible
- Check if migrations table exists
- Try running migrations manually

## Next Steps

1. Implement file upload/download
2. Add Redis caching for presence
3. Integrate NATS for message queue
4. Set up monitoring with Prometheus & Grafana
5. Configure rate limiting
6. Enable message encryption
7. Deploy to production

## Support

For issues and questions:
- Check logs: `docker-compose logs -f messaging`
- Review README.md for detailed documentation
- See progress.md for implementation details

## Production Checklist

Before deploying to production:

- [ ] Change JWT secret key
- [ ] Enable SSL/TLS
- [ ] Configure proper CORS origins
- [ ] Set up database backups
- [ ] Enable monitoring and alerts
- [ ] Configure log aggregation
- [ ] Set up rate limiting
- [ ] Review security settings
- [ ] Load test the service
- [ ] Prepare rollback plan
