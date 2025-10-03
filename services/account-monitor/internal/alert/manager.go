package alert

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/yourorg/b25/services/account-monitor/internal/config"
	"github.com/yourorg/b25/services/account-monitor/internal/metrics"
)

type AlertType string

const (
	AlertLowBalance    AlertType = "LOW_BALANCE"
	AlertHighDrawdown  AlertType = "HIGH_DRAWDOWN"
	AlertMarginRatio   AlertType = "HIGH_MARGIN_RATIO"
	AlertBalanceDrift  AlertType = "BALANCE_DRIFT"
	AlertPositionDrift AlertType = "POSITION_DRIFT"
)

type Manager struct {
	config          config.AlertsConfig
	redis           *redis.Client
	natsConn        *nats.Conn
	db              *pgxpool.Pool
	logger          *zap.Logger
	suppressionMap  map[AlertType]time.Time
	suppressionLock sync.RWMutex
}

type Alert struct {
	Type      AlertType       `json:"type"`
	Severity  string          `json:"severity"` // INFO, WARNING, CRITICAL
	Symbol    string          `json:"symbol,omitempty"`
	Message   string          `json:"message"`
	Value     decimal.Decimal `json:"value"`
	Threshold decimal.Decimal `json:"threshold,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

func NewManager(
	cfg config.AlertsConfig,
	redis *redis.Client,
	natsConn *nats.Conn,
	db *pgxpool.Pool,
	logger *zap.Logger,
) *Manager {
	return &Manager{
		config:         cfg,
		redis:          redis,
		natsConn:       natsConn,
		db:             db,
		logger:         logger,
		suppressionMap: make(map[AlertType]time.Time),
	}
}

// Start begins alert monitoring (placeholder for future expansion)
func (m *Manager) Start(ctx context.Context) {
	m.logger.Info("Alert manager started")
	<-ctx.Done()
	m.logger.Info("Alert manager stopped")
}

// PublishAlert publishes an alert
func (m *Manager) PublishAlert(alert Alert) error {
	if !m.config.Enabled {
		return nil
	}

	// Check suppression
	if m.shouldSuppress(alert.Type) {
		m.logger.Debug("Alert suppressed", zap.String("type", string(alert.Type)))
		return nil
	}

	// Update suppression map
	m.suppressionLock.Lock()
	m.suppressionMap[alert.Type] = time.Now()
	m.suppressionLock.Unlock()

	// Update metrics
	metrics.AlertsTriggered.WithLabelValues(string(alert.Type), alert.Severity).Inc()

	// Store in database
	if err := m.storeAlert(context.Background(), alert); err != nil {
		m.logger.Error("Failed to store alert", zap.Error(err))
	}

	// Publish to NATS
	data, err := json.Marshal(alert)
	if err != nil {
		return err
	}

	if err := m.natsConn.Publish("trading.alerts", data); err != nil {
		return fmt.Errorf("failed to publish alert to NATS: %w", err)
	}

	m.logger.Info("Alert published",
		zap.String("type", string(alert.Type)),
		zap.String("severity", alert.Severity),
		zap.String("message", alert.Message),
	)

	return nil
}

// shouldSuppress checks if alert should be suppressed
func (m *Manager) shouldSuppress(alertType AlertType) bool {
	m.suppressionLock.RLock()
	defer m.suppressionLock.RUnlock()

	lastTime, exists := m.suppressionMap[alertType]
	if !exists {
		return false
	}

	return time.Since(lastTime) < m.config.SuppressionDuration
}

// storeAlert stores alert in database
func (m *Manager) storeAlert(ctx context.Context, alert Alert) error {
	query := `
		INSERT INTO alerts (type, severity, symbol, message, value, threshold, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := m.db.Exec(ctx, query,
		alert.Type,
		alert.Severity,
		alert.Symbol,
		alert.Message,
		alert.Value,
		alert.Threshold,
		alert.Timestamp,
	)

	return err
}

// GetRecentAlerts retrieves recent alerts
func (m *Manager) GetRecentAlerts(ctx context.Context, limit int) ([]Alert, error) {
	query := `
		SELECT type, severity, symbol, message, value, threshold, timestamp
		FROM alerts
		ORDER BY timestamp DESC
		LIMIT $1
	`

	rows, err := m.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	alerts := []Alert{}
	for rows.Next() {
		var alert Alert
		err := rows.Scan(
			&alert.Type,
			&alert.Severity,
			&alert.Symbol,
			&alert.Message,
			&alert.Value,
			&alert.Threshold,
			&alert.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// CheckBalanceThreshold checks if balance is below threshold
func (m *Manager) CheckBalanceThreshold(balance, threshold decimal.Decimal) *Alert {
	if balance.LessThan(threshold) {
		return &Alert{
			Type:      AlertLowBalance,
			Severity:  "WARNING",
			Message:   "Balance below threshold",
			Value:     balance,
			Threshold: threshold,
			Timestamp: time.Now(),
		}
	}
	return nil
}

// CheckDrawdown checks if drawdown exceeds threshold
func (m *Manager) CheckDrawdown(realizedPnL, initialBalance, threshold decimal.Decimal) *Alert {
	drawdownPct := realizedPnL.Div(initialBalance).Mul(decimal.NewFromInt(100))
	if drawdownPct.LessThan(threshold) {
		return &Alert{
			Type:      AlertHighDrawdown,
			Severity:  "CRITICAL",
			Message:   fmt.Sprintf("Drawdown %.2f%% exceeds threshold %.2f%%", drawdownPct.InexactFloat64(), threshold.InexactFloat64()),
			Value:     drawdownPct,
			Threshold: threshold,
			Timestamp: time.Now(),
		}
	}
	return nil
}

// CheckMarginRatio checks if margin ratio exceeds threshold
func (m *Manager) CheckMarginRatio(marginRatio, threshold decimal.Decimal) *Alert {
	if marginRatio.GreaterThan(threshold) {
		return &Alert{
			Type:      AlertMarginRatio,
			Severity:  "CRITICAL",
			Message:   fmt.Sprintf("Margin ratio %.2f%% exceeds threshold %.2f%%", marginRatio.InexactFloat64(), threshold.InexactFloat64()),
			Value:     marginRatio,
			Threshold: threshold,
			Timestamp: time.Now(),
		}
	}
	return nil
}
