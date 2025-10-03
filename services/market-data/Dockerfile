# Multi-stage build for minimal image size
FROM rust:1.75-slim as builder

# Install dependencies
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

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
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

# Run the service
CMD ["./market-data-service"]
