package generators

import (
	"fmt"
	"math/rand"
	"time"
)

// OrderGenerator generates test orders
type OrderGenerator struct {
	rand *rand.Rand
}

// NewOrderGenerator creates a new order generator
func NewOrderGenerator() *OrderGenerator {
	return &OrderGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateOrder generates a random order
func (g *OrderGenerator) GenerateOrder(symbol string) map[string]interface{} {
	sides := []string{"BUY", "SELL"}
	types := []string{"MARKET", "LIMIT"}

	side := sides[g.rand.Intn(len(sides))]
	orderType := types[g.rand.Intn(len(types))]

	order := map[string]interface{}{
		"symbol":          symbol,
		"side":            side,
		"type":            orderType,
		"quantity":        g.randomQuantity(),
		"client_order_id": g.randomClientOrderID(),
	}

	if orderType == "LIMIT" {
		order["price"] = g.randomPrice(symbol)
		order["time_in_force"] = "GTC"
	}

	return order
}

// GenerateMarketOrder generates a market order
func (g *OrderGenerator) GenerateMarketOrder(symbol, side string, quantity float64) map[string]interface{} {
	return map[string]interface{}{
		"symbol":          symbol,
		"side":            side,
		"type":            "MARKET",
		"quantity":        quantity,
		"client_order_id": g.randomClientOrderID(),
	}
}

// GenerateLimitOrder generates a limit order
func (g *OrderGenerator) GenerateLimitOrder(symbol, side string, price, quantity float64) map[string]interface{} {
	return map[string]interface{}{
		"symbol":          symbol,
		"side":            side,
		"type":            "LIMIT",
		"price":           price,
		"quantity":        quantity,
		"time_in_force":   "GTC",
		"client_order_id": g.randomClientOrderID(),
	}
}

// randomPrice generates a random price based on symbol
func (g *OrderGenerator) randomPrice(symbol string) float64 {
	basePrices := map[string]float64{
		"BTCUSDT": 50000,
		"ETHUSDT": 3000,
	}

	basePrice, ok := basePrices[symbol]
	if !ok {
		basePrice = 100
	}

	// Random price within ±5% of base
	variation := (g.rand.Float64() - 0.5) * 0.1 * basePrice
	return basePrice + variation
}

// randomQuantity generates a random quantity
func (g *OrderGenerator) randomQuantity() float64 {
	return g.rand.Float64()*10 + 0.01 // 0.01 to 10.01
}

// randomClientOrderID generates a random client order ID
func (g *OrderGenerator) randomClientOrderID() string {
	return fmt.Sprintf("test_order_%d_%d", time.Now().UnixNano(), g.rand.Intn(1000))
}

// MarketDataGenerator generates test market data
type MarketDataGenerator struct {
	rand       *rand.Rand
	basePrices map[string]float64
}

// NewMarketDataGenerator creates a new market data generator
func NewMarketDataGenerator() *MarketDataGenerator {
	return &MarketDataGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
		basePrices: map[string]float64{
			"BTCUSDT": 50000,
			"ETHUSDT": 3000,
		},
	}
}

// GenerateTick generates a market data tick
func (g *MarketDataGenerator) GenerateTick(symbol string) map[string]interface{} {
	basePrice := g.getBasePrice(symbol)

	// Simulate price movement
	priceChange := (g.rand.Float64() - 0.5) * 0.001 * basePrice
	currentPrice := basePrice + priceChange

	spread := basePrice * 0.0001 // 0.01% spread
	bidPrice := currentPrice - spread/2
	askPrice := currentPrice + spread/2

	return map[string]interface{}{
		"symbol":      symbol,
		"timestamp":   time.Now().UnixMilli(),
		"last_price":  currentPrice,
		"bid_price":   bidPrice,
		"ask_price":   askPrice,
		"bid_size":    g.rand.Float64()*10 + 1,
		"ask_size":    g.rand.Float64()*10 + 1,
		"volume":      g.rand.Float64()*1000 + 100,
		"volume_quote": (g.rand.Float64()*1000 + 100) * currentPrice,
	}
}

// GenerateOrderBook generates a complete order book
func (g *MarketDataGenerator) GenerateOrderBook(symbol string, levels int) map[string]interface{} {
	basePrice := g.getBasePrice(symbol)
	midPrice := basePrice + (g.rand.Float64()-0.5)*0.001*basePrice

	bids := make([]map[string]float64, levels)
	asks := make([]map[string]float64, levels)

	for i := 0; i < levels; i++ {
		bids[i] = map[string]float64{
			"price":    midPrice - float64(i+1)*0.01*basePrice,
			"quantity": g.rand.Float64()*5 + 0.1,
		}
		asks[i] = map[string]float64{
			"price":    midPrice + float64(i+1)*0.01*basePrice,
			"quantity": g.rand.Float64()*5 + 0.1,
		}
	}

	return map[string]interface{}{
		"symbol":    symbol,
		"timestamp": time.Now().UnixMilli(),
		"bids":      bids,
		"asks":      asks,
	}
}

// GenerateCandle generates OHLCV candle data
func (g *MarketDataGenerator) GenerateCandle(symbol string, interval time.Duration) map[string]interface{} {
	basePrice := g.getBasePrice(symbol)

	open := basePrice + (g.rand.Float64()-0.5)*0.01*basePrice
	high := open + g.rand.Float64()*0.005*basePrice
	low := open - g.rand.Float64()*0.005*basePrice
	close := low + g.rand.Float64()*(high-low)
	volume := g.rand.Float64()*1000 + 100

	return map[string]interface{}{
		"symbol":       symbol,
		"timestamp":    time.Now().UnixMilli(),
		"open":         open,
		"high":         high,
		"low":          low,
		"close":        close,
		"volume":       volume,
		"volume_quote": volume * close,
	}
}

// getBasePrice returns the base price for a symbol
func (g *MarketDataGenerator) getBasePrice(symbol string) float64 {
	if price, ok := g.basePrices[symbol]; ok {
		return price
	}
	return 100.0
}

// AccountDataGenerator generates test account data
type AccountDataGenerator struct {
	rand *rand.Rand
}

// NewAccountDataGenerator creates a new account data generator
func NewAccountDataGenerator() *AccountDataGenerator {
	return &AccountDataGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GeneratePosition generates a random position
func (g *AccountDataGenerator) GeneratePosition(symbol string) map[string]interface{} {
	sides := []string{"long", "short", "flat"}
	side := sides[g.rand.Intn(len(sides))]

	quantity := 0.0
	avgEntryPrice := 0.0
	unrealizedPnL := 0.0

	if side != "flat" {
		quantity = g.rand.Float64()*10 + 0.1
		avgEntryPrice = g.randomPrice(symbol)
		currentPrice := avgEntryPrice + (g.rand.Float64()-0.5)*0.1*avgEntryPrice

		if side == "long" {
			unrealizedPnL = (currentPrice - avgEntryPrice) * quantity
		} else {
			unrealizedPnL = (avgEntryPrice - currentPrice) * quantity
		}
	}

	return map[string]interface{}{
		"symbol":          symbol,
		"side":            side,
		"quantity":        quantity,
		"avg_entry_price": avgEntryPrice,
		"current_price":   avgEntryPrice + (g.rand.Float64()-0.5)*0.1*avgEntryPrice,
		"unrealized_pnl":  unrealizedPnL,
		"realized_pnl":    g.rand.Float64()*1000 - 500,
		"timestamp":       time.Now().UnixMilli(),
	}
}

// GenerateBalance generates account balance
func (g *AccountDataGenerator) GenerateBalance() map[string]interface{} {
	return map[string]interface{}{
		"total_equity":      g.rand.Float64()*100000 + 10000,
		"available_balance": g.rand.Float64()*50000 + 5000,
		"margin_used":       g.rand.Float64()*30000,
		"unrealized_pnl":    g.rand.Float64()*2000 - 1000,
		"timestamp":         time.Now().UnixMilli(),
	}
}

// randomPrice generates a random price
func (g *AccountDataGenerator) randomPrice(symbol string) float64 {
	basePrices := map[string]float64{
		"BTCUSDT": 50000,
		"ETHUSDT": 3000,
	}

	basePrice, ok := basePrices[symbol]
	if !ok {
		basePrice = 100
	}

	variation := (g.rand.Float64() - 0.5) * 0.1 * basePrice
	return basePrice + variation
}

// StrategyDataGenerator generates test strategy data
type StrategyDataGenerator struct {
	rand *rand.Rand
}

// NewStrategyDataGenerator creates a new strategy data generator
func NewStrategyDataGenerator() *StrategyDataGenerator {
	return &StrategyDataGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateSignal generates a trading signal
func (g *StrategyDataGenerator) GenerateSignal(strategy, symbol string) map[string]interface{} {
	sides := []string{"buy", "sell"}
	orderTypes := []string{"market", "limit"}

	side := sides[g.rand.Intn(len(sides))]
	orderType := orderTypes[g.rand.Intn(len(orderTypes))]

	signal := map[string]interface{}{
		"id":         fmt.Sprintf("signal_%d", time.Now().UnixNano()),
		"strategy":   strategy,
		"symbol":     symbol,
		"side":       side,
		"order_type": orderType,
		"quantity":   g.rand.Float64()*5 + 0.1,
		"priority":   g.rand.Intn(10) + 1,
		"timestamp":  time.Now().UnixMilli(),
	}

	if orderType == "limit" {
		signal["price"] = g.randomPrice(symbol)
	}

	return signal
}

// GenerateStrategyMetrics generates strategy performance metrics
func (g *StrategyDataGenerator) GenerateStrategyMetrics(strategyName string) map[string]interface{} {
	return map[string]interface{}{
		"strategy_name":    strategyName,
		"total_trades":     g.rand.Intn(1000) + 100,
		"winning_trades":   g.rand.Intn(600) + 60,
		"losing_trades":    g.rand.Intn(400) + 40,
		"total_pnl":        g.rand.Float64()*10000 - 5000,
		"win_rate":         g.rand.Float64()*0.3 + 0.4, // 40-70%
		"avg_win":          g.rand.Float64()*100 + 50,
		"avg_loss":         -(g.rand.Float64()*80 + 20),
		"sharpe_ratio":     g.rand.Float64()*2 + 0.5,
		"max_drawdown":     -(g.rand.Float64()*2000 + 500),
		"timestamp":        time.Now().UnixMilli(),
	}
}

// randomPrice generates a random price
func (g *StrategyDataGenerator) randomPrice(symbol string) float64 {
	basePrices := map[string]float64{
		"BTCUSDT": 50000,
		"ETHUSDT": 3000,
	}

	basePrice, ok := basePrices[symbol]
	if !ok {
		basePrice = 100
	}

	variation := (g.rand.Float64() - 0.5) * 0.1 * basePrice
	return basePrice + variation
}

// ScenarioGenerator generates test scenarios
type ScenarioGenerator struct {
	orderGen      *OrderGenerator
	marketDataGen *MarketDataGenerator
	accountGen    *AccountDataGenerator
	strategyGen   *StrategyDataGenerator
}

// NewScenarioGenerator creates a new scenario generator
func NewScenarioGenerator() *ScenarioGenerator {
	return &ScenarioGenerator{
		orderGen:      NewOrderGenerator(),
		marketDataGen: NewMarketDataGenerator(),
		accountGen:    NewAccountDataGenerator(),
		strategyGen:   NewStrategyDataGenerator(),
	}
}

// GenerateTradingScenario generates a complete trading scenario
func (s *ScenarioGenerator) GenerateTradingScenario(symbol string, numOrders int) map[string]interface{} {
	orders := make([]map[string]interface{}, numOrders)
	for i := 0; i < numOrders; i++ {
		orders[i] = s.orderGen.GenerateOrder(symbol)
	}

	return map[string]interface{}{
		"symbol":       symbol,
		"market_data":  s.marketDataGen.GenerateTick(symbol),
		"order_book":   s.marketDataGen.GenerateOrderBook(symbol, 5),
		"orders":       orders,
		"position":     s.accountGen.GeneratePosition(symbol),
		"balance":      s.accountGen.GenerateBalance(),
		"signal":       s.strategyGen.GenerateSignal("momentum", symbol),
		"timestamp":    time.Now().UnixMilli(),
	}
}

// GenerateHighFrequencyScenario generates a high-frequency trading scenario
func (s *ScenarioGenerator) GenerateHighFrequencyScenario(symbol string, duration time.Duration, tickInterval time.Duration) []map[string]interface{} {
	scenarios := []map[string]interface{}{}

	start := time.Now()
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for range ticker.C {
		if time.Since(start) >= duration {
			break
		}

		scenario := map[string]interface{}{
			"timestamp":   time.Now().UnixMilli(),
			"market_data": s.marketDataGen.GenerateTick(symbol),
			"order_book":  s.marketDataGen.GenerateOrderBook(symbol, 10),
		}
		scenarios = append(scenarios, scenario)
	}

	return scenarios
}
