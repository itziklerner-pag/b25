# Configuration Service - Development Plan

**Service Name:** Configuration Service
**Purpose:** Centralized configuration management with versioning, validation, and hot-reload capabilities
**Last Updated:** 2025-10-02
**Version:** 1.0

---

## Table of Contents

1. [Technology Stack Recommendation](#1-technology-stack-recommendation)
2. [Architecture Design](#2-architecture-design)
3. [Development Phases](#3-development-phases)
4. [Implementation Details](#4-implementation-details)
5. [Testing Strategy](#5-testing-strategy)
6. [Deployment](#6-deployment)
7. [Observability](#7-observability)

---

## 1. Technology Stack Recommendation

### Primary Technology Stack (Recommended)

**Language:** Go (Golang)
- **Rationale:** Fast compilation, excellent concurrency, strong typing, built-in testing, single binary deployment
- **Alternatives:** Python (FastAPI), Node.js (NestJS), Rust (Actix-web)

**Database:** PostgreSQL 15+
- **Rationale:** ACID compliance, JSONB support for flexible configs, excellent versioning support, mature ecosystem
- **Schema Features:** Temporal tables, triggers, row-level security
- **Alternatives:** MySQL 8+, SQLite (for development)

**API Framework:**
- **gRPC:** For high-performance service-to-service communication
  - Library: `google.golang.org/grpc`
  - Code generation: `protoc` with Go plugins
- **REST:** For admin UI and external integrations
  - Library: `gin-gonic/gin` or `labstack/echo`
  - OpenAPI/Swagger documentation

**Configuration Formats:**
- **Primary:** YAML (human-readable, comment support)
- **Secondary:** JSON (machine-friendly, strict validation)
- **Advanced:** TOML (for complex nested configs)
- **Library:** `gopkg.in/yaml.v3`, `encoding/json`

**Validation Library:**
- **go-playground/validator/v10** - struct tag-based validation
- **xeipuuv/gojsonschema** - JSON Schema validation
- **Custom validators** for domain-specific rules

**Testing Frameworks:**
- **Unit Testing:** `testing` (standard library)
- **Mocking:** `golang/mock` or `stretchr/testify/mock`
- **Integration:** `testcontainers-go` for PostgreSQL
- **API Testing:** `httptest` (standard library)
- **Load Testing:** `k6` or `vegeta`

**Additional Libraries:**
- **Database ORM:** `sqlc` (type-safe SQL) or `gorm` (full-featured ORM)
- **Migration:** `golang-migrate/migrate`
- **Pub/Sub:** `go-redis/redis` for Redis, `nats-io/nats.go` for NATS
- **Observability:** `prometheus/client_golang`, `uber-go/zap` (logging)
- **Configuration:** `spf13/viper` for service config

---

## 2. Architecture Design

### 2.1 System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                   Configuration Service                      │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   gRPC API   │  │   REST API   │  │   Web UI     │     │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘     │
│         │                  │                  │              │
│         └──────────────────┴──────────────────┘              │
│                            │                                 │
│                   ┌────────▼────────┐                       │
│                   │  Business Logic  │                       │
│                   │                  │                       │
│                   │  - Validation    │                       │
│                   │  - Versioning    │                       │
│                   │  - Access Ctrl   │                       │
│                   └────────┬────────┘                       │
│                            │                                 │
│         ┌──────────────────┼──────────────────┐             │
│         │                  │                  │             │
│    ┌────▼────┐      ┌─────▼──────┐    ┌─────▼──────┐      │
│    │ Storage │      │ Versioning │    │  Pub/Sub   │      │
│    │ Layer   │      │  Engine    │    │  Notifier  │      │
│    └────┬────┘      └─────┬──────┘    └─────┬──────┘      │
│         │                  │                  │             │
└─────────┼──────────────────┼──────────────────┼─────────────┘
          │                  │                  │
     ┌────▼────┐        ┌────▼────┐        ┌───▼────┐
     │PostgreSQL│        │PostgreSQL│        │ Redis/ │
     │ Configs  │        │ Versions │        │  NATS  │
     └──────────┘        └──────────┘        └────────┘
```

### 2.2 Configuration Schema Design

**Hierarchical Structure:**
```
namespace/
├── service_name/
│   ├── environment (dev/staging/prod)
│   │   ├── version (v1, v2, etc.)
│   │   │   ├── config_key_1
│   │   │   ├── config_key_2
│   │   │   └── ...
```

**Configuration Types:**
1. **Strategy Configurations** - Trading strategy parameters
2. **Risk Limits** - Position limits, drawdown thresholds, leverage limits
3. **Symbol Metadata** - Trading pairs, tick sizes, min/max order sizes
4. **System Settings** - Service endpoints, timeouts, retry policies
5. **Feature Flags** - Enable/disable features dynamically

### 2.3 Versioning System

**Versioning Strategy:**
- **Semantic Versioning:** Major.Minor.Patch (e.g., 1.2.3)
- **Automatic Versioning:** Auto-increment on every change
- **Immutable Versions:** Once published, versions cannot be modified
- **Rollback Support:** Activate any previous version instantly
- **Diff Tracking:** Store diffs between versions for audit

**Version States:**
- `DRAFT` - Under construction, mutable
- `ACTIVE` - Currently in use, immutable
- `ARCHIVED` - Previously active, available for rollback
- `DEPRECATED` - Marked for deletion

### 2.4 Validation Pipeline

**Multi-Layer Validation:**

1. **Syntax Validation** (Level 1)
   - YAML/JSON parsing
   - Schema validation against JSON Schema
   - Data type checking

2. **Semantic Validation** (Level 2)
   - Business rule validation
   - Cross-field dependencies
   - Range and boundary checks

3. **Integration Validation** (Level 3)
   - Compatibility checks with other configs
   - External dependency validation (e.g., symbol exists in exchange)
   - Dry-run simulation

4. **Safety Validation** (Level 4)
   - Risk limit sanity checks
   - Production safety gates
   - Mandatory approval workflow (optional)

### 2.5 Hot-Reload Notification Mechanism

**Event-Driven Architecture:**

```
Config Change Event:
{
  "event_id": "uuid",
  "namespace": "strategy",
  "service": "strategy-engine",
  "environment": "prod",
  "version": "2.5.0",
  "change_type": "UPDATE|CREATE|DELETE|ROLLBACK",
  "timestamp": "2025-10-02T10:30:00Z",
  "changed_by": "admin@example.com",
  "checksum": "sha256:abc123..."
}
```

**Notification Channels:**
1. **Pub/Sub Topic:** `config.changes.{namespace}.{service}`
2. **WebHook:** HTTP POST to registered endpoints
3. **Long-Polling:** REST API for polling clients
4. **WebSocket:** Real-time push for UI clients

**Consumer Behavior:**
- Subscribe to specific namespaces/services
- Receive event with new version number
- Fetch full config via API
- Validate and apply changes
- Send acknowledgment back

### 2.6 Access Control (RBAC)

**Roles:**
- **ADMIN** - Full access to all configs
- **OPERATOR** - Read/write to specific namespaces
- **DEVELOPER** - Read/write to dev/staging, read-only to prod
- **SERVICE** - Read-only access via API keys
- **VIEWER** - Read-only access to all

**Permissions:**
- `config:read` - View configurations
- `config:write` - Create/update configurations
- `config:delete` - Delete configurations
- `config:publish` - Promote draft to active
- `config:rollback` - Rollback to previous version
- `config:admin` - Manage access control

**Implementation:**
- JWT tokens with role claims
- API key authentication for services
- Row-level security in PostgreSQL (optional)

---

## 3. Development Phases

### Phase 1: Database Schema and Migrations (Week 1)

**Deliverables:**
- PostgreSQL database schema
- Migration scripts (up/down)
- Seed data for development
- Connection pooling setup

**Tasks:**
1. Design normalized schema with version history
2. Create migration files using `golang-migrate`
3. Setup development database with Docker Compose
4. Create initial seed data scripts
5. Document schema design decisions

**Acceptance Criteria:**
- Database can be provisioned from scratch
- Migrations are reversible
- Seed data creates working examples
- Schema passes normalization review

### Phase 2: CRUD API (gRPC + REST) (Week 2-3)

**Deliverables:**
- gRPC service definition (.proto files)
- REST API endpoints with OpenAPI spec
- CRUD operations for configs
- Basic error handling

**Tasks:**
1. Define gRPC protobuf schemas
2. Implement gRPC server with CRUD operations
3. Build REST gateway using grpc-gateway or Gin
4. Create OpenAPI/Swagger documentation
5. Implement pagination, filtering, sorting
6. Add comprehensive error responses

**Acceptance Criteria:**
- gRPC service can be called by clients
- REST API documented with Swagger UI
- All CRUD operations functional
- Proper HTTP status codes returned
- API follows RESTful conventions

### Phase 3: Validation System (Week 4)

**Deliverables:**
- JSON Schema definitions for config types
- Validation middleware
- Custom validators for business rules
- Validation error reporting

**Tasks:**
1. Define JSON Schemas for each config type
2. Implement schema validation pipeline
3. Create custom validators for domain rules
4. Build validation error aggregation
5. Add pre-flight validation endpoint
6. Create validation test suite

**Acceptance Criteria:**
- Invalid configs are rejected with clear errors
- Validation happens before persistence
- Custom rules are enforced
- Validation is extensible
- Error messages are actionable

### Phase 4: Versioning and Rollback (Week 5)

**Deliverables:**
- Version management system
- Rollback functionality
- Version diff calculation
- Version history API

**Tasks:**
1. Implement version state machine (draft/active/archived)
2. Create version promotion workflow
3. Build rollback mechanism
4. Implement diff calculation and storage
5. Add version comparison API
6. Create version audit log

**Acceptance Criteria:**
- Versions are immutable once active
- Rollback works instantly
- Version history is preserved
- Diffs are accurate
- Audit trail is complete

### Phase 5: Change Notification Pub/Sub (Week 6)

**Deliverables:**
- Pub/Sub integration (Redis/NATS)
- Event publisher service
- WebHook support
- Subscription management

**Tasks:**
1. Integrate Redis or NATS client
2. Implement event publishing on config changes
3. Create subscription management API
4. Add WebHook registration and delivery
5. Build retry and dead-letter queue
6. Add delivery confirmation tracking

**Acceptance Criteria:**
- Events published on all config changes
- Subscribers receive notifications reliably
- WebHooks are delivered with retries
- Failed deliveries are logged
- Event format is documented

### Phase 6: Configuration UI and Testing (Week 7-8)

**Deliverables:**
- Web-based configuration browser
- Version history viewer
- Configuration editor with validation
- Comprehensive test suite

**Tasks:**
1. Build React/Vue configuration browser
2. Create version history timeline UI
3. Implement diff viewer
4. Add configuration editor with live validation
5. Create search and filter UI
6. Complete integration and load testing
7. Performance optimization
8. Security audit

**Acceptance Criteria:**
- UI is responsive and intuitive
- Real-time validation in editor
- Version comparison is visual
- Search works across all configs
- Load tests pass (1000+ req/s)
- Security vulnerabilities addressed

---

## 4. Implementation Details

### 4.1 Database Schema

```sql
-- ============================================================================
-- Configuration Service Database Schema
-- ============================================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- Main Configuration Table
-- ============================================================================
CREATE TABLE configurations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Hierarchical key structure
    namespace VARCHAR(100) NOT NULL,        -- e.g., 'strategy', 'risk', 'system'
    service_name VARCHAR(100) NOT NULL,     -- e.g., 'strategy-engine', 'order-execution'
    environment VARCHAR(50) NOT NULL,       -- e.g., 'dev', 'staging', 'prod'
    config_key VARCHAR(255) NOT NULL,       -- e.g., 'max_position_size'

    -- Configuration data
    config_value JSONB NOT NULL,            -- Flexible JSON storage
    config_format VARCHAR(20) NOT NULL DEFAULT 'json', -- 'json', 'yaml', 'toml'
    data_type VARCHAR(50),                  -- 'string', 'number', 'boolean', 'object', 'array'

    -- Validation
    json_schema JSONB,                      -- JSON Schema for validation
    validation_rules JSONB,                 -- Custom validation rules

    -- Metadata
    description TEXT,
    tags TEXT[],
    is_sensitive BOOLEAN DEFAULT FALSE,     -- Mark secrets for encryption
    is_required BOOLEAN DEFAULT FALSE,
    default_value JSONB,

    -- Version tracking
    version_id UUID NOT NULL REFERENCES config_versions(id),
    version_number VARCHAR(20) NOT NULL,    -- e.g., '1.2.3'

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),

    -- Constraints
    UNIQUE(namespace, service_name, environment, config_key, version_number)
);

-- Indexes for common queries
CREATE INDEX idx_configurations_namespace ON configurations(namespace);
CREATE INDEX idx_configurations_service ON configurations(service_name);
CREATE INDEX idx_configurations_env ON configurations(environment);
CREATE INDEX idx_configurations_version ON configurations(version_id);
CREATE INDEX idx_configurations_lookup ON configurations(namespace, service_name, environment);
CREATE INDEX idx_configurations_tags ON configurations USING GIN(tags);
CREATE INDEX idx_configurations_value ON configurations USING GIN(config_value);

-- ============================================================================
-- Version Management Table
-- ============================================================================
CREATE TYPE version_state AS ENUM ('DRAFT', 'ACTIVE', 'ARCHIVED', 'DEPRECATED');

CREATE TABLE config_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Version identification
    namespace VARCHAR(100) NOT NULL,
    service_name VARCHAR(100) NOT NULL,
    environment VARCHAR(50) NOT NULL,
    version_number VARCHAR(20) NOT NULL,    -- Semantic version: 1.2.3

    -- Version state
    state version_state NOT NULL DEFAULT 'DRAFT',

    -- Version metadata
    description TEXT,
    change_summary TEXT,
    change_type VARCHAR(50),                -- 'MAJOR', 'MINOR', 'PATCH', 'HOTFIX'

    -- Relationships
    parent_version_id UUID REFERENCES config_versions(id), -- Previous version

    -- Checksums for integrity
    checksum VARCHAR(64),                   -- SHA-256 of entire config set

    -- Activation tracking
    activated_at TIMESTAMP WITH TIME ZONE,
    activated_by VARCHAR(255),
    archived_at TIMESTAMP WITH TIME ZONE,
    archived_by VARCHAR(255),

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255),

    -- Constraints
    UNIQUE(namespace, service_name, environment, version_number)
);

-- Indexes
CREATE INDEX idx_versions_state ON config_versions(state);
CREATE INDEX idx_versions_lookup ON config_versions(namespace, service_name, environment);
CREATE INDEX idx_versions_created ON config_versions(created_at DESC);

-- ============================================================================
-- Version History and Audit Log
-- ============================================================================
CREATE TABLE config_audit_log (
    id BIGSERIAL PRIMARY KEY,

    -- What changed
    config_id UUID REFERENCES configurations(id),
    version_id UUID REFERENCES config_versions(id),

    -- Action details
    action VARCHAR(50) NOT NULL,            -- 'CREATE', 'UPDATE', 'DELETE', 'ROLLBACK', 'PUBLISH'

    -- Change tracking
    old_value JSONB,
    new_value JSONB,
    diff JSONB,                             -- JSON diff

    -- Context
    reason TEXT,
    metadata JSONB,                         -- Additional context

    -- Actor
    actor_id VARCHAR(255),
    actor_type VARCHAR(50),                 -- 'USER', 'SERVICE', 'SYSTEM'
    ip_address INET,
    user_agent TEXT,

    -- Timestamp
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for audit queries
CREATE INDEX idx_audit_config ON config_audit_log(config_id);
CREATE INDEX idx_audit_version ON config_audit_log(version_id);
CREATE INDEX idx_audit_action ON config_audit_log(action);
CREATE INDEX idx_audit_actor ON config_audit_log(actor_id);
CREATE INDEX idx_audit_created ON config_audit_log(created_at DESC);

-- ============================================================================
-- Subscriptions and Notifications
-- ============================================================================
CREATE TABLE config_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Subscriber details
    subscriber_name VARCHAR(255) NOT NULL,
    subscriber_type VARCHAR(50) NOT NULL,   -- 'SERVICE', 'WEBHOOK', 'EMAIL'

    -- Subscription scope
    namespace VARCHAR(100),                 -- NULL = all namespaces
    service_name VARCHAR(100),              -- NULL = all services
    environment VARCHAR(50),                -- NULL = all environments

    -- Delivery configuration
    endpoint_url TEXT,                      -- For webhooks
    delivery_method VARCHAR(50) NOT NULL,   -- 'PUBSUB', 'WEBHOOK', 'POLLING'

    -- Filters
    event_types TEXT[],                     -- ['UPDATE', 'ROLLBACK']

    -- Status
    is_active BOOLEAN DEFAULT TRUE,

    -- Metadata
    metadata JSONB,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255)
);

-- Indexes
CREATE INDEX idx_subscriptions_lookup ON config_subscriptions(namespace, service_name, environment);
CREATE INDEX idx_subscriptions_active ON config_subscriptions(is_active) WHERE is_active = TRUE;

-- ============================================================================
-- Notification Delivery Log
-- ============================================================================
CREATE TABLE notification_log (
    id BIGSERIAL PRIMARY KEY,

    -- Subscription
    subscription_id UUID REFERENCES config_subscriptions(id),

    -- Event details
    event_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    version_id UUID REFERENCES config_versions(id),

    -- Delivery status
    status VARCHAR(50) NOT NULL,            -- 'PENDING', 'DELIVERED', 'FAILED', 'RETRYING'
    attempt_count INT DEFAULT 0,
    last_attempt_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,

    -- Error tracking
    error_message TEXT,
    error_details JSONB,

    -- Payload
    payload JSONB,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_notification_subscription ON notification_log(subscription_id);
CREATE INDEX idx_notification_status ON notification_log(status);
CREATE INDEX idx_notification_event ON notification_log(event_id);
CREATE INDEX idx_notification_created ON notification_log(created_at DESC);

-- ============================================================================
-- Access Control (Optional)
-- ============================================================================
CREATE TABLE access_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Principal (who)
    principal_type VARCHAR(50) NOT NULL,    -- 'USER', 'SERVICE', 'ROLE', 'API_KEY'
    principal_id VARCHAR(255) NOT NULL,

    -- Resource (what)
    namespace VARCHAR(100),                 -- NULL = all
    service_name VARCHAR(100),              -- NULL = all
    environment VARCHAR(50),                -- NULL = all

    -- Permissions (actions)
    permissions TEXT[] NOT NULL,            -- ['config:read', 'config:write', etc.]

    -- Conditions
    conditions JSONB,                       -- Additional constraints

    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP WITH TIME ZONE,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255)
);

-- Indexes
CREATE INDEX idx_access_principal ON access_policies(principal_type, principal_id);
CREATE INDEX idx_access_scope ON access_policies(namespace, service_name, environment);
CREATE INDEX idx_access_active ON access_policies(is_active) WHERE is_active = TRUE;

-- ============================================================================
-- Functions and Triggers
-- ============================================================================

-- Update timestamp trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to tables
CREATE TRIGGER update_configurations_updated_at BEFORE UPDATE ON configurations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_subscriptions_updated_at BEFORE UPDATE ON config_subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Audit log trigger function
CREATE OR REPLACE FUNCTION log_config_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO config_audit_log (config_id, action, new_value, actor_type)
        VALUES (NEW.id, 'CREATE', to_jsonb(NEW), 'SYSTEM');
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO config_audit_log (config_id, action, old_value, new_value, actor_type)
        VALUES (NEW.id, 'UPDATE', to_jsonb(OLD), to_jsonb(NEW), 'SYSTEM');
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO config_audit_log (config_id, action, old_value, actor_type)
        VALUES (OLD.id, 'DELETE', to_jsonb(OLD), 'SYSTEM');
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Apply audit trigger
CREATE TRIGGER audit_configurations AFTER INSERT OR UPDATE OR DELETE ON configurations
    FOR EACH ROW EXECUTE FUNCTION log_config_changes();

-- ============================================================================
-- Views for Common Queries
-- ============================================================================

-- Active configurations view
CREATE VIEW v_active_configurations AS
SELECT
    c.*,
    v.state as version_state,
    v.activated_at,
    v.activated_by
FROM configurations c
JOIN config_versions v ON c.version_id = v.id
WHERE v.state = 'ACTIVE';

-- Latest version view
CREATE VIEW v_latest_versions AS
SELECT DISTINCT ON (namespace, service_name, environment)
    *
FROM config_versions
ORDER BY namespace, service_name, environment, created_at DESC;

-- Configuration summary view
CREATE VIEW v_config_summary AS
SELECT
    namespace,
    service_name,
    environment,
    COUNT(*) as config_count,
    MAX(updated_at) as last_updated,
    array_agg(DISTINCT config_key) as config_keys
FROM configurations
GROUP BY namespace, service_name, environment;
```

### 4.2 Configuration Validation Rules

**JSON Schema Example for Risk Limits:**

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "Risk Limits Configuration",
  "properties": {
    "max_position_size": {
      "type": "number",
      "minimum": 0,
      "maximum": 1000000,
      "description": "Maximum position size in base currency"
    },
    "max_leverage": {
      "type": "number",
      "minimum": 1,
      "maximum": 100,
      "description": "Maximum allowed leverage"
    },
    "max_drawdown_percent": {
      "type": "number",
      "minimum": 0,
      "maximum": 100,
      "description": "Maximum drawdown percentage before emergency stop"
    },
    "daily_loss_limit": {
      "type": "number",
      "minimum": 0,
      "description": "Daily loss limit in USD"
    },
    "symbols": {
      "type": "array",
      "items": {
        "type": "string",
        "pattern": "^[A-Z]+/[A-Z]+$"
      },
      "minItems": 1,
      "description": "Allowed trading symbols"
    }
  },
  "required": ["max_position_size", "max_leverage", "max_drawdown_percent"],
  "additionalProperties": false
}
```

**Custom Validation Rules:**

```go
// Custom validator for risk limits
func ValidateRiskLimits(config map[string]interface{}) error {
    // Business rule: max_leverage must decrease as max_position_size increases
    leverage := config["max_leverage"].(float64)
    positionSize := config["max_position_size"].(float64)

    if positionSize > 100000 && leverage > 10 {
        return errors.New("leverage cannot exceed 10x for positions over 100k")
    }

    // Daily loss limit should be reasonable
    dailyLoss := config["daily_loss_limit"].(float64)
    if dailyLoss > positionSize * 0.5 {
        return errors.New("daily loss limit too high relative to position size")
    }

    return nil
}

// Validator registry
type ValidatorFunc func(map[string]interface{}) error

var validators = map[string]ValidatorFunc{
    "risk_limits": ValidateRiskLimits,
    "strategy_params": ValidateStrategyParams,
    "symbol_metadata": ValidateSymbolMetadata,
}
```

### 4.3 Version Control Logic

**Version Promotion Workflow:**

```go
// Version state machine
type VersionState string

const (
    StateDraft      VersionState = "DRAFT"
    StateActive     VersionState = "ACTIVE"
    StateArchived   VersionState = "ARCHIVED"
    StateDeprecated VersionState = "DEPRECATED"
)

// PromoteVersion activates a draft version
func (s *ConfigService) PromoteVersion(ctx context.Context, versionID uuid.UUID) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 1. Get version and validate state
    version, err := s.getVersion(ctx, tx, versionID)
    if err != nil {
        return err
    }

    if version.State != StateDraft {
        return ErrInvalidStateTransition
    }

    // 2. Archive currently active version
    err = s.archiveActiveVersion(ctx, tx, version.Namespace, version.ServiceName, version.Environment)
    if err != nil {
        return err
    }

    // 3. Activate new version
    _, err = tx.ExecContext(ctx, `
        UPDATE config_versions
        SET state = $1,
            activated_at = NOW(),
            activated_by = $2
        WHERE id = $3
    `, StateActive, getCurrentUser(ctx), versionID)
    if err != nil {
        return err
    }

    // 4. Publish change event
    event := ConfigChangeEvent{
        EventID:     uuid.New(),
        Namespace:   version.Namespace,
        Service:     version.ServiceName,
        Environment: version.Environment,
        Version:     version.VersionNumber,
        ChangeType:  "ACTIVATE",
        Timestamp:   time.Now(),
        ChangedBy:   getCurrentUser(ctx),
    }

    err = s.publishEvent(ctx, event)
    if err != nil {
        return err
    }

    return tx.Commit()
}

// RollbackToVersion reverts to a previous version
func (s *ConfigService) RollbackToVersion(ctx context.Context, versionID uuid.UUID) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 1. Get target version
    targetVersion, err := s.getVersion(ctx, tx, versionID)
    if err != nil {
        return err
    }

    // Only archived versions can be rolled back to
    if targetVersion.State != StateArchived {
        return ErrInvalidStateTransition
    }

    // 2. Archive current active version
    err = s.archiveActiveVersion(ctx, tx, targetVersion.Namespace,
        targetVersion.ServiceName, targetVersion.Environment)
    if err != nil {
        return err
    }

    // 3. Reactivate target version
    _, err = tx.ExecContext(ctx, `
        UPDATE config_versions
        SET state = $1,
            activated_at = NOW(),
            activated_by = $2
        WHERE id = $3
    `, StateActive, getCurrentUser(ctx), versionID)
    if err != nil {
        return err
    }

    // 4. Publish rollback event
    event := ConfigChangeEvent{
        EventID:     uuid.New(),
        Namespace:   targetVersion.Namespace,
        Service:     targetVersion.ServiceName,
        Environment: targetVersion.Environment,
        Version:     targetVersion.VersionNumber,
        ChangeType:  "ROLLBACK",
        Timestamp:   time.Now(),
        ChangedBy:   getCurrentUser(ctx),
    }

    err = s.publishEvent(ctx, event)
    if err != nil {
        return err
    }

    return tx.Commit()
}
```

### 4.4 Change Notification Mechanism

**Event Publisher:**

```go
type ConfigChangeEvent struct {
    EventID     uuid.UUID `json:"event_id"`
    Namespace   string    `json:"namespace"`
    Service     string    `json:"service"`
    Environment string    `json:"environment"`
    Version     string    `json:"version"`
    ChangeType  string    `json:"change_type"` // UPDATE, CREATE, DELETE, ROLLBACK
    Timestamp   time.Time `json:"timestamp"`
    ChangedBy   string    `json:"changed_by"`
    Checksum    string    `json:"checksum"`
}

// Pub/Sub implementation using Redis
type RedisPubSub struct {
    client *redis.Client
}

func (p *RedisPubSub) PublishConfigChange(ctx context.Context, event ConfigChangeEvent) error {
    // Create topic name
    topic := fmt.Sprintf("config.changes.%s.%s", event.Namespace, event.Service)

    // Serialize event
    payload, err := json.Marshal(event)
    if err != nil {
        return err
    }

    // Publish to Redis
    err = p.client.Publish(ctx, topic, payload).Err()
    if err != nil {
        return err
    }

    // Also publish to wildcard topic for global listeners
    globalTopic := "config.changes.*"
    err = p.client.Publish(ctx, globalTopic, payload).Err()

    return err
}

// WebHook delivery with retry
type WebHookDelivery struct {
    httpClient *http.Client
    maxRetries int
}

func (w *WebHookDelivery) DeliverToWebHook(ctx context.Context,
    subscription ConfigSubscription, event ConfigChangeEvent) error {

    payload, err := json.Marshal(event)
    if err != nil {
        return err
    }

    // Retry logic with exponential backoff
    var lastErr error
    for attempt := 0; attempt < w.maxRetries; attempt++ {
        req, err := http.NewRequestWithContext(ctx, "POST",
            subscription.EndpointURL, bytes.NewReader(payload))
        if err != nil {
            return err
        }

        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-Event-ID", event.EventID.String())
        req.Header.Set("X-Event-Type", event.ChangeType)

        resp, err := w.httpClient.Do(req)
        if err != nil {
            lastErr = err
            time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * time.Second)
            continue
        }
        defer resp.Body.Close()

        if resp.StatusCode >= 200 && resp.StatusCode < 300 {
            // Success
            w.logDelivery(subscription.ID, event.EventID, "DELIVERED", attempt+1, nil)
            return nil
        }

        lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
        time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * time.Second)
    }

    // All retries failed
    w.logDelivery(subscription.ID, event.EventID, "FAILED", w.maxRetries, lastErr)
    return lastErr
}
```

### 4.5 API Endpoints Specification

**gRPC Service Definition (config.proto):**

```protobuf
syntax = "proto3";

package config.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/struct.proto";

service ConfigService {
  // Configuration CRUD
  rpc CreateConfig(CreateConfigRequest) returns (ConfigResponse);
  rpc GetConfig(GetConfigRequest) returns (ConfigResponse);
  rpc UpdateConfig(UpdateConfigRequest) returns (ConfigResponse);
  rpc DeleteConfig(DeleteConfigRequest) returns (DeleteConfigResponse);
  rpc ListConfigs(ListConfigsRequest) returns (ListConfigsResponse);

  // Batch operations
  rpc GetConfigSet(GetConfigSetRequest) returns (GetConfigSetResponse);
  rpc UpdateConfigSet(UpdateConfigSetRequest) returns (UpdateConfigSetResponse);

  // Version management
  rpc CreateVersion(CreateVersionRequest) returns (VersionResponse);
  rpc GetVersion(GetVersionRequest) returns (VersionResponse);
  rpc ListVersions(ListVersionsRequest) returns (ListVersionsResponse);
  rpc PromoteVersion(PromoteVersionRequest) returns (VersionResponse);
  rpc RollbackVersion(RollbackVersionRequest) returns (VersionResponse);
  rpc CompareVersions(CompareVersionsRequest) returns (CompareVersionsResponse);

  // Validation
  rpc ValidateConfig(ValidateConfigRequest) returns (ValidateConfigResponse);

  // Subscriptions
  rpc Subscribe(SubscribeRequest) returns (SubscribeResponse);
  rpc Unsubscribe(UnsubscribeRequest) returns (UnsubscribeResponse);
  rpc ListSubscriptions(ListSubscriptionsRequest) returns (ListSubscriptionsResponse);

  // Health check
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}

message Config {
  string id = 1;
  string namespace = 2;
  string service_name = 3;
  string environment = 4;
  string config_key = 5;
  google.protobuf.Struct config_value = 6;
  string config_format = 7;
  string version_id = 8;
  string version_number = 9;
  google.protobuf.Timestamp created_at = 10;
  google.protobuf.Timestamp updated_at = 11;
  string created_by = 12;
  string updated_by = 13;
}

message Version {
  string id = 1;
  string namespace = 2;
  string service_name = 3;
  string environment = 4;
  string version_number = 5;
  string state = 6;
  string description = 7;
  string checksum = 8;
  google.protobuf.Timestamp created_at = 9;
  google.protobuf.Timestamp activated_at = 10;
  string created_by = 11;
}

message CreateConfigRequest {
  string namespace = 1;
  string service_name = 2;
  string environment = 3;
  string config_key = 4;
  google.protobuf.Struct config_value = 5;
  string version_id = 6;
}

message ConfigResponse {
  Config config = 1;
  string message = 2;
}

// ... additional message definitions
```

**REST API Endpoints:**

```
Base URL: /api/v1

Configuration Management:
  POST   /configs                          - Create configuration
  GET    /configs/:id                      - Get configuration by ID
  PUT    /configs/:id                      - Update configuration
  DELETE /configs/:id                      - Delete configuration
  GET    /configs                          - List configurations (with filters)
  GET    /configs/search                   - Search configurations

  GET    /namespaces/:namespace/services/:service/envs/:env/configs
                                            - Get all configs for service+env
  PUT    /namespaces/:namespace/services/:service/envs/:env/configs
                                            - Bulk update configs

Version Management:
  POST   /versions                         - Create new version (draft)
  GET    /versions/:id                     - Get version details
  GET    /versions                         - List versions (with filters)
  POST   /versions/:id/promote             - Promote draft to active
  POST   /versions/:id/rollback            - Rollback to this version
  GET    /versions/:id/configs             - Get all configs in version
  GET    /versions/:id1/compare/:id2       - Compare two versions
  GET    /versions/:id/diff                - Diff with previous version

Validation:
  POST   /validate                         - Validate configuration
  GET    /schemas                          - List validation schemas
  GET    /schemas/:type                    - Get schema for config type

Subscriptions:
  POST   /subscriptions                    - Create subscription
  GET    /subscriptions                    - List subscriptions
  GET    /subscriptions/:id                - Get subscription details
  PUT    /subscriptions/:id                - Update subscription
  DELETE /subscriptions/:id                - Delete subscription
  GET    /subscriptions/:id/deliveries     - Get delivery history

Utilities:
  GET    /health                           - Health check
  GET    /metrics                          - Prometheus metrics
  GET    /namespaces                       - List all namespaces
  GET    /services                         - List all services
  GET    /environments                     - List all environments
```

**Example REST Response:**

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "namespace": "strategy",
    "service_name": "strategy-engine",
    "environment": "prod",
    "config_key": "max_position_size",
    "config_value": {
      "value": 50000,
      "unit": "USD"
    },
    "version_id": "660e8400-e29b-41d4-a716-446655440001",
    "version_number": "2.1.0",
    "created_at": "2025-10-02T10:00:00Z",
    "updated_at": "2025-10-02T10:30:00Z"
  },
  "meta": {
    "request_id": "req-abc123",
    "timestamp": "2025-10-02T10:30:15Z"
  }
}
```

---

## 5. Testing Strategy

### 5.1 Unit Testing

**Coverage Requirements:**
- Minimum 80% code coverage
- 100% coverage for critical paths (validation, versioning)

**Test Categories:**

```go
// Validation tests
func TestValidateRiskLimits(t *testing.T) {
    tests := []struct {
        name      string
        config    map[string]interface{}
        wantError bool
    }{
        {
            name: "valid config",
            config: map[string]interface{}{
                "max_position_size": 100000.0,
                "max_leverage": 5.0,
                "max_drawdown_percent": 20.0,
            },
            wantError: false,
        },
        {
            name: "invalid leverage for large position",
            config: map[string]interface{}{
                "max_position_size": 200000.0,
                "max_leverage": 20.0,
            },
            wantError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateRiskLimits(tt.config)
            if (err != nil) != tt.wantError {
                t.Errorf("ValidateRiskLimits() error = %v, wantError %v",
                    err, tt.wantError)
            }
        })
    }
}

// Version state machine tests
func TestVersionStateTransitions(t *testing.T) {
    // Test DRAFT -> ACTIVE
    // Test ACTIVE -> ARCHIVED
    // Test invalid transitions
}

// Database tests with testcontainers
func TestConfigRepository(t *testing.T) {
    ctx := context.Background()

    // Start PostgreSQL container
    postgres, err := testcontainers.GenericContainer(ctx, /* ... */)
    require.NoError(t, err)
    defer postgres.Terminate(ctx)

    // Run migrations
    // Run tests
}
```

### 5.2 Integration Testing

**Test Scenarios:**

1. **Full Configuration Lifecycle**
   - Create draft version
   - Add configurations
   - Validate
   - Promote to active
   - Verify event published
   - Rollback
   - Verify rollback event

2. **Concurrent Update Handling**
   - Multiple services updating same config
   - Optimistic locking verification
   - Conflict resolution

3. **Event Delivery**
   - Pub/Sub message delivery
   - WebHook delivery with retries
   - Dead letter queue handling

4. **Database Transactions**
   - ACID compliance verification
   - Rollback on validation failure
   - Isolation level testing

### 5.3 Performance Testing

**Load Test Scenarios:**

```javascript
// k6 load test script
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  stages: [
    { duration: '2m', target: 100 },  // Ramp up to 100 users
    { duration: '5m', target: 100 },  // Stay at 100 users
    { duration: '2m', target: 200 },  // Ramp up to 200 users
    { duration: '5m', target: 200 },  // Stay at 200 users
    { duration: '2m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p95<500', 'p99<1000'], // 95% < 500ms, 99% < 1s
    http_req_failed: ['rate<0.01'],             // Error rate < 1%
  },
};

export default function() {
  // Read config
  let res = http.get('http://localhost:8080/api/v1/configs/search?namespace=strategy&service=strategy-engine&env=prod');
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  // Write config (10% of requests)
  if (Math.random() < 0.1) {
    http.post('http://localhost:8080/api/v1/configs', JSON.stringify({
      namespace: 'strategy',
      service_name: 'strategy-engine',
      environment: 'dev',
      config_key: `test_key_${Date.now()}`,
      config_value: { value: Math.random() * 100 },
    }), {
      headers: { 'Content-Type': 'application/json' },
    });
  }
}
```

**Performance Targets:**
- Read operations: p95 < 100ms, p99 < 200ms
- Write operations: p95 < 500ms, p99 < 1000ms
- Throughput: 1000+ req/s sustained
- Event delivery: < 100ms latency

### 5.4 Access Control Tests

```go
func TestAccessControl(t *testing.T) {
    tests := []struct {
        name        string
        user        User
        action      string
        resource    Resource
        shouldAllow bool
    }{
        {
            name: "admin can write prod",
            user: User{Role: "ADMIN"},
            action: "config:write",
            resource: Resource{Environment: "prod"},
            shouldAllow: true,
        },
        {
            name: "developer cannot write prod",
            user: User{Role: "DEVELOPER"},
            action: "config:write",
            resource: Resource{Environment: "prod"},
            shouldAllow: false,
        },
        {
            name: "service can only read",
            user: User{Type: "SERVICE"},
            action: "config:write",
            resource: Resource{},
            shouldAllow: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            allowed := CheckPermission(tt.user, tt.action, tt.resource)
            if allowed != tt.shouldAllow {
                t.Errorf("CheckPermission() = %v, want %v", allowed, tt.shouldAllow)
            }
        })
    }
}
```

---

## 6. Deployment

### 6.1 Dockerfile

```dockerfile
# Multi-stage build for minimal image size
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o /config-service \
    ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /config-service .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run
ENTRYPOINT ["./config-service"]
CMD ["serve"]
```

### 6.2 Docker Compose

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: configdb
      POSTGRES_USER: configuser
      POSTGRES_PASSWORD: configpass
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U configuser -d configdb"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5

  config-service:
    build:
      context: .
      dockerfile: Dockerfile
    environment:
      DATABASE_URL: postgres://configuser:configpass@postgres:5432/configdb?sslmode=disable
      REDIS_URL: redis://redis:6379
      GRPC_PORT: 9090
      HTTP_PORT: 8080
      LOG_LEVEL: info
      ENVIRONMENT: production
    ports:
      - "8080:8080"  # REST API
      - "9090:9090"  # gRPC
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  # Optional: Prometheus for metrics
  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    ports:
      - "9091:9090"
    depends_on:
      - config-service

  # Optional: Grafana for visualization
  grafana:
    image: grafana/grafana:latest
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin
      GF_USERS_ALLOW_SIGN_UP: false
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
    ports:
      - "3000:3000"
    depends_on:
      - prometheus

volumes:
  postgres_data:
  prometheus_data:
  grafana_data:

networks:
  default:
    name: trading-net
```

### 6.3 Database Setup Script

```bash
#!/bin/bash
# setup-db.sh

set -e

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-configdb}"
DB_USER="${DB_USER:-configuser}"
DB_PASSWORD="${DB_PASSWORD:-configpass}"

echo "Waiting for PostgreSQL to be ready..."
until PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c '\q'; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done

echo "PostgreSQL is up - creating database..."

PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres <<-EOSQL
  SELECT 'CREATE DATABASE $DB_NAME'
  WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$DB_NAME')\gexec
EOSQL

echo "Running migrations..."
migrate -path ./migrations \
        -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable" \
        up

echo "Database setup complete!"
```

### 6.4 Migration Strategy

**Zero-Downtime Migration Approach:**

1. **Backward-Compatible Changes First**
   - Add new columns with defaults
   - Add new tables
   - Deploy new service version

2. **Data Migration**
   - Run background job to migrate data
   - Monitor progress

3. **Cleanup**
   - Remove old columns/tables in next release
   - Only after confirming rollback window passed

**Migration File Example:**

```sql
-- migrations/000001_initial_schema.up.sql
-- See schema in section 4.1

-- migrations/000001_initial_schema.down.sql
DROP TABLE IF EXISTS notification_log;
DROP TABLE IF EXISTS config_subscriptions;
DROP TABLE IF EXISTS config_audit_log;
DROP TABLE IF EXISTS access_policies;
DROP TABLE IF EXISTS config_versions;
DROP TABLE IF EXISTS configurations;
DROP TYPE IF EXISTS version_state;
DROP EXTENSION IF EXISTS "uuid-ossp";
```

### 6.5 Kubernetes Deployment (Optional)

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: config-service
  labels:
    app: config-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: config-service
  template:
    metadata:
      labels:
        app: config-service
    spec:
      containers:
      - name: config-service
        image: config-service:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: grpc
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: config-service-secrets
              key: database-url
        - name: REDIS_URL
          valueFrom:
            configMapKeyRef:
              name: config-service-config
              key: redis-url
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: config-service
spec:
  selector:
    app: config-service
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: grpc
    port: 9090
    targetPort: 9090
  type: ClusterIP
```

---

## 7. Observability

### 7.1 Configuration Browser UI

**Technology Stack:**
- Frontend: React 18 + TypeScript
- UI Framework: Material-UI or Ant Design
- State Management: React Query + Zustand
- Code Editor: Monaco Editor (VSCode editor)

**Key Features:**

1. **Configuration Explorer**
   - Tree view: Namespace > Service > Environment
   - Search and filter
   - Bulk operations
   - Export/import configs

2. **Configuration Editor**
   - Syntax highlighting (YAML/JSON)
   - Live validation
   - Schema-aware autocomplete
   - Diff viewer for comparing versions

3. **Version Management**
   - Version timeline
   - Visual diff between versions
   - One-click rollback
   - Promote draft to active

4. **Real-time Updates**
   - WebSocket connection for live changes
   - Notifications for config updates
   - Conflict detection

**UI Mockup Structure:**

```
┌─────────────────────────────────────────────────────────────┐
│  Configuration Service - Dashboard                 [User]   │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐  ┌────────────────────────────────────┐  │
│  │ Namespaces   │  │  Active Configuration              │  │
│  │              │  │                                     │  │
│  │ ▼ strategy   │  │  Service: strategy-engine          │  │
│  │   ▶ risk     │  │  Environment: production           │  │
│  │   ▶ account  │  │  Version: 2.1.0 (ACTIVE)          │  │
│  │              │  │                                     │  │
│  │ ▼ risk       │  │  ┌──────────────────────────────┐  │  │
│  │   ▶ limits   │  │  │ max_position_size: 50000    │  │  │
│  │   ▶ policies │  │  │ max_leverage: 5.0           │  │  │
│  │              │  │  │ max_drawdown_percent: 20.0  │  │  │
│  │ ▼ system     │  │  │ symbols: [BTC/USD, ...]     │  │  │
│  │   ▶ api      │  │  └──────────────────────────────┘  │  │
│  │   ▶ database │  │                                     │  │
│  │              │  │  [Edit] [Rollback] [History]       │  │
│  └──────────────┘  └────────────────────────────────────┘  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 7.2 Version History Viewer

**Features:**
- Timeline visualization
- Commit-style history
- Author and timestamp tracking
- Change summary
- Side-by-side diff view
- Cherry-pick specific changes

**Component Structure:**

```typescript
interface VersionHistoryProps {
  namespace: string;
  service: string;
  environment: string;
}

const VersionHistory: React.FC<VersionHistoryProps> = ({
  namespace, service, environment
}) => {
  const { data: versions } = useQuery(['versions', namespace, service, environment],
    () => fetchVersions(namespace, service, environment)
  );

  return (
    <Timeline>
      {versions?.map(version => (
        <TimelineItem key={version.id}>
          <TimelineLabel>
            {version.version_number} - {version.state}
            <br />
            {formatDate(version.created_at)} by {version.created_by}
          </TimelineLabel>
          <TimelineContent>
            <Card>
              <CardHeader>
                {version.description}
              </CardHeader>
              <CardActions>
                <Button onClick={() => viewDiff(version.id)}>View Changes</Button>
                {version.state === 'ARCHIVED' && (
                  <Button onClick={() => rollback(version.id)}>Rollback</Button>
                )}
              </CardActions>
            </Card>
          </TimelineContent>
        </TimelineItem>
      ))}
    </Timeline>
  );
};
```

### 7.3 Metrics and Monitoring

**Prometheus Metrics:**

```go
var (
    // Request metrics
    requestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "config_service_request_duration_seconds",
            Help: "Duration of HTTP requests",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint", "status"},
    )

    // Configuration metrics
    configCount = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "config_service_configurations_total",
            Help: "Total number of configurations",
        },
        []string{"namespace", "service", "environment"},
    )

    versionCount = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "config_service_version_changes_total",
            Help: "Total number of version changes",
        },
        []string{"namespace", "service", "environment", "change_type"},
    )

    // Event delivery metrics
    eventDeliveryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "config_service_event_delivery_duration_seconds",
            Help: "Duration of event delivery",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
        },
        []string{"delivery_method", "status"},
    )

    // Validation metrics
    validationErrors = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "config_service_validation_errors_total",
            Help: "Total number of validation errors",
        },
        []string{"namespace", "service", "error_type"},
    )
)
```

**Grafana Dashboard Panels:**

1. **Overview Panel**
   - Total configurations by namespace
   - Active versions count
   - Recent changes timeline
   - System health status

2. **Performance Panel**
   - Request rate (req/s)
   - Request duration (p50, p95, p99)
   - Error rate
   - Database query performance

3. **Events Panel**
   - Event delivery success rate
   - Event delivery latency
   - Failed deliveries
   - Retry queue depth

4. **Business Metrics Panel**
   - Configurations by environment
   - Most frequently updated configs
   - Rollback frequency
   - User activity

### 7.4 Logging Strategy

**Structured Logging with Zap:**

```go
import "go.uber.org/zap"

type Logger struct {
    *zap.Logger
}

func (l *Logger) LogConfigChange(ctx context.Context, event ConfigChangeEvent) {
    l.Info("configuration changed",
        zap.String("event_id", event.EventID.String()),
        zap.String("namespace", event.Namespace),
        zap.String("service", event.Service),
        zap.String("environment", event.Environment),
        zap.String("version", event.Version),
        zap.String("change_type", event.ChangeType),
        zap.String("changed_by", event.ChangedBy),
        zap.String("trace_id", getTraceID(ctx)),
    )
}

func (l *Logger) LogValidationError(ctx context.Context, config Config, err error) {
    l.Error("validation failed",
        zap.String("namespace", config.Namespace),
        zap.String("service", config.ServiceName),
        zap.String("config_key", config.ConfigKey),
        zap.Error(err),
        zap.String("trace_id", getTraceID(ctx)),
    )
}
```

**Log Levels:**
- `DEBUG` - Detailed diagnostic information
- `INFO` - General informational messages
- `WARN` - Warning messages (validation failures, retries)
- `ERROR` - Error messages (system failures, critical issues)
- `FATAL` - Fatal errors causing service shutdown

### 7.5 Health Check Endpoints

```go
type HealthStatus struct {
    Status      string            `json:"status"`      // "healthy", "degraded", "unhealthy"
    Version     string            `json:"version"`
    Uptime      string            `json:"uptime"`
    Timestamp   time.Time         `json:"timestamp"`
    Checks      map[string]Check  `json:"checks"`
}

type Check struct {
    Status  string                 `json:"status"`
    Message string                 `json:"message,omitempty"`
    Details map[string]interface{} `json:"details,omitempty"`
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    health := HealthStatus{
        Status:    "healthy",
        Version:   s.version,
        Uptime:    time.Since(s.startTime).String(),
        Timestamp: time.Now(),
        Checks:    make(map[string]Check),
    }

    // Check database
    dbCheck := s.checkDatabase(ctx)
    health.Checks["database"] = dbCheck
    if dbCheck.Status != "healthy" {
        health.Status = "degraded"
    }

    // Check Redis
    redisCheck := s.checkRedis(ctx)
    health.Checks["redis"] = redisCheck
    if redisCheck.Status != "healthy" {
        health.Status = "degraded"
    }

    // Set appropriate HTTP status
    statusCode := http.StatusOK
    if health.Status == "unhealthy" {
        statusCode = http.StatusServiceUnavailable
    } else if health.Status == "degraded" {
        statusCode = http.StatusOK // Still accept traffic
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(health)
}
```

---

## 8. API Client SDK (Bonus)

### 8.1 Go Client Example

```go
package configclient

import (
    "context"
    "fmt"

    pb "github.com/yourorg/config-service/proto/config/v1"
    "google.golang.org/grpc"
)

type Client struct {
    conn   *grpc.ClientConn
    client pb.ConfigServiceClient
}

func NewClient(address string) (*Client, error) {
    conn, err := grpc.Dial(address, grpc.WithInsecure())
    if err != nil {
        return nil, fmt.Errorf("failed to connect: %w", err)
    }

    return &Client{
        conn:   conn,
        client: pb.NewConfigServiceClient(conn),
    }, nil
}

func (c *Client) GetConfig(ctx context.Context, namespace, service, env, key string) (*pb.Config, error) {
    req := &pb.GetConfigRequest{
        Namespace:   namespace,
        ServiceName: service,
        Environment: env,
        ConfigKey:   key,
    }

    resp, err := c.client.GetConfig(ctx, req)
    if err != nil {
        return nil, err
    }

    return resp.Config, nil
}

// Watch for configuration changes
func (c *Client) WatchChanges(ctx context.Context, namespace, service, env string) (<-chan ConfigChange, error) {
    changes := make(chan ConfigChange)

    // Subscribe to Redis pub/sub
    // Parse and forward events to channel

    return changes, nil
}

func (c *Client) Close() error {
    return c.conn.Close()
}
```

---

## 9. Security Considerations

### 9.1 Encryption

- **At Rest:** Encrypt sensitive config values in database (AES-256)
- **In Transit:** TLS 1.3 for all communications
- **Secrets Management:** Integration with HashiCorp Vault or AWS Secrets Manager

### 9.2 Authentication

- **Service-to-Service:** mTLS certificates
- **User Access:** JWT tokens with short expiration
- **API Keys:** For external integrations, with rotation

### 9.3 Authorization

- RBAC with principle of least privilege
- Row-level security in PostgreSQL
- Audit all access attempts

### 9.4 Input Validation

- Strict schema validation
- SQL injection prevention (parameterized queries)
- XSS prevention (output encoding)
- CSRF protection for web UI

---

## 10. Operational Runbook

### 10.1 Common Operations

**Deploying a Configuration Change:**
1. Create draft version
2. Upload configurations
3. Validate
4. Promote to active
5. Monitor for errors
6. Rollback if needed

**Rollback Procedure:**
1. Identify target version
2. Call rollback API
3. Verify services received event
4. Monitor for issues

**Adding New Service:**
1. Create namespace (if new)
2. Define JSON schema
3. Create initial configuration
4. Document configuration keys

### 10.2 Troubleshooting

**Issue: Events not delivered**
- Check Redis/NATS connectivity
- Verify subscription active
- Check notification logs
- Review retry queue

**Issue: Validation failing**
- Check JSON schema
- Review validation errors in logs
- Test with validation endpoint

**Issue: Database slow**
- Check connection pool
- Review slow query logs
- Analyze query plans
- Add indexes if needed

---

## Appendix A: Example Configuration Files

### A.1 Strategy Configuration Example

```yaml
# namespace: strategy
# service: strategy-engine
# environment: prod

strategy_name: "momentum_scalper"
enabled: true

parameters:
  timeframe: "1m"
  lookback_periods: 20
  momentum_threshold: 0.02
  max_position_size: 50000
  take_profit_percent: 1.5
  stop_loss_percent: 0.8

risk_management:
  max_daily_trades: 100
  max_concurrent_positions: 5
  max_loss_per_trade: 1000

symbols:
  - "BTC/USDT"
  - "ETH/USDT"
  - "SOL/USDT"

exchange_settings:
  order_type: "LIMIT"
  time_in_force: "GTC"
  post_only: true
```

### A.2 Risk Limits Configuration Example

```json
{
  "namespace": "risk",
  "service": "risk-manager",
  "environment": "prod",
  "limits": {
    "max_position_size_usd": 100000,
    "max_leverage": 5.0,
    "max_drawdown_percent": 20.0,
    "daily_loss_limit_usd": 5000,
    "max_open_orders": 50,
    "max_orders_per_second": 10
  },
  "alerts": {
    "drawdown_warning_percent": 15.0,
    "position_warning_percent": 80.0,
    "loss_warning_percent": 75.0
  },
  "emergency_stop": {
    "enabled": true,
    "max_drawdown_trigger": 25.0,
    "daily_loss_trigger": 6000,
    "manual_override_required": true
  }
}
```

---

## Appendix B: Migration from Hardcoded Configs

**Migration Strategy:**

1. **Audit Phase**
   - Inventory all hardcoded configs
   - Map to namespaces and services
   - Define schemas

2. **Migration Phase**
   - Import configs to service
   - Test with validation
   - Deploy config-aware service versions
   - Switch to dynamic config loading

3. **Validation Phase**
   - Verify all services loading correctly
   - Test hot-reload
   - Remove hardcoded fallbacks

---

**End of Configuration Service Development Plan**

This comprehensive plan provides a complete roadmap for implementing a production-ready Configuration Service with versioning, validation, hot-reload, and full observability. The service is designed to be the central source of truth for all trading system configurations, enabling rapid iteration and safe deployments.

**Estimated Timeline:** 8 weeks for full implementation with 1-2 developers

**Next Steps:**
1. Review and approve architecture
2. Set up development environment
3. Begin Phase 1 (Database schema)
4. Establish CI/CD pipeline
5. Weekly progress reviews
