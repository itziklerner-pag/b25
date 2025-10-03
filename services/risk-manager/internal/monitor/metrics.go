package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsCollector collects Prometheus metrics
type MetricsCollector struct {
	// Pre-trade check metrics
	orderChecksTotal      *prometheus.CounterVec
	orderCheckDuration    prometheus.Histogram
	ordersApproved        prometheus.Counter
	ordersRejected        *prometheus.CounterVec

	// Risk metrics
	currentLeverage       prometheus.Gauge
	currentMarginRatio    prometheus.Gauge
	currentDrawdown       prometheus.Gauge
	currentEquity         prometheus.Gauge
	openPositionsCount    prometheus.Gauge
	pendingOrdersCount    prometheus.Gauge

	// Violation metrics
	violationsTotal       *prometheus.CounterVec
	emergencyStopsTotal   prometheus.Counter
	emergencyStopActive   prometheus.Gauge

	// Alert metrics
	alertsPublished       *prometheus.CounterVec
	alertsDeduplicated    prometheus.Counter

	// Circuit breaker metrics
	circuitBreakerTrips   prometheus.Counter
	circuitBreakerCount   prometheus.Gauge
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		// Pre-trade check metrics
		orderChecksTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "risk_order_checks_total",
				Help: "Total number of order risk checks performed",
			},
			[]string{"result"}, // approved, rejected
		),
		orderCheckDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "risk_order_check_duration_microseconds",
				Help:    "Duration of order risk checks in microseconds",
				Buckets: []float64{100, 500, 1000, 2500, 5000, 10000, 25000, 50000},
			},
		),
		ordersApproved: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "risk_orders_approved_total",
				Help: "Total number of orders approved",
			},
		),
		ordersRejected: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "risk_orders_rejected_total",
				Help: "Total number of orders rejected",
			},
			[]string{"reason"},
		),

		// Risk metrics
		currentLeverage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "risk_current_leverage",
				Help: "Current account leverage",
			},
		),
		currentMarginRatio: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "risk_current_margin_ratio",
				Help: "Current margin ratio (equity / margin used)",
			},
		),
		currentDrawdown: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "risk_current_drawdown",
				Help: "Current drawdown percentage",
			},
		),
		currentEquity: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "risk_current_equity",
				Help: "Current account equity",
			},
		),
		openPositionsCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "risk_open_positions",
				Help: "Number of open positions",
			},
		),
		pendingOrdersCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "risk_pending_orders",
				Help: "Number of pending orders",
			},
		),

		// Violation metrics
		violationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "risk_violations_total",
				Help: "Total number of policy violations",
			},
			[]string{"policy_type", "policy_name"},
		),
		emergencyStopsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "risk_emergency_stops_total",
				Help: "Total number of emergency stops triggered",
			},
		),
		emergencyStopActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "risk_emergency_stop_active",
				Help: "Whether emergency stop is currently active (1=active, 0=inactive)",
			},
		),

		// Alert metrics
		alertsPublished: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "risk_alerts_published_total",
				Help: "Total number of alerts published",
			},
			[]string{"level"},
		),
		alertsDeduplicated: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "risk_alerts_deduplicated_total",
				Help: "Total number of alerts deduplicated",
			},
		),

		// Circuit breaker metrics
		circuitBreakerTrips: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "risk_circuit_breaker_trips_total",
				Help: "Total number of circuit breaker trips",
			},
		),
		circuitBreakerCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "risk_circuit_breaker_violation_count",
				Help: "Current number of violations in circuit breaker window",
			},
		),
	}
}

// RecordOrderCheck records an order check
func (m *MetricsCollector) RecordOrderCheck(approved bool, durationUs int64, rejectionReason string) {
	m.orderCheckDuration.Observe(float64(durationUs))

	if approved {
		m.orderChecksTotal.WithLabelValues("approved").Inc()
		m.ordersApproved.Inc()
	} else {
		m.orderChecksTotal.WithLabelValues("rejected").Inc()
		m.ordersRejected.WithLabelValues(rejectionReason).Inc()
	}
}

// UpdateRiskMetrics updates current risk metrics
func (m *MetricsCollector) UpdateRiskMetrics(leverage, marginRatio, drawdown, equity float64, openPositions, pendingOrders int) {
	m.currentLeverage.Set(leverage)
	m.currentMarginRatio.Set(marginRatio)
	m.currentDrawdown.Set(drawdown)
	m.currentEquity.Set(equity)
	m.openPositionsCount.Set(float64(openPositions))
	m.pendingOrdersCount.Set(float64(pendingOrders))
}

// RecordViolation records a policy violation
func (m *MetricsCollector) RecordViolation(policyType, policyName string) {
	m.violationsTotal.WithLabelValues(policyType, policyName).Inc()
}

// RecordEmergencyStop records an emergency stop
func (m *MetricsCollector) RecordEmergencyStop() {
	m.emergencyStopsTotal.Inc()
	m.emergencyStopActive.Set(1)
}

// ClearEmergencyStop clears the emergency stop flag
func (m *MetricsCollector) ClearEmergencyStop() {
	m.emergencyStopActive.Set(0)
}

// RecordAlert records an alert publication
func (m *MetricsCollector) RecordAlert(level string) {
	m.alertsPublished.WithLabelValues(level).Inc()
}

// RecordAlertDeduplicated records a deduplicated alert
func (m *MetricsCollector) RecordAlertDeduplicated() {
	m.alertsDeduplicated.Inc()
}

// RecordCircuitBreakerTrip records a circuit breaker trip
func (m *MetricsCollector) RecordCircuitBreakerTrip() {
	m.circuitBreakerTrips.Inc()
}

// UpdateCircuitBreakerCount updates the circuit breaker violation count
func (m *MetricsCollector) UpdateCircuitBreakerCount(count int) {
	m.circuitBreakerCount.Set(float64(count))
}
