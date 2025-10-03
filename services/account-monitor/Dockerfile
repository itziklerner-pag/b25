<<<<<<< HEAD
# Multi-stage build for Account Monitor Service
=======
# Multi-stage build for Go Account Monitor Service

# Builder stage - compiles the application
>>>>>>> refs/remotes/origin/main
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o account-monitor ./cmd/server

<<<<<<< HEAD
# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/account-monitor .

# Copy config
COPY config.example.yaml ./config.yaml

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
=======
# Development stage - includes dev tools and supports hot reload
FROM golang:1.21-alpine AS development

# Install development dependencies
RUN apk add --no-cache git gcc musl-dev curl wget

# Install air for hot reload
RUN go install github.com/cosmtrek/air@latest

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download
>>>>>>> refs/remotes/origin/main

# Expose ports
# 50051: gRPC
# 8080: HTTP/WebSocket/Dashboard
# 9093: Metrics
EXPOSE 50051 8080 9093

<<<<<<< HEAD
=======
# Health check for development
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Note: Source code should be mounted as volume in docker-compose
# Development command with hot reload
CMD ["air", "-c", ".air.toml"]

# Production stage - minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/account-monitor .

# Copy config
COPY config.example.yaml ./config.yaml

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

# Expose ports
# 50051: gRPC
# 8080: HTTP/WebSocket/Dashboard
# 9093: Metrics
EXPOSE 50051 8080 9093

# Health check for production
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

>>>>>>> refs/remotes/origin/main
CMD ["./account-monitor"]
