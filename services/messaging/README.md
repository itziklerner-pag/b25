# Messaging Service

Real-time messaging service with WebSocket support for direct messages, group chats, and team collaboration.

## Features

- **Real-time Messaging**: WebSocket-based instant message delivery
- **Direct Messages**: One-on-one conversations
- **Group Chats**: Multi-user conversations with member management
- **Message Persistence**: PostgreSQL-backed message history
- **Read Receipts**: Track message read status
- **Typing Indicators**: Real-time typing status
- **File Attachments**: Support for file sharing in messages
- **Message Search**: Full-text search across conversations
- **Online Presence**: Real-time user online/offline status
- **Message Threading**: Reply to specific messages
- **Message Reactions**: Emoji reactions to messages

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                 Messaging Service                        │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │            WebSocket Server                       │  │
│  │  - Real-time message delivery                     │  │
│  │  - Typing indicators                              │  │
│  │  - Presence tracking                              │  │
│  └──────────────────────────────────────────────────┘  │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │            REST API                               │  │
│  │  - Conversation management                        │  │
│  │  - Message history                                │  │
│  │  - User management                                │  │
│  │  - File uploads                                   │  │
│  └──────────────────────────────────────────────────┘  │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │            Message Queue (NATS)                   │  │
│  │  - Reliable message delivery                      │  │
│  │  - Message persistence                            │  │
│  │  - Event streaming                                │  │
│  └──────────────────────────────────────────────────┘  │
│                                                          │
└────────┬──────────────────────────────────────┬─────────┘
         │                                       │
    ┌────▼────────┐                        ┌────▼─────────┐
    │ PostgreSQL  │                        │   Redis      │
    │             │                        │              │
    │ - Messages  │                        │ - Presence   │
    │ - Convos    │                        │ - Typing     │
    │ - Users     │                        │ - Cache      │
    └─────────────┘                        └──────────────┘
```

## API Endpoints

### REST API

#### Conversations
- `POST /api/v1/conversations` - Create conversation
- `GET /api/v1/conversations` - List conversations
- `GET /api/v1/conversations/:id` - Get conversation
- `PUT /api/v1/conversations/:id` - Update conversation
- `DELETE /api/v1/conversations/:id` - Delete conversation
- `POST /api/v1/conversations/:id/members` - Add member
- `DELETE /api/v1/conversations/:id/members/:userId` - Remove member

#### Messages
- `POST /api/v1/conversations/:id/messages` - Send message
- `GET /api/v1/conversations/:id/messages` - Get messages (paginated)
- `GET /api/v1/messages/:id` - Get specific message
- `PUT /api/v1/messages/:id` - Edit message
- `DELETE /api/v1/messages/:id` - Delete message
- `POST /api/v1/messages/:id/reactions` - Add reaction
- `DELETE /api/v1/messages/:id/reactions/:emoji` - Remove reaction
- `POST /api/v1/messages/:id/read` - Mark as read

#### Search
- `GET /api/v1/search/messages?q=query` - Search messages
- `GET /api/v1/search/conversations?q=query` - Search conversations

#### Files
- `POST /api/v1/files` - Upload file
- `GET /api/v1/files/:id` - Download file
- `DELETE /api/v1/files/:id` - Delete file

#### Users
- `GET /api/v1/users/:id/presence` - Get user presence
- `GET /api/v1/users/online` - Get online users

### WebSocket API

Connect: `ws://localhost:9097/ws?token=<jwt>`

#### Client -> Server Messages

```json
{
  "type": "message.send",
  "conversation_id": "uuid",
  "content": "Hello world",
  "metadata": {}
}

{
  "type": "typing.start",
  "conversation_id": "uuid"
}

{
  "type": "typing.stop",
  "conversation_id": "uuid"
}

{
  "type": "presence.update",
  "status": "online|away|busy|offline"
}

{
  "type": "message.read",
  "message_id": "uuid"
}
```

#### Server -> Client Messages

```json
{
  "type": "message.new",
  "data": {
    "id": "uuid",
    "conversation_id": "uuid",
    "sender_id": "uuid",
    "content": "Hello",
    "created_at": "2025-10-03T..."
  }
}

{
  "type": "typing.indicator",
  "conversation_id": "uuid",
  "user_id": "uuid",
  "is_typing": true
}

{
  "type": "presence.changed",
  "user_id": "uuid",
  "status": "online"
}

{
  "type": "message.read",
  "message_id": "uuid",
  "user_id": "uuid"
}
```

## Configuration

Configuration via environment variables or `config.yaml`:

```yaml
server:
  http_port: 9097
  websocket_path: /ws

database:
  host: localhost
  port: 5432
  name: messaging
  user: postgres
  password: postgres

redis:
  host: localhost
  port: 6379
  db: 0

nats:
  url: nats://localhost:4222

storage:
  type: local  # or s3
  path: ./uploads
  max_file_size: 10485760  # 10MB

auth:
  jwt_secret: your-secret-key
  token_expiry: 24h

logging:
  level: info
  format: json
```

## Running

### Development

```bash
# Install dependencies
go mod download

# Run migrations
go run cmd/migrate/main.go up

# Run server
go run cmd/server/main.go
```

### Docker

```bash
# Build
docker build -t messaging-service .

# Run
docker-compose up
```

### Production

```bash
# Build binary
go build -o messaging-server cmd/server/main.go

# Run
./messaging-server
```

## Database Schema

### Tables

- `users` - User information
- `conversations` - Conversation metadata
- `conversation_members` - Conversation membership
- `messages` - Message content
- `message_reactions` - Message reactions
- `message_read_receipts` - Read tracking
- `files` - File metadata

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./...

# Load test
go run tests/load/main.go
```

## Metrics

Prometheus metrics available at `/metrics`:

- `messaging_websocket_connections` - Active WebSocket connections
- `messaging_messages_sent_total` - Total messages sent
- `messaging_messages_received_total` - Total messages received
- `messaging_message_latency_seconds` - Message delivery latency
- `messaging_conversations_total` - Total conversations
- `messaging_users_online` - Currently online users

## Performance

- Supports 10,000+ concurrent WebSocket connections
- Message delivery latency < 100ms p99
- Handles 1,000+ messages/second
- Search response time < 200ms

## Security

- JWT-based authentication
- TLS/SSL support
- Rate limiting per user
- Input validation and sanitization
- SQL injection prevention
- XSS protection
- File upload validation

## License

See LICENSE file
