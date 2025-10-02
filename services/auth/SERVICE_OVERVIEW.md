# Authentication Service - Technical Overview

## Service Summary

Production-ready JWT-based authentication service built with Node.js/TypeScript for the B25 HFT Trading System.

**Status**: ✅ Build Complete - Ready for Deployment
**Port**: 9097
**Database**: PostgreSQL
**Language**: TypeScript (Node.js 20)

## File Structure

```
services/auth/
├── src/
│   ├── config/
│   │   └── index.ts                    # Environment configuration
│   ├── database/
│   │   ├── pool.ts                     # PostgreSQL connection pool
│   │   ├── schema.sql                  # Database schema
│   │   ├── migrations.ts               # Migration runner
│   │   └── repositories/
│   │       ├── user.repository.ts      # User CRUD operations
│   │       └── token.repository.ts     # Token management
│   ├── middleware/
│   │   ├── auth.ts                     # JWT authentication middleware
│   │   ├── validation.ts               # Input validation rules
│   │   └── error-handler.ts            # Global error handler
│   ├── routes/
│   │   ├── auth.routes.ts              # Authentication endpoints
│   │   └── health.routes.ts            # Health check endpoints
│   ├── services/
│   │   ├── auth.service.ts             # Authentication business logic
│   │   └── jwt.service.ts              # JWT token operations
│   ├── types/
│   │   └── index.ts                    # TypeScript interfaces
│   ├── utils/
│   │   ├── logger.ts                   # Structured logging
│   │   ├── password.ts                 # Password hashing/validation
│   │   └── response.ts                 # API response helpers
│   ├── app.ts                          # Express app configuration
│   └── server.ts                       # Server entry point
├── Dockerfile                           # Multi-stage production build
├── docker-compose.yml                   # Docker Compose configuration
├── .env.example                         # Environment variables template
├── .dockerignore                        # Docker ignore rules
├── .gitignore                           # Git ignore rules
├── package.json                         # Dependencies and scripts
├── tsconfig.json                        # TypeScript configuration
├── README.md                            # User documentation
└── progress.md                          # Build progress tracker
```

## API Endpoints

### Authentication
- `POST /auth/register` - User registration with email/password
- `POST /auth/login` - User login with JWT token issuance
- `POST /auth/refresh` - Refresh access token using refresh token
- `POST /auth/logout` - Logout and revoke refresh token
- `GET /auth/verify` - Verify access token (requires authentication)

### Health Checks
- `GET /health` - Comprehensive health check with DB status
- `GET /health/ready` - Kubernetes readiness probe
- `GET /health/live` - Kubernetes liveness probe

## Security Features

### Password Security
- **Bcrypt hashing**: 12 salt rounds for password hashing
- **Password strength validation**: Min 8 chars, uppercase, lowercase, number, special char
- **No plaintext storage**: All passwords stored as bcrypt hashes

### Token Security
- **JWT implementation**: Access tokens (15min) + Refresh tokens (7 days)
- **Token rotation**: Refresh tokens are rotated on each use
- **Token revocation**: Refresh tokens can be revoked
- **SHA-256 hashing**: Refresh tokens hashed before database storage
- **Stateless authentication**: No session storage required

### API Security
- **Helmet.js**: Security headers (XSS, clickjacking, etc.)
- **CORS**: Configurable origin whitelist
- **Rate limiting**: 100 requests per 15 minutes per IP
- **Input validation**: express-validator for all inputs
- **SQL injection prevention**: Parameterized queries
- **Error sanitization**: No sensitive data in error responses

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE
);
```

### Refresh Tokens Table
```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    revoked BOOLEAN DEFAULT FALSE
);
```

## Technology Stack

### Core
- **Runtime**: Node.js 20
- **Language**: TypeScript 5.9
- **Framework**: Express.js 5

### Security
- **Authentication**: jsonwebtoken 9.0
- **Password Hashing**: bcrypt 6.0
- **Security Headers**: helmet 8.1
- **CORS**: cors 2.8
- **Rate Limiting**: express-rate-limit 8.1

### Database
- **Database**: PostgreSQL (via pg 8.16)
- **Connection Pooling**: Built-in pg pool
- **Migrations**: SQL-based migrations

### Validation
- **Input Validation**: express-validator 7.2
- **Type Safety**: TypeScript strict mode

### DevOps
- **Build**: TypeScript compiler
- **Development**: nodemon + ts-node
- **Containerization**: Docker multi-stage build
- **Environment**: dotenv

## Key Design Patterns

### Repository Pattern
Separate data access logic from business logic:
- `user.repository.ts` - User data operations
- `token.repository.ts` - Token data operations

### Service Layer
Business logic separated from HTTP handlers:
- `auth.service.ts` - Authentication logic
- `jwt.service.ts` - Token generation/validation

### Middleware Pattern
Cross-cutting concerns as Express middleware:
- Authentication, validation, error handling, logging

### Dependency Injection
Services injected into route handlers for testability

## Configuration

### Environment Variables
All configuration via environment variables:
- Database connection settings
- JWT secrets and expiry times
- Server port and CORS origins
- Rate limiting parameters

### Configuration Files
- `.env.example` - Template for environment variables
- `tsconfig.json` - TypeScript compiler options
- `package.json` - Dependencies and scripts

## Development Workflow

### Local Development
```bash
npm install              # Install dependencies
cp .env.example .env    # Create environment file
# Edit .env with your settings
npm run dev             # Start with hot reload
```

### Production Build
```bash
npm run build           # Compile TypeScript
npm start               # Start production server
```

### Docker Deployment
```bash
docker build -t b25/auth-service .
docker run -p 9097:9097 b25/auth-service
```

## Testing Strategy

### Unit Tests (Ready for Implementation)
- Repository layer tests
- Service layer tests
- Utility function tests
- Middleware tests

### Integration Tests (Ready for Implementation)
- API endpoint tests
- Database integration tests
- Token flow tests

### Security Tests (Ready for Implementation)
- Password validation tests
- Token security tests
- Rate limiting tests
- Input validation tests

## Performance Optimizations

### Database
- Connection pooling (20 connections)
- Indexes on email and token lookups
- Prepared statements for queries

### Caching
- JWT tokens cached client-side
- No database lookup for access token validation

### Async Operations
- All I/O operations use async/await
- Non-blocking event loop

## Monitoring & Observability

### Structured Logging
JSON logs with:
- Timestamp
- Log level (ERROR, WARN, INFO, DEBUG)
- Context
- Message
- Metadata

### Health Checks
- Database connectivity
- Service uptime
- Response time metrics

### Metrics (Ready for Integration)
- Request count
- Response times
- Error rates
- Token issuance rates

## Error Handling

### Error Types
- `VALIDATION_ERROR` - Input validation failures
- `DUPLICATE_USER` - Email already registered
- `INVALID_CREDENTIALS` - Wrong email/password
- `INVALID_TOKEN` - Token is invalid
- `TOKEN_EXPIRED` - Token has expired
- `TOKEN_REVOKED` - Token has been revoked
- `USER_NOT_FOUND` - User doesn't exist
- `RATE_LIMIT_EXCEEDED` - Too many requests
- `INTERNAL_ERROR` - Server errors

### Error Response Format
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable message"
  },
  "timestamp": "2025-10-03T12:00:00.000Z"
}
```

## Integration with B25 System

### Service Discovery
- Fixed port: 9097
- Health checks for service registration

### Authentication Flow
1. User logs in → receives access + refresh tokens
2. Client includes access token in Authorization header
3. Services validate token using shared JWT secret
4. Token expires → client uses refresh token
5. Refresh token rotated on each use

### Cross-Service Auth
Other B25 services can:
- Validate access tokens using JWT secret
- Verify tokens via `/auth/verify` endpoint
- Implement their own JWT validation

## Production Readiness

### Checklist
- ✅ Secure password hashing
- ✅ JWT implementation with refresh tokens
- ✅ Token rotation and revocation
- ✅ Input validation
- ✅ Error handling
- ✅ Security middleware
- ✅ Rate limiting
- ✅ Health checks
- ✅ Graceful shutdown
- ✅ Docker support
- ✅ Structured logging
- ✅ Connection pooling
- ✅ Environment configuration
- ✅ API documentation
- ✅ TypeScript strict mode
- ✅ Non-root Docker user
- ✅ Multi-stage Docker build

### Deployment Considerations
- Set strong JWT secrets in production
- Configure database connection properly
- Set appropriate rate limits
- Configure CORS for actual domains
- Enable HTTPS in production
- Set up log aggregation
- Monitor health endpoints
- Regular token cleanup jobs

## Maintenance

### Database Maintenance
```sql
-- Cleanup expired tokens (run periodically)
SELECT cleanup_expired_tokens();
```

### Monitoring
- Watch error rates in logs
- Monitor health check failures
- Track token generation rates
- Alert on database connectivity issues

## Future Enhancements

### Potential Improvements
- OAuth2/Social login integration
- Multi-factor authentication (MFA)
- Email verification workflow
- Password reset functionality
- User profile management
- Role-based access control (RBAC)
- API key management
- Audit logging
- Metrics export (Prometheus)
- OpenTelemetry tracing

### Testing
- Comprehensive unit tests
- Integration test suite
- End-to-end tests
- Load testing
- Security penetration testing

## Support & Documentation

- **README.md**: User-facing documentation
- **progress.md**: Build progress and notes
- **This file**: Technical implementation details
- **Code comments**: Inline documentation

---

**Built for**: B25 HFT Trading System
**Service Version**: 1.0.0
**Build Date**: 2025-10-03
**Status**: Production Ready
