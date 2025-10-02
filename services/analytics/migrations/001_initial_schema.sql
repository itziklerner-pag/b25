-- Analytics Service Database Schema
-- Optimized for high-throughput event ingestion and time-series queries

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- Events table: Raw event storage with time-series partitioning
CREATE TABLE events (
    id UUID DEFAULT uuid_generate_v4(),
    event_type VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    session_id VARCHAR(255),
    properties JSONB DEFAULT '{}'::jsonb,
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

-- Create indexes on events table
CREATE INDEX idx_events_event_type ON events(event_type, timestamp DESC);
CREATE INDEX idx_events_user_id ON events(user_id, timestamp DESC) WHERE user_id IS NOT NULL;
CREATE INDEX idx_events_session_id ON events(session_id, timestamp DESC) WHERE session_id IS NOT NULL;
CREATE INDEX idx_events_timestamp ON events(timestamp DESC);
CREATE INDEX idx_events_properties ON events USING GIN (properties);

-- Create partitions for events (last 7 days + future 7 days)
-- In production, use automated partition management
CREATE TABLE events_default PARTITION OF events DEFAULT;

-- User behavior aggregation table
CREATE TABLE user_behavior (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    session_start TIMESTAMPTZ NOT NULL,
    session_end TIMESTAMPTZ,
    duration INTEGER, -- seconds
    page_views INTEGER DEFAULT 0,
    event_count INTEGER DEFAULT 0,
    unique_event_types INTEGER DEFAULT 0,
    properties JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_behavior_user_id ON user_behavior(user_id, session_start DESC);
CREATE INDEX idx_user_behavior_session_id ON user_behavior(session_id);
CREATE INDEX idx_user_behavior_session_start ON user_behavior(session_start DESC);
CREATE INDEX idx_user_behavior_properties ON user_behavior USING GIN (properties);

-- Metric aggregations table with time-series optimization
CREATE TABLE metric_aggregations (
    id UUID DEFAULT uuid_generate_v4(),
    metric_name VARCHAR(255) NOT NULL,
    interval VARCHAR(10) NOT NULL, -- 1m, 5m, 15m, 1h, 1d
    time_bucket TIMESTAMPTZ NOT NULL,
    count BIGINT NOT NULL DEFAULT 0,
    sum DOUBLE PRECISION,
    avg DOUBLE PRECISION,
    min DOUBLE PRECISION,
    max DOUBLE PRECISION,
    p50 DOUBLE PRECISION,
    p95 DOUBLE PRECISION,
    p99 DOUBLE PRECISION,
    dimensions JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (metric_name, interval, time_bucket, id)
) PARTITION BY RANGE (time_bucket);

-- Create indexes on metric_aggregations
CREATE INDEX idx_metric_agg_time_bucket ON metric_aggregations(metric_name, interval, time_bucket DESC);
CREATE INDEX idx_metric_agg_dimensions ON metric_aggregations USING GIN (dimensions);

-- Create default partition for metric_aggregations
CREATE TABLE metric_aggregations_default PARTITION OF metric_aggregations DEFAULT;

-- Custom event definitions table
CREATE TABLE custom_event_definitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    schema JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_custom_events_name ON custom_event_definitions(name) WHERE is_active = true;
CREATE INDEX idx_custom_events_active ON custom_event_definitions(is_active, created_at DESC);

-- Trading-specific analytics tables

-- Order analytics
CREATE TABLE order_analytics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    symbol VARCHAR(50) NOT NULL,
    side VARCHAR(10) NOT NULL, -- BUY, SELL
    order_type VARCHAR(20) NOT NULL, -- LIMIT, MARKET, etc.
    price DOUBLE PRECISION,
    quantity DOUBLE PRECISION NOT NULL,
    filled_quantity DOUBLE PRECISION DEFAULT 0,
    status VARCHAR(50) NOT NULL,
    strategy_id VARCHAR(255),
    latency_ms DOUBLE PRECISION, -- Order placement latency
    is_maker BOOLEAN,
    commission DOUBLE PRECISION,
    pnl DOUBLE PRECISION,
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (timestamp);

CREATE INDEX idx_order_analytics_order_id ON order_analytics(order_id);
CREATE INDEX idx_order_analytics_user_id ON order_analytics(user_id, timestamp DESC);
CREATE INDEX idx_order_analytics_symbol ON order_analytics(symbol, timestamp DESC);
CREATE INDEX idx_order_analytics_strategy ON order_analytics(strategy_id, timestamp DESC);
CREATE INDEX idx_order_analytics_timestamp ON order_analytics(timestamp DESC);

CREATE TABLE order_analytics_default PARTITION OF order_analytics DEFAULT;

-- Strategy performance analytics
CREATE TABLE strategy_performance (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    strategy_id VARCHAR(255) NOT NULL,
    strategy_name VARCHAR(255) NOT NULL,
    time_bucket TIMESTAMPTZ NOT NULL,
    interval VARCHAR(10) NOT NULL, -- 1m, 5m, 15m, 1h, 1d
    total_trades INTEGER DEFAULT 0,
    winning_trades INTEGER DEFAULT 0,
    losing_trades INTEGER DEFAULT 0,
    total_pnl DOUBLE PRECISION DEFAULT 0,
    gross_profit DOUBLE PRECISION DEFAULT 0,
    gross_loss DOUBLE PRECISION DEFAULT 0,
    avg_win DOUBLE PRECISION,
    avg_loss DOUBLE PRECISION,
    largest_win DOUBLE PRECISION,
    largest_loss DOUBLE PRECISION,
    win_rate DOUBLE PRECISION,
    profit_factor DOUBLE PRECISION,
    sharpe_ratio DOUBLE PRECISION,
    max_drawdown DOUBLE PRECISION,
    avg_latency_ms DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_strategy_perf_strategy ON strategy_performance(strategy_id, interval, time_bucket DESC);
CREATE INDEX idx_strategy_perf_time ON strategy_performance(time_bucket DESC);

-- Market data analytics
CREATE TABLE market_analytics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    symbol VARCHAR(50) NOT NULL,
    time_bucket TIMESTAMPTZ NOT NULL,
    interval VARCHAR(10) NOT NULL,
    open DOUBLE PRECISION,
    high DOUBLE PRECISION,
    low DOUBLE PRECISION,
    close DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    trades_count INTEGER,
    avg_spread DOUBLE PRECISION,
    volatility DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (time_bucket);

CREATE INDEX idx_market_analytics_symbol ON market_analytics(symbol, interval, time_bucket DESC);
CREATE INDEX idx_market_analytics_time ON market_analytics(time_bucket DESC);

CREATE TABLE market_analytics_default PARTITION OF market_analytics DEFAULT;

-- System performance metrics
CREATE TABLE system_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_name VARCHAR(255) NOT NULL,
    metric_type VARCHAR(100) NOT NULL, -- latency, throughput, error_rate, cpu, memory
    time_bucket TIMESTAMPTZ NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    dimensions JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (time_bucket);

CREATE INDEX idx_system_metrics_service ON system_metrics(service_name, metric_type, time_bucket DESC);
CREATE INDEX idx_system_metrics_time ON system_metrics(time_bucket DESC);

CREATE TABLE system_metrics_default PARTITION OF system_metrics DEFAULT;

-- Materialized view for real-time dashboard metrics
CREATE MATERIALIZED VIEW dashboard_metrics AS
SELECT
    COUNT(DISTINCT user_id) FILTER (WHERE timestamp > NOW() - INTERVAL '5 minutes') as active_users,
    COUNT(*) FILTER (WHERE timestamp > NOW() - INTERVAL '1 minute') / 60.0 as events_per_second,
    COUNT(*) FILTER (WHERE event_type = 'order.placed' AND timestamp > NOW() - INTERVAL '1 hour') as orders_placed_1h,
    COUNT(*) FILTER (WHERE event_type = 'order.filled' AND timestamp > NOW() - INTERVAL '1 hour') as orders_filled_1h,
    COUNT(DISTINCT properties->>'strategy_id') FILTER (WHERE event_type LIKE 'strategy.%' AND timestamp > NOW() - INTERVAL '5 minutes') as active_strategies,
    NOW() as last_updated
FROM events
WHERE timestamp > NOW() - INTERVAL '1 hour';

CREATE UNIQUE INDEX idx_dashboard_metrics_refresh ON dashboard_metrics(last_updated);

-- Function to auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for auto-updating timestamps
CREATE TRIGGER update_user_behavior_updated_at
    BEFORE UPDATE ON user_behavior
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_custom_event_definitions_updated_at
    BEFORE UPDATE ON custom_event_definitions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to clean up old data based on retention policies
CREATE OR REPLACE FUNCTION cleanup_old_data() RETURNS void AS $$
BEGIN
    -- Clean up old raw events (older than 90 days by default)
    DELETE FROM events WHERE timestamp < NOW() - INTERVAL '90 days';

    -- Clean up old minute aggregations (older than 1 year)
    DELETE FROM metric_aggregations
    WHERE interval = '1m' AND time_bucket < NOW() - INTERVAL '365 days';

    -- Clean up old hourly aggregations (older than 2 years)
    DELETE FROM metric_aggregations
    WHERE interval = '1h' AND time_bucket < NOW() - INTERVAL '730 days';

    -- Clean up old order analytics (older than 90 days)
    DELETE FROM order_analytics WHERE timestamp < NOW() - INTERVAL '90 days';

    -- Clean up old market analytics (older than 90 days for raw data)
    DELETE FROM market_analytics WHERE time_bucket < NOW() - INTERVAL '90 days' AND interval IN ('1m', '5m');

    -- Clean up old system metrics (older than 30 days)
    DELETE FROM system_metrics WHERE time_bucket < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;

-- Refresh materialized view function
CREATE OR REPLACE FUNCTION refresh_dashboard_metrics() RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY dashboard_metrics;
END;
$$ LANGUAGE plpgsql;

-- Grant permissions (adjust as needed)
-- CREATE USER analytics_user WITH PASSWORD 'secure_password';
-- GRANT CONNECT ON DATABASE analytics TO analytics_user;
-- GRANT USAGE ON SCHEMA public TO analytics_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO analytics_user;
-- GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO analytics_user;

-- Comments for documentation
COMMENT ON TABLE events IS 'Raw event storage with time-series partitioning for high-throughput ingestion';
COMMENT ON TABLE user_behavior IS 'Aggregated user behavior and session analytics';
COMMENT ON TABLE metric_aggregations IS 'Pre-computed metric aggregations at various time intervals';
COMMENT ON TABLE custom_event_definitions IS 'User-defined custom event type definitions';
COMMENT ON TABLE order_analytics IS 'Trading order analytics and performance metrics';
COMMENT ON TABLE strategy_performance IS 'Trading strategy performance metrics over time';
COMMENT ON TABLE market_analytics IS 'Market data analytics and OHLCV aggregations';
COMMENT ON TABLE system_metrics IS 'System performance and health metrics';
COMMENT ON MATERIALIZED VIEW dashboard_metrics IS 'Real-time dashboard metrics (refreshed periodically)';
