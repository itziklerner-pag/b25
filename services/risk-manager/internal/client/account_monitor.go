package client

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/b25/services/risk-manager/internal/risk"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/yourorg/b25/services/account-monitor/pkg/proto"
)

// AccountMonitorClient wraps the gRPC client for Account Monitor service
type AccountMonitorClient struct {
	conn   *grpc.ClientConn
	client pb.AccountMonitorClient
	logger *zap.Logger
	userID string // The user/account ID to query
}

// NewAccountMonitorClient creates a new Account Monitor client
func NewAccountMonitorClient(address string, userID string, logger *zap.Logger) (*AccountMonitorClient, error) {
	// Create gRPC connection with keepalive and timeout options
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to account monitor at %s: %w", address, err)
	}

	client := pb.NewAccountMonitorClient(conn)

	logger.Info("account monitor client connected", zap.String("address", address))

	return &AccountMonitorClient{
		conn:   conn,
		client: client,
		logger: logger,
		userID: userID,
	}, nil
}

// Close closes the gRPC connection
func (c *AccountMonitorClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetAccountState retrieves the current account state from the Account Monitor
func (c *AccountMonitorClient) GetAccountState(ctx context.Context, accountID string) (risk.AccountState, error) {
	// Use accountID if provided, otherwise use the default userID
	userID := c.userID
	if accountID != "" {
		userID = accountID
	}

	// Call Account Monitor gRPC service
	req := &pb.AccountRequest{
		UserId: userID,
	}

	resp, err := c.client.GetAccountState(ctx, req)
	if err != nil {
		c.logger.Error("failed to get account state",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return risk.AccountState{}, fmt.Errorf("account monitor error: %w", err)
	}

	// Convert AccountMonitor protobuf response to risk.AccountState
	accountState, err := c.convertToAccountState(resp)
	if err != nil {
		c.logger.Error("failed to convert account state",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return risk.AccountState{}, fmt.Errorf("failed to convert account state: %w", err)
	}

	return accountState, nil
}

// convertToAccountState converts Account Monitor protobuf to risk.AccountState
func (c *AccountMonitorClient) convertToAccountState(resp *pb.AccountState) (risk.AccountState, error) {
	state := risk.AccountState{
		Positions:     []risk.Position{},
		PendingOrders: []risk.Order{},
	}

	// Calculate total balance and equity from all asset balances
	var totalBalance float64
	var totalEquity float64
	var availableMargin float64

	for _, bal := range resp.Balances {
		// Parse USD values
		usdValue, err := strconv.ParseFloat(bal.UsdValue, 64)
		if err != nil {
			c.logger.Warn("failed to parse balance USD value",
				zap.String("asset", bal.Asset),
				zap.String("usd_value", bal.UsdValue),
				zap.Error(err),
			)
			continue
		}

		// Parse free balance for available margin calculation
		freeBalance, err := strconv.ParseFloat(bal.Free, 64)
		if err == nil {
			freeUSD, _ := strconv.ParseFloat(bal.UsdValue, 64)
			// Approximate available margin as free balance in USD
			if bal.Asset == "USDT" || bal.Asset == "USDC" || bal.Asset == "USD" {
				availableMargin += freeBalance
			} else {
				// For other assets, use proportion of free/total * usd_value
				total, _ := strconv.ParseFloat(bal.Total, 64)
				if total > 0 {
					availableMargin += (freeBalance / total) * freeUSD
				}
			}
		}

		totalEquity += usdValue
		totalBalance += usdValue
	}

	state.Equity = totalEquity
	state.Balance = totalBalance
	state.AvailableMargin = availableMargin

	// Convert PnL data
	if resp.Pnl != nil {
		realizedPnL, _ := strconv.ParseFloat(resp.Pnl.RealizedPnl, 64)
		unrealizedPnL, _ := strconv.ParseFloat(resp.Pnl.UnrealizedPnl, 64)

		state.UnrealizedPnL = unrealizedPnL

		// Calculate margin used (equity - available margin)
		state.MarginUsed = totalEquity - availableMargin

		// Set peak equity (using current equity for now, should be tracked separately)
		state.PeakEquity = totalEquity

		// DailyStartEquity = current equity - (realized + unrealized PnL today)
		state.DailyStartEquity = totalEquity - (realizedPnL + unrealizedPnL)
	}

	// Convert positions
	for _, pos := range resp.Positions {
		quantity, err := strconv.ParseFloat(pos.Quantity, 64)
		if err != nil {
			c.logger.Warn("failed to parse position quantity",
				zap.String("symbol", pos.Symbol),
				zap.Error(err),
			)
			continue
		}

		entryPrice, _ := strconv.ParseFloat(pos.EntryPrice, 64)
		currentPrice, _ := strconv.ParseFloat(pos.CurrentPrice, 64)
		unrealizedPnL, _ := strconv.ParseFloat(pos.UnrealizedPnl, 64)
		realizedPnL, _ := strconv.ParseFloat(pos.RealizedPnl, 64)

		position := risk.Position{
			Symbol:        pos.Symbol,
			Quantity:      quantity,
			Side:          determineSide(quantity),
			EntryPrice:    entryPrice,
			CurrentPrice:  currentPrice,
			UnrealizedPnL: unrealizedPnL,
			RealizedPnL:   realizedPnL,
			MarginUsed:    calculatePositionMargin(quantity, entryPrice),
		}

		state.Positions = append(state.Positions, position)
	}

	return state, nil
}

// determineSide determines if position is long or short based on quantity
func determineSide(quantity float64) string {
	if quantity > 0 {
		return "buy"
	}
	return "sell"
}

// calculatePositionMargin calculates margin used for a position (simplified)
func calculatePositionMargin(quantity, entryPrice float64) float64 {
	// Simplified: margin = notional / leverage
	// Assuming 10x leverage for now
	notional := abs(quantity) * entryPrice
	return notional / 10.0
}

// abs returns absolute value of float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
