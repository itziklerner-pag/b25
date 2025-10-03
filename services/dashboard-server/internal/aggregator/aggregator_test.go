package aggregator

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/yourusername/b25/services/dashboard-server/internal/types"
)

func TestNewAggregator(t *testing.T) {
	logger := zerolog.Nop()
	agg := NewAggregator(logger, "localhost:6379")

	assert.NotNil(t, agg)
	assert.NotNil(t, agg.marketData)
	assert.NotNil(t, agg.orders)
	assert.NotNil(t, agg.positions)
	assert.NotNil(t, agg.account)
	assert.NotNil(t, agg.strategies)
}

func TestUpdateMarketData(t *testing.T) {
	logger := zerolog.Nop()
	agg := NewAggregator(logger, "localhost:6379")

	md := &types.MarketData{
		Symbol:    "BTCUSDT",
		LastPrice: 50000.0,
		BidPrice:  49999.0,
		AskPrice:  50001.0,
		Volume24h: 1000000.0,
	}

	agg.UpdateMarketData("BTCUSDT", md)

	state := agg.GetFullState()
	assert.Equal(t, 1, len(state.MarketData))
	assert.Equal(t, "BTCUSDT", state.MarketData["BTCUSDT"].Symbol)
	assert.Equal(t, 50000.0, state.MarketData["BTCUSDT"].LastPrice)
}

func TestUpdateOrder(t *testing.T) {
	logger := zerolog.Nop()
	agg := NewAggregator(logger, "localhost:6379")

	order := &types.Order{
		ID:       "order-1",
		Symbol:   "BTCUSDT",
		Side:     "BUY",
		Type:     "LIMIT",
		Price:    50000.0,
		Quantity: 0.5,
		Status:   "NEW",
	}

	agg.UpdateOrder(order)

	state := agg.GetFullState()
	assert.Equal(t, 1, len(state.Orders))
	assert.Equal(t, "order-1", state.Orders[0].ID)

	// Update same order
	order.Status = "FILLED"
	agg.UpdateOrder(order)

	state = agg.GetFullState()
	assert.Equal(t, 1, len(state.Orders))
	assert.Equal(t, "FILLED", state.Orders[0].Status)
}

func TestUpdatePosition(t *testing.T) {
	logger := zerolog.Nop()
	agg := NewAggregator(logger, "localhost:6379")

	position := &types.Position{
		Symbol:        "BTCUSDT",
		Side:          "LONG",
		Quantity:      0.5,
		EntryPrice:    50000.0,
		MarkPrice:     50500.0,
		UnrealizedPnL: 250.0,
	}

	agg.UpdatePosition("BTCUSDT", position)

	state := agg.GetFullState()
	assert.Equal(t, 1, len(state.Positions))
	assert.Equal(t, "BTCUSDT", state.Positions["BTCUSDT"].Symbol)
	assert.Equal(t, 250.0, state.Positions["BTCUSDT"].UnrealizedPnL)
}

func TestUpdateAccount(t *testing.T) {
	logger := zerolog.Nop()
	agg := NewAggregator(logger, "localhost:6379")

	account := &types.Account{
		TotalBalance:     10000.0,
		AvailableBalance: 8000.0,
		MarginUsed:       2000.0,
		UnrealizedPnL:    150.0,
		Balances: map[string]float64{
			"USDT": 10000.0,
			"BTC":  0.1,
		},
	}

	agg.UpdateAccount(account)

	state := agg.GetFullState()
	assert.NotNil(t, state.Account)
	assert.Equal(t, 10000.0, state.Account.TotalBalance)
	assert.Equal(t, 2, len(state.Account.Balances))
}

func TestUpdateStrategy(t *testing.T) {
	logger := zerolog.Nop()
	agg := NewAggregator(logger, "localhost:6379")

	strategy := &types.Strategy{
		ID:      "strat-1",
		Name:    "MeanReversion",
		Status:  "ACTIVE",
		PnL:     150.0,
		Trades:  10,
		WinRate: 0.7,
	}

	agg.UpdateStrategy("strat-1", strategy)

	state := agg.GetFullState()
	assert.Equal(t, 1, len(state.Strategies))
	assert.Equal(t, "strat-1", state.Strategies["strat-1"].ID)
	assert.Equal(t, "MeanReversion", state.Strategies["strat-1"].Name)
}

func TestGetFullState(t *testing.T) {
	logger := zerolog.Nop()
	agg := NewAggregator(logger, "localhost:6379")

	// Add various data
	agg.UpdateMarketData("BTCUSDT", &types.MarketData{Symbol: "BTCUSDT", LastPrice: 50000.0})
	agg.UpdateOrder(&types.Order{ID: "order-1", Symbol: "BTCUSDT", Status: "NEW"})
	agg.UpdatePosition("BTCUSDT", &types.Position{Symbol: "BTCUSDT", Quantity: 0.5})
	agg.UpdateAccount(&types.Account{TotalBalance: 10000.0, Balances: make(map[string]float64)})

	state := agg.GetFullState()

	assert.NotNil(t, state)
	assert.Equal(t, 1, len(state.MarketData))
	assert.Equal(t, 1, len(state.Orders))
	assert.Equal(t, 1, len(state.Positions))
	assert.NotNil(t, state.Account)
	assert.NotZero(t, state.Timestamp)
}
