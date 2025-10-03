package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/b25/strategy-engine/internal/config"
	"github.com/b25/strategy-engine/internal/strategies"
	"github.com/b25/strategy-engine/pkg/logger"
	"github.com/b25/strategy-engine/pkg/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OrderExecutionClient wraps gRPC client for order execution
type OrderExecutionClient struct {
	conn    *grpc.ClientConn
	cfg     *config.GRPCConfig
	logger  *logger.Logger
	metrics *metrics.Collector
}

// NewOrderExecutionClient creates a new order execution client
func NewOrderExecutionClient(cfg *config.GRPCConfig, log *logger.Logger, m *metrics.Collector) (*OrderExecutionClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, cfg.OrderExecutionAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order execution service: %w", err)
	}

	log.Info("Connected to order execution service", "addr", cfg.OrderExecutionAddr)

	return &OrderExecutionClient{
		conn:    conn,
		cfg:     cfg,
		logger:  log,
		metrics: m,
	}, nil
}

// SubmitOrder submits an order to the execution service
func (c *OrderExecutionClient) SubmitOrder(ctx context.Context, signal *strategies.Signal) error {
	startTime := time.Now()
	defer func() {
		latency := time.Since(startTime).Microseconds()
		c.metrics.OrderLatency.WithLabelValues(signal.Strategy).Observe(float64(latency))
	}()

	// In a real implementation, this would use the generated protobuf client
	// For now, we'll simulate the call
	c.logger.Info("Submitting order",
		"signal_id", signal.ID,
		"strategy", signal.Strategy,
		"symbol", signal.Symbol,
		"side", signal.Side,
		"type", signal.OrderType,
		"quantity", signal.Quantity,
		"price", signal.Price,
	)

	// Simulate network delay
	select {
	case <-time.After(1 * time.Millisecond):
		// Order submitted successfully
		c.metrics.OrdersSubmitted.WithLabelValues(
			signal.Strategy,
			signal.Symbol,
			signal.Side,
			signal.OrderType,
		).Inc()

		return nil

	case <-ctx.Done():
		return fmt.Errorf("order submission cancelled: %w", ctx.Err())
	}
}

// SubmitOrders submits multiple orders in batch
func (c *OrderExecutionClient) SubmitOrders(ctx context.Context, signals []*strategies.Signal) error {
	for _, signal := range signals {
		if err := c.SubmitOrder(ctx, signal); err != nil {
			c.logger.Error("Failed to submit order",
				"error", err,
				"signal_id", signal.ID,
			)
			c.metrics.OrdersRejected.WithLabelValues(
				signal.Strategy,
				signal.Symbol,
				"grpc_error",
			).Inc()
			// Continue with other orders even if one fails
			continue
		}
	}

	return nil
}

// CancelOrder cancels an order
func (c *OrderExecutionClient) CancelOrder(ctx context.Context, orderID string) error {
	c.logger.Info("Canceling order", "order_id", orderID)

	// In a real implementation, this would use the generated protobuf client
	return nil
}

// Close closes the gRPC connection
func (c *OrderExecutionClient) Close() error {
	return c.conn.Close()
}

// OrderRequest represents an order request (placeholder for protobuf message)
type OrderRequest struct {
	Symbol      string
	Side        string
	OrderType   string
	Quantity    float64
	Price       float64
	StopPrice   float64
	TimeInForce string
	StrategyID  string
	Metadata    map[string]string
}

// OrderResponse represents an order response (placeholder for protobuf message)
type OrderResponse struct {
	OrderID   string
	Status    string
	Message   string
	Timestamp time.Time
}
