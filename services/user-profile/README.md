# User Profile Service

A production-ready microservice for managing user profiles with CRUD operations, privacy settings, and authentication.

## JavaScript-Only Policy

**IMPORTANT:** This service follows a strict JavaScript-only policy:

- All code must be written in pure JavaScript (ES6+)
- NO TypeScript syntax or type annotations allowed
- Use `.js` file extensions only (no `.ts` files)
- Use JSDoc comments for type documentation when needed
- Focus on clean, well-documented JavaScript code

See the main [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

## Features

- **CRUD Operations**: Create, read, update, and delete user profiles
- **Privacy Settings**: Granular control over profile visibility and data sharing
- **Authentication**: JWT-based authentication with role-based access control
- **Validation**: Comprehensive input validation using Joi schemas
- **Search**: Full-text search across profile names and bios
- **Metrics**: Prometheus metrics endpoint for monitoring
- **Health Checks**: Kubernetes-ready health and readiness probes
- **Rate Limiting**: Built-in rate limiting for API protection
- **Logging**: Structured logging with Winston
- **JSDoc**: Type documentation using JSDoc comments

## Tech Stack

- **Runtime**: Node.js 18+
- **Language**: JavaScript (ES6+)
- **Framework**: Express.js
- **Database**: PostgreSQL 15+
- **Authentication**: JWT (jsonwebtoken)
- **Validation**: Joi
- **Logging**: Winston
- **Metrics**: prom-client
- **Testing**: Jest + Supertest

## Prerequisites

- Node.js 18 or higher
- PostgreSQL 15 or higher
- npm or yarn

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f user-profile

# Stop services
docker-compose down
```

The service will be available at:
- API: http://localhost:9100
- Metrics: http://localhost:9101/metrics
- Health: http://localhost:9100/health

### Local Development

1. **Install dependencies**
```bash
npm install
```

2. **Set up environment variables**
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. **Start PostgreSQL**
```bash
# Using Docker
docker run -d \
  --name postgres \
  -e POSTGRES_DB=b25_user_profiles \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:15-alpine
```

4. **Run database migrations**
```bash
npm run build
npm run migrate
```

5. **Start development server**
```bash
npm run dev
```

## API Documentation

### Base URL
```
http://localhost:9100/api/v1
```

### Authentication
Most endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

### Endpoints

#### Create Profile
```http
POST /api/v1/profiles
Authorization: Bearer <token>
Content-Type: application/json

{
  "userId": "user-123",
  "name": "John Doe",
  "bio": "Software developer",
  "avatarUrl": "https://example.com/avatar.jpg",
  "preferences": {
    "theme": "dark",
    "language": "en",
    "timezone": "America/New_York"
  },
  "privacySettings": {
    "profileVisibility": "public",
    "showEmail": false,
    "showBio": true
  }
}
```

#### Get Profile by ID
```http
GET /api/v1/profiles/:id
```

#### Get Profile by User ID
```http
GET /api/v1/profiles/user/:userId
```

#### Get Current User's Profile
```http
GET /api/v1/profiles/me
Authorization: Bearer <token>
```

#### List Profiles
```http
GET /api/v1/profiles?page=1&limit=20
```

#### Search Profiles
```http
GET /api/v1/profiles/search?q=john&page=1&limit=20
```

#### Update Profile
```http
PUT /api/v1/profiles/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Jane Doe",
  "bio": "Updated bio"
}
```

#### Update Privacy Settings
```http
PATCH /api/v1/profiles/:id/privacy
Authorization: Bearer <token>
Content-Type: application/json

{
  "profileVisibility": "private",
  "showEmail": false,
  "allowMessaging": true
}
```

#### Delete Profile
```http
DELETE /api/v1/profiles/:id
Authorization: Bearer <token>
```

### Response Format

#### Success Response
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "userId": "user-123",
    "name": "John Doe",
    "bio": "Software developer",
    "avatarUrl": "https://example.com/avatar.jpg",
    "preferences": {},
    "privacySettings": {},
    "createdAt": "2025-10-03T00:00:00.000Z",
    "updatedAt": "2025-10-03T00:00:00.000Z"
  },
  "meta": {
    "timestamp": "2025-10-03T00:00:00.000Z",
    "requestId": "req-123"
  }
}
```

#### Error Response
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      {
        "field": "name",
        "message": "Name is required"
      }
    ]
  },
  "meta": {
    "timestamp": "2025-10-03T00:00:00.000Z"
  }
}
```

### Privacy Levels

- **public**: Profile visible to everyone
- **friends**: Profile visible to friends only
- **private**: Profile visible only to owner

## Database Schema

The service uses PostgreSQL with the following main table:

```sql
user_profiles (
  id UUID PRIMARY KEY,
  user_id VARCHAR(255) UNIQUE NOT NULL,
  name VARCHAR(255) NOT NULL,
  bio TEXT,
  avatar_url TEXT,
  preferences JSONB,
  privacy_settings JSONB,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
)
```

See `src/db/schema.sql` for the complete schema.

## Configuration

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| NODE_ENV | Environment (development/production) | development |
| PORT | HTTP server port | 9100 |
| HOST | Server host | 0.0.0.0 |
| DB_HOST | PostgreSQL host | localhost |
| DB_PORT | PostgreSQL port | 5432 |
| DB_NAME | Database name | b25_user_profiles |
| DB_USER | Database user | postgres |
| DB_PASSWORD | Database password | postgres |
| JWT_SECRET | JWT signing secret | (required in production) |
| JWT_EXPIRY | JWT expiration time | 24h |
| LOG_LEVEL | Logging level | info |
| RATE_LIMIT_WINDOW_MS | Rate limit window | 900000 (15 min) |
| RATE_LIMIT_MAX_REQUESTS | Max requests per window | 100 |
| CORS_ORIGIN | Allowed CORS origins | http://localhost:3000 |

## Development

### Scripts

```bash
# Development with hot reload
npm run dev

# Build (if needed for any transpilation)
npm run build

# Run tests
npm test

# Run tests with coverage
npm run test:coverage

# Run tests in watch mode
npm run test:watch

# Lint code
npm run lint

# Format code
npm run format

# Database migration
npm run migrate
```

### Project Structure

```
user-profile/
├── src/
│   ├── config/          # Configuration management
│   ├── controllers/     # Request handlers
│   ├── db/              # Database connection and schema
│   ├── middleware/      # Express middleware
│   ├── models/          # Data models and repositories
│   ├── routes/          # API routes
│   ├── types/           # JSDoc type definitions and helpers
│   ├── utils/           # Utility functions
│   ├── validators/      # Input validation schemas
│   ├── app.ts           # Express app configuration
│   └── index.ts         # Entry point
├── __tests__/           # Test files
├── Dockerfile           # Docker image definition
├── docker-compose.yml   # Docker Compose configuration
└── package.json         # Dependencies and scripts
```

## Health Checks

### Basic Health Check
```bash
curl http://localhost:9100/health
```

### Detailed Health Check
```bash
curl http://localhost:9100/health/detailed
```

### Readiness Probe (K8s)
```bash
curl http://localhost:9100/health/ready
```

### Liveness Probe (K8s)
```bash
curl http://localhost:9100/health/live
```

## Metrics

Prometheus metrics are available at:
```
http://localhost:9101/metrics
```

Available metrics:
- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - HTTP request duration
- `http_active_connections` - Active HTTP connections
- `db_query_duration_seconds` - Database query duration
- `db_connection_pool_size` - Database connection pool size
- Node.js default metrics (memory, CPU, etc.)

## Security

### Best Practices Implemented

- **JWT Authentication**: Secure token-based authentication
- **Input Validation**: All inputs validated with Joi schemas
- **Rate Limiting**: Protection against brute force attacks
- **Helmet**: Security headers with helmet.js
- **CORS**: Configurable CORS policy
- **SQL Injection Prevention**: Parameterized queries
- **Password Hashing**: bcrypt for password hashing (if needed)
- **Environment Variables**: Sensitive data in environment variables
- **Non-root Docker User**: Container runs as non-root user

### Production Checklist

- [ ] Set strong `JWT_SECRET` in production
- [ ] Configure proper `CORS_ORIGIN` values
- [ ] Use environment variables for all secrets
- [ ] Enable HTTPS/TLS
- [ ] Set up database backups
- [ ] Configure log aggregation
- [ ] Set up monitoring and alerting
- [ ] Review and adjust rate limits
- [ ] Enable database connection pooling
- [ ] Set up database read replicas (if needed)

## Testing

```bash
# Run all tests
npm test

# Run with coverage
npm run test:coverage

# Watch mode for development
npm run test:watch
```

Tests cover:
- Profile CRUD operations
- Authentication and authorization
- Input validation
- Privacy settings
- Error handling
- Health checks
- Metrics endpoints

## Troubleshooting

### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check database logs
docker logs postgres

# Test connection manually
psql -h localhost -U postgres -d b25_user_profiles
```

### Port Already in Use

```bash
# Find process using port 9100
lsof -i :9100

# Kill the process
kill -9 <PID>
```

## Contributing

See the main repository [CONTRIBUTING.md](../../CONTRIBUTING.md) for contribution guidelines.

## License

MIT

## Support

For issues and questions:
- Create an issue in the repository
- Check existing documentation
- Review API documentation above

## Roadmap

- [ ] Rate limiting per user
- [ ] Profile picture upload support
- [ ] Email verification
- [ ] Profile activity logging
- [ ] GraphQL API
- [ ] WebSocket support for real-time updates
- [ ] Redis caching layer
- [ ] Multi-region deployment support
