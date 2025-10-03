package exchange

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// MockExchange represents a mock exchange server for testing
type MockExchange struct {
	HTTPServer *http.Server
	WSServer   *websocket.Upgrader

	orders      map[string]*Order
	ordersMutex sync.RWMutex

	fills       []*Fill
	fillsMutex  sync.RWMutex

	wsConnections []*websocket.Conn
	wsMutex       sync.RWMutex

	config *MockExchangeConfig
}

// MockExchangeConfig holds configuration for the mock exchange
type MockExchangeConfig struct {
	HTTPAddr           string
	WSAddr             string
	OrderLatency       time.Duration // Simulated order processing latency
	FillDelay          time.Duration // Delay before order fills
	RejectRate         float64       // 0.0-1.0, probability of order rejection
	PartialFillEnabled bool          // Enable partial fills
	MarketDataEnabled  bool          // Enable market data streaming
}

// Order represents an exchange order
type Order struct {
	OrderID       string    `json:"orderId"`
	ClientOrderID string    `json:"clientOrderId"`
	Symbol        string    `json:"symbol"`
	Side          string    `json:"side"`
	Type          string    `json:"type"`
	Status        string    `json:"status"`
	Price         string    `json:"price"`
	Quantity      string    `json:"origQty"`
	ExecutedQty   string    `json:"executedQty"`
	AvgPrice      string    `json:"avgPrice"`
	UpdateTime    int64     `json:"updateTime"`
}

// Fill represents an order fill
type Fill struct {
	FillID    string    `json:"fillId"`
	OrderID   string    `json:"orderId"`
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"`
	Price     string    `json:"price"`
	Quantity  string    `json:"qty"`
	Fee       string    `json:"commission"`
	Timestamp int64     `json:"time"`
}

// MarketDataTick represents a market data tick
type MarketDataTick struct {
	EventType string  `json:"e"`
	EventTime int64   `json:"E"`
	Symbol    string  `json:"s"`
	Price     string  `json:"p"`
	Quantity  string  `json:"q"`
	BidPrice  string  `json:"b"`
	BidQty    string  `json:"B"`
	AskPrice  string  `json:"a"`
	AskQty    string  `json:"A"`
}

// NewMockExchange creates a new mock exchange
func NewMockExchange(config *MockExchangeConfig) *MockExchange {
	if config == nil {
		config = DefaultMockExchangeConfig()
	}

	return &MockExchange{
		orders:        make(map[string]*Order),
		fills:         make([]*Fill, 0),
		wsConnections: make([]*websocket.Conn, 0),
		config:        config,
		WSServer: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// DefaultMockExchangeConfig returns default configuration
func DefaultMockExchangeConfig() *MockExchangeConfig {
	return &MockExchangeConfig{
		HTTPAddr:           ":8545",
		WSAddr:             ":8546",
		OrderLatency:       10 * time.Millisecond,
		FillDelay:          50 * time.Millisecond,
		RejectRate:         0.0,
		PartialFillEnabled: false,
		MarketDataEnabled:  true,
	}
}

// Start starts the mock exchange servers
func (m *MockExchange) Start() error {
	// Setup HTTP routes
	mux := http.NewServeMux()

	// Order endpoints
	mux.HandleFunc("/api/v3/order", m.handleCreateOrder)
	mux.HandleFunc("/api/v3/order/cancel", m.handleCancelOrder)
	mux.HandleFunc("/api/v3/order/status", m.handleOrderStatus)
	mux.HandleFunc("/api/v3/exchangeInfo", m.handleExchangeInfo)

	// Account endpoints
	mux.HandleFunc("/api/v3/account", m.handleAccount)
	mux.HandleFunc("/api/v3/userDataStream", m.handleUserDataStream)

	// WebSocket endpoint
	mux.HandleFunc("/ws", m.handleWebSocket)

	m.HTTPServer = &http.Server{
		Addr:    m.config.HTTPAddr,
		Handler: mux,
	}

	// Start HTTP server
	go func() {
		if err := m.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	// Start market data generator if enabled
	if m.config.MarketDataEnabled {
		go m.generateMarketData()
	}

	return nil
}

// handleCreateOrder handles order creation
func (m *MockExchange) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	// Simulate processing latency
	time.Sleep(m.config.OrderLatency)

	// Parse request
	symbol := r.FormValue("symbol")
	side := r.FormValue("side")
	orderType := r.FormValue("type")
	quantity := r.FormValue("quantity")
	price := r.FormValue("price")

	// Generate order ID
	orderID := fmt.Sprintf("%d", time.Now().UnixNano())
	clientOrderID := r.FormValue("newClientOrderId")

	// Create order
	order := &Order{
		OrderID:       orderID,
		ClientOrderID: clientOrderID,
		Symbol:        symbol,
		Side:          side,
		Type:          orderType,
		Status:        "NEW",
		Price:         price,
		Quantity:      quantity,
		ExecutedQty:   "0",
		AvgPrice:      "0",
		UpdateTime:    time.Now().UnixMilli(),
	}

	// Store order
	m.ordersMutex.Lock()
	m.orders[orderID] = order
	m.ordersMutex.Unlock()

	// Respond
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)

	// Simulate order fill
	go m.simulateFill(order)
}

// handleCancelOrder handles order cancellation
func (m *MockExchange) handleCancelOrder(w http.ResponseWriter, r *http.Request) {
	orderID := r.FormValue("orderId")
	symbol := r.FormValue("symbol")

	m.ordersMutex.Lock()
	defer m.ordersMutex.Unlock()

	order, exists := m.orders[orderID]
	if !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	if order.Symbol != symbol {
		http.Error(w, "Symbol mismatch", http.StatusBadRequest)
		return
	}

	order.Status = "CANCELED"
	order.UpdateTime = time.Now().UnixMilli()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// handleOrderStatus handles order status query
func (m *MockExchange) handleOrderStatus(w http.ResponseWriter, r *http.Request) {
	orderID := r.FormValue("orderId")

	m.ordersMutex.RLock()
	order, exists := m.orders[orderID]
	m.ordersMutex.RUnlock()

	if !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// handleExchangeInfo handles exchange info request
func (m *MockExchange) handleExchangeInfo(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"timezone":   "UTC",
		"serverTime": time.Now().UnixMilli(),
		"symbols": []map[string]interface{}{
			{
				"symbol":            "BTCUSDT",
				"status":            "TRADING",
				"baseAsset":         "BTC",
				"quoteAsset":        "USDT",
				"pricePrecision":    2,
				"quantityPrecision": 3,
				"filters": []map[string]interface{}{
					{
						"filterType": "PRICE_FILTER",
						"minPrice":   "0.01",
						"maxPrice":   "100000",
						"tickSize":   "0.01",
					},
					{
						"filterType": "LOT_SIZE",
						"minQty":     "0.001",
						"maxQty":     "100",
						"stepSize":   "0.001",
					},
					{
						"filterType":  "MIN_NOTIONAL",
						"minNotional": "10",
					},
				},
			},
			{
				"symbol":            "ETHUSDT",
				"status":            "TRADING",
				"baseAsset":         "ETH",
				"quoteAsset":        "USDT",
				"pricePrecision":    2,
				"quantityPrecision": 2,
				"filters": []map[string]interface{}{
					{
						"filterType": "PRICE_FILTER",
						"minPrice":   "0.01",
						"maxPrice":   "10000",
						"tickSize":   "0.01",
					},
					{
						"filterType": "LOT_SIZE",
						"minQty":     "0.01",
						"maxQty":     "1000",
						"stepSize":   "0.01",
					},
					{
						"filterType":  "MIN_NOTIONAL",
						"minNotional": "10",
					},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// handleAccount handles account info request
func (m *MockExchange) handleAccount(w http.ResponseWriter, r *http.Request) {
	account := map[string]interface{}{
		"canTrade":  true,
		"canDeposit": true,
		"canWithdraw": true,
		"updateTime": time.Now().UnixMilli(),
		"balances": []map[string]string{
			{"asset": "BTC", "free": "10.000", "locked": "0.000"},
			{"asset": "ETH", "free": "100.00", "locked": "0.00"},
			{"asset": "USDT", "free": "100000.00", "locked": "0.00"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

// handleUserDataStream handles user data stream creation
func (m *MockExchange) handleUserDataStream(w http.ResponseWriter, r *http.Request) {
	listenKey := fmt.Sprintf("key_%d", time.Now().UnixNano())
	response := map[string]string{"listenKey": listenKey}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleWebSocket handles WebSocket connections
func (m *MockExchange) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := m.WSServer.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade error: %v\n", err)
		return
	}

	m.wsMutex.Lock()
	m.wsConnections = append(m.wsConnections, conn)
	m.wsMutex.Unlock()

	// Keep connection alive
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				m.removeConnection(conn)
				break
			}
		}
	}()
}

// simulateFill simulates order filling
func (m *MockExchange) simulateFill(order *Order) {
	time.Sleep(m.config.FillDelay)

	m.ordersMutex.Lock()
	defer m.ordersMutex.Unlock()

	// Update order status
	order.Status = "FILLED"
	order.ExecutedQty = order.Quantity
	order.AvgPrice = order.Price
	order.UpdateTime = time.Now().UnixMilli()

	// Create fill
	fill := &Fill{
		FillID:    fmt.Sprintf("fill_%d", time.Now().UnixNano()),
		OrderID:   order.OrderID,
		Symbol:    order.Symbol,
		Side:      order.Side,
		Price:     order.Price,
		Quantity:  order.Quantity,
		Fee:       "0.001",
		Timestamp: time.Now().UnixMilli(),
	}

	m.fillsMutex.Lock()
	m.fills = append(m.fills, fill)
	m.fillsMutex.Unlock()

	// Broadcast order update via WebSocket
	m.broadcastOrderUpdate(order)
}

// generateMarketData generates simulated market data
func (m *MockExchange) generateMarketData() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	basePrice := 50000.0

	for range ticker.C {
		// Simulate price movement
		priceChange := (float64(time.Now().UnixNano()%1000) - 500) / 100
		currentPrice := basePrice + priceChange

		tick := &MarketDataTick{
			EventType: "24hrTicker",
			EventTime: time.Now().UnixMilli(),
			Symbol:    "BTCUSDT",
			Price:     fmt.Sprintf("%.2f", currentPrice),
			Quantity:  "0.1",
			BidPrice:  fmt.Sprintf("%.2f", currentPrice-0.5),
			BidQty:    "1.5",
			AskPrice:  fmt.Sprintf("%.2f", currentPrice+0.5),
			AskQty:    "2.0",
		}

		m.broadcastMarketData(tick)
	}
}

// broadcastOrderUpdate broadcasts order update to all WebSocket connections
func (m *MockExchange) broadcastOrderUpdate(order *Order) {
	message := map[string]interface{}{
		"e": "executionReport",
		"E": time.Now().UnixMilli(),
		"s": order.Symbol,
		"o": order.OrderID,
		"x": order.Status,
		"q": order.Quantity,
		"p": order.Price,
		"z": order.ExecutedQty,
	}

	m.broadcast(message)
}

// broadcastMarketData broadcasts market data to all WebSocket connections
func (m *MockExchange) broadcastMarketData(tick *MarketDataTick) {
	m.broadcast(tick)
}

// broadcast sends message to all WebSocket connections
func (m *MockExchange) broadcast(message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		return
	}

	m.wsMutex.RLock()
	defer m.wsMutex.RUnlock()

	for _, conn := range m.wsConnections {
		conn.WriteMessage(websocket.TextMessage, data)
	}
}

// removeConnection removes a WebSocket connection
func (m *MockExchange) removeConnection(conn *websocket.Conn) {
	m.wsMutex.Lock()
	defer m.wsMutex.Unlock()

	for i, c := range m.wsConnections {
		if c == conn {
			m.wsConnections = append(m.wsConnections[:i], m.wsConnections[i+1:]...)
			break
		}
	}
	conn.Close()
}

// GetOrder returns an order by ID
func (m *MockExchange) GetOrder(orderID string) (*Order, bool) {
	m.ordersMutex.RLock()
	defer m.ordersMutex.RUnlock()

	order, exists := m.orders[orderID]
	return order, exists
}

// GetFills returns all fills
func (m *MockExchange) GetFills() []*Fill {
	m.fillsMutex.RLock()
	defer m.fillsMutex.RUnlock()

	fills := make([]*Fill, len(m.fills))
	copy(fills, m.fills)
	return fills
}

// Stop stops the mock exchange
func (m *MockExchange) Stop() error {
	m.wsMutex.Lock()
	for _, conn := range m.wsConnections {
		conn.Close()
	}
	m.wsConnections = nil
	m.wsMutex.Unlock()

	if m.HTTPServer != nil {
		return m.HTTPServer.Close()
	}
	return nil
}
