# Multi-stage build for Analytics Service
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /analytics-server \
    ./cmd/server

# Final stage
FROM alpine:3.18

# Install ca-certificates for HTTPS
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 analytics && \
    adduser -D -u 1000 -G analytics analytics

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /analytics-server .

# Copy configuration template
COPY config.example.yaml ./config.example.yaml

# Change ownership
RUN chown -R analytics:analytics /app

# Switch to non-root user
USER analytics

# Expose ports
EXPOSE 9097 9098 9099

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9097/health || exit 1

# Run the application
ENTRYPOINT ["./analytics-server"]
CMD ["-config", "config.yaml"]
