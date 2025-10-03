<<<<<<< HEAD
# Multi-stage build for minimal image size
=======
# Multi-stage build for Go Configuration Service

# Builder stage - compiles the application
>>>>>>> refs/remotes/origin/main
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o /configuration-service \
    ./cmd/server

<<<<<<< HEAD
# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
=======
# Development stage - includes dev tools and supports hot reload
FROM golang:1.21-alpine AS development

# Install development dependencies
RUN apk add --no-cache git ca-certificates curl wget

# Install air for hot reload
RUN go install github.com/cosmtrek/air@latest

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Expose ports
EXPOSE 9096 9097

# Health check for development
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9096/health || exit 1

# Note: Source code and migrations should be mounted as volume in docker-compose
# Development command with hot reload
CMD ["air", "-c", ".air.toml"]

# Production stage - minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget
>>>>>>> refs/remotes/origin/main

WORKDIR /app

# Copy binary from builder
COPY --from=builder /configuration-service .

# Copy migrations
COPY migrations ./migrations

# Copy example config
COPY config.example.yaml ./config.yaml

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

# Expose ports
EXPOSE 9096 9097

<<<<<<< HEAD
# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:9096/health || exit 1
=======
# Health check for production
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9096/health || exit 1
>>>>>>> refs/remotes/origin/main

# Run
ENTRYPOINT ["./configuration-service"]
