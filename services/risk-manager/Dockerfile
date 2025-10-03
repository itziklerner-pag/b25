# Multi-stage build for Go Risk Manager Service

# Builder stage - compiles the application
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git make protobuf-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/risk-manager ./cmd/server

# Development stage - includes dev tools and supports hot reload
FROM golang:1.22-alpine AS development

# Install development dependencies
RUN apk add --no-cache git make protobuf-dev curl wget

# Install air for hot reload
RUN go install github.com/cosmtrek/air@latest

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Expose ports
EXPOSE 50051 9095

# Health check for development
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9095/health || exit 1

# Note: Source code and migrations should be mounted as volume in docker-compose
# Development command with hot reload
CMD ["air", "-c", ".air.toml"]

# Production stage - minimal runtime image
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user
RUN addgroup -g 1000 riskmanager && \
    adduser -D -u 1000 -G riskmanager riskmanager

# Copy binary from builder
COPY --from=builder /app/risk-manager /app/risk-manager

# Copy migrations
COPY --from=builder /app/migrations /app/migrations

# Change ownership
RUN chown -R riskmanager:riskmanager /app

# Switch to non-root user
USER riskmanager

# Expose ports
EXPOSE 50051 9095

# Health check for production
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9095/health || exit 1

# Run the application
ENTRYPOINT ["/app/risk-manager"]
