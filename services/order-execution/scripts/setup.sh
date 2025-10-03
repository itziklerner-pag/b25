#!/bin/bash
set -e

echo "Setting up Order Execution Service..."

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed"
    echo "Install protoc: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

# Install Go protobuf plugins
echo "Installing protobuf Go plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Download dependencies
echo "Downloading Go dependencies..."
go mod download
go mod tidy

# Generate protobuf code
echo "Generating protobuf code..."
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/order.proto

# Create directories
echo "Creating directories..."
mkdir -p bin
mkdir -p logs

# Copy example config if config doesn't exist
if [ ! -f config.yaml ]; then
    echo "Creating config.yaml from example..."
    cp config.example.yaml config.yaml
    echo "NOTE: Edit config.yaml with your Binance API credentials"
fi

echo "Setup complete!"
echo ""
echo "Next steps:"
echo "1. Edit config.yaml with your API credentials"
echo "2. Run 'make build' to build the service"
echo "3. Run 'make run' to start the service"
