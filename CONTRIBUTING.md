# Contributing to B25 Trading System

Thank you for contributing! This guide will help you get started.

## 📋 Table of Contents

- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Style](#code-style)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Service-Specific Guidelines](#service-specific-guidelines)

## 🚀 Getting Started

### Prerequisites

- **Docker & Docker Compose** - For running infrastructure
- **Git** - Version control
- **Language-specific tools**:
  - Rust 1.75+ (for market-data, terminal-ui)
  - Go 1.21+ (for most backend services)
  - Node.js 20+ (for web dashboard)
  - Python 3.11+ (for strategy plugins)

### Setup

```bash
# Clone repository
git clone <repo-url>
cd b25

# Start infrastructure
docker-compose -f docker/docker-compose.dev.yml up -d redis postgres timescaledb nats

# Install dependencies per service (see service READMEs)
```

## 🔄 Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/bug-description
```

### 2. Make Changes

- Each service can be developed independently
- Follow service-specific development guides in `services/*/README.md`
- Update documentation as needed

### 3. Test Your Changes

```bash
# Test specific service
cd services/order-execution
go test ./...

# Test all services
./scripts/test-all.sh

# Run integration tests
./scripts/test-integration.sh
```

### 4. Commit Your Changes

```bash
git add .
git commit -m "feat(order-execution): add rate limiting"
```

**Commit Message Format:**
```
<type>(<scope>): <subject>

<body (optional)>

<footer (optional)>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding tests
- `chore`: Maintenance tasks

**Scopes:**
- Service names: `market-data`, `order-execution`, etc.
- `shared`: Shared code changes
- `ci`: CI/CD changes
- `docker`: Docker configuration changes

### 5. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

## 💅 Code Style

### Rust

- Use `rustfmt`: `cargo fmt`
- Use `clippy`: `cargo clippy -- -D warnings`
- Follow [Rust API Guidelines](https://rust-lang.github.io/api-guidelines/)

### Go

- Use `gofmt`: `gofmt -w .`
- Use `golangci-lint`: `golangci-lint run`
- Follow [Effective Go](https://go.dev/doc/effective_go)

### TypeScript/JavaScript

- Use ESLint: `npm run lint`
- Use Prettier: `npm run format`
- Follow project's ESLint config

### Python

- Use `black`: `black .`
- Use `mypy`: `mypy .`
- Follow PEP 8

## 🧪 Testing

### Unit Tests

Write unit tests for all new functionality:

```bash
# Rust
cargo test

# Go
go test ./...

# TypeScript
npm test

# Python
pytest
```

### Integration Tests

Test interactions between services:

```bash
./scripts/test-integration.sh
```

### Performance Tests

For latency-critical services (market-data, order-execution):

```bash
# Rust
cargo bench

# Go
go test -bench=.
```

### Test Coverage

Aim for 80%+ test coverage:

```bash
# Rust
cargo tarpaulin

# Go
go test -cover ./...

# TypeScript
npm run test:coverage
```

## 📤 Submitting Changes

### Pull Request Process

1. **Update Documentation**
   - Update service README if needed
   - Update API documentation
   - Add comments for complex code

2. **Ensure CI Passes**
   - All tests pass
   - Code is formatted
   - No linting errors

3. **Write Clear PR Description**
   ```markdown
   ## Description
   Brief description of changes

   ## Changes
   - Change 1
   - Change 2

   ## Testing
   How to test these changes

   ## Related Issues
   Closes #123
   ```

4. **Request Review**
   - Tag relevant reviewers
   - Address feedback promptly

5. **Merge**
   - Use "Squash and Merge" for feature branches
   - Use "Rebase and Merge" for small fixes

## 🔧 Service-Specific Guidelines

### Market Data Service (Rust)

- **Latency Critical**: Target <100μs processing time
- **Benchmarks Required**: Add benchmarks for hot paths
- **Zero-Copy**: Use zero-copy serialization where possible

### Order Execution Service (Go)

- **Idempotency**: All operations must be idempotent
- **Error Handling**: Comprehensive error handling for exchange API
- **Rate Limiting**: Respect exchange rate limits

### Strategy Engine (Go + Python)

- **Plugin Interface**: Don't break plugin API
- **Hot-Reload**: Test hot-reload functionality
- **Example Strategies**: Update examples if interface changes

### Web Dashboard (React/TypeScript)

- **Mobile-First**: Ensure mobile responsiveness
- **Accessibility**: Follow WCAG guidelines
- **Performance**: Bundle size <500KB

## 🔍 Code Review Guidelines

### For Reviewers

- Focus on logic, not style (CI handles that)
- Check for edge cases
- Verify tests cover changes
- Consider performance implications
- Ensure documentation is updated

### For Contributors

- Respond to feedback constructively
- Make requested changes promptly
- Ask questions if unclear
- Be patient during review process

## 🐛 Reporting Bugs

Use GitHub Issues with this template:

```markdown
## Bug Description
Clear description of the bug

## Steps to Reproduce
1. Step 1
2. Step 2
3. ...

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS: Ubuntu 22.04
- Service: order-execution
- Version: v1.2.3

## Logs
Relevant log output
```

## 💡 Suggesting Features

Use GitHub Issues with this template:

```markdown
## Feature Description
Clear description of the feature

## Use Case
Why this feature is needed

## Proposed Solution
How this could be implemented

## Alternatives
Other approaches considered
```

## 📚 Additional Resources

- [Architecture Documentation](docs/SYSTEM_ARCHITECTURE.md)
- [Component Specifications](docs/COMPONENT_SPECIFICATIONS.md)
- [Implementation Guide](docs/IMPLEMENTATION_GUIDE.md)
- [Service Development Plans](docs/service-plans/)

## ❓ Questions?

- Open a GitHub Discussion
- Check existing documentation
- Ask in PR comments

---

**Thank you for contributing to B25!** 🚀
