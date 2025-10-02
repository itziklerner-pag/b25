# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Install FFmpeg and libvips for media processing
RUN apk add --no-cache ffmpeg vips-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o media-service ./cmd/server

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates ffmpeg vips tzdata

# Create non-root user
RUN addgroup -g 1000 mediauser && \
    adduser -D -u 1000 -G mediauser mediauser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/media-service .
COPY --from=builder /build/config.example.yaml ./config.yaml

# Create directories
RUN mkdir -p /var/media /tmp/media-uploads && \
    chown -R mediauser:mediauser /var/media /tmp/media-uploads /app

USER mediauser

EXPOSE 9097

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9097/health || exit 1

ENTRYPOINT ["./media-service"]
