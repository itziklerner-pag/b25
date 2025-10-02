# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o search-service ./cmd/server

# Final stage
FROM alpine:3.19

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 search && \
    adduser -D -u 1000 -G search search

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/search-service .
COPY --from=builder /app/config.example.yaml ./config.yaml

# Change ownership
RUN chown -R search:search /app

# Switch to non-root user
USER search

# Expose ports
EXPOSE 9097 9098

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9097/health || exit 1

# Run the application
ENTRYPOINT ["./search-service"]
CMD ["-config", "config.yaml"]
