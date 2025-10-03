-- Initialize test database schema

-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    order_id VARCHAR(64) PRIMARY KEY,
    client_order_id VARCHAR(64),
    exchange_order_id VARCHAR(64),
    user_id VARCHAR(64),
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL,
    order_type VARCHAR(20) NOT NULL,
    state VARCHAR(20) NOT NULL,
    time_in_force VARCHAR(10),
    quantity DECIMAL(20, 8) NOT NULL,
    price DECIMAL(20, 8),
    stop_price DECIMAL(20, 8),
    filled_quantity DECIMAL(20, 8) DEFAULT 0,
    average_price DECIMAL(20, 8) DEFAULT 0,
    fee DECIMAL(20, 8) DEFAULT 0,
    fee_asset VARCHAR(10),
    reduce_only BOOLEAN DEFAULT FALSE,
    post_only BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_symbol ON orders(symbol);
CREATE INDEX IF NOT EXISTS idx_orders_state ON orders(state);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);

-- Create fills table
CREATE TABLE IF NOT EXISTS fills (
    fill_id VARCHAR(64) PRIMARY KEY,
    order_id VARCHAR(64) NOT NULL REFERENCES orders(order_id),
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL,
    price DECIMAL(20, 8) NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL,
    fee DECIMAL(20, 8) DEFAULT 0,
    fee_asset VARCHAR(10),
    is_maker BOOLEAN DEFAULT FALSE,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fills_order_id ON fills(order_id);
CREATE INDEX IF NOT EXISTS idx_fills_symbol ON fills(symbol);
CREATE INDEX IF NOT EXISTS idx_fills_timestamp ON fills(timestamp);

-- Create positions table
CREATE TABLE IF NOT EXISTS positions (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL,
    avg_entry_price DECIMAL(20, 8) NOT NULL,
    current_price DECIMAL(20, 8),
    unrealized_pnl DECIMAL(20, 8) DEFAULT 0,
    realized_pnl DECIMAL(20, 8) DEFAULT 0,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, symbol)
);

CREATE INDEX IF NOT EXISTS idx_positions_user_id ON positions(user_id);
CREATE INDEX IF NOT EXISTS idx_positions_symbol ON positions(symbol);

-- Create balances table
CREATE TABLE IF NOT EXISTS balances (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    asset VARCHAR(10) NOT NULL,
    free DECIMAL(20, 8) NOT NULL DEFAULT 0,
    locked DECIMAL(20, 8) NOT NULL DEFAULT 0,
    total DECIMAL(20, 8) GENERATED ALWAYS AS (free + locked) STORED,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, asset)
);

CREATE INDEX IF NOT EXISTS idx_balances_user_id ON balances(user_id);

-- Create account_history table for reconciliation
CREATE TABLE IF NOT EXISTS account_history (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    asset VARCHAR(10) NOT NULL,
    amount DECIMAL(20, 8) NOT NULL,
    balance_before DECIMAL(20, 8),
    balance_after DECIMAL(20, 8),
    reference_id VARCHAR(64),
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_account_history_user_id ON account_history(user_id);
CREATE INDEX IF NOT EXISTS idx_account_history_timestamp ON account_history(timestamp);

-- Create strategy_signals table
CREATE TABLE IF NOT EXISTS strategy_signals (
    id VARCHAR(64) PRIMARY KEY,
    strategy_name VARCHAR(50) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL,
    order_type VARCHAR(20) NOT NULL,
    price DECIMAL(20, 8),
    quantity DECIMAL(20, 8) NOT NULL,
    stop_price DECIMAL(20, 8),
    priority INT DEFAULT 5,
    status VARCHAR(20) DEFAULT 'PENDING',
    order_id VARCHAR(64),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    executed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_strategy_signals_strategy ON strategy_signals(strategy_name);
CREATE INDEX IF NOT EXISTS idx_strategy_signals_status ON strategy_signals(status);
CREATE INDEX IF NOT EXISTS idx_strategy_signals_created_at ON strategy_signals(created_at);

-- Create strategy_performance table
CREATE TABLE IF NOT EXISTS strategy_performance (
    id SERIAL PRIMARY KEY,
    strategy_name VARCHAR(50) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    total_trades INT DEFAULT 0,
    winning_trades INT DEFAULT 0,
    losing_trades INT DEFAULT 0,
    total_pnl DECIMAL(20, 8) DEFAULT 0,
    win_rate DECIMAL(5, 2) DEFAULT 0,
    avg_win DECIMAL(20, 8) DEFAULT 0,
    avg_loss DECIMAL(20, 8) DEFAULT 0,
    sharpe_ratio DECIMAL(10, 4) DEFAULT 0,
    max_drawdown DECIMAL(20, 8) DEFAULT 0,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_strategy_performance_strategy ON strategy_performance(strategy_name);
CREATE INDEX IF NOT EXISTS idx_strategy_performance_timestamp ON strategy_performance(timestamp);

-- Insert test data
INSERT INTO balances (user_id, asset, free, locked) VALUES
    ('test_user_1', 'BTC', 10.0, 0.0),
    ('test_user_1', 'ETH', 100.0, 0.0),
    ('test_user_1', 'USDT', 100000.0, 0.0)
ON CONFLICT (user_id, asset) DO NOTHING;

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO testuser;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO testuser;
