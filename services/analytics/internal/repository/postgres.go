package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/b25/analytics/internal/config"
	"github.com/b25/analytics/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles database operations
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new repository instance
func NewRepository(cfg *config.DatabaseConfig) (*Repository, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&pool_max_conns=%d",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.SSLMode,
		cfg.MaxConnections,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MaxIdleConnections)
	poolConfig.MaxConnLifetime = cfg.ConnectionLifetime

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Repository{pool: pool}, nil
}

// Close closes the database connection pool
func (r *Repository) Close() {
	r.pool.Close()
}

// InsertEvent inserts a single event
func (r *Repository) InsertEvent(ctx context.Context, event *models.Event) error {
	query := `
		INSERT INTO events (id, event_type, user_id, session_id, properties, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query,
		event.ID,
		event.EventType,
		event.UserID,
		event.SessionID,
		event.Properties,
		event.Timestamp,
		event.CreatedAt,
	)

	return err
}

// InsertEventsBatch inserts multiple events in a batch
func (r *Repository) InsertEventsBatch(ctx context.Context, events []*models.Event) error {
	if len(events) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO events (id, event_type, user_id, session_id, properties, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	for _, event := range events {
		batch.Queue(query,
			event.ID,
			event.EventType,
			event.UserID,
			event.SessionID,
			event.Properties,
			event.Timestamp,
			event.CreatedAt,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range events {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed to execute batch: %w", err)
		}
	}

	return nil
}

// GetEventsByTimeRange retrieves events within a time range
func (r *Repository) GetEventsByTimeRange(ctx context.Context, startTime, endTime time.Time, eventType string, limit int) ([]*models.Event, error) {
	query := `
		SELECT id, event_type, user_id, session_id, properties, timestamp, created_at
		FROM events
		WHERE timestamp >= $1 AND timestamp <= $2
	`

	args := []interface{}{startTime, endTime}
	if eventType != "" {
		query += " AND event_type = $3"
		args = append(args, eventType)
	}

	query += " ORDER BY timestamp DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event := &models.Event{}
		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.UserID,
			&event.SessionID,
			&event.Properties,
			&event.Timestamp,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

// GetEventCountByType returns event counts grouped by type
func (r *Repository) GetEventCountByType(ctx context.Context, startTime, endTime time.Time) (map[string]int64, error) {
	query := `
		SELECT event_type, COUNT(*) as count
		FROM events
		WHERE timestamp >= $1 AND timestamp <= $2
		GROUP BY event_type
		ORDER BY count DESC
	`

	rows, err := r.pool.Query(ctx, query, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int64)
	for rows.Next() {
		var eventType string
		var count int64
		if err := rows.Scan(&eventType, &count); err != nil {
			return nil, err
		}
		counts[eventType] = count
	}

	return counts, rows.Err()
}

// UpsertUserBehavior inserts or updates user behavior data
func (r *Repository) UpsertUserBehavior(ctx context.Context, behavior *models.UserBehavior) error {
	query := `
		INSERT INTO user_behavior (id, user_id, session_id, session_start, session_end, duration,
			page_views, event_count, unique_event_types, properties, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			session_end = EXCLUDED.session_end,
			duration = EXCLUDED.duration,
			page_views = EXCLUDED.page_views,
			event_count = EXCLUDED.event_count,
			unique_event_types = EXCLUDED.unique_event_types,
			properties = EXCLUDED.properties,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query,
		behavior.ID,
		behavior.UserID,
		behavior.SessionID,
		behavior.SessionStart,
		behavior.SessionEnd,
		behavior.Duration,
		behavior.PageViews,
		behavior.EventCount,
		behavior.UniqueEventTypes,
		behavior.Properties,
		behavior.CreatedAt,
		behavior.UpdatedAt,
	)

	return err
}

// InsertMetricAggregation inserts a metric aggregation
func (r *Repository) InsertMetricAggregation(ctx context.Context, metric *models.MetricAggregation) error {
	query := `
		INSERT INTO metric_aggregations (id, metric_name, interval, time_bucket, count, sum, avg,
			min, max, p50, p95, p99, dimensions, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.pool.Exec(ctx, query,
		metric.ID,
		metric.MetricName,
		metric.Interval,
		metric.TimeBucket,
		metric.Count,
		metric.Sum,
		metric.Avg,
		metric.Min,
		metric.Max,
		metric.Percentile50,
		metric.Percentile95,
		metric.Percentile99,
		metric.Dimensions,
		metric.CreatedAt,
	)

	return err
}

// GetMetricAggregations retrieves metric aggregations for a time range
func (r *Repository) GetMetricAggregations(ctx context.Context, metricName, interval string, startTime, endTime time.Time) ([]*models.MetricAggregation, error) {
	query := `
		SELECT id, metric_name, interval, time_bucket, count, sum, avg, min, max,
			p50, p95, p99, dimensions, created_at
		FROM metric_aggregations
		WHERE metric_name = $1 AND interval = $2
			AND time_bucket >= $3 AND time_bucket <= $4
		ORDER BY time_bucket ASC
	`

	rows, err := r.pool.Query(ctx, query, metricName, interval, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*models.MetricAggregation
	for rows.Next() {
		metric := &models.MetricAggregation{}
		err := rows.Scan(
			&metric.ID,
			&metric.MetricName,
			&metric.Interval,
			&metric.TimeBucket,
			&metric.Count,
			&metric.Sum,
			&metric.Avg,
			&metric.Min,
			&metric.Max,
			&metric.Percentile50,
			&metric.Percentile95,
			&metric.Percentile99,
			&metric.Dimensions,
			&metric.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	return metrics, rows.Err()
}

// CreateCustomEventDefinition creates a new custom event definition
func (r *Repository) CreateCustomEventDefinition(ctx context.Context, def *models.CustomEventDefinition) error {
	query := `
		INSERT INTO custom_event_definitions (id, name, display_name, description, schema, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		def.ID,
		def.Name,
		def.DisplayName,
		def.Description,
		def.Schema,
		def.IsActive,
		def.CreatedAt,
		def.UpdatedAt,
	)

	return err
}

// GetCustomEventDefinition retrieves a custom event definition by name
func (r *Repository) GetCustomEventDefinition(ctx context.Context, name string) (*models.CustomEventDefinition, error) {
	query := `
		SELECT id, name, display_name, description, schema, is_active, created_at, updated_at
		FROM custom_event_definitions
		WHERE name = $1
	`

	def := &models.CustomEventDefinition{}
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&def.ID,
		&def.Name,
		&def.DisplayName,
		&def.Description,
		&def.Schema,
		&def.IsActive,
		&def.CreatedAt,
		&def.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return def, nil
}

// GetDashboardMetrics retrieves real-time dashboard metrics
func (r *Repository) GetDashboardMetrics(ctx context.Context) (*models.DashboardMetrics, error) {
	query := `
		SELECT
			COUNT(DISTINCT user_id) FILTER (WHERE timestamp > NOW() - INTERVAL '5 minutes') as active_users,
			COUNT(*) FILTER (WHERE timestamp > NOW() - INTERVAL '1 minute')::float / 60.0 as events_per_second,
			COUNT(*) FILTER (WHERE event_type = 'order.placed' AND timestamp > NOW() - INTERVAL '1 hour') as orders_placed,
			COUNT(*) FILTER (WHERE event_type = 'order.filled' AND timestamp > NOW() - INTERVAL '1 hour') as orders_filled,
			COUNT(DISTINCT properties->>'strategy_id') FILTER (WHERE event_type LIKE 'strategy.%' AND timestamp > NOW() - INTERVAL '5 minutes') as active_strategies
		FROM events
		WHERE timestamp > NOW() - INTERVAL '1 hour'
	`

	metrics := &models.DashboardMetrics{
		CustomMetrics: make(map[string]float64),
		LastUpdated:   time.Now(),
	}

	err := r.pool.QueryRow(ctx, query).Scan(
		&metrics.ActiveUsers,
		&metrics.EventsPerSecond,
		&metrics.OrdersPlaced,
		&metrics.OrdersFilled,
		&metrics.ActiveStrategies,
	)

	if err != nil {
		return nil, err
	}

	return metrics, nil
}

// InsertOrderAnalytics inserts order analytics data
func (r *Repository) InsertOrderAnalytics(ctx context.Context, orderID, userID, symbol, side, orderType, status, strategyID string,
	price, quantity, filledQuantity, latencyMs, commission, pnl *float64, isMaker *bool, timestamp time.Time) error {

	query := `
		INSERT INTO order_analytics (id, order_id, user_id, symbol, side, order_type, price, quantity,
			filled_quantity, status, strategy_id, latency_ms, is_maker, commission, pnl, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err := r.pool.Exec(ctx, query,
		uuid.New(),
		orderID,
		userID,
		symbol,
		side,
		orderType,
		price,
		quantity,
		filledQuantity,
		status,
		strategyID,
		latencyMs,
		isMaker,
		commission,
		pnl,
		timestamp,
		time.Now(),
	)

	return err
}

// RefreshDashboardMaterializedView refreshes the dashboard metrics materialized view
func (r *Repository) RefreshDashboardMaterializedView(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY dashboard_metrics")
	return err
}

// CleanupOldData runs the cleanup function to remove old data
func (r *Repository) CleanupOldData(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, "SELECT cleanup_old_data()")
	return err
}

// GetHealth checks database health
func (r *Repository) GetHealth(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
