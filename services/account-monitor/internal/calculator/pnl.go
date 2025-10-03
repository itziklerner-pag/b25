package calculator

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/yourorg/b25/services/account-monitor/internal/balance"
	"github.com/yourorg/b25/services/account-monitor/internal/position"
)

type PnLCalculator struct {
	positionMgr *position.Manager
	balanceMgr  *balance.Manager
	db          *pgxpool.Pool
	logger      *zap.Logger
}

type PnLReport struct {
	Timestamp     time.Time       `json:"timestamp"`
	RealizedPnL   decimal.Decimal `json:"realized_pnl"`
	UnrealizedPnL decimal.Decimal `json:"unrealized_pnl"`
	TotalPnL      decimal.Decimal `json:"total_pnl"`
	TotalFees     decimal.Decimal `json:"total_fees"`
	NetPnL        decimal.Decimal `json:"net_pnl"`
	WinRate       decimal.Decimal `json:"win_rate"`
	TotalTrades   int             `json:"total_trades"`
	WinningTrades int             `json:"winning_trades"`
	LosingTrades  int             `json:"losing_trades"`
	AverageWin    decimal.Decimal `json:"average_win"`
	AverageLoss   decimal.Decimal `json:"average_loss"`
	ProfitFactor  decimal.Decimal `json:"profit_factor"`
}

type PnLSnapshot struct {
	Timestamp     time.Time       `json:"timestamp"`
	Symbol        string          `json:"symbol"`
	RealizedPnL   decimal.Decimal `json:"realized_pnl"`
	UnrealizedPnL decimal.Decimal `json:"unrealized_pnl"`
	TotalPnL      decimal.Decimal `json:"total_pnl"`
	TotalFees     decimal.Decimal `json:"total_fees"`
	Equity        decimal.Decimal `json:"equity"`
}

type TradeStatistics struct {
	WinRate       decimal.Decimal
	TotalTrades   int
	WinningTrades int
	LosingTrades  int
	AverageWin    decimal.Decimal
	AverageLoss   decimal.Decimal
	ProfitFactor  decimal.Decimal
}

type PriceProvider interface {
	GetPrice(symbol string) (decimal.Decimal, error)
}

func NewPnLCalculator(positionMgr *position.Manager, balanceMgr *balance.Manager, db *pgxpool.Pool, logger *zap.Logger) *PnLCalculator {
	return &PnLCalculator{
		positionMgr: positionMgr,
		balanceMgr:  balanceMgr,
		db:          db,
		logger:      logger,
	}
}

// GetCurrentPnL calculates current P&L across all positions
func (p *PnLCalculator) GetCurrentPnL(ctx context.Context, priceMap map[string]decimal.Decimal) (*PnLReport, error) {
	positions := p.positionMgr.GetAllPositions()

	var totalRealizedPnL decimal.Decimal
	var totalUnrealizedPnL decimal.Decimal
	var totalFees decimal.Decimal

	for symbol, pos := range positions {
		// Get current price
		currentPrice, ok := priceMap[symbol]
		if !ok {
			p.logger.Warn("Price not available for symbol", zap.String("symbol", symbol))
			continue
		}

		// Calculate unrealized P&L
		unrealizedPnL, err := p.positionMgr.CalculateUnrealizedPnL(symbol, currentPrice)
		if err != nil {
			return nil, err
		}

		totalRealizedPnL = totalRealizedPnL.Add(pos.RealizedPnL)
		totalUnrealizedPnL = totalUnrealizedPnL.Add(unrealizedPnL)
		totalFees = totalFees.Add(pos.TotalFees)
	}

	totalPnL := totalRealizedPnL.Add(totalUnrealizedPnL)
	netPnL := totalPnL.Sub(totalFees)

	// Calculate statistics
	stats := p.calculateStatistics(positions)

	report := &PnLReport{
		Timestamp:     time.Now(),
		RealizedPnL:   totalRealizedPnL,
		UnrealizedPnL: totalUnrealizedPnL,
		TotalPnL:      totalPnL,
		TotalFees:     totalFees,
		NetPnL:        netPnL,
		WinRate:       stats.WinRate,
		TotalTrades:   stats.TotalTrades,
		WinningTrades: stats.WinningTrades,
		LosingTrades:  stats.LosingTrades,
		AverageWin:    stats.AverageWin,
		AverageLoss:   stats.AverageLoss,
		ProfitFactor:  stats.ProfitFactor,
	}

	return report, nil
}

// calculateStatistics calculates trade statistics
func (p *PnLCalculator) calculateStatistics(positions map[string]*position.Position) TradeStatistics {
	var totalWins decimal.Decimal
	var totalLosses decimal.Decimal
	winCount := 0
	lossCount := 0

	// Count closed trades based on realized P&L
	for _, pos := range positions {
		if pos.RealizedPnL.IsZero() {
			continue
		}

		if pos.RealizedPnL.IsPositive() {
			totalWins = totalWins.Add(pos.RealizedPnL)
			winCount++
		} else {
			totalLosses = totalLosses.Add(pos.RealizedPnL.Abs())
			lossCount++
		}
	}

	totalTrades := winCount + lossCount
	var winRate decimal.Decimal
	if totalTrades > 0 {
		winRate = decimal.NewFromInt(int64(winCount)).Div(decimal.NewFromInt(int64(totalTrades))).Mul(decimal.NewFromInt(100))
	}

	var avgWin decimal.Decimal
	if winCount > 0 {
		avgWin = totalWins.Div(decimal.NewFromInt(int64(winCount)))
	}

	var avgLoss decimal.Decimal
	if lossCount > 0 {
		avgLoss = totalLosses.Div(decimal.NewFromInt(int64(lossCount)))
	}

	var profitFactor decimal.Decimal
	if !totalLosses.IsZero() {
		profitFactor = totalWins.Div(totalLosses)
	}

	return TradeStatistics{
		WinRate:       winRate,
		TotalTrades:   totalTrades,
		WinningTrades: winCount,
		LosingTrades:  lossCount,
		AverageWin:    avgWin,
		AverageLoss:   avgLoss,
		ProfitFactor:  profitFactor,
	}
}

// StorePnLSnapshot stores P&L snapshot to TimescaleDB
func (p *PnLCalculator) StorePnLSnapshot(ctx context.Context, report *PnLReport) error {
	query := `
		INSERT INTO pnl_snapshots (
			timestamp, realized_pnl, unrealized_pnl, total_pnl,
			total_fees, net_pnl, win_rate, total_trades
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := p.db.Exec(ctx, query,
		report.Timestamp,
		report.RealizedPnL,
		report.UnrealizedPnL,
		report.TotalPnL,
		report.TotalFees,
		report.NetPnL,
		report.WinRate,
		report.TotalTrades,
	)

	if err != nil {
		return fmt.Errorf("failed to store P&L snapshot: %w", err)
	}

	p.logger.Debug("P&L snapshot stored",
		zap.Time("timestamp", report.Timestamp),
		zap.String("total_pnl", report.TotalPnL.String()),
	)

	return nil
}

// GetPnLHistory retrieves P&L history from TimescaleDB
func (p *PnLCalculator) GetPnLHistory(ctx context.Context, from, to time.Time, interval string) ([]*PnLReport, error) {
	query := `
		SELECT
			time_bucket($1, timestamp) as bucket,
			AVG(realized_pnl) as avg_realized,
			AVG(unrealized_pnl) as avg_unrealized,
			AVG(total_pnl) as avg_total,
			SUM(total_fees) as sum_fees,
			AVG(win_rate) as avg_win_rate,
			MAX(total_trades) as total_trades
		FROM pnl_snapshots
		WHERE timestamp >= $2 AND timestamp <= $3
		GROUP BY bucket
		ORDER BY bucket DESC
	`

	rows, err := p.db.Query(ctx, query, interval, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query P&L history: %w", err)
	}
	defer rows.Close()

	reports := []*PnLReport{}
	for rows.Next() {
		var r PnLReport
		err := rows.Scan(
			&r.Timestamp,
			&r.RealizedPnL,
			&r.UnrealizedPnL,
			&r.TotalPnL,
			&r.TotalFees,
			&r.WinRate,
			&r.TotalTrades,
		)
		if err != nil {
			return nil, err
		}
		r.NetPnL = r.TotalPnL.Sub(r.TotalFees)
		reports = append(reports, &r)
	}

	return reports, nil
}

// GetSymbolPnL gets P&L for a specific symbol
func (p *PnLCalculator) GetSymbolPnL(symbol string, currentPrice decimal.Decimal) (*PnLSnapshot, error) {
	pos, err := p.positionMgr.GetPosition(symbol)
	if err != nil {
		return nil, err
	}

	unrealizedPnL, err := p.positionMgr.CalculateUnrealizedPnL(symbol, currentPrice)
	if err != nil {
		return nil, err
	}

	snapshot := &PnLSnapshot{
		Timestamp:     time.Now(),
		Symbol:        symbol,
		RealizedPnL:   pos.RealizedPnL,
		UnrealizedPnL: unrealizedPnL,
		TotalPnL:      pos.RealizedPnL.Add(unrealizedPnL),
		TotalFees:     pos.TotalFees,
	}

	return snapshot, nil
}
