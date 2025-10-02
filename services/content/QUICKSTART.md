# Content Management Service - Quick Start Guide

## TL;DR

```bash
# 1. Start PostgreSQL
docker run -d --name postgres \
  -e POSTGRES_DB=content_db \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 postgres:15-alpine

# 2. Run migrations
migrate -path migrations \
  -database "postgres://postgres:postgres@localhost:5432/content_db?sslmode=disable" up

# 3. Start the service
go run cmd/server/main.go

# 4. Test it
curl http://localhost:8080/health
```

## Using Docker Compose (Recommended)

```bash
# Start everything
docker-compose up -d

# View logs
docker-compose logs -f content-service

# Stop everything
docker-compose down
```

## API Examples

### 1. Register a User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "author@example.com",
    "username": "author",
    "password": "password123",
    "role": "author"
  }'
```

### 2. Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "author@example.com",
    "password": "password123"
  }'

# Save the token from response
export TOKEN="your_jwt_token_here"
```

### 3. Create a Post

```bash
curl -X POST http://localhost:8080/api/v1/content \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "type": "post",
    "title": "My First Post",
    "body": "This is my first blog post!",
    "excerpt": "Introduction to content management",
    "status": "draft",
    "tags": ["tutorial", "cms"],
    "categories": ["technology"]
  }'

# Save the content ID from response
export CONTENT_ID="content_id_here"
```

### 4. Update Content

```bash
curl -X PUT http://localhost:8080/api/v1/content/$CONTENT_ID \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "My Updated Post",
    "body": "Updated content body"
  }'
```

### 5. Publish Content

```bash
curl -X POST http://localhost:8080/api/v1/content/$CONTENT_ID/publish \
  -H "Authorization: Bearer $TOKEN"
```

### 6. Search Content

```bash
# Search published posts
curl "http://localhost:8080/api/v1/content?status=published&type=post"

# Full-text search
curl "http://localhost:8080/api/v1/content?query=blog&page=1&page_size=10"

# Filter by tags
curl "http://localhost:8080/api/v1/content?tags=tutorial,cms"
```

### 7. Upload Media

```bash
curl -X POST http://localhost:8080/api/v1/media/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/path/to/image.jpg"
```

### 8. Get Content by Slug

```bash
curl http://localhost:8080/api/v1/content/slug/my-updated-post-1234567890
```

### 9. Get Version History

```bash
curl http://localhost:8080/api/v1/content/$CONTENT_ID/versions
```

## Configuration

Copy and edit the config file:

```bash
cp config.example.yaml config.yaml
# Edit config.yaml with your settings
```

Or use environment variables:

```bash
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export DATABASE_USER=postgres
export DATABASE_PASSWORD=postgres
export DATABASE_DBNAME=content_db
export JWT_SECRET=your-secret-key
```

## Makefile Commands

```bash
make help          # Show all available commands
make build         # Build the binary
make run           # Run the service
make test          # Run tests
make migrate-up    # Apply migrations
make migrate-down  # Rollback migrations
make docker-build  # Build Docker image
make docker-run    # Run Docker container
```

## Default Credentials

The migration creates a default admin user:
- **Email**: admin@example.com
- **Password**: admin123
- **Role**: admin

**IMPORTANT**: Change this password in production!

## Service URLs

- **API**: http://localhost:8080/api/v1
- **Health**: http://localhost:8080/health
- **Metrics**: http://localhost:8080/metrics
- **Uploads**: http://localhost:8080/uploads

## User Roles & Permissions

- **Admin**: Full access to everything
- **Editor**: Read, create, update all content
- **Author**: Read, create, update own content
- **Viewer**: Read-only access

## Content Types

- **post**: Blog posts
- **article**: Long-form articles
- **media**: Uploaded files (images, videos, PDFs)

## Content Status

- **draft**: Not publicly visible
- **published**: Publicly accessible
- **archived**: Hidden from public but preserved

## Troubleshooting

### Database Connection Error

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Test connection
psql -h localhost -U postgres -d content_db
```

### Migration Errors

```bash
# Check current migration version
migrate -path migrations \
  -database "postgres://postgres:postgres@localhost:5432/content_db?sslmode=disable" version

# Force to specific version (use with caution)
migrate -path migrations \
  -database "postgres://postgres:postgres@localhost:5432/content_db?sslmode=disable" force 1
```

### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or change port in config.yaml
server:
  port: 8081
```

## Production Checklist

- [ ] Change JWT secret in config
- [ ] Change default admin password
- [ ] Set up proper database backups
- [ ] Configure SSL/TLS
- [ ] Set up reverse proxy (nginx)
- [ ] Enable rate limiting
- [ ] Configure log rotation
- [ ] Set up monitoring alerts
- [ ] Review file upload limits
- [ ] Configure CORS properly
- [ ] Set up CI/CD pipeline

## Performance Tips

1. **Database Indexes**: Already optimized with GIN indexes
2. **Connection Pool**: Adjust in main.go based on load
3. **File Storage**: Use S3/cloud storage for production
4. **Caching**: Add Redis for frequently accessed content
5. **CDN**: Use CDN for media files

## Support

See the main README.md for detailed documentation.
