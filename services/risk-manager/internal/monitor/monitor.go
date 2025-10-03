package monitor

import (
	"context"
	"time"

	"github.com/b25/services/risk-manager/internal/emergency"
	"github.com/b25/services/risk-manager/internal/limits"
	"github.com/b25/services/risk-manager/internal/repository"
	"github.com/b25/services/risk-manager/internal/risk"
	"go.uber.org/zap"
)

// RiskMonitor continuously monitors risk metrics
type RiskMonitor struct {
	logger         *zap.Logger
	calculator     *risk.Calculator
	policyEngine   *limits.PolicyEngine
	repository     *repository.PolicyRepository
	stopManager    *emergency.StopManager
	alertPublisher AlertPublisher
	interval       time.Duration
	circuitBreaker *emergency.CircuitBreaker
}

// AlertPublisher publishes alerts
type AlertPublisher interface {
	PublishAlert(ctx context.Context, level string, violation *limits.PolicyViolation) error
	PublishMetrics(ctx context.Context, metrics risk.RiskMetrics) error
}

// NewRiskMonitor creates a new risk monitor
func NewRiskMonitor(
	logger *zap.Logger,
	calculator *risk.Calculator,
	policyEngine *limits.PolicyEngine,
	repository *repository.PolicyRepository,
	stopManager *emergency.StopManager,
	alertPublisher AlertPublisher,
	interval time.Duration,
) *RiskMonitor {
	monitor := &RiskMonitor{
		logger:         logger,
		calculator:     calculator,
		policyEngine:   policyEngine,
		repository:     repository,
		stopManager:    stopManager,
		alertPublisher: alertPublisher,
		interval:       interval,
	}

	// Create circuit breaker for emergency stops
	// Trigger emergency stop if 5 violations in 1 minute
	monitor.circuitBreaker = emergency.NewCircuitBreaker(5, 1*time.Minute, func() {
		monitor.logger.Error("circuit breaker tripped - triggering emergency stop")
		if err := stopManager.Trigger(context.Background(), "Circuit breaker tripped due to repeated violations", "risk_monitor", false); err != nil {
			monitor.logger.Error("failed to trigger emergency stop from circuit breaker", zap.Error(err))
		}
	})

	return monitor
}

// Run starts the monitoring loop
func (m *RiskMonitor) Run(ctx context.Context) error {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	m.logger.Info("risk monitor started", zap.Duration("interval", m.interval))

	for {
		select {
		case <-ticker.C:
			if err := m.checkRisk(ctx); err != nil {
				m.logger.Error("risk check failed", zap.Error(err))
			}

		case <-ctx.Done():
			m.logger.Info("risk monitor stopped")
			return ctx.Err()
		}
	}
}

// checkRisk performs a single risk check
func (m *RiskMonitor) checkRisk(ctx context.Context) error {
	// Get account state (mock for now)
	accountState := m.getMockAccountState()

	// Calculate risk metrics
	metrics := m.calculator.CalculateMetrics(accountState)

	// Publish metrics to dashboard
	if err := m.alertPublisher.PublishMetrics(ctx, metrics); err != nil {
		m.logger.Error("failed to publish metrics", zap.Error(err))
	}

	// Convert metrics to map for policy evaluation
	metricsMap := limits.MetricsFromRiskMetrics(
		metrics.Leverage,
		metrics.MarginRatio,
		metrics.DrawdownDaily,
		metrics.DrawdownMax,
		metrics.PositionConcentration,
	)

	// Evaluate all active policies
	violations := m.policyEngine.EvaluateAll(metricsMap, "", "")

	// Handle violations
	if len(violations) > 0 {
		m.handleViolations(ctx, violations, accountState)
	}

	return nil
}

// handleViolations processes detected violations
func (m *RiskMonitor) handleViolations(ctx context.Context, violations []*limits.PolicyViolation, accountState risk.AccountState) {
	// Group violations by type
	grouped := limits.GetViolationsByType(violations)

	// Handle emergency violations
	if emergencyViolations, ok := grouped[limits.PolicyTypeEmergency]; ok && len(emergencyViolations) > 0 {
		m.logger.Error("emergency violations detected",
			zap.Int("count", len(emergencyViolations)),
		)

		// Log all emergency violations
		for _, v := range emergencyViolations {
			m.logger.Error("emergency violation",
				zap.String("policy", v.Policy.Name),
				zap.String("message", v.Message),
			)

			// Record violation
			m.recordViolation(ctx, v, accountState, "emergency_stop_triggered")

			// Publish critical alert
			if err := m.alertPublisher.PublishAlert(ctx, "critical", v); err != nil {
				m.logger.Error("failed to publish emergency alert", zap.Error(err))
			}
		}

		// Trigger emergency stop
		reason := limits.FormatViolations(emergencyViolations)[0]
		if err := m.stopManager.Trigger(ctx, reason, "risk_monitor", false); err != nil {
			m.logger.Error("failed to trigger emergency stop", zap.Error(err))
		}

		return
	}

	// Handle hard violations
	if hardViolations, ok := grouped[limits.PolicyTypeHard]; ok && len(hardViolations) > 0 {
		m.logger.Warn("hard violations detected",
			zap.Int("count", len(hardViolations)),
		)

		for _, v := range hardViolations {
			m.logger.Warn("hard violation",
				zap.String("policy", v.Policy.Name),
				zap.String("message", v.Message),
			)

			// Record violation
			m.recordViolation(ctx, v, accountState, "alert_published")

			// Publish critical alert
			if err := m.alertPublisher.PublishAlert(ctx, "critical", v); err != nil {
				m.logger.Error("failed to publish hard violation alert", zap.Error(err))
			}

			// Record in circuit breaker
			if m.circuitBreaker.RecordViolation() {
				m.logger.Error("circuit breaker tripped")
			}
		}
	}

	// Handle soft violations
	if softViolations, ok := grouped[limits.PolicyTypeSoft]; ok && len(softViolations) > 0 {
		m.logger.Info("soft violations detected",
			zap.Int("count", len(softViolations)),
		)

		for _, v := range softViolations {
			// Record violation
			m.recordViolation(ctx, v, accountState, "warning_published")

			// Publish warning alert
			if err := m.alertPublisher.PublishAlert(ctx, "warning", v); err != nil {
				m.logger.Error("failed to publish soft violation alert", zap.Error(err))
			}
		}
	}
}

// recordViolation records a violation to the database
func (m *RiskMonitor) recordViolation(ctx context.Context, violation *limits.PolicyViolation, accountState risk.AccountState, action string) {
	contextData := map[string]interface{}{
		"timestamp":      violation.Timestamp.Unix(),
		"equity":         accountState.Equity,
		"margin_used":    accountState.MarginUsed,
		"positions":      len(accountState.Positions),
		"pending_orders": len(accountState.PendingOrders),
	}

	if err := m.repository.RecordViolation(
		ctx,
		violation.Policy.ID,
		violation.MetricValue,
		violation.ThresholdValue,
		contextData,
		action,
	); err != nil {
		m.logger.Error("failed to record violation", zap.Error(err))
	}
}

// getMockAccountState returns mock account state (replace with real data)
func (m *RiskMonitor) getMockAccountState() risk.AccountState {
	return risk.AccountState{
		Equity:           100000.0,
		Balance:          100000.0,
		UnrealizedPnL:    0.0,
		MarginUsed:       10000.0,
		AvailableMargin:  90000.0,
		Positions:        []risk.Position{},
		PendingOrders:    []risk.Order{},
		PeakEquity:       105000.0,
		DailyStartEquity: 98000.0,
	}
}

// AlertDeduplicator prevents duplicate alerts
type AlertDeduplicator struct {
	recentAlerts map[string]time.Time
	window       time.Duration
}

// NewAlertDeduplicator creates a new alert deduplicator
func NewAlertDeduplicator(window time.Duration) *AlertDeduplicator {
	return &AlertDeduplicator{
		recentAlerts: make(map[string]time.Time),
		window:       window,
	}
}

// ShouldAlert checks if an alert should be sent
func (d *AlertDeduplicator) ShouldAlert(alertKey string) bool {
	lastSent, exists := d.recentAlerts[alertKey]

	if !exists || time.Since(lastSent) > d.window {
		d.recentAlerts[alertKey] = time.Now()
		return true
	}

	return false
}

// Cleanup removes old alerts from the map
func (d *AlertDeduplicator) Cleanup() {
	for key, t := range d.recentAlerts {
		if time.Since(t) > d.window*2 {
			delete(d.recentAlerts, key)
		}
	}
}
