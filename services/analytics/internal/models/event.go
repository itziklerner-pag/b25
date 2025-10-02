package models

import (
	"time"

	"github.com/google/uuid"
)

// Event represents a tracked event in the system
type Event struct {
	ID         uuid.UUID              `json:"id" db:"id"`
	EventType  string                 `json:"event_type" db:"event_type"`
	UserID     *string                `json:"user_id,omitempty" db:"user_id"`
	SessionID  *string                `json:"session_id,omitempty" db:"session_id"`
	Properties map[string]interface{} `json:"properties" db:"properties"`
	Timestamp  time.Time              `json:"timestamp" db:"timestamp"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

// EventType constants for trading system events
const (
	EventTypeOrderPlaced      = "order.placed"
	EventTypeOrderFilled      = "order.filled"
	EventTypeOrderCanceled    = "order.canceled"
	EventTypeOrderRejected    = "order.rejected"
	EventTypeStrategyStarted  = "strategy.started"
	EventTypeStrategyStopped  = "strategy.stopped"
	EventTypeSignalGenerated  = "signal.generated"
	EventTypePositionOpened   = "position.opened"
	EventTypePositionClosed   = "position.closed"
	EventTypeBalanceUpdated   = "balance.updated"
	EventTypeMarketDataUpdate = "market.data.update"
	EventTypeAlertTriggered   = "alert.triggered"
	EventTypeUserLogin        = "user.login"
	EventTypeUserLogout       = "user.logout"
	EventTypePageView         = "page.view"
	EventTypeButtonClick      = "button.click"
	EventTypeCustomEvent      = "custom.event"
)

// UserBehavior represents aggregated user behavior data
type UserBehavior struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	UserID          string                 `json:"user_id" db:"user_id"`
	SessionID       string                 `json:"session_id" db:"session_id"`
	SessionStart    time.Time              `json:"session_start" db:"session_start"`
	SessionEnd      *time.Time             `json:"session_end,omitempty" db:"session_end"`
	Duration        *int64                 `json:"duration,omitempty" db:"duration"` // seconds
	PageViews       int                    `json:"page_views" db:"page_views"`
	EventCount      int                    `json:"event_count" db:"event_count"`
	UniqueEventTypes int                   `json:"unique_event_types" db:"unique_event_types"`
	Properties      map[string]interface{} `json:"properties" db:"properties"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// MetricAggregation represents time-based metric aggregations
type MetricAggregation struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	MetricName    string                 `json:"metric_name" db:"metric_name"`
	Interval      string                 `json:"interval" db:"interval"` // 1m, 5m, 15m, 1h, 1d
	TimeBucket    time.Time              `json:"time_bucket" db:"time_bucket"`
	Count         int64                  `json:"count" db:"count"`
	Sum           *float64               `json:"sum,omitempty" db:"sum"`
	Avg           *float64               `json:"avg,omitempty" db:"avg"`
	Min           *float64               `json:"min,omitempty" db:"min"`
	Max           *float64               `json:"max,omitempty" db:"max"`
	Percentile50  *float64               `json:"p50,omitempty" db:"p50"`
	Percentile95  *float64               `json:"p95,omitempty" db:"p95"`
	Percentile99  *float64               `json:"p99,omitempty" db:"p99"`
	Dimensions    map[string]interface{} `json:"dimensions" db:"dimensions"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
}

// CustomEventDefinition represents a custom event type definition
type CustomEventDefinition struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	DisplayName string                 `json:"display_name" db:"display_name"`
	Description string                 `json:"description" db:"description"`
	Schema      map[string]interface{} `json:"schema" db:"schema"`
	IsActive    bool                   `json:"is_active" db:"is_active"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// TimeSeriesPoint represents a single point in a time series
type TimeSeriesPoint struct {
	Timestamp  time.Time              `json:"timestamp"`
	Value      float64                `json:"value"`
	Dimensions map[string]interface{} `json:"dimensions,omitempty"`
}

// QueryResult represents the result of an analytics query
type QueryResult struct {
	MetricName  string            `json:"metric_name"`
	Interval    string            `json:"interval"`
	StartTime   time.Time         `json:"start_time"`
	EndTime     time.Time         `json:"end_time"`
	DataPoints  []TimeSeriesPoint `json:"data_points"`
	TotalCount  int               `json:"total_count"`
	Aggregation string            `json:"aggregation"` // count, sum, avg, min, max
}

// DashboardMetrics represents real-time dashboard metrics
type DashboardMetrics struct {
	ActiveUsers        int64              `json:"active_users"`
	EventsPerSecond    float64            `json:"events_per_second"`
	OrdersPlaced       int64              `json:"orders_placed"`
	OrdersFilled       int64              `json:"orders_filled"`
	ActiveStrategies   int64              `json:"active_strategies"`
	TotalVolume        float64            `json:"total_volume"`
	AverageLatency     float64            `json:"average_latency"`
	ErrorRate          float64            `json:"error_rate"`
	CustomMetrics      map[string]float64 `json:"custom_metrics"`
	LastUpdated        time.Time          `json:"last_updated"`
}
