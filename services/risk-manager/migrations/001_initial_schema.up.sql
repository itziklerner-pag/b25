-- Risk Manager Service - Initial Schema

-- Risk policies table
CREATE TABLE IF NOT EXISTS risk_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('hard', 'soft', 'emergency')),
    metric VARCHAR(100) NOT NULL,
    operator VARCHAR(50) NOT NULL CHECK (operator IN ('less_than', 'less_than_or_equal', 'greater_than', 'greater_than_or_equal', 'equal', 'not_equal')),
    threshold NUMERIC(20, 8) NOT NULL,
    scope VARCHAR(50) NOT NULL CHECK (scope IN ('account', 'symbol', 'strategy')),
    scope_id VARCHAR(100),
    action VARCHAR(50),
    enabled BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    version INTEGER DEFAULT 1
);

CREATE INDEX idx_policies_enabled ON risk_policies(enabled);
CREATE INDEX idx_policies_scope ON risk_policies(scope, scope_id);
CREATE INDEX idx_policies_type ON risk_policies(type);

-- Risk violations table
CREATE TABLE IF NOT EXISTS risk_violations (
    id BIGSERIAL PRIMARY KEY,
    policy_id UUID REFERENCES risk_policies(id) ON DELETE SET NULL,
    violation_time TIMESTAMPTZ DEFAULT NOW(),
    metric_value NUMERIC(20, 8),
    threshold_value NUMERIC(20, 8),
    context JSONB,
    action_taken VARCHAR(100),
    resolved BOOLEAN DEFAULT false,
    resolved_at TIMESTAMPTZ
);

CREATE INDEX idx_violations_time ON risk_violations(violation_time DESC);
CREATE INDEX idx_violations_policy ON risk_violations(policy_id);
CREATE INDEX idx_violations_resolved ON risk_violations(resolved);

-- Emergency stops table
CREATE TABLE IF NOT EXISTS emergency_stops (
    id BIGSERIAL PRIMARY KEY,
    trigger_time TIMESTAMPTZ DEFAULT NOW(),
    trigger_reason TEXT NOT NULL,
    triggered_by VARCHAR(100) NOT NULL,
    account_state JSONB,
    positions_snapshot JSONB,
    orders_canceled INTEGER DEFAULT 0,
    positions_closed INTEGER DEFAULT 0,
    completed_at TIMESTAMPTZ,
    re_enabled_at TIMESTAMPTZ,
    re_enabled_by VARCHAR(100)
);

CREATE INDEX idx_emergency_stops_time ON emergency_stops(trigger_time DESC);
CREATE INDEX idx_emergency_stops_completed ON emergency_stops(completed_at);

-- Insert default policies
INSERT INTO risk_policies (id, name, type, metric, operator, threshold, scope, enabled, priority) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Max Account Leverage', 'hard', 'leverage', 'less_than_or_equal', 10.0, 'account', true, 100),
    ('22222222-2222-2222-2222-222222222222', 'Min Margin Ratio', 'hard', 'margin_ratio', 'greater_than_or_equal', 1.0, 'account', true, 100),
    ('33333333-3333-3333-3333-333333333333', 'Daily Drawdown Warning', 'soft', 'drawdown_daily', 'greater_than', 0.10, 'account', true, 50),
    ('44444444-4444-4444-4444-444444444444', 'Max Drawdown Hard Limit', 'hard', 'drawdown_max', 'greater_than', 0.20, 'account', true, 150),
    ('55555555-5555-5555-5555-555555555555', 'Emergency Drawdown Stop', 'emergency', 'drawdown_max', 'greater_than', 0.25, 'account', true, 200)
ON CONFLICT (id) DO NOTHING;

-- Add comments
COMMENT ON TABLE risk_policies IS 'Risk management policies and limits';
COMMENT ON TABLE risk_violations IS 'Historical record of policy violations';
COMMENT ON TABLE emergency_stops IS 'Emergency stop events and their details';

COMMENT ON COLUMN risk_policies.type IS 'Policy severity: hard (blocks orders), soft (warning only), emergency (triggers stop)';
COMMENT ON COLUMN risk_policies.metric IS 'Risk metric to evaluate: leverage, margin_ratio, drawdown_daily, drawdown_max, concentration_<symbol>';
COMMENT ON COLUMN risk_policies.operator IS 'Comparison operator for threshold evaluation';
COMMENT ON COLUMN risk_policies.scope IS 'What the policy applies to: account, symbol, or strategy';
COMMENT ON COLUMN risk_policies.scope_id IS 'Specific ID for symbol or strategy scope (NULL for account)';
