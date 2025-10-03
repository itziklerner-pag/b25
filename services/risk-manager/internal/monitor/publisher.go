package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/b25/services/risk-manager/internal/limits"
	"github.com/b25/services/risk-manager/internal/risk"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// NATSAlertPublisher publishes alerts via NATS
type NATSAlertPublisher struct {
	nc             *nats.Conn
	logger         *zap.Logger
	alertSubject   string
	metricsSubject string
	deduplicator   *AlertDeduplicator
}

// NewNATSAlertPublisher creates a new NATS alert publisher
func NewNATSAlertPublisher(nc *nats.Conn, logger *zap.Logger, alertSubject string, alertWindow time.Duration) *NATSAlertPublisher {
	return &NATSAlertPublisher{
		nc:             nc,
		logger:         logger,
		alertSubject:   alertSubject,
		metricsSubject: "risk.metrics",
		deduplicator:   NewAlertDeduplicator(alertWindow),
	}
}

// PublishAlert publishes a risk violation alert
func (p *NATSAlertPublisher) PublishAlert(ctx context.Context, level string, violation *limits.PolicyViolation) error {
	// Create alert key for deduplication
	alertKey := fmt.Sprintf("%s:%s", violation.Policy.ID, violation.Policy.Metric)

	// Check if we should send this alert
	if !p.deduplicator.ShouldAlert(alertKey) {
		p.logger.Debug("alert deduplicated",
			zap.String("policy", violation.Policy.Name),
			zap.String("metric", violation.Policy.Metric),
		)
		return nil
	}

	// Create alert message
	alert := map[string]interface{}{
		"level":           level,
		"policy_id":       violation.Policy.ID,
		"policy_name":     violation.Policy.Name,
		"policy_type":     string(violation.Policy.Type),
		"metric":          violation.Policy.Metric,
		"metric_value":    violation.MetricValue,
		"threshold_value": violation.ThresholdValue,
		"message":         violation.Message,
		"timestamp":       violation.Timestamp.Unix(),
	}

	data, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("marshal alert: %w", err)
	}

	// Publish to NATS
	subject := fmt.Sprintf("%s.%s", p.alertSubject, level)
	if err := p.nc.Publish(subject, data); err != nil {
		return fmt.Errorf("publish to NATS: %w", err)
	}

	p.logger.Info("alert published",
		zap.String("level", level),
		zap.String("policy", violation.Policy.Name),
		zap.String("subject", subject),
	)

	return nil
}

// PublishMetrics publishes risk metrics
func (p *NATSAlertPublisher) PublishMetrics(ctx context.Context, metrics risk.RiskMetrics) error {
	metricsData := map[string]interface{}{
		"margin_ratio":           metrics.MarginRatio,
		"leverage":               metrics.Leverage,
		"drawdown_daily":         metrics.DrawdownDaily,
		"drawdown_max":           metrics.DrawdownMax,
		"daily_pnl":              metrics.DailyPnL,
		"unrealized_pnl":         metrics.UnrealizedPnL,
		"total_equity":           metrics.TotalEquity,
		"total_margin_used":      metrics.TotalMarginUsed,
		"position_concentration": metrics.PositionConcentration,
		"open_positions":         metrics.OpenPositions,
		"pending_orders":         metrics.PendingOrders,
		"timestamp":              time.Now().Unix(),
	}

	data, err := json.Marshal(metricsData)
	if err != nil {
		return fmt.Errorf("marshal metrics: %w", err)
	}

	if err := p.nc.Publish(p.metricsSubject, data); err != nil {
		return fmt.Errorf("publish metrics: %w", err)
	}

	return nil
}

// PublishEmergencyAlert publishes an emergency stop alert
func (p *NATSAlertPublisher) PublishEmergencyAlert(ctx context.Context, reason, triggeredBy string) error {
	alert := map[string]interface{}{
		"level":        "emergency",
		"type":         "emergency_stop",
		"reason":       reason,
		"triggered_by": triggeredBy,
		"timestamp":    time.Now().Unix(),
	}

	data, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("marshal emergency alert: %w", err)
	}

	subject := fmt.Sprintf("%s.emergency", p.alertSubject)
	if err := p.nc.Publish(subject, data); err != nil {
		return fmt.Errorf("publish emergency alert: %w", err)
	}

	p.logger.Error("emergency alert published",
		zap.String("reason", reason),
		zap.String("triggered_by", triggeredBy),
	)

	return nil
}

// StartCleanup starts periodic cleanup of deduplicator
func (p *NATSAlertPublisher) StartCleanup(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.deduplicator.Cleanup()
		case <-ctx.Done():
			return
		}
	}
}
