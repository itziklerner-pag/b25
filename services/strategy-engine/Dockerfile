<<<<<<< HEAD
# Build stage
=======
# Multi-stage build for Go Strategy Engine Service

# Builder stage - compiles the application and plugins
>>>>>>> refs/remotes/origin/main
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o strategy-engine ./cmd/server

# Build Go plugins (if any)
RUN if [ -d "plugins/go" ] && [ -n "$(ls -A plugins/go/*.go 2>/dev/null)" ]; then \
    cd plugins/go && \
    for plugin in *.go; do \
        if [ "$plugin" != "*.go" ]; then \
            name="${plugin%.go}"; \
            echo "Building plugin: $name"; \
            go build -buildmode=plugin -o "${name}.so" "$plugin" || echo "Failed to build $plugin"; \
        fi; \
    done; \
fi

<<<<<<< HEAD
# Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata python3 py3-pip
=======
# Development stage - includes dev tools and supports hot reload
FROM golang:1.21-alpine AS development

# Install development dependencies
RUN apk add --no-cache git make gcc musl-dev python3 py3-pip curl wget

# Install air for hot reload
RUN go install github.com/cosmtrek/air@latest

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Expose ports
EXPOSE 9092

# Health check for development
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9092/health || exit 1

# Note: Source code and plugins should be mounted as volume in docker-compose
# Development command with hot reload
CMD ["air", "-c", ".air.toml"]

# Production stage - minimal runtime image
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata python3 py3-pip wget
>>>>>>> refs/remotes/origin/main

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/strategy-engine .

# Copy configuration
COPY config.example.yaml ./config.yaml

# Copy plugins
COPY --from=builder /app/plugins ./plugins

# Install Python dependencies for Python strategies
RUN if [ -f "plugins/python/requirements.txt" ]; then \
    pip3 install --no-cache-dir -r plugins/python/requirements.txt; \
fi

# Create non-root user
RUN addgroup -g 1000 strategy && \
    adduser -D -u 1000 -G strategy strategy && \
    chown -R strategy:strategy /app

USER strategy

# Expose ports
EXPOSE 9092

<<<<<<< HEAD
# Health check
=======
# Health check for production
>>>>>>> refs/remotes/origin/main
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9092/health || exit 1

# Run the application
ENTRYPOINT ["/app/strategy-engine"]
