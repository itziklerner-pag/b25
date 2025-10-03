.PHONY: all proto go rust clean test

all: proto

# Generate protobuf code for all languages
proto: proto-go proto-rust

# Generate Go protobuf code
proto-go:
	@echo "Generating Go protobuf code..."
	@mkdir -p proto/gen/go
	@protoc \
		--go_out=proto/gen/go \
		--go_opt=paths=source_relative \
		--go-grpc_out=proto/gen/go \
		--go-grpc_opt=paths=source_relative \
		-I proto \
		proto/*.proto
	@echo "Go protobuf code generated successfully"

# Generate Rust protobuf code
proto-rust:
	@echo "Generating Rust protobuf code..."
	@mkdir -p proto/gen/rust
	@cd proto && buf generate
	@echo "Rust protobuf code generated successfully"

# Alternative: use buf for all generation
proto-buf:
	@echo "Generating protobuf code with buf..."
	@cd proto && buf generate
	@echo "Protobuf code generated successfully"

# Build Go libraries
go-build:
	@echo "Building Go libraries..."
	@cd lib/go && go build ./...
	@echo "Go libraries built successfully"

# Build Rust libraries
rust-build:
	@echo "Building Rust libraries..."
	@cd lib/rust/common && cargo build
	@echo "Rust libraries built successfully"

# Test Go libraries
go-test:
	@echo "Testing Go libraries..."
	@cd lib/go && go test -v ./...

# Test Rust libraries
rust-test:
	@echo "Testing Rust libraries..."
	@cd lib/rust/common && cargo test

# Run all tests
test: go-test rust-test

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	@rm -rf proto/gen/go proto/gen/rust
	@cd lib/rust/common && cargo clean
	@echo "Clean complete"

# Format Go code
go-fmt:
	@cd lib/go && go fmt ./...

# Format Rust code
rust-fmt:
	@cd lib/rust/common && cargo fmt

# Format all code
fmt: go-fmt rust-fmt

# Lint Go code
go-lint:
	@cd lib/go && golangci-lint run

# Lint Rust code
rust-lint:
	@cd lib/rust/common && cargo clippy

# Lint all code
lint: go-lint rust-lint

# Install protobuf tools
install-tools:
	@echo "Installing protobuf tools..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@curl -sSL "https://github.com/bufbuild/buf/releases/download/v1.30.0/buf-$(shell uname -s)-$(shell uname -m)" \
		-o /usr/local/bin/buf
	@chmod +x /usr/local/bin/buf
	@echo "Tools installed successfully"

# Initialize Go module
go-init:
	@cd lib/go && go mod download
	@cd lib/go && go mod tidy

# Update dependencies
update-deps:
	@cd lib/go && go get -u ./...
	@cd lib/go && go mod tidy
	@cd lib/rust/common && cargo update

help:
	@echo "Available targets:"
	@echo "  all          - Generate all protobuf code (default)"
	@echo "  proto        - Generate protobuf code for all languages"
	@echo "  proto-go     - Generate Go protobuf code"
	@echo "  proto-rust   - Generate Rust protobuf code"
	@echo "  proto-buf    - Generate protobuf code using buf"
	@echo "  go-build     - Build Go libraries"
	@echo "  rust-build   - Build Rust libraries"
	@echo "  test         - Run all tests"
	@echo "  go-test      - Test Go libraries"
	@echo "  rust-test    - Test Rust libraries"
	@echo "  fmt          - Format all code"
	@echo "  lint         - Lint all code"
	@echo "  clean        - Clean generated files"
	@echo "  install-tools - Install required protobuf tools"
	@echo "  help         - Show this help message"
