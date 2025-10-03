#!/bin/bash

# Script to generate protobuf code for Go and Rust

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SHARED_DIR="$(dirname "$SCRIPT_DIR")"
PROTO_DIR="$SHARED_DIR/proto"
GEN_DIR="$PROTO_DIR/gen"

echo "=== B25 Protobuf Code Generation ==="
echo "Proto directory: $PROTO_DIR"
echo "Output directory: $GEN_DIR"

# Check for protoc
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc not found. Please install Protocol Buffers compiler."
    echo "See: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

# Create output directories
mkdir -p "$GEN_DIR/go"
mkdir -p "$GEN_DIR/rust"

echo ""
echo "=== Generating Go Code ==="
protoc \
    --go_out="$GEN_DIR/go" \
    --go_opt=paths=source_relative \
    --go-grpc_out="$GEN_DIR/go" \
    --go-grpc_opt=paths=source_relative \
    -I "$PROTO_DIR" \
    "$PROTO_DIR"/*.proto

if [ $? -eq 0 ]; then
    echo "✓ Go code generated successfully"
else
    echo "✗ Failed to generate Go code"
    exit 1
fi

echo ""
echo "=== Generating Rust Code ==="

# Check if buf is available
if command -v buf &> /dev/null; then
    echo "Using buf for Rust generation..."
    cd "$PROTO_DIR"
    buf generate
    if [ $? -eq 0 ]; then
        echo "✓ Rust code generated successfully with buf"
    else
        echo "✗ Failed to generate Rust code with buf"
        exit 1
    fi
else
    echo "Warning: buf not found. Skipping Rust generation."
    echo "Install buf from: https://docs.buf.build/installation"
    echo "Or use: make install-tools"
fi

echo ""
echo "=== Generation Complete ==="
echo ""
echo "Generated files:"
echo "  Go:   $GEN_DIR/go/"
echo "  Rust: $GEN_DIR/rust/"
echo ""
echo "To use in your services:"
echo "  Go:   import \"github.com/b25/shared/proto/gen/go/common\""
echo "  Rust: Add dependency to b25-proto in Cargo.toml"
