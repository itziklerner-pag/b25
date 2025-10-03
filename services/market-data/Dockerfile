<<<<<<< HEAD
# Multi-stage build for minimal image size
FROM rust:1.75-slim as builder

# Install dependencies
=======
# Multi-stage build for Rust Market Data Service

# Builder stage - compiles the application
FROM rust:1.75-slim AS builder

# Install build dependencies
>>>>>>> refs/remotes/origin/main
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build

# Copy manifests
COPY Cargo.toml ./

# Create dummy main to cache dependencies
RUN mkdir src && \
    echo "fn main() {}" > src/main.rs && \
    cargo build --release && \
    rm -rf src

# Copy source code
COPY src ./src

# Build for release (with full optimizations)
RUN touch src/main.rs && \
    cargo build --release

<<<<<<< HEAD
# Runtime stage
=======
# Development stage - includes dev tools and supports hot reload
FROM rust:1.75-slim AS development

# Install development dependencies
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    curl \
    git \
    && cargo install cargo-watch \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy manifests for dependency installation
COPY Cargo.toml ./

# Pre-download dependencies
RUN mkdir src && \
    echo "fn main() {}" > src/main.rs && \
    cargo build && \
    rm -rf src

# Expose ports
EXPOSE 9090

# Health check for development
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:9090/health || exit 1

# Note: Source code should be mounted as volume in docker-compose
# Development command with hot reload
CMD ["cargo", "watch", "-x", "run"]

# Production stage - minimal runtime image
>>>>>>> refs/remotes/origin/main
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
<<<<<<< HEAD
=======
    curl \
>>>>>>> refs/remotes/origin/main
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -m -u 1000 marketdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/target/release/market-data-service /app/

# Copy example config
COPY config.example.yaml /app/

# Change ownership
RUN chown -R marketdata:marketdata /app

USER marketdata

# Expose health check port
EXPOSE 9090

<<<<<<< HEAD
=======
# Health check for production
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:9090/health || exit 1

>>>>>>> refs/remotes/origin/main
# Run the service
CMD ["./market-data-service"]
