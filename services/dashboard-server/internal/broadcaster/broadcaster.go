package broadcaster

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/yourusername/b25/services/dashboard-server/internal/aggregator"
	"github.com/yourusername/b25/services/dashboard-server/internal/metrics"
	"github.com/yourusername/b25/services/dashboard-server/internal/types"
)

// ClientInfo stores information about a connected client
type ClientInfo struct {
	ID            string
	Type          types.ClientType
	SendChan      chan []byte
	Format        types.SerializationFormat
	Subscriptions map[string]bool
	LastState     *types.State
}

// Broadcaster manages broadcasting state updates to clients
type Broadcaster struct {
	clients     map[string]*ClientInfo
	clientsMu   sync.RWMutex
	aggregator  *aggregator.Aggregator
	logger      zerolog.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	tuiSequence uint64
	webSequence uint64
}

// NewBroadcaster creates a new broadcaster
func NewBroadcaster(logger zerolog.Logger, aggregator *aggregator.Aggregator) *Broadcaster {
	ctx, cancel := context.WithCancel(context.Background())
	return &Broadcaster{
		clients:    make(map[string]*ClientInfo),
		aggregator: aggregator,
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start begins the broadcasting process
func (b *Broadcaster) Start() {
	b.logger.Info().Msg("Broadcaster started")
	go b.tuiBroadcaster()
	go b.webBroadcaster()
}

// Stop stops the broadcasting process
func (b *Broadcaster) Stop() {
	b.cancel()
	b.logger.Info().Msg("Broadcaster stopped")
}

// RegisterClient registers a new client for broadcasting
func (b *Broadcaster) RegisterClient(id string, clientType types.ClientType, sendChan chan []byte, format types.SerializationFormat) {
	b.clientsMu.Lock()
	defer b.clientsMu.Unlock()

	b.clients[id] = &ClientInfo{
		ID:            id,
		Type:          clientType,
		SendChan:      sendChan,
		Format:        format,
		Subscriptions: make(map[string]bool),
	}

	b.logger.Info().
		Str("client_id", id).
		Str("type", clientType.String()).
		Str("format", format.String()).
		Msg("Client registered with broadcaster")
}

// UnregisterClient removes a client from broadcasting
func (b *Broadcaster) UnregisterClient(id string) {
	b.clientsMu.Lock()
	defer b.clientsMu.Unlock()

	if client, exists := b.clients[id]; exists {
		// Decrement subscription metrics
		for channel := range client.Subscriptions {
			metrics.DecrementClientSubscriptions(channel)
		}
		delete(b.clients, id)
	}

	b.logger.Debug().
		Str("client_id", id).
		Msg("Client unregistered from broadcaster")
}

// UpdateSubscriptions updates a client's subscriptions
func (b *Broadcaster) UpdateSubscriptions(id string, subscriptions map[string]bool) {
	b.clientsMu.Lock()
	defer b.clientsMu.Unlock()

	if client, exists := b.clients[id]; exists {
		client.Subscriptions = make(map[string]bool)
		for k, v := range subscriptions {
			client.Subscriptions[k] = v
		}
	}
}

// tuiBroadcaster broadcasts updates to TUI clients every 100ms
func (b *Broadcaster) tuiBroadcaster() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			b.tuiSequence++
			b.broadcastToClients(types.ClientTypeTUI, b.tuiSequence)
		}
	}
}

// webBroadcaster broadcasts updates to Web clients every 250ms
func (b *Broadcaster) webBroadcaster() {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			b.webSequence++
			b.broadcastToClients(types.ClientTypeWeb, b.webSequence)
		}
	}
}

// broadcastToClients broadcasts state updates to clients of a specific type
func (b *Broadcaster) broadcastToClients(clientType types.ClientType, sequence uint64) {
	start := time.Now()

	currentState := b.aggregator.GetFullState()
	currentState.Sequence = sequence

	b.clientsMu.RLock()
	defer b.clientsMu.RUnlock()

	notifiedCount := 0
	clientCount := 0
	updatesSent := 0
	snapshotsSent := 0
	skippedNoChange := 0

	for _, client := range b.clients {
		if client.Type != clientType {
			continue
		}
		clientCount++

		// Filter state based on subscriptions (or send all if no subscriptions)
		filteredState := b.filterStateBySubscriptions(currentState, client.Subscriptions)

		// Generate differential update if possible
		var message []byte
		var err error
		var messageType string
		var diffCount int

		if client.LastState != nil {
			// FIXED: Always compute diff and check if there are actual changes
			diff := b.computeDiff(client.LastState, filteredState)
			diffCount = len(diff)
			if diffCount > 0 {
				// Send incremental update with changes
				msg := types.ServerMessage{
					Type:      "snapshot",  // Always send full snapshot
					Sequence:  sequence,
					Timestamp: time.Now(),
					Data: filteredState,  // Send full state instead of diff
				}
				message, err = b.serializeMessage(client.Format, msg)
				messageType = "update"
				if err == nil {
					metrics.RecordMessageSent(client.Type.String(), "update")
					metrics.RecordMessageSize(client.Format.String(), "update", len(message))
					updatesSent++
				}
			} else {
				// No changes detected, skip this client
				skippedNoChange++
				continue
			}
		} else {
			// Send full snapshot for first update
			msg := types.ServerMessage{
				Type:      "snapshot",
				Sequence:  sequence,
				Timestamp: time.Now(),
				Data:      filteredState,
			}
			message, err = b.serializeMessage(client.Format, msg)
			messageType = "snapshot"
			if err == nil {
				metrics.RecordMessageSent(client.Type.String(), "snapshot")
				metrics.RecordMessageSize(client.Format.String(), "snapshot", len(message))
				snapshotsSent++
			}
		}

		if err != nil {
			b.logger.Error().Err(err).Str("client_id", client.ID).Msg("Failed to serialize message")
			continue
		}

		if message != nil {
			select {
			case client.SendChan <- message:
				client.LastState = filteredState
				notifiedCount++

				// Log details for debugging
				if messageType == "update" {
					b.logger.Debug().
						Str("client_id", client.ID).
						Str("client_type", clientType.String()).
						Uint64("sequence", sequence).
						Int("changes", diffCount).
						Int("message_bytes", len(message)).
						Msg("Sent incremental update to client")
				}
			default:
				b.logger.Warn().Str("client_id", client.ID).Msg("Client send buffer full, dropping message")
			}
		}
	}

	duration := time.Since(start)
	metrics.RecordBroadcastLatency(clientType.String(), duration.Seconds())

	// ENHANCED LOGGING: Always log broadcast attempts to help debugging
	if clientCount > 0 {
		b.logger.Info().
			Str("client_type", clientType.String()).
			Uint64("sequence", sequence).
			Int("clients_total", clientCount).
			Int("clients_notified", notifiedCount).
			Int("updates_sent", updatesSent).
			Int("snapshots_sent", snapshotsSent).
			Int("skipped_no_change", skippedNoChange).
			Dur("duration", duration).
			Msg("Broadcasting to clients")
	}
}

// filterStateBySubscriptions filters state based on client subscriptions
func (b *Broadcaster) filterStateBySubscriptions(state *types.State, subscriptions map[string]bool) *types.State {
	filtered := &types.State{
		Timestamp: state.Timestamp,
		Sequence:  state.Sequence,
		// Initialize all maps to prevent null in JSON serialization
		MarketData: make(map[string]*types.MarketData),
		Orders:     make([]*types.Order, 0),
		Positions:  make(map[string]*types.Position),
		Strategies: make(map[string]*types.Strategy),
	}

	// If no subscriptions are set, send ALL data (opt-out instead of opt-in)
	sendAll := len(subscriptions) == 0

	// Copy data based on subscriptions or send all if no subscriptions
	if (sendAll || subscriptions["market_data"]) && state.MarketData != nil {
		filtered.MarketData = state.MarketData
	}
	if (sendAll || subscriptions["orders"]) && state.Orders != nil {
		filtered.Orders = state.Orders
	}
	if (sendAll || subscriptions["positions"]) && state.Positions != nil {
		filtered.Positions = state.Positions
	}
	if (sendAll || subscriptions["account"]) && state.Account != nil {
		filtered.Account = state.Account
	}
	if (sendAll || subscriptions["strategies"]) && state.Strategies != nil {
		filtered.Strategies = state.Strategies
	}

	return filtered
}

// computeDiff computes the differences between two states
func (b *Broadcaster) computeDiff(oldState, newState *types.State) map[string]interface{} {
	diff := make(map[string]interface{})

	// Compare market data
	if oldState.MarketData != nil && newState.MarketData != nil {
		for symbol, newMD := range newState.MarketData {
			oldMD, exists := oldState.MarketData[symbol]
			if !exists || oldMD.LastPrice != newMD.LastPrice {
				diff[fmt.Sprintf("market_data.%s.last_price", symbol)] = newMD.LastPrice
			}
			if !exists || oldMD.BidPrice != newMD.BidPrice {
				diff[fmt.Sprintf("market_data.%s.bid_price", symbol)] = newMD.BidPrice
			}
			if !exists || oldMD.AskPrice != newMD.AskPrice {
				diff[fmt.Sprintf("market_data.%s.ask_price", symbol)] = newMD.AskPrice
			}
			if !exists || oldMD.Volume24h != newMD.Volume24h {
				diff[fmt.Sprintf("market_data.%s.volume_24h", symbol)] = newMD.Volume24h
			}
		}
	}

	// Compare positions
	if oldState.Positions != nil && newState.Positions != nil {
		for symbol, newPos := range newState.Positions {
			oldPos, exists := oldState.Positions[symbol]
			if !exists || oldPos.UnrealizedPnL != newPos.UnrealizedPnL {
				diff[fmt.Sprintf("positions.%s.unrealized_pnl", symbol)] = newPos.UnrealizedPnL
			}
			if !exists || oldPos.MarkPrice != newPos.MarkPrice {
				diff[fmt.Sprintf("positions.%s.mark_price", symbol)] = newPos.MarkPrice
			}
			if !exists || oldPos.Quantity != newPos.Quantity {
				diff[fmt.Sprintf("positions.%s.quantity", symbol)] = newPos.Quantity
			}
		}
	}

	// Compare account
	if oldState.Account != nil && newState.Account != nil {
		if oldState.Account.TotalBalance != newState.Account.TotalBalance {
			diff["account.total_balance"] = newState.Account.TotalBalance
		}
		if oldState.Account.AvailableBalance != newState.Account.AvailableBalance {
			diff["account.available_balance"] = newState.Account.AvailableBalance
		}
		if oldState.Account.UnrealizedPnL != newState.Account.UnrealizedPnL {
			diff["account.unrealized_pnl"] = newState.Account.UnrealizedPnL
		}
		if oldState.Account.MarginUsed != newState.Account.MarginUsed {
			diff["account.margin_used"] = newState.Account.MarginUsed
		}
	}

	// Compare orders count (simplified)
	if len(oldState.Orders) != len(newState.Orders) {
		diff["orders.count"] = len(newState.Orders)
		// Also send the full orders array when count changes
		diff["orders"] = newState.Orders
	}

	// Compare strategies
	if oldState.Strategies != nil && newState.Strategies != nil {
		for id, newStrat := range newState.Strategies {
			oldStrat, exists := oldState.Strategies[id]
			if !exists || oldStrat.PnL != newStrat.PnL {
				diff[fmt.Sprintf("strategies.%s.pnl", id)] = newStrat.PnL
			}
			if !exists || oldStrat.Status != newStrat.Status {
				diff[fmt.Sprintf("strategies.%s.status", id)] = newStrat.Status
			}
		}
	}

	return diff
}

// serializeMessage serializes a server message
func (b *Broadcaster) serializeMessage(format types.SerializationFormat, msg types.ServerMessage) ([]byte, error) {
	start := time.Now()
	defer func() {
		metrics.RecordSerializationDuration(format.String(), time.Since(start).Seconds())
	}()

	switch format {
	case types.FormatMessagePack:
		return msgpack.Marshal(msg)
	case types.FormatJSON:
		return json.Marshal(msg)
	default:
		return nil, fmt.Errorf("unsupported serialization format: %v", format)
	}
}
