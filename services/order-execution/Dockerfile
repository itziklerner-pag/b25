# Multi-stage build for Go Order Execution Service

# Builder stage - compiles the application
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git protobuf-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags '-extldflags "-static" -s -w' \
    -o order-execution ./cmd/server

# Development stage - includes dev tools and supports hot reload
FROM golang:1.21-alpine AS development

# Install development dependencies
RUN apk add --no-cache git protobuf-dev make curl wget

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Install air for hot reload
RUN go install github.com/cosmtrek/air@latest

# Expose ports
EXPOSE 50051 9091

# Health check for development
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9091/health/live || exit 1

# Note: Source code should be mounted as volume in docker-compose
# Development command with hot reload using air
CMD ["air", "-c", ".air.toml"]

# Production stage - minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/order-execution .
COPY --from=builder /build/config.example.yaml ./config.example.yaml

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

# Expose ports
EXPOSE 50051 9091

# Health check for production
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9091/health/live || exit 1

# Run the service
ENTRYPOINT ["/app/order-execution"]
