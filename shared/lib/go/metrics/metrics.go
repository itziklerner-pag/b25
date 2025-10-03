package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Common metric labels
const (
	LabelService  = "service"
	LabelExchange = "exchange"
	LabelSymbol   = "symbol"
	LabelStrategy = "strategy"
	LabelSide     = "side"
	LabelStatus   = "status"
	LabelType     = "type"
)

// Metrics holds all Prometheus metrics for the B25 trading system.
type Metrics struct {
	// Latency metrics
	MarketDataLatency    *prometheus.HistogramVec
	StrategyLatency      *prometheus.HistogramVec
	OrderValidationLatency *prometheus.HistogramVec
	OrderSubmissionLatency *prometheus.HistogramVec
	EndToEndLatency      *prometheus.HistogramVec

	// Throughput metrics
	MarketDataEventsTotal prometheus.Counter
	OrdersSubmittedTotal  *prometheus.CounterVec
	FillsTotal            *prometheus.CounterVec
	SignalsTotal          *prometheus.CounterVec

	// Business metrics
	PnLGauge              *prometheus.GaugeVec
	RealizedPnL           *prometheus.GaugeVec
	UnrealizedPnL         *prometheus.GaugeVec
	DailyPnL              *prometheus.GaugeVec
	PositionSize          *prometheus.GaugeVec
	AccountBalance        *prometheus.GaugeVec

	// Performance metrics
	WinRate               *prometheus.GaugeVec
	MakerFillRate         *prometheus.GaugeVec
	FillRate              *prometheus.GaugeVec

	// System metrics
	WebSocketConnected    *prometheus.GaugeVec
	CircuitBreakerState   *prometheus.GaugeVec
	ErrorRate             *prometheus.CounterVec
	OrderRejectRate       *prometheus.CounterVec
	ReconciliationMismatch *prometheus.CounterVec

	// Order book metrics
	SpreadBPS             *prometheus.GaugeVec
	Imbalance             *prometheus.GaugeVec
	BidDepth              *prometheus.GaugeVec
	AskDepth              *prometheus.GaugeVec

	// Rate limiter metrics
	RateLimitRemaining    *prometheus.GaugeVec
	RateLimitHits         *prometheus.CounterVec
}

// NewMetrics creates and registers all Prometheus metrics.
func NewMetrics(namespace string) *Metrics {
	return &Metrics{
		// Latency metrics (microseconds)
		MarketDataLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "market_data_latency_microseconds",
				Help:      "Market data processing latency in microseconds",
				Buckets:   prometheus.ExponentialBuckets(10, 2, 10), // 10μs to 5ms
			},
			[]string{LabelExchange, LabelSymbol},
		),

		StrategyLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "strategy_latency_microseconds",
				Help:      "Strategy decision latency in microseconds",
				Buckets:   prometheus.ExponentialBuckets(50, 2, 10), // 50μs to 25ms
			},
			[]string{LabelStrategy, LabelSymbol},
		),

		OrderValidationLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "order_validation_latency_microseconds",
				Help:      "Order validation latency in microseconds",
				Buckets:   prometheus.ExponentialBuckets(10, 2, 10),
			},
			[]string{LabelExchange},
		),

		OrderSubmissionLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "order_submission_latency_milliseconds",
				Help:      "Order submission latency in milliseconds",
				Buckets:   prometheus.ExponentialBuckets(10, 2, 8), // 10ms to 1.28s
			},
			[]string{LabelExchange, LabelSymbol},
		),

		EndToEndLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "end_to_end_latency_milliseconds",
				Help:      "End-to-end latency from signal to order in milliseconds",
				Buckets:   prometheus.ExponentialBuckets(1, 2, 10), // 1ms to 512ms
			},
			[]string{LabelStrategy, LabelExchange, LabelSymbol},
		),

		// Throughput metrics
		MarketDataEventsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "market_data_events_total",
				Help:      "Total number of market data events processed",
			},
		),

		OrdersSubmittedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "orders_submitted_total",
				Help:      "Total number of orders submitted",
			},
			[]string{LabelExchange, LabelSymbol, LabelSide, LabelType},
		),

		FillsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "fills_total",
				Help:      "Total number of fills",
			},
			[]string{LabelExchange, LabelSymbol, LabelSide},
		),

		SignalsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "signals_total",
				Help:      "Total number of trading signals generated",
			},
			[]string{LabelStrategy, LabelSymbol, LabelSide},
		),

		// Business metrics
		PnLGauge: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "pnl_usdt",
				Help:      "Current profit/loss in USDT",
			},
			[]string{LabelStrategy, LabelSymbol},
		),

		RealizedPnL: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "realized_pnl_usdt",
				Help:      "Realized profit/loss in USDT",
			},
			[]string{LabelStrategy, LabelSymbol},
		),

		UnrealizedPnL: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "unrealized_pnl_usdt",
				Help:      "Unrealized profit/loss in USDT",
			},
			[]string{LabelStrategy, LabelSymbol},
		),

		DailyPnL: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "daily_pnl_usdt",
				Help:      "Daily profit/loss in USDT",
			},
			[]string{LabelStrategy},
		),

		PositionSize: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "position_size",
				Help:      "Current position size",
			},
			[]string{LabelExchange, LabelSymbol, LabelSide},
		),

		AccountBalance: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "account_balance",
				Help:      "Account balance by asset",
			},
			[]string{LabelExchange, "asset"},
		),

		// Performance metrics
		WinRate: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "win_rate_percent",
				Help:      "Percentage of profitable trades",
			},
			[]string{LabelStrategy},
		),

		MakerFillRate: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "maker_fill_rate_percent",
				Help:      "Percentage of orders filled as maker",
			},
			[]string{LabelExchange, LabelSymbol},
		),

		FillRate: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "fill_rate_percent",
				Help:      "Percentage of orders that get filled",
			},
			[]string{LabelExchange, LabelSymbol},
		),

		// System metrics
		WebSocketConnected: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "websocket_connected",
				Help:      "WebSocket connection status (1=connected, 0=disconnected)",
			},
			[]string{LabelExchange, "stream"},
		),

		CircuitBreakerState: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "circuit_breaker_state",
				Help:      "Circuit breaker state (0=closed, 1=open, 2=half-open)",
			},
			[]string{LabelService, "circuit"},
		),

		ErrorRate: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "errors_total",
				Help:      "Total number of errors",
			},
			[]string{LabelService, LabelType},
		),

		OrderRejectRate: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "order_rejects_total",
				Help:      "Total number of order rejections",
			},
			[]string{LabelExchange, LabelSymbol, "reason"},
		),

		ReconciliationMismatch: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "reconciliation_mismatches_total",
				Help:      "Total number of reconciliation mismatches",
			},
			[]string{LabelExchange, LabelType},
		),

		// Order book metrics
		SpreadBPS: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "spread_basis_points",
				Help:      "Order book spread in basis points",
			},
			[]string{LabelExchange, LabelSymbol},
		),

		Imbalance: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "imbalance",
				Help:      "Order book imbalance (-1 to 1)",
			},
			[]string{LabelExchange, LabelSymbol},
		),

		BidDepth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "bid_depth",
				Help:      "Total bid volume",
			},
			[]string{LabelExchange, LabelSymbol},
		),

		AskDepth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "ask_depth",
				Help:      "Total ask volume",
			},
			[]string{LabelExchange, LabelSymbol},
		),

		// Rate limiter metrics
		RateLimitRemaining: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "rate_limit_remaining",
				Help:      "Remaining rate limit quota",
			},
			[]string{LabelService, "limiter"},
		),

		RateLimitHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "rate_limit_hits_total",
				Help:      "Total number of rate limit hits",
			},
			[]string{LabelService, "limiter"},
		),
	}
}
