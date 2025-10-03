package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	// Order metrics
	OrdersCreated   prometheus.Counter
	OrdersCanceled  prometheus.Counter
	OrdersRejected  prometheus.Counter
	OrdersFilled    prometheus.Counter
	OrdersPartial   prometheus.Counter

	// Latency metrics
	OrderLatency    prometheus.Histogram
	CancelLatency   prometheus.Histogram

	// Exchange metrics
	ExchangeRequests prometheus.Counter
	ExchangeErrors   prometheus.Counter
	ExchangeLatency  prometheus.Histogram

	// State metrics
	OrderStateGauge *prometheus.GaugeVec

	// Rate limit metrics
	RateLimitHits   prometheus.Counter

	// Circuit breaker metrics
	CircuitBreakerState *prometheus.GaugeVec

	// Event metrics
	EventsPublished prometheus.Counter
	EventsFailed    prometheus.Counter

	// Cache metrics
	CacheHits   prometheus.Counter
	CacheMisses prometheus.Counter
}

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		OrdersCreated: promauto.NewCounter(prometheus.CounterOpts{
			Name: "order_execution_orders_created_total",
			Help: "Total number of orders created",
		}),
		OrdersCanceled: promauto.NewCounter(prometheus.CounterOpts{
			Name: "order_execution_orders_canceled_total",
			Help: "Total number of orders canceled",
		}),
		OrdersRejected: promauto.NewCounter(prometheus.CounterOpts{
			Name: "order_execution_orders_rejected_total",
			Help: "Total number of orders rejected",
		}),
		OrdersFilled: promauto.NewCounter(prometheus.CounterOpts{
			Name: "order_execution_orders_filled_total",
			Help: "Total number of orders filled",
		}),
		OrdersPartial: promauto.NewCounter(prometheus.CounterOpts{
			Name: "order_execution_orders_partial_total",
			Help: "Total number of partially filled orders",
		}),
		OrderLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "order_execution_order_latency_seconds",
			Help:    "Order creation latency in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}),
		CancelLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "order_execution_cancel_latency_seconds",
			Help:    "Order cancellation latency in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}),
		ExchangeRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: "order_execution_exchange_requests_total",
			Help: "Total number of exchange API requests",
		}),
		ExchangeErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "order_execution_exchange_errors_total",
			Help: "Total number of exchange API errors",
		}),
		ExchangeLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "order_execution_exchange_latency_seconds",
			Help:    "Exchange API latency in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}),
		OrderStateGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "order_execution_order_state",
			Help: "Current number of orders in each state",
		}, []string{"state"}),
		RateLimitHits: promauto.NewCounter(prometheus.CounterOpts{
			Name: "order_execution_rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		}),
		CircuitBreakerState: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "order_execution_circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=half-open, 2=open)",
		}, []string{"name"}),
		EventsPublished: promauto.NewCounter(prometheus.CounterOpts{
			Name: "order_execution_events_published_total",
			Help: "Total number of events published",
		}),
		EventsFailed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "order_execution_events_failed_total",
			Help: "Total number of failed event publications",
		}),
		CacheHits: promauto.NewCounter(prometheus.CounterOpts{
			Name: "order_execution_cache_hits_total",
			Help: "Total number of cache hits",
		}),
		CacheMisses: promauto.NewCounter(prometheus.CounterOpts{
			Name: "order_execution_cache_misses_total",
			Help: "Total number of cache misses",
		}),
	}
}

// RecordOrderState records the current state of an order
func (m *Metrics) RecordOrderState(state string) {
	m.OrderStateGauge.WithLabelValues(state).Inc()
}

// RecordCircuitBreakerState records circuit breaker state
func (m *Metrics) RecordCircuitBreakerState(name string, state int) {
	m.CircuitBreakerState.WithLabelValues(name).Set(float64(state))
}
