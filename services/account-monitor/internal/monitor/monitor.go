package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/yourorg/b25/services/account-monitor/internal/alert"
	"github.com/yourorg/b25/services/account-monitor/internal/balance"
	"github.com/yourorg/b25/services/account-monitor/internal/calculator"
	"github.com/yourorg/b25/services/account-monitor/internal/exchange"
	"github.com/yourorg/b25/services/account-monitor/internal/position"
	"github.com/yourorg/b25/services/account-monitor/internal/reconciliation"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Configure properly for production
	},
}

type AccountMonitor struct {
	positionMgr *position.Manager
	balanceMgr  *balance.Manager
	pnlCalc     *calculator.PnLCalculator
	reconciler  *reconciliation.Reconciler
	alertMgr    *alert.Manager
	wsClient    *exchange.WebSocketClient
	natsConn    *nats.Conn
	logger      *zap.Logger

	priceMap   map[string]decimal.Decimal
	priceMutex sync.RWMutex
}

func NewAccountMonitor(
	positionMgr *position.Manager,
	balanceMgr *balance.Manager,
	pnlCalc *calculator.PnLCalculator,
	reconciler *reconciliation.Reconciler,
	alertMgr *alert.Manager,
	wsClient *exchange.WebSocketClient,
	natsConn *nats.Conn,
	logger *zap.Logger,
) *AccountMonitor {
	return &AccountMonitor{
		positionMgr: positionMgr,
		balanceMgr:  balanceMgr,
		pnlCalc:     pnlCalc,
		reconciler:  reconciler,
		alertMgr:    alertMgr,
		wsClient:    wsClient,
		natsConn:    natsConn,
		logger:      logger,
		priceMap:    make(map[string]decimal.Decimal),
	}
}

// Start initializes and starts all monitoring components
func (am *AccountMonitor) Start(ctx context.Context) error {
	var wg sync.WaitGroup

	// Restore state from Redis
	am.logger.Info("Restoring state from Redis")
	if err := am.positionMgr.RestoreFromRedis(ctx); err != nil {
		am.logger.Warn("Failed to restore positions from Redis", zap.Error(err))
	}
	if err := am.balanceMgr.RestoreFromRedis(ctx); err != nil {
		am.logger.Warn("Failed to restore balances from Redis", zap.Error(err))
	}

	// 1. Subscribe to fill events from NATS
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := am.subscribeFillEvents(ctx); err != nil {
			am.logger.Error("Fill events subscription failed", zap.Error(err))
		}
	}()

	// 2. Start WebSocket client for user data stream
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := am.wsClient.Start(ctx, am.handleUserDataEvent); err != nil {
			am.logger.Error("WebSocket client failed", zap.Error(err))
		}
	}()

	// 3. Start reconciliation
	wg.Add(1)
	go func() {
		defer wg.Done()
		am.reconciler.Start(ctx)
	}()

	// 4. Start periodic P&L snapshot
	wg.Add(1)
	go func() {
		defer wg.Done()
		am.periodicPnLSnapshot(ctx)
	}()

	// 5. Start alert monitoring
	wg.Add(1)
	go func() {
		defer wg.Done()
		am.alertMgr.Start(ctx)
	}()

	wg.Wait()
	return nil
}

// subscribeFillEvents subscribes to fill events from NATS
func (am *AccountMonitor) subscribeFillEvents(ctx context.Context) error {
	sub, err := am.natsConn.Subscribe("trading.fills", func(msg *nats.Msg) {
		var fill position.Fill
		if err := json.Unmarshal(msg.Data, &fill); err != nil {
			am.logger.Error("Failed to unmarshal fill event", zap.Error(err))
			return
		}

		am.logger.Info("Received fill event",
			zap.String("symbol", fill.Symbol),
			zap.String("side", fill.Side),
			zap.String("quantity", fill.Quantity.String()),
			zap.String("price", fill.Price.String()),
		)

		// Update position
		if err := am.positionMgr.UpdatePosition(fill); err != nil {
			am.logger.Error("Failed to update position", zap.Error(err))
			return
		}

		// Store current price
		am.priceMutex.Lock()
		am.priceMap[fill.Symbol] = fill.Price
		am.priceMutex.Unlock()
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to fill events: %w", err)
	}

	am.logger.Info("Subscribed to fill events on NATS")

	<-ctx.Done()
	sub.Unsubscribe()
	return nil
}

// handleUserDataEvent processes WebSocket user data events
func (am *AccountMonitor) handleUserDataEvent(event exchange.UserDataEvent) {
	switch event.Type {
	case "balanceUpdate":
		am.handleBalanceUpdate(event)
	case "executionReport":
		am.handleExecutionReport(event)
	case "outboundAccountPosition":
		am.handleAccountPosition(event)
	default:
		am.logger.Debug("Unknown user data event type", zap.String("type", event.Type))
	}
}

// handleBalanceUpdate processes balance update events
func (am *AccountMonitor) handleBalanceUpdate(event exchange.UserDataEvent) {
	if balanceData, ok := event.Data.(map[string]interface{}); ok {
		asset := balanceData["asset"].(string)
		free, _ := decimal.NewFromString(balanceData["free"].(string))
		locked, _ := decimal.NewFromString(balanceData["locked"].(string))

		if err := am.balanceMgr.UpdateBalance(asset, free, locked); err != nil {
			am.logger.Error("Failed to update balance", zap.Error(err))
		}
	}
}

// handleExecutionReport processes execution report (fill) events
func (am *AccountMonitor) handleExecutionReport(event exchange.UserDataEvent) {
	// Already handled via NATS fills, this is backup/validation
	am.logger.Debug("Execution report received via WebSocket")
}

// handleAccountPosition processes account position update events
func (am *AccountMonitor) handleAccountPosition(event exchange.UserDataEvent) {
	am.logger.Debug("Account position update received")
}

// periodicPnLSnapshot stores P&L snapshots periodically
func (am *AccountMonitor) periodicPnLSnapshot(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			am.priceMutex.RLock()
			priceMap := make(map[string]decimal.Decimal)
			for k, v := range am.priceMap {
				priceMap[k] = v
			}
			am.priceMutex.RUnlock()

			report, err := am.pnlCalc.GetCurrentPnL(ctx, priceMap)
			if err != nil {
				am.logger.Error("Failed to calculate P&L", zap.Error(err))
				continue
			}

			if err := am.pnlCalc.StorePnLSnapshot(ctx, report); err != nil {
				am.logger.Error("Failed to store P&L snapshot", zap.Error(err))
			}

		case <-ctx.Done():
			return
		}
	}
}

// HTTP Handlers

// HandleAccountState returns current account state
func (am *AccountMonitor) HandleAccountState(w http.ResponseWriter, r *http.Request) {
	am.priceMutex.RLock()
	priceMap := make(map[string]decimal.Decimal)
	for k, v := range am.priceMap {
		priceMap[k] = v
	}
	am.priceMutex.RUnlock()

	report, err := am.pnlCalc.GetCurrentPnL(r.Context(), priceMap)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	balances := am.balanceMgr.GetAllBalances()
	positions := am.positionMgr.GetAllPositions()

	response := map[string]interface{}{
		"pnl":       report,
		"balances":  balances,
		"positions": positions,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandlePositions returns all positions
func (am *AccountMonitor) HandlePositions(w http.ResponseWriter, r *http.Request) {
	positions := am.positionMgr.GetAllPositions()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(positions)
}

// HandlePnL returns P&L report
func (am *AccountMonitor) HandlePnL(w http.ResponseWriter, r *http.Request) {
	am.priceMutex.RLock()
	priceMap := make(map[string]decimal.Decimal)
	for k, v := range am.priceMap {
		priceMap[k] = v
	}
	am.priceMutex.RUnlock()

	report, err := am.pnlCalc.GetCurrentPnL(r.Context(), priceMap)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// HandleBalance returns balances
func (am *AccountMonitor) HandleBalance(w http.ResponseWriter, r *http.Request) {
	balances := am.balanceMgr.GetAllBalances()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balances)
}

// HandleAlerts returns recent alerts
func (am *AccountMonitor) HandleAlerts(w http.ResponseWriter, r *http.Request) {
	alerts, err := am.alertMgr.GetRecentAlerts(r.Context(), 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

// HandleWebSocket handles WebSocket connections for real-time updates
func (am *AccountMonitor) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		am.logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			am.priceMutex.RLock()
			priceMap := make(map[string]decimal.Decimal)
			for k, v := range am.priceMap {
				priceMap[k] = v
			}
			am.priceMutex.RUnlock()

			report, err := am.pnlCalc.GetCurrentPnL(r.Context(), priceMap)
			if err != nil {
				continue
			}

			data := map[string]interface{}{
				"type":      "pnl_update",
				"timestamp": time.Now(),
				"data":      report,
			}

			if err := conn.WriteJSON(data); err != nil {
				return
			}
		}
	}
}
