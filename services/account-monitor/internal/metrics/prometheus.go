package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Position metrics
	PositionCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "account_positions_total",
			Help: "Current number of open positions",
		},
		[]string{"symbol"},
	)

	PositionValue = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "account_position_value_usd",
			Help: "Position value in USD",
		},
		[]string{"symbol"},
	)

	// P&L metrics
	RealizedPnL = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "account_realized_pnl_usd",
			Help: "Total realized P&L in USD",
		},
	)

	UnrealizedPnL = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "account_unrealized_pnl_usd",
			Help: "Unrealized P&L per symbol",
		},
		[]string{"symbol"},
	)

	// Balance metrics
	AccountBalance = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "account_balance",
			Help: "Account balance by asset",
		},
		[]string{"asset"},
	)

	AccountEquity = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "account_equity_usd",
			Help: "Total account equity in USD",
		},
	)

	// Reconciliation metrics
	ReconciliationDrift = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "reconciliation_drift_abs",
			Help:    "Absolute drift values during reconciliation",
			Buckets: prometheus.ExponentialBuckets(0.0001, 10, 8),
		},
	)

	ReconciliationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "reconciliation_duration_seconds",
			Help:    "Duration of reconciliation process",
			Buckets: prometheus.DefBuckets,
		},
	)

	// Alert metrics
	AlertsTriggered = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "alerts_triggered_total",
			Help: "Number of alerts triggered by type",
		},
		[]string{"type", "severity"},
	)

	// WebSocket metrics
	WebSocketReconnects = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "websocket_reconnects_total",
			Help: "Number of WebSocket reconnections",
		},
	)

	WebSocketMessagesReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "websocket_messages_received_total",
			Help: "Number of WebSocket messages received by type",
		},
		[]string{"type"},
	)

	// gRPC metrics
	GRPCRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	GRPCDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_duration_seconds",
			Help:    "gRPC request duration",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)
