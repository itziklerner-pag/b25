package reconciliation

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/yourorg/b25/services/account-monitor/internal/balance"
	"github.com/yourorg/b25/services/account-monitor/internal/config"
	"github.com/yourorg/b25/services/account-monitor/internal/metrics"
	"github.com/yourorg/b25/services/account-monitor/internal/position"
)

type Reconciler struct {
	positionMgr       *position.Manager
	balanceMgr        *balance.Manager
	exchangeClient    ExchangeClient
	config            config.ReconciliationConfig
	logger            *zap.Logger
	balanceTolerance  decimal.Decimal
	positionTolerance decimal.Decimal
}

type ExchangeClient interface {
	GetAccountInfo(ctx context.Context) (*ExchangeAccount, error)
}

type ExchangeAccount struct {
	Balances  map[string]ExchangeBalance
	Positions map[string]ExchangePosition
}

type ExchangeBalance struct {
	Total decimal.Decimal
	Free  decimal.Decimal
	Locked decimal.Decimal
}

type ExchangePosition struct {
	Quantity   decimal.Decimal
	EntryPrice decimal.Decimal
}

type ReconciliationReport struct {
	Timestamp      time.Time
	BalanceDrifts  []BalanceDrift
	PositionDrifts []PositionDrift
	Corrected      bool
	Error          error
}

type BalanceDrift struct {
	Asset           string
	LocalBalance    decimal.Decimal
	ExchangeBalance decimal.Decimal
	Drift           decimal.Decimal
	DriftPercent    decimal.Decimal
}

type PositionDrift struct {
	Symbol           string
	LocalQuantity    decimal.Decimal
	ExchangeQuantity decimal.Decimal
	Drift            decimal.Decimal
	DriftPercent     decimal.Decimal
}

func NewReconciler(
	positionMgr *position.Manager,
	balanceMgr *balance.Manager,
	exchangeClient ExchangeClient,
	cfg config.ReconciliationConfig,
	logger *zap.Logger,
) *Reconciler {
	return &Reconciler{
		positionMgr:       positionMgr,
		balanceMgr:        balanceMgr,
		exchangeClient:    exchangeClient,
		config:            cfg,
		logger:            logger,
		balanceTolerance:  cfg.BalanceTolerance,
		positionTolerance: cfg.PositionTolerance,
	}
}

// Start begins periodic reconciliation
func (r *Reconciler) Start(ctx context.Context) {
	if !r.config.Enabled {
		r.logger.Info("Reconciliation is disabled")
		return
	}

	ticker := time.NewTicker(r.config.Interval)
	defer ticker.Stop()

	r.logger.Info("Reconciliation started", zap.Duration("interval", r.config.Interval))

	for {
		select {
		case <-ticker.C:
			start := time.Now()
			report, err := r.ReconcileNow(ctx)
			duration := time.Since(start)

			metrics.ReconciliationDuration.Observe(duration.Seconds())

			if err != nil {
				r.logger.Error("Reconciliation failed", zap.Error(err))
				continue
			}

			if len(report.BalanceDrifts) > 0 || len(report.PositionDrifts) > 0 {
				r.logger.Warn("Drifts detected during reconciliation",
					zap.Int("balance_drifts", len(report.BalanceDrifts)),
					zap.Int("position_drifts", len(report.PositionDrifts)),
				)
			} else {
				r.logger.Debug("Reconciliation completed successfully, no drifts detected")
			}

		case <-ctx.Done():
			r.logger.Info("Reconciliation stopped")
			return
		}
	}
}

// ReconcileNow performs immediate reconciliation
func (r *Reconciler) ReconcileNow(ctx context.Context) (*ReconciliationReport, error) {
	report := &ReconciliationReport{
		Timestamp: time.Now(),
	}

	// Fetch exchange account state
	exchangeAccount, err := r.exchangeClient.GetAccountInfo(ctx)
	if err != nil {
		report.Error = err
		return report, fmt.Errorf("failed to get exchange account info: %w", err)
	}

	// 1. Reconcile Balances
	balanceDrifts := r.reconcileBalances(exchangeAccount.Balances)
	report.BalanceDrifts = balanceDrifts

	// 2. Reconcile Positions
	positionDrifts := r.reconcilePositions(exchangeAccount.Positions)
	report.PositionDrifts = positionDrifts

	// 3. Auto-correct if drifts detected
	if len(balanceDrifts) > 0 || len(positionDrifts) > 0 {
		if err := r.correctDrifts(balanceDrifts, positionDrifts); err != nil {
			report.Error = err
			return report, fmt.Errorf("failed to correct drifts: %w", err)
		}
		report.Corrected = true
	}

	return report, nil
}

// reconcileBalances checks for balance discrepancies
func (r *Reconciler) reconcileBalances(exchangeBalances map[string]ExchangeBalance) []BalanceDrift {
	drifts := []BalanceDrift{}
	localBalances := r.balanceMgr.GetAllBalances()

	// Check existing local balances for drifts
	for asset, localBal := range localBalances {
		exchangeBal, exists := exchangeBalances[asset]
		if !exists {
			continue
		}

		drift := exchangeBal.Total.Sub(localBal.Total)
		driftPct := decimal.Zero
		if !localBal.Total.IsZero() {
			driftPct = drift.Div(localBal.Total).Mul(decimal.NewFromInt(100))
		}

		if drift.Abs().GreaterThan(r.balanceTolerance) {
			drifts = append(drifts, BalanceDrift{
				Asset:           asset,
				LocalBalance:    localBal.Total,
				ExchangeBalance: exchangeBal.Total,
				Drift:           drift,
				DriftPercent:    driftPct,
			})

			metrics.ReconciliationDrift.Observe(drift.Abs().InexactFloat64())
		}
	}

	// Check for exchange balances that don't exist locally (initialization)
	for asset, exchangeBal := range exchangeBalances {
		if _, exists := localBalances[asset]; !exists && !exchangeBal.Total.IsZero() {
			// Treat missing local balance as a drift that needs correction
			drifts = append(drifts, BalanceDrift{
				Asset:           asset,
				LocalBalance:    decimal.Zero,
				ExchangeBalance: exchangeBal.Total,
				Drift:           exchangeBal.Total,
				DriftPercent:    decimal.NewFromInt(100),
			})
		}
	}

	return drifts
}

// reconcilePositions checks for position discrepancies
func (r *Reconciler) reconcilePositions(exchangePositions map[string]ExchangePosition) []PositionDrift {
	drifts := []PositionDrift{}
	localPositions := r.positionMgr.GetAllPositions()

	// Check existing local positions for drifts
	for symbol, localPos := range localPositions {
		exchangePos, exists := exchangePositions[symbol]
		if !exists {
			continue
		}

		drift := exchangePos.Quantity.Sub(localPos.Quantity)
		driftPct := decimal.Zero
		if !localPos.Quantity.IsZero() {
			driftPct = drift.Div(localPos.Quantity.Abs()).Mul(decimal.NewFromInt(100))
		}

		if drift.Abs().GreaterThan(r.positionTolerance) {
			drifts = append(drifts, PositionDrift{
				Symbol:           symbol,
				LocalQuantity:    localPos.Quantity,
				ExchangeQuantity: exchangePos.Quantity,
				Drift:            drift,
				DriftPercent:     driftPct,
			})

			metrics.ReconciliationDrift.Observe(drift.Abs().InexactFloat64())
		}
	}

	// Check for exchange positions that don't exist locally (initialization)
	for symbol, exchangePos := range exchangePositions {
		if _, exists := localPositions[symbol]; !exists && !exchangePos.Quantity.IsZero() {
			// Treat missing local position as a drift that needs correction
			drifts = append(drifts, PositionDrift{
				Symbol:           symbol,
				LocalQuantity:    decimal.Zero,
				ExchangeQuantity: exchangePos.Quantity,
				Drift:            exchangePos.Quantity,
				DriftPercent:     decimal.NewFromInt(100),
			})
		}
	}

	return drifts
}

// correctDrifts corrects detected drifts
func (r *Reconciler) correctDrifts(balanceDrifts []BalanceDrift, positionDrifts []PositionDrift) error {
	// Correct balance drifts
	for _, drift := range balanceDrifts {
		if err := r.balanceMgr.SetBalance(drift.Asset, drift.ExchangeBalance); err != nil {
			return err
		}

		r.logger.Info("Corrected balance drift",
			zap.String("asset", drift.Asset),
			zap.String("old_balance", drift.LocalBalance.String()),
			zap.String("new_balance", drift.ExchangeBalance.String()),
			zap.String("drift", drift.Drift.String()),
			zap.String("drift_pct", drift.DriftPercent.StringFixed(2)),
		)
	}

	// Correct position drifts
	for _, drift := range positionDrifts {
		if err := r.positionMgr.SetPosition(drift.Symbol, drift.ExchangeQuantity); err != nil {
			return err
		}

		r.logger.Info("Corrected position drift",
			zap.String("symbol", drift.Symbol),
			zap.String("old_qty", drift.LocalQuantity.String()),
			zap.String("new_qty", drift.ExchangeQuantity.String()),
			zap.String("drift", drift.Drift.String()),
			zap.String("drift_pct", drift.DriftPercent.StringFixed(2)),
		)
	}

	return nil
}
