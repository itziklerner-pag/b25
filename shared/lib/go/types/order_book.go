package types

import (
	"container/heap"
	"sync"
)

// OrderBookSide represents one side of the order book (bids or asks).
type OrderBookSide struct {
	levels map[string]*OrderBookLevel // price -> level
	sorted *PriceLevelHeap            // sorted price levels
	mu     sync.RWMutex
}

// OrderBookLevel represents a price level in the order book.
type OrderBookLevel struct {
	Price      *Decimal
	Quantity   *Decimal
	OrderCount int
}

// PriceLevelHeap implements heap.Interface for maintaining sorted price levels.
type PriceLevelHeap struct {
	levels     []*OrderBookLevel
	isAscending bool // true for asks, false for bids
}

func (h PriceLevelHeap) Len() int { return len(h.levels) }

func (h PriceLevelHeap) Less(i, j int) bool {
	cmp := h.levels[i].Price.Cmp(h.levels[j].Price)
	if h.isAscending {
		return cmp < 0
	}
	return cmp > 0
}

func (h PriceLevelHeap) Swap(i, j int) {
	h.levels[i], h.levels[j] = h.levels[j], h.levels[i]
}

func (h *PriceLevelHeap) Push(x interface{}) {
	h.levels = append(h.levels, x.(*OrderBookLevel))
}

func (h *PriceLevelHeap) Pop() interface{} {
	old := h.levels
	n := len(old)
	x := old[n-1]
	h.levels = old[0 : n-1]
	return x
}

// NewOrderBookSide creates a new order book side.
func NewOrderBookSide(isAsk bool) *OrderBookSide {
	return &OrderBookSide{
		levels: make(map[string]*OrderBookLevel),
		sorted: &PriceLevelHeap{
			levels:      make([]*OrderBookLevel, 0),
			isAscending: isAsk,
		},
	}
}

// Update updates a price level in the order book.
func (s *OrderBookSide) Update(price, quantity *Decimal, orderCount int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	priceStr := price.String()

	if quantity.IsZero() {
		// Remove the level
		delete(s.levels, priceStr)
		s.rebuildHeap()
	} else {
		// Update or add the level
		level := &OrderBookLevel{
			Price:      price,
			Quantity:   quantity,
			OrderCount: orderCount,
		}
		s.levels[priceStr] = level
		s.rebuildHeap()
	}
}

// GetBest returns the best price level (highest bid or lowest ask).
func (s *OrderBookSide) GetBest() *OrderBookLevel {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.sorted.levels) == 0 {
		return nil
	}
	return s.sorted.levels[0]
}

// GetDepth returns the top n levels.
func (s *OrderBookSide) GetDepth(n int) []*OrderBookLevel {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n > len(s.sorted.levels) {
		n = len(s.sorted.levels)
	}

	result := make([]*OrderBookLevel, n)
	copy(result, s.sorted.levels[:n])
	return result
}

// GetTotalVolume returns the total volume up to depth n.
func (s *OrderBookSide) GetTotalVolume(n int) *Decimal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := Zero()
	levels := s.sorted.levels
	if n > len(levels) {
		n = len(levels)
	}

	for i := 0; i < n; i++ {
		total = total.Add(levels[i].Quantity)
	}
	return total
}

// Clear removes all levels.
func (s *OrderBookSide) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.levels = make(map[string]*OrderBookLevel)
	s.sorted.levels = make([]*OrderBookLevel, 0)
}

// rebuildHeap rebuilds the sorted heap from the levels map.
func (s *OrderBookSide) rebuildHeap() {
	s.sorted.levels = make([]*OrderBookLevel, 0, len(s.levels))
	for _, level := range s.levels {
		s.sorted.levels = append(s.sorted.levels, level)
	}
	heap.Init(s.sorted)
}

// OrderBook represents a complete order book with bids and asks.
type OrderBook struct {
	Symbol         string
	Bids           *OrderBookSide
	Asks           *OrderBookSide
	SequenceNumber int64
	mu             sync.RWMutex
}

// NewOrderBook creates a new order book.
func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol: symbol,
		Bids:   NewOrderBookSide(false),
		Asks:   NewOrderBookSide(true),
	}
}

// GetMidPrice returns the mid price (average of best bid and ask).
func (ob *OrderBook) GetMidPrice() *Decimal {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	bestBid := ob.Bids.GetBest()
	bestAsk := ob.Asks.GetBest()

	if bestBid == nil || bestAsk == nil {
		return nil
	}

	return bestBid.Price.Add(bestAsk.Price).Div(NewDecimalFromInt64(2))
}

// GetSpread returns the spread in basis points.
func (ob *OrderBook) GetSpread() *Decimal {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	bestBid := ob.Bids.GetBest()
	bestAsk := ob.Asks.GetBest()

	if bestBid == nil || bestAsk == nil {
		return nil
	}

	spread := bestAsk.Price.Sub(bestBid.Price)
	mid := bestBid.Price.Add(bestAsk.Price).Div(NewDecimalFromInt64(2))

	if mid.IsZero() {
		return Zero()
	}

	// Spread in basis points = (spread / mid) * 10000
	bps := spread.Div(mid).Mul(NewDecimalFromInt64(10000))
	return bps
}

// GetMicroPrice returns the volume-weighted mid price.
func (ob *OrderBook) GetMicroPrice() *Decimal {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	bestBid := ob.Bids.GetBest()
	bestAsk := ob.Asks.GetBest()

	if bestBid == nil || bestAsk == nil {
		return nil
	}

	bidQty := bestBid.Quantity
	askQty := bestAsk.Quantity
	totalQty := bidQty.Add(askQty)

	if totalQty.IsZero() {
		return ob.GetMidPrice()
	}

	// Microprice = (bid_price * ask_qty + ask_price * bid_qty) / (bid_qty + ask_qty)
	numerator := bestBid.Price.Mul(askQty).Add(bestAsk.Price.Mul(bidQty))
	return numerator.Div(totalQty)
}

// GetImbalance returns the order book imbalance (-1 to 1).
func (ob *OrderBook) GetImbalance(depth int) *Decimal {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	bidVolume := ob.Bids.GetTotalVolume(depth)
	askVolume := ob.Asks.GetTotalVolume(depth)

	totalVolume := bidVolume.Add(askVolume)
	if totalVolume.IsZero() {
		return Zero()
	}

	// Imbalance = (bid_volume - ask_volume) / (bid_volume + ask_volume)
	return bidVolume.Sub(askVolume).Div(totalVolume)
}
