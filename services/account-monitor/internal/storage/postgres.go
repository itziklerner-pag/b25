package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/b25/services/account-monitor/internal/config"
)

// NewPostgresPool creates a new PostgreSQL connection pool
func NewPostgresPool(ctx context.Context, cfg config.PostgresConfig) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxConnections)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// RunMigrations runs database migrations
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []string{
		// Create TimescaleDB extension if not exists
		`CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;`,

		// Create P&L snapshots table
		`CREATE TABLE IF NOT EXISTS pnl_snapshots (
			id BIGSERIAL,
			timestamp TIMESTAMPTZ NOT NULL,
			symbol VARCHAR(20),
			realized_pnl DECIMAL(20, 8) NOT NULL DEFAULT 0,
			unrealized_pnl DECIMAL(20, 8) NOT NULL DEFAULT 0,
			total_pnl DECIMAL(20, 8) NOT NULL DEFAULT 0,
			total_fees DECIMAL(20, 8) NOT NULL DEFAULT 0,
			net_pnl DECIMAL(20, 8) NOT NULL DEFAULT 0,
			win_rate DECIMAL(5, 2) DEFAULT 0,
			total_trades INT DEFAULT 0
		);`,

		// Create hypertable
		`SELECT create_hypertable('pnl_snapshots', 'timestamp', if_not_exists => TRUE);`,

		// Create indexes
		`CREATE INDEX IF NOT EXISTS idx_pnl_snapshots_timestamp ON pnl_snapshots(timestamp DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_pnl_snapshots_symbol ON pnl_snapshots(symbol, timestamp DESC);`,

		// Create alerts table
		`CREATE TABLE IF NOT EXISTS alerts (
			id BIGSERIAL PRIMARY KEY,
			type VARCHAR(50) NOT NULL,
			severity VARCHAR(20) NOT NULL,
			symbol VARCHAR(20),
			message TEXT NOT NULL,
			value DECIMAL(20, 8),
			threshold DECIMAL(20, 8),
			timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,

		// Create alert index
		`CREATE INDEX IF NOT EXISTS idx_alerts_timestamp ON alerts(timestamp DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_type ON alerts(type, timestamp DESC);`,
	}

	for i, migration := range migrations {
		if _, err := pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	return nil
}
