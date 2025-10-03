-- Create configurations table
CREATE TABLE IF NOT EXISTS configurations (
    id UUID PRIMARY KEY,
    key VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,
    value JSONB NOT NULL,
    format VARCHAR(20) NOT NULL DEFAULT 'json',
    description TEXT,
    version INTEGER NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on key for fast lookups
CREATE INDEX idx_configurations_key ON configurations(key);

-- Create index on type for filtering
CREATE INDEX idx_configurations_type ON configurations(type);

-- Create index on is_active for filtering
CREATE INDEX idx_configurations_is_active ON configurations(is_active);

-- Create configuration_versions table for version history
CREATE TABLE IF NOT EXISTS configuration_versions (
    id UUID PRIMARY KEY,
    configuration_id UUID NOT NULL REFERENCES configurations(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    value JSONB NOT NULL,
    format VARCHAR(20) NOT NULL DEFAULT 'json',
    changed_by VARCHAR(255) NOT NULL,
    change_reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(configuration_id, version)
);

-- Create index on configuration_id for fast version lookups
CREATE INDEX idx_config_versions_config_id ON configuration_versions(configuration_id);

-- Create index on version for sorting
CREATE INDEX idx_config_versions_version ON configuration_versions(configuration_id, version DESC);

-- Create audit_logs table for tracking all changes
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY,
    configuration_id UUID NOT NULL REFERENCES configurations(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    actor_id VARCHAR(255) NOT NULL,
    actor_name VARCHAR(255) NOT NULL,
    old_value JSONB,
    new_value JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on configuration_id for audit log lookups
CREATE INDEX idx_audit_logs_config_id ON audit_logs(configuration_id);

-- Create index on timestamp for sorting
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);

-- Create index on actor_id for filtering
CREATE INDEX idx_audit_logs_actor_id ON audit_logs(actor_id);

-- Insert some example configurations
INSERT INTO configurations (id, key, type, value, format, description, version, is_active, created_by, created_at, updated_at)
VALUES
    (
        gen_random_uuid(),
        'default_strategy',
        'strategy',
        '{"name": "Market Making", "type": "market_making", "enabled": true, "parameters": {"spread": 0.002, "order_size": 100}}'::jsonb,
        'json',
        'Default market making strategy configuration',
        1,
        true,
        'system',
        NOW(),
        NOW()
    ),
    (
        gen_random_uuid(),
        'default_risk_limits',
        'risk_limit',
        '{"max_position_size": 10000, "max_loss_per_trade": 500, "max_daily_loss": 2000, "max_leverage": 10, "stop_loss_percent": 5}'::jsonb,
        'json',
        'Default risk limit configuration',
        1,
        true,
        'system',
        NOW(),
        NOW()
    ),
    (
        gen_random_uuid(),
        'btc_usdt_pair',
        'trading_pair',
        '{"symbol": "BTC/USDT", "base_currency": "BTC", "quote_currency": "USDT", "min_order_size": 0.001, "max_order_size": 10, "price_precision": 2, "quantity_precision": 8, "enabled": true}'::jsonb,
        'json',
        'Bitcoin/USDT trading pair configuration',
        1,
        true,
        'system',
        NOW(),
        NOW()
    );
