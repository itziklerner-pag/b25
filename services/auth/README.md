# Authentication Service

Production-ready authentication service for the B25 HFT Trading System. Provides secure user authentication using JWT tokens with refresh token rotation.

## JavaScript-Only Policy

**IMPORTANT:** This service follows a strict JavaScript-only policy:

- All code must be written in pure JavaScript (ES6+)
- NO TypeScript syntax or type annotations allowed
- Use `.js` file extensions only (no `.ts` files)
- Use JSDoc comments for type documentation when needed
- Focus on clean, well-documented JavaScript code

See the main [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

## Features

- User registration with email/password
- Secure password hashing with bcrypt (12 rounds)
- JWT-based authentication with access and refresh tokens
- Token refresh mechanism with rotation
- Password strength validation
- Input validation and sanitization
- Rate limiting
- CORS support
- Comprehensive error handling
- Health check endpoints
- Structured logging
- Graceful shutdown
- Docker support

## Tech Stack

- **Runtime**: Node.js 20
- **Language**: JavaScript (ES6+)
- **Framework**: Express.js
- **Database**: PostgreSQL
- **Authentication**: JWT (jsonwebtoken)
- **Password Hashing**: bcrypt
- **Validation**: express-validator
- **Security**: helmet, cors, rate-limiting

## API Endpoints

### Authentication

#### POST /auth/register
Register a new user account.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
    "expiresIn": 900
  },
  "timestamp": "2025-10-03T12:00:00.000Z"
}
```

#### POST /auth/login
Login with existing credentials.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
    "expiresIn": 900
  },
  "timestamp": "2025-10-03T12:00:00.000Z"
}
```

#### POST /auth/refresh
Refresh access token using refresh token.

**Request:**
```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
    "expiresIn": 900
  },
  "timestamp": "2025-10-03T12:00:00.000Z"
}
```

#### POST /auth/logout
Logout and revoke refresh token.

**Request:**
```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Logged out successfully"
  },
  "timestamp": "2025-10-03T12:00:00.000Z"
}
```

#### GET /auth/verify
Verify access token and get user information.

**Headers:**
```
Authorization: Bearer <access-token>
```

**Response:**
```json
{
  "success": true,
  "data": {
    "userId": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com"
  },
  "timestamp": "2025-10-03T12:00:00.000Z"
}
```

### Health Checks

#### GET /health
Comprehensive health check including database connectivity.

#### GET /health/ready
Kubernetes readiness probe.

#### GET /health/live
Kubernetes liveness probe.

## Password Requirements

Passwords must meet the following criteria:
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one number
- At least one special character

## Environment Variables

See `.env.example` for all available configuration options.

### Required Variables

```bash
JWT_ACCESS_SECRET=your-secret-key
JWT_REFRESH_SECRET=your-secret-key
```

### Database Configuration

```bash
DB_HOST=localhost
DB_PORT=5432
DB_NAME=b25_auth
DB_USER=postgres
DB_PASSWORD=postgres
```

## Getting Started

### Prerequisites

- Node.js 20+
- PostgreSQL 14+
- npm or yarn

### Development Setup

1. Install dependencies:
```bash
npm install
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Update `.env` with your configuration, especially JWT secrets:
```bash
# Generate secrets (Linux/Mac)
openssl rand -base64 64
```

4. Start PostgreSQL database (or use Docker Compose)

5. Run in development mode:
```bash
npm run dev
```

The service will:
- Start on port 9097 (or PORT from .env)
- Automatically run database migrations
- Enable hot-reloading with nodemon

### Production Build

```bash
# Build (if needed for any transpilation)
npm run build

# Start production server
npm start
```

## Docker Deployment

### Build Image

```bash
docker build -t b25/auth-service:latest .
```

### Run Container

```bash
docker run -d \
  --name auth-service \
  -p 9097:9097 \
  -e DB_HOST=postgres \
  -e DB_PASSWORD=yourpassword \
  -e JWT_ACCESS_SECRET=your-secret \
  -e JWT_REFRESH_SECRET=your-secret \
  b25/auth-service:latest
```

### Docker Compose

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:14-alpine
    environment:
      POSTGRES_DB: b25_auth
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  auth-service:
    build: .
    ports:
      - "9097:9097"
    environment:
      DB_HOST: postgres
      DB_NAME: b25_auth
      DB_USER: postgres
      DB_PASSWORD: postgres
      JWT_ACCESS_SECRET: ${JWT_ACCESS_SECRET}
      JWT_REFRESH_SECRET: ${JWT_REFRESH_SECRET}
    depends_on:
      - postgres

volumes:
  postgres_data:
```

## Database Schema

The service automatically creates the following tables:

### users
- `id` (UUID, Primary Key)
- `email` (VARCHAR, Unique)
- `password_hash` (VARCHAR)
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)
- `last_login` (TIMESTAMP)
- `is_active` (BOOLEAN)

### refresh_tokens
- `id` (UUID, Primary Key)
- `user_id` (UUID, Foreign Key)
- `token_hash` (VARCHAR)
- `expires_at` (TIMESTAMP)
- `created_at` (TIMESTAMP)
- `revoked` (BOOLEAN)

## Security Features

### Password Security
- Bcrypt hashing with 12 salt rounds
- Password strength validation
- No plain-text password storage

### Token Security
- JWT with HS256 algorithm
- Short-lived access tokens (15 minutes default)
- Long-lived refresh tokens (7 days default)
- Refresh token rotation on use
- Token revocation support
- Refresh tokens hashed (SHA-256) before storage

### API Security
- Helmet.js for security headers
- CORS configuration
- Rate limiting (100 requests per 15 minutes)
- Input validation and sanitization
- SQL injection prevention (parameterized queries)
- XSS protection

## Error Handling

All errors follow a consistent format:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {}
  },
  "timestamp": "2025-10-03T12:00:00.000Z"
}
```

### Error Codes

- `VALIDATION_ERROR` - Input validation failed
- `DUPLICATE_USER` - Email already registered
- `INVALID_CREDENTIALS` - Wrong email or password
- `INVALID_TOKEN` - Token is invalid
- `TOKEN_EXPIRED` - Token has expired
- `TOKEN_REVOKED` - Token has been revoked
- `USER_NOT_FOUND` - User does not exist
- `RATE_LIMIT_EXCEEDED` - Too many requests
- `INTERNAL_ERROR` - Server error

## Monitoring

### Logs

Structured JSON logs:
```json
{
  "timestamp": "2025-10-03T12:00:00.000Z",
  "level": "INFO",
  "context": "AuthService",
  "message": "User login successful",
  "meta": {
    "userId": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com"
  }
}
```

### Health Checks

- `/health` - Overall health status
- `/health/ready` - Readiness for traffic
- `/health/live` - Service liveness

## Integration with B25 System

This service integrates with the B25 monorepo:

- Port: 9097 (default)
- Database: PostgreSQL (shared or dedicated)
- Used by: Web dashboard, Terminal UI, Other services
- Shared libraries: Can use shared types from `/shared/lib/`

## Performance

- Connection pooling (20 connections default)
- Async/await for non-blocking operations
- Efficient bcrypt rounds (12)
- Database indexes on email and token lookups
- Token cleanup jobs for expired tokens

## Testing

```bash
# Run tests (when implemented)
npm test

# Test coverage
npm run test:coverage
```

## Maintenance

### Token Cleanup

Periodically clean up expired tokens:

```sql
SELECT cleanup_expired_tokens();
```

Consider setting up a cron job or scheduled task.

## Contributing

Follow the B25 monorepo contribution guidelines in `/CONTRIBUTING.md`.

## License

See LICENSE file in the repository root.

## Support

For issues and questions, please refer to the B25 project documentation.

---

**Service Status**: Production Ready
**Version**: 1.0.0
**Maintained by**: B25 Team
