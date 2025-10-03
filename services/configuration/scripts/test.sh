#!/bin/bash

# Test script for Configuration Service

set -e

echo "=== Configuration Service Test Suite ==="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if required tools are installed
check_requirements() {
    echo "Checking requirements..."

    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: Go is not installed${NC}"
        exit 1
    fi

    if ! command -v docker &> /dev/null; then
        echo -e "${YELLOW}Warning: Docker is not installed (optional)${NC}"
    fi

    echo -e "${GREEN}✓ Requirements check passed${NC}"
    echo ""
}

# Run unit tests
run_unit_tests() {
    echo "Running unit tests..."
    go test -v -race -coverprofile=coverage.out ./...

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Unit tests passed${NC}"
    else
        echo -e "${RED}✗ Unit tests failed${NC}"
        exit 1
    fi
    echo ""
}

# Generate coverage report
generate_coverage() {
    echo "Generating coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    echo -e "${GREEN}✓ Coverage report generated: coverage.html${NC}"
    echo ""

    # Show coverage summary
    go tool cover -func=coverage.out | tail -1
    echo ""
}

# Run linting
run_linting() {
    echo "Running linting..."

    if command -v golangci-lint &> /dev/null; then
        golangci-lint run
        echo -e "${GREEN}✓ Linting passed${NC}"
    else
        echo -e "${YELLOW}⚠ golangci-lint not installed, skipping${NC}"
    fi
    echo ""
}

# Build the service
build_service() {
    echo "Building service..."
    go build -o bin/configuration-service ./cmd/server

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Build successful${NC}"
    else
        echo -e "${RED}✗ Build failed${NC}"
        exit 1
    fi
    echo ""
}

# Check code formatting
check_formatting() {
    echo "Checking code formatting..."

    unformatted=$(gofmt -l .)
    if [ -n "$unformatted" ]; then
        echo -e "${RED}The following files need formatting:${NC}"
        echo "$unformatted"
        echo ""
        echo "Run 'make fmt' to format the code"
        exit 1
    else
        echo -e "${GREEN}✓ Code formatting is correct${NC}"
    fi
    echo ""
}

# Run all checks
run_all() {
    check_requirements
    check_formatting
    run_unit_tests
    generate_coverage
    run_linting
    build_service

    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  All tests passed! ✓${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# Parse command line arguments
case "${1:-all}" in
    requirements)
        check_requirements
        ;;
    unit)
        run_unit_tests
        ;;
    coverage)
        run_unit_tests
        generate_coverage
        ;;
    lint)
        run_linting
        ;;
    build)
        build_service
        ;;
    format)
        check_formatting
        ;;
    all)
        run_all
        ;;
    *)
        echo "Usage: $0 {requirements|unit|coverage|lint|build|format|all}"
        exit 1
        ;;
esac
