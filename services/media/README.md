# Media Service

A production-ready media management service for the B25 platform, providing efficient file upload, storage, processing, and delivery capabilities.

## Features

### Core Capabilities
- **File Upload & Storage**: Support for single and multipart uploads with S3-compatible storage
- **Image Processing**: Automatic image optimization, resizing, and thumbnail generation
- **Video Transcoding**: Multi-quality video processing with FFmpeg (480p, 720p, 1080p)
- **Streaming Support**: HTTP range request support for video streaming
- **CDN Integration**: Ready for CDN deployment with caching support
- **Media Metadata**: Comprehensive metadata extraction and management
- **Storage Quota**: User and organization-level storage quota management

### Security Features
- **File Validation**: MIME type and file size validation
- **Virus Scanning**: Optional ClamAV integration for malware detection
- **Filename Sanitization**: Protection against path traversal attacks
- **Signed URLs**: Temporary access URLs for secure file sharing

### Performance
- **Redis Caching**: Fast metadata and CDN cache management
- **Async Processing**: Background image and video processing
- **Connection Pooling**: Optimized database and storage connections
- **Horizontal Scaling**: Stateless design for easy scaling

## Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│       Media Service (Go)            │
│  ┌──────────────────────────────┐  │
│  │  RESTful API (Gin)           │  │
│  └──────────────────────────────┘  │
│  ┌──────┬──────────┬─────────────┐ │
│  │Upload│Processing│  Streaming  │ │
│  └──────┴──────────┴─────────────┘ │
└─────┬────────┬────────┬───────────┘
      │        │        │
      ▼        ▼        ▼
┌──────────┐ ┌────┐ ┌─────────┐
│PostgreSQL│ │S3  │ │  Redis  │
│(Metadata)│ │    │ │ (Cache) │
└──────────┘ └────┘ └─────────┘
```

## Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- FFmpeg (for video processing)
- PostgreSQL 15+
- Redis 7+
- MinIO or AWS S3

### Development Setup

1. **Clone and navigate to the service:**
```bash
cd services/media
```

2. **Copy configuration files:**
```bash
cp config.example.yaml config.yaml
cp .env.example .env
```

3. **Start infrastructure services:**
```bash
docker-compose up -d postgres redis minio
```

4. **Run the service:**
```bash
make run
```

The service will be available at `http://localhost:9097`

### Docker Deployment

**Build and run everything with Docker:**
```bash
docker-compose up -d
```

**Access services:**
- Media Service: http://localhost:9097
- MinIO Console: http://localhost:9001
- PostgreSQL: localhost:5433

## API Endpoints

### Health & Metrics
```
GET  /health              - Health check
GET  /metrics             - Prometheus metrics
```

### Media Management
```
POST   /api/v1/media/upload                    - Upload single file
GET    /api/v1/media/:id                       - Get media details
GET    /api/v1/media/:id/download              - Download file
GET    /api/v1/media/:id/stream                - Stream file (with range support)
GET    /api/v1/media                           - List media (with filters)
DELETE /api/v1/media/:id                       - Delete media
GET    /api/v1/media/:id/metadata              - Get metadata
```

### Image Operations
```
GET  /api/v1/media/:id/thumbnail/:size    - Get thumbnail (small/medium/large)
POST /api/v1/media/:id/resize             - Resize image
POST /api/v1/media/:id/convert            - Convert image format
```

### Video Operations
```
GET  /api/v1/media/:id/variants           - Get video variants
GET  /api/v1/media/:id/playlist           - Get HLS playlist
```

### Quota Management
```
GET  /api/v1/quota/user/:user_id          - Get user quota
GET  /api/v1/quota/org/:org_id            - Get org quota
PUT  /api/v1/quota/user/:user_id          - Set user quota limit
PUT  /api/v1/quota/org/:org_id            - Set org quota limit
```

### Statistics
```
GET  /api/v1/stats/user/:user_id          - Get user statistics
GET  /api/v1/stats/org/:org_id            - Get org statistics
```

### Admin
```
POST /api/v1/admin/quota/recalculate/:entity_id  - Recalculate quota
GET  /api/v1/admin/storage/status                - Storage status
```

## Usage Examples

### Upload a File

```bash
curl -X POST http://localhost:9097/api/v1/media/upload \
  -H "X-User-ID: user123" \
  -F "file=@image.jpg"
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "user123",
  "file_name": "550e8400-e29b-41d4-a716-446655440000.jpg",
  "original_name": "image.jpg",
  "mime_type": "image/jpeg",
  "media_type": "image",
  "size": 1048576,
  "status": "pending",
  "public_url": "http://localhost:9000/b25-media/...",
  "created_at": "2025-10-03T10:00:00Z"
}
```

### Get Media Details

```bash
curl http://localhost:9097/api/v1/media/550e8400-e29b-41d4-a716-446655440000 \
  -H "X-User-ID: user123"
```

### Download a File

```bash
curl -O http://localhost:9097/api/v1/media/550e8400-e29b-41d4-a716-446655440000/download \
  -H "X-User-ID: user123"
```

### List Media

```bash
curl "http://localhost:9097/api/v1/media?user_id=user123&type=image&limit=10" \
  -H "X-User-ID: user123"
```

### Get User Quota

```bash
curl http://localhost:9097/api/v1/quota/user/user123 \
  -H "X-User-ID: user123"
```

Response:
```json
{
  "id": "quota-id",
  "entity_id": "user123",
  "type": "user",
  "used": 52428800,
  "limit": 1073741824,
  "updated_at": "2025-10-03T10:00:00Z"
}
```

## Configuration

### Environment Variables

```bash
# Server
SERVER_PORT=9097
SERVER_HOST=0.0.0.0
SERVER_MODE=development

# Storage
STORAGE_TYPE=s3                      # s3 or local
S3_ENDPOINT=http://localhost:9000
S3_BUCKET=b25-media
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=b25_media
DB_PASSWORD=b25_media_pass
DB_NAME=b25_media

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Security
ENABLE_VIRUS_SCAN=false
CLAMAV_HOST=localhost
CLAMAV_PORT=3310

# Quota
QUOTA_ENABLED=true
DEFAULT_USER_QUOTA=1073741824    # 1GB
DEFAULT_ORG_QUOTA=10737418240    # 10GB
```

### Configuration File (config.yaml)

See `config.example.yaml` for detailed configuration options including:
- Image processing settings (quality, max dimensions, thumbnail sizes)
- Video transcoding profiles
- Security settings (allowed MIME types, file size limits)
- CDN configuration
- Logging settings

## Processing Pipeline

### Image Upload Flow
1. File uploaded via API
2. MIME type and size validation
3. Optional virus scanning
4. Upload to S3/storage
5. Create database record (status: pending)
6. **Async processing:**
   - Optimize and resize image
   - Generate thumbnails (small, medium, large)
   - Upload variants to storage
   - Update metadata
   - Set status to ready

### Video Upload Flow
1. File uploaded via API
2. Validation and virus scanning
3. Upload to S3/storage
4. Create database record (status: pending)
5. **Async processing:**
   - Extract metadata (duration, resolution)
   - Transcode to multiple qualities (480p, 720p, 1080p)
   - Generate thumbnail
   - Upload variants to storage
   - Update metadata
   - Set status to ready

## Storage Backends

### S3-Compatible Storage
- AWS S3
- MinIO (for development)
- DigitalOcean Spaces
- Any S3-compatible service

Configuration:
```yaml
storage:
  type: s3
  s3:
    endpoint: http://localhost:9000
    region: us-east-1
    bucket: b25-media
    access_key: your-access-key
    secret_key: your-secret-key
```

### Local Filesystem
For development or single-server deployments:
```yaml
storage:
  type: local
  local:
    base_path: /var/media/b25
```

## Quota Management

The service supports storage quotas at two levels:
- **User Quota**: Per-user storage limits
- **Organization Quota**: Per-organization storage limits

### Quota Enforcement
- Checked before upload
- Updated after successful upload
- Automatically recalculated on schedule
- Manual recalculation available via admin API

### Default Quotas
- User: 1GB (configurable)
- Organization: 10GB (configurable)

## Testing

### Run Tests
```bash
make test
```

### Run Tests with Coverage
```bash
make test-coverage
```

### Integration Tests
```bash
# Start test infrastructure
docker-compose up -d

# Run integration tests
go test -tags=integration ./tests/...
```

## Monitoring

### Prometheus Metrics
Available at `/metrics`:
- Upload count and size
- Processing duration
- Storage usage
- Quota utilization
- Error rates

### Health Check
Available at `/health`:
```json
{
  "status": "healthy",
  "time": "2025-10-03T10:00:00Z"
}
```

## Performance Tuning

### Database Connection Pool
```yaml
database:
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300
```

### Redis Connection Pool
```yaml
redis:
  pool_size: 10
```

### Upload Limits
```yaml
server:
  max_upload_size: 104857600  # 100MB

security:
  max_file_size: 104857600    # 100MB
```

## Security Best Practices

1. **Enable virus scanning in production:**
   ```yaml
   security:
     enable_virus_scan: true
   ```

2. **Use signed URLs for sensitive content:**
   ```go
   signedURL, err := storage.GetSignedURL(path, 3600) // 1 hour expiry
   ```

3. **Implement rate limiting** (via reverse proxy or API gateway)

4. **Use HTTPS in production** (via reverse proxy)

5. **Validate file types strictly:**
   ```yaml
   security:
     allowed_mime_types:
       - image/jpeg
       - image/png
       - video/mp4
   ```

## Troubleshooting

### Common Issues

**FFmpeg not found:**
```bash
# Install FFmpeg
# Ubuntu/Debian
sudo apt-get install ffmpeg

# Alpine (Docker)
apk add ffmpeg
```

**Database connection failed:**
- Check PostgreSQL is running
- Verify connection credentials
- Check firewall/network settings

**Storage upload failed:**
- Verify S3 credentials
- Check bucket permissions
- Ensure bucket exists

**Quota not enforced:**
- Enable quota in config: `quota.enabled: true`
- Initialize quota for users: `POST /api/v1/admin/quota/recalculate/:entity_id`

## Development

### Project Structure
```
services/media/
├── cmd/
│   └── server/          # Main entry point
├── internal/
│   ├── api/             # HTTP handlers and routing
│   ├── cache/           # Redis cache layer
│   ├── config/          # Configuration management
│   ├── database/        # Database layer and migrations
│   ├── models/          # Data models
│   ├── processing/      # Image and video processing
│   ├── quota/           # Quota management
│   ├── security/        # File validation and scanning
│   └── storage/         # Storage backends (S3, local)
├── tests/               # Tests
├── Dockerfile           # Docker image definition
├── docker-compose.yml   # Development environment
├── Makefile            # Build and development tasks
└── README.md           # This file
```

### Adding a New Storage Backend

1. Implement the `Storage` interface in `internal/storage/`:
```go
type Storage interface {
    Upload(path string, reader io.Reader, contentType string) (string, error)
    Download(path string) (io.ReadCloser, error)
    Delete(path string) error
    Exists(path string) (bool, error)
    GetURL(path string) string
    GetSignedURL(path string, expiry int) (string, error)
    Copy(sourcePath, destPath string) error
    GetSize(path string) (int64, error)
}
```

2. Add to `NewStorage()` factory function
3. Update configuration structure
4. Add tests

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for contribution guidelines.

## License

See [LICENSE](../../LICENSE) for license information.

## Support

For issues and questions:
- GitHub Issues: https://github.com/yourorg/b25/issues
- Documentation: https://docs.b25.dev

---

**Built with:**
- Go 1.21
- Gin Web Framework
- PostgreSQL
- Redis
- AWS S3 SDK
- FFmpeg
- Docker
