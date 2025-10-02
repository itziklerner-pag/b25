-- Notification Service Database Schema
-- Version: 1.0.0

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Notification channels enum
CREATE TYPE notification_channel AS ENUM ('email', 'sms', 'push', 'webhook');

-- Notification status enum
CREATE TYPE notification_status AS ENUM (
    'pending',
    'queued',
    'sending',
    'sent',
    'delivered',
    'failed',
    'retrying',
    'cancelled'
);

-- Notification priority enum
CREATE TYPE notification_priority AS ENUM ('low', 'normal', 'high', 'critical');

-- Template types enum
CREATE TYPE template_type AS ENUM (
    'trading_alert',
    'risk_violation',
    'order_fill',
    'account_update',
    'system_notification',
    'custom'
);

-- Users table (notification recipients)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_user_id VARCHAR(255) UNIQUE NOT NULL, -- ID from main system
    email VARCHAR(255),
    phone_number VARCHAR(50),
    timezone VARCHAR(50) DEFAULT 'UTC',
    language VARCHAR(10) DEFAULT 'en',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT valid_email CHECK (email IS NULL OR email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

CREATE INDEX idx_users_external_id ON users(external_user_id);
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;

-- User devices (for push notifications)
CREATE TABLE user_devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_token VARCHAR(500) NOT NULL,
    device_type VARCHAR(50) NOT NULL, -- ios, android, web
    device_name VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT unique_device_token UNIQUE(device_token)
);

CREATE INDEX idx_user_devices_user_id ON user_devices(user_id);
CREATE INDEX idx_user_devices_active ON user_devices(is_active) WHERE is_active = true;

-- User notification preferences
CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel notification_channel NOT NULL,
    category VARCHAR(100) NOT NULL, -- trading_alerts, account_updates, etc.
    is_enabled BOOLEAN DEFAULT true,
    quiet_hours_enabled BOOLEAN DEFAULT false,
    quiet_hours_start TIME,
    quiet_hours_end TIME,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT unique_user_channel_category UNIQUE(user_id, channel, category)
);

CREATE INDEX idx_preferences_user_id ON notification_preferences(user_id);
CREATE INDEX idx_preferences_enabled ON notification_preferences(is_enabled);

-- Notification templates
CREATE TABLE notification_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    type template_type NOT NULL,
    channel notification_channel NOT NULL,
    subject VARCHAR(500), -- for email
    body_template TEXT NOT NULL,
    variables JSONB, -- available template variables
    is_active BOOLEAN DEFAULT true,
    version INT DEFAULT 1,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT valid_template_name CHECK (name ~* '^[a-z0-9_-]+$')
);

CREATE INDEX idx_templates_name ON notification_templates(name);
CREATE INDEX idx_templates_type_channel ON notification_templates(type, channel);
CREATE INDEX idx_templates_active ON notification_templates(is_active) WHERE is_active = true;

-- Notifications table
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel notification_channel NOT NULL,
    template_id UUID REFERENCES notification_templates(id),
    priority notification_priority DEFAULT 'normal',
    status notification_status DEFAULT 'pending',

    -- Content
    subject VARCHAR(500),
    body TEXT NOT NULL,
    metadata JSONB, -- additional context

    -- Delivery details
    recipient_address VARCHAR(500), -- email, phone, or device token

    -- Timing
    scheduled_at TIMESTAMP WITH TIME ZONE,
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,

    -- Retry logic
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    next_retry_at TIMESTAMP WITH TIME ZONE,

    -- Error tracking
    error_message TEXT,
    error_code VARCHAR(100),

    -- External references
    external_message_id VARCHAR(255), -- ID from provider (SendGrid, Twilio, etc.)
    correlation_id VARCHAR(255), -- for tracking related notifications

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_channel ON notifications(channel);
CREATE INDEX idx_notifications_priority ON notifications(priority);
CREATE INDEX idx_notifications_scheduled_at ON notifications(scheduled_at) WHERE scheduled_at IS NOT NULL;
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX idx_notifications_correlation_id ON notifications(correlation_id) WHERE correlation_id IS NOT NULL;
CREATE INDEX idx_notifications_retry ON notifications(status, next_retry_at) WHERE status = 'retrying';

-- Notification events (delivery tracking)
CREATE TABLE notification_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL, -- queued, sent, delivered, opened, clicked, bounced, failed
    event_data JSONB,
    provider_event_id VARCHAR(255),
    occurred_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_notification_events_notification_id ON notification_events(notification_id);
CREATE INDEX idx_notification_events_type ON notification_events(event_type);
CREATE INDEX idx_notification_events_occurred_at ON notification_events(occurred_at DESC);

-- Notification batches (for bulk sending)
CREATE TABLE notification_batches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255),
    description TEXT,
    channel notification_channel NOT NULL,
    template_id UUID REFERENCES notification_templates(id),
    total_count INT DEFAULT 0,
    sent_count INT DEFAULT 0,
    failed_count INT DEFAULT 0,
    status VARCHAR(50) DEFAULT 'pending', -- pending, processing, completed, failed
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_batches_status ON notification_batches(status);
CREATE INDEX idx_batches_created_at ON notification_batches(created_at DESC);

-- Rate limiting tracking
CREATE TABLE rate_limits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    channel notification_channel NOT NULL,
    window_start TIMESTAMP WITH TIME ZONE NOT NULL,
    window_end TIMESTAMP WITH TIME ZONE NOT NULL,
    count INT DEFAULT 0,
    limit_value INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT unique_rate_limit UNIQUE(user_id, channel, window_start)
);

CREATE INDEX idx_rate_limits_window ON rate_limits(window_start, window_end);
CREATE INDEX idx_rate_limits_user_channel ON rate_limits(user_id, channel);

-- Webhook configurations
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(1000) NOT NULL,
    secret VARCHAR(255), -- for signature verification
    event_types TEXT[], -- array of event types to receive
    is_active BOOLEAN DEFAULT true,
    retry_on_failure BOOLEAN DEFAULT true,
    max_retries INT DEFAULT 3,
    timeout_seconds INT DEFAULT 30,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_webhooks_user_id ON webhooks(user_id);
CREATE INDEX idx_webhooks_active ON webhooks(is_active) WHERE is_active = true;

-- Webhook delivery logs
CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    notification_id UUID REFERENCES notifications(id) ON DELETE SET NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    response_status_code INT,
    response_body TEXT,
    retry_count INT DEFAULT 0,
    delivered_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_webhook_deliveries_webhook_id ON webhook_deliveries(webhook_id);
CREATE INDEX idx_webhook_deliveries_created_at ON webhook_deliveries(created_at DESC);

-- Alert rules (for automated notifications)
CREATE TABLE alert_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    event_source VARCHAR(100) NOT NULL, -- trading, risk, account, system
    event_type VARCHAR(100) NOT NULL, -- specific event to trigger on
    conditions JSONB, -- conditions to evaluate
    template_id UUID REFERENCES notification_templates(id),
    channels notification_channel[] NOT NULL,
    priority notification_priority DEFAULT 'normal',
    recipient_groups TEXT[], -- user groups to notify
    is_active BOOLEAN DEFAULT true,
    cooldown_minutes INT DEFAULT 0, -- prevent spam
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_alert_rules_active ON alert_rules(is_active) WHERE is_active = true;
CREATE INDEX idx_alert_rules_event ON alert_rules(event_source, event_type);

-- Alert rule executions (audit log)
CREATE TABLE alert_rule_executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rule_id UUID NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    event_data JSONB,
    matched BOOLEAN NOT NULL,
    notifications_created INT DEFAULT 0,
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_rule_executions_rule_id ON alert_rule_executions(rule_id);
CREATE INDEX idx_rule_executions_executed_at ON alert_rule_executions(executed_at DESC);

-- Functions and triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_devices_updated_at BEFORE UPDATE ON user_devices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notification_preferences_updated_at BEFORE UPDATE ON notification_preferences
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notification_templates_updated_at BEFORE UPDATE ON notification_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notifications_updated_at BEFORE UPDATE ON notifications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_webhooks_updated_at BEFORE UPDATE ON webhooks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_alert_rules_updated_at BEFORE UPDATE ON alert_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create views for common queries
CREATE VIEW notification_stats_by_channel AS
SELECT
    channel,
    status,
    priority,
    COUNT(*) as count,
    DATE_TRUNC('hour', created_at) as hour
FROM notifications
WHERE created_at > NOW() - INTERVAL '24 hours'
GROUP BY channel, status, priority, DATE_TRUNC('hour', created_at);

CREATE VIEW user_notification_summary AS
SELECT
    u.id as user_id,
    u.email,
    COUNT(n.id) as total_notifications,
    COUNT(CASE WHEN n.status = 'delivered' THEN 1 END) as delivered_count,
    COUNT(CASE WHEN n.status = 'failed' THEN 1 END) as failed_count,
    MAX(n.created_at) as last_notification_at
FROM users u
LEFT JOIN notifications n ON u.id = n.user_id
WHERE u.deleted_at IS NULL
GROUP BY u.id, u.email;

-- Comments for documentation
COMMENT ON TABLE notifications IS 'Core notification records with delivery tracking';
COMMENT ON TABLE notification_templates IS 'Reusable notification templates';
COMMENT ON TABLE notification_preferences IS 'User preferences for notification delivery';
COMMENT ON TABLE alert_rules IS 'Automated notification rules triggered by events';
COMMENT ON COLUMN notifications.correlation_id IS 'Links related notifications together';
COMMENT ON COLUMN notifications.metadata IS 'Additional context and custom fields';
