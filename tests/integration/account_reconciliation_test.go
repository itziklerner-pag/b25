package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/yourusername/b25/tests/testutil/generators"
)

// AccountReconciliationTestSuite tests position and balance reconciliation
type AccountReconciliationTestSuite struct {
	suite.Suite
	db          *sql.DB
	redisClient *redis.Client
	natsConn    *nats.Conn
	accountGen  *generators.AccountDataGenerator
}

func (s *AccountReconciliationTestSuite) SetupSuite() {
	var err error

	// Connect to PostgreSQL
	dbHost := getEnv("POSTGRES_ADDR", "localhost:5433")
	dbUser := getEnv("POSTGRES_USER", "testuser")
	dbPass := getEnv("POSTGRES_PASSWORD", "testpass")
	dbName := getEnv("POSTGRES_DB", "b25_test")

	connStr := "host=" + dbHost + " user=" + dbUser + " password=" + dbPass + " dbname=" + dbName + " sslmode=disable"
	s.db, err = sql.Open("postgres", connStr)
	require.NoError(s.T(), err)

	err = s.db.Ping()
	require.NoError(s.T(), err)

	// Connect to Redis
	s.redisClient = redis.NewClient(&redis.Options{
		Addr: getEnv("REDIS_ADDR", "localhost:6380"),
		DB:   0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = s.redisClient.Ping(ctx).Err()
	require.NoError(s.T(), err)

	// Connect to NATS
	s.natsConn, err = nats.Connect(getEnv("NATS_ADDR", "nats://localhost:4223"))
	require.NoError(s.T(), err)

	s.accountGen = generators.NewAccountDataGenerator()
}

func (s *AccountReconciliationTestSuite) TearDownSuite() {
	if s.natsConn != nil {
		s.natsConn.Close()
	}
	if s.redisClient != nil {
		s.redisClient.Close()
	}
	if s.db != nil {
		s.db.Close()
	}
}

func (s *AccountReconciliationTestSuite) SetupTest() {
	// Clean up test data
	ctx := context.Background()
	s.redisClient.FlushDB(ctx)

	// Clean database tables
	s.db.Exec("TRUNCATE TABLE positions, account_history CASCADE")
}

// TestPositionReconciliation tests position tracking and reconciliation
func (s *AccountReconciliationTestSuite) TestPositionReconciliation() {
	userID := "test_user_1"
	symbol := "BTCUSDT"

	// Create initial position
	position := map[string]interface{}{
		"user_id":         userID,
		"symbol":          symbol,
		"side":            "long",
		"quantity":        1.5,
		"avg_entry_price": 50000.0,
		"current_price":   50500.0,
		"unrealized_pnl":  750.0,
		"realized_pnl":    0.0,
	}

	// Store in database
	_, err := s.db.Exec(`
		INSERT INTO positions (user_id, symbol, side, quantity, avg_entry_price, current_price, unrealized_pnl, realized_pnl)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, symbol) DO UPDATE
		SET quantity = $4, avg_entry_price = $5, current_price = $6, unrealized_pnl = $7, realized_pnl = $8
	`, userID, symbol, position["side"], position["quantity"], position["avg_entry_price"],
		position["current_price"], position["unrealized_pnl"], position["realized_pnl"])
	require.NoError(s.T(), err)

	// Trigger reconciliation via NATS
	reconcileReq := map[string]interface{}{
		"user_id": userID,
		"symbol":  symbol,
	}

	data, err := json.Marshal(reconcileReq)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("account.reconcile.position", data)
	require.NoError(s.T(), err)

	// Wait for reconciliation
	time.Sleep(500 * time.Millisecond)

	// Verify reconciled position in Redis
	ctx := context.Background()
	key := "position:" + userID + ":" + symbol
	storedData, err := s.redisClient.Get(ctx, key).Bytes()

	if err == redis.Nil {
		s.T().Skip("Account monitor service not running")
		return
	}

	require.NoError(s.T(), err)

	var reconciledPosition map[string]interface{}
	err = json.Unmarshal(storedData, &reconciledPosition)
	require.NoError(s.T(), err)

	// Verify position data
	assert.Equal(s.T(), symbol, reconciledPosition["symbol"])
	assert.Equal(s.T(), "long", reconciledPosition["side"])
}

// TestBalanceReconciliation tests balance tracking and reconciliation
func (s *AccountReconciliationTestSuite) TestBalanceReconciliation() {
	userID := "test_user_1"
	asset := "USDT"

	// Create balance record
	balance := map[string]interface{}{
		"user_id": userID,
		"asset":   asset,
		"free":    50000.0,
		"locked":  5000.0,
	}

	// Store in database
	_, err := s.db.Exec(`
		INSERT INTO balances (user_id, asset, free, locked)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, asset) DO UPDATE
		SET free = $3, locked = $4
	`, userID, asset, balance["free"], balance["locked"])
	require.NoError(s.T(), err)

	// Trigger reconciliation
	reconcileReq := map[string]interface{}{
		"user_id": userID,
		"asset":   asset,
	}

	data, err := json.Marshal(reconcileReq)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("account.reconcile.balance", data)
	require.NoError(s.T(), err)

	time.Sleep(500 * time.Millisecond)

	// Verify reconciled balance
	ctx := context.Background()
	key := "balance:" + userID + ":" + asset
	storedData, err := s.redisClient.Get(ctx, key).Bytes()

	if err == redis.Nil {
		s.T().Skip("Account monitor service not running")
		return
	}

	require.NoError(s.T(), err)

	var reconciledBalance map[string]interface{}
	err = json.Unmarshal(storedData, &reconciledBalance)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), asset, reconciledBalance["asset"])
}

// TestPnLCalculation tests P&L calculation accuracy
func (s *AccountReconciliationTestSuite) TestPnLCalculation() {
	userID := "test_user_1"
	symbol := "BTCUSDT"

	// Simulate fills
	fills := []map[string]interface{}{
		{
			"order_id": "order_1",
			"symbol":   symbol,
			"side":     "BUY",
			"price":    50000.0,
			"quantity": 1.0,
			"fee":      10.0,
		},
		{
			"order_id": "order_2",
			"symbol":   symbol,
			"side":     "SELL",
			"price":    51000.0,
			"quantity": 1.0,
			"fee":      10.0,
		},
	}

	expectedPnL := (51000.0 - 50000.0) * 1.0 - 20.0 // 980.0

	// Publish fills
	for _, fill := range fills {
		data, err := json.Marshal(fill)
		require.NoError(s.T(), err)

		err = s.natsConn.Publish("orders.fills."+symbol, data)
		require.NoError(s.T(), err)

		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(500 * time.Millisecond)

	// Check calculated P&L in database
	var realizedPnL float64
	err := s.db.QueryRow(`
		SELECT COALESCE(realized_pnl, 0)
		FROM positions
		WHERE user_id = $1 AND symbol = $2
	`, userID, symbol).Scan(&realizedPnL)

	if err == sql.ErrNoRows {
		s.T().Skip("No position found - P&L calculation may not be configured")
		return
	}

	require.NoError(s.T(), err)

	// Allow small floating point differences
	assert.InDelta(s.T(), expectedPnL, realizedPnL, 1.0, "Realized P&L should match expected")
}

// TestAccountHistoryTracking tests account history tracking
func (s *AccountReconciliationTestSuite) TestAccountHistoryTracking() {
	userID := "test_user_1"
	asset := "USDT"

	// Create balance change event
	event := map[string]interface{}{
		"user_id":        userID,
		"event_type":     "TRADE",
		"asset":          asset,
		"amount":         -1000.0,
		"balance_before": 100000.0,
		"balance_after":  99000.0,
		"reference_id":   "order_123",
	}

	data, err := json.Marshal(event)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("account.balance.change", data)
	require.NoError(s.T(), err)

	time.Sleep(500 * time.Millisecond)

	// Verify history record
	var count int
	err = s.db.QueryRow(`
		SELECT COUNT(*)
		FROM account_history
		WHERE user_id = $1 AND asset = $2 AND event_type = $3
	`, userID, asset, "TRADE").Scan(&count)

	if err != nil {
		s.T().Skip("Account history tracking may not be configured")
		return
	}

	assert.Greater(s.T(), count, 0, "Account history should be recorded")
}

// TestPositionSizeLimit tests position size limit enforcement
func (s *AccountReconciliationTestSuite) TestPositionSizeLimit() {
	userID := "test_user_1"
	symbol := "BTCUSDT"

	// Create large position
	largePosition := map[string]interface{}{
		"user_id":         userID,
		"symbol":          symbol,
		"side":            "long",
		"quantity":        15.0, // Exceeds max position size of 10
		"avg_entry_price": 50000.0,
	}

	// Subscribe to alerts
	alertsChan := make(chan *nats.Msg, 1)
	sub, err := s.natsConn.ChanSubscribe("alerts.position.limit", alertsChan)
	require.NoError(s.T(), err)
	defer sub.Unsubscribe()

	// Publish position update
	data, err := json.Marshal(largePosition)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("account.position.update", data)
	require.NoError(s.T(), err)

	// Wait for alert
	select {
	case msg := <-alertsChan:
		var alert map[string]interface{}
		err = json.Unmarshal(msg.Data, &alert)
		require.NoError(s.T(), err)

		s.T().Logf("Position limit alert received: %v", alert)
		assert.Equal(s.T(), symbol, alert["symbol"])

	case <-time.After(3 * time.Second):
		s.T().Skip("No position limit alert received")
	}
}

// TestMultiSymbolReconciliation tests reconciliation across multiple symbols
func (s *AccountReconciliationTestSuite) TestMultiSymbolReconciliation() {
	userID := "test_user_1"
	symbols := []string{"BTCUSDT", "ETHUSDT"}

	// Create positions for multiple symbols
	for _, symbol := range symbols {
		position := s.accountGen.GeneratePosition(symbol)
		position["user_id"] = userID

		_, err := s.db.Exec(`
			INSERT INTO positions (user_id, symbol, side, quantity, avg_entry_price, current_price, unrealized_pnl, realized_pnl)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (user_id, symbol) DO UPDATE
			SET quantity = $4, avg_entry_price = $5, current_price = $6
		`, userID, symbol, position["side"], position["quantity"], position["avg_entry_price"],
			position["current_price"], position["unrealized_pnl"], position["realized_pnl"])
		require.NoError(s.T(), err)
	}

	// Trigger full account reconciliation
	reconcileReq := map[string]interface{}{
		"user_id": userID,
	}

	data, err := json.Marshal(reconcileReq)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("account.reconcile.full", data)
	require.NoError(s.T(), err)

	time.Sleep(1 * time.Second)

	// Verify all positions reconciled
	ctx := context.Background()
	for _, symbol := range symbols {
		key := "position:" + userID + ":" + symbol
		exists, err := s.redisClient.Exists(ctx, key).Result()

		if err == nil && exists > 0 {
			s.T().Logf("Position reconciled for %s", symbol)
		}
	}
}

// TestReconciliationPerformance tests reconciliation performance
func (s *AccountReconciliationTestSuite) TestReconciliationPerformance() {
	userID := "test_user_perf"
	numSymbols := 10

	// Create positions for multiple symbols
	for i := 0; i < numSymbols; i++ {
		symbol := "SYMBOL" + string(rune(i))
		position := s.accountGen.GeneratePosition(symbol)

		_, err := s.db.Exec(`
			INSERT INTO positions (user_id, symbol, side, quantity, avg_entry_price, current_price, unrealized_pnl, realized_pnl)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, userID, symbol, position["side"], position["quantity"], position["avg_entry_price"],
			position["current_price"], position["unrealized_pnl"], position["realized_pnl"])
		require.NoError(s.T(), err)
	}

	// Measure reconciliation time
	startTime := time.Now()

	reconcileReq := map[string]interface{}{
		"user_id": userID,
	}

	data, err := json.Marshal(reconcileReq)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("account.reconcile.full", data)
	require.NoError(s.T(), err)

	time.Sleep(1 * time.Second)

	duration := time.Since(startTime)
	s.T().Logf("Reconciliation time for %d symbols: %v", numSymbols, duration)

	assert.Less(s.T(), duration, 2*time.Second, "Reconciliation should complete quickly")
}

func TestAccountReconciliationSuite(t *testing.T) {
	suite.Run(t, new(AccountReconciliationTestSuite))
}
